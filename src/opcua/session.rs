use std::collections::HashMap;
use std::sync::Arc;
use std::time::Duration;
use std::{fmt, io};

use opcua::client::{Client, IdentityToken, Session};
use opcua::types::{DateTime, EndpointDescription, MessageSecurityMode, StatusCode, Variant};
use parking_lot::Mutex;
use tokio::sync::mpsc;
use tokio::task::{JoinError, JoinHandle};
use tokio_util::task::AbortOnDropHandle;
use tracing::{Instrument, error, info, info_span};

use crate::centrifugo::TagChangeMessage;
use crate::opcua::config::PartnerConfig;
use crate::opcua::subscription::CentrifugoSubscriber;
use crate::opcua::tag_set::tag_set_from_config_groups;

use super::utils::encode_tag_changes;

/// The maximum time allowed for the session to be connected.
const CONNECT_TIMEOUT: Duration = Duration::from_secs(5);

/// Wraps errors that can be encountered when stopping a session.
#[derive(Debug)]
pub(super) enum SessionStopError {
    Disconnect(StatusCode),
    JoinLoopTask(JoinError),
}

impl fmt::Display for SessionStopError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Disconnect(status_code) => write!(f, "session disconnect error: {status_code}"),
            Self::JoinLoopTask(err) => write!(f, "session event loop join error: {err}"),
        }
    }
}

impl std::error::Error for SessionStopError {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        match self {
            Self::Disconnect(e) => e.source(),
            Self::JoinLoopTask(e) => e.source(),
        }
    }
}

/// Represents an active OPC-UA session.
pub(super) struct OpcUaSession {
    /// The inner OPC-UA session.
    session: Arc<Session>,
    /// The handle to the session runtime task.
    event_loop_handle: JoinHandle<StatusCode>,
    /// Storage of current tag values.
    current_values: Arc<Mutex<HashMap<String, (Variant, DateTime)>>>,
}

impl OpcUaSession {
    /// Spawn a session from provided client and configuration, and start polling its event loop.
    ///
    /// Upon success, the created session is stored in the provided registry.
    pub(super) fn spawn(
        client: Arc<Client>,
        config: PartnerConfig,
        tag_change_channel: mpsc::Sender<TagChangeMessage>,
        session_registry: Arc<Mutex<HashMap<String, OpcUaSession>>>,
    ) {
        let task_span =
            info_span!(parent: None, "session_spawn", partner_id = config.partner_id.clone());
        let partner_id = config.partner_id.clone();

        tokio::spawn(
            async move {
                info!(msg = "spawning OPC-UA session");

                let (endpoint, identity_token) = connection_params(&config);

                // Workaround for `Client::connect_to_matching_endpoint` unnecessarily taking
                // an exclusive reference to the client.
                let endpoints = match client
                    .get_server_endpoints_from_url(endpoint.endpoint_url.as_ref())
                    .await
                {
                    Ok(e) => e,
                    Err(err) => {
                        error!(during = "getting server endpoints", %err);
                        return;
                    }
                };
                let session_builder = match client
                    .session_builder()
                    .with_endpoints(endpoints)
                    .user_identity_token(identity_token)
                    .connect_to_matching_endpoint(endpoint)
                {
                    Ok(b) => b,
                    Err(err) => {
                        error!(during = "adding endpoint to session builder", %err);
                        return;
                    }
                };
                let cert_store = Arc::clone(client.certificate_store());
                let (session, event_loop) = match session_builder.build(cert_store) {
                    Ok(t) => t,
                    Err(err) => {
                        error!(during = "building the session and session event loop", %err);
                        return;
                    }
                };

                // Start polling the event loop to bring the session alive.
                let event_loop_handle = tokio::spawn(
                    event_loop
                        .run()
                        .instrument(info_span!(parent: None, "event_loop", partner_id)),
                );

                // Allow the event loop handling task to be aborted if anything goes wrong before the end of this scope.
                let loop_abort_handle = AbortOnDropHandle::new(event_loop_handle);

                match tokio::time::timeout(CONNECT_TIMEOUT, session.wait_for_connection()).await {
                    Ok(true) => {}
                    Ok(false) => {
                        error!(kind = "session connection failure");
                        return;
                    }
                    Err(_) => {
                        error!(kind = "session connection timed out");
                        return;
                    }
                }

                let tag_set = match tag_set_from_config_groups(&session, &config.tags).await {
                    Ok(t) => t,
                    Err(err) => {
                        error!(during = "building tag set", %err);
                        return;
                    }
                };

                let current_values = Default::default();

                info!(msg = "subscribing to tag changes");
                let subscriber = CentrifugoSubscriber::new(
                    Arc::clone(&session),
                    config.partner_id,
                    tag_set,
                    Arc::clone(&current_values),
                    tag_change_channel,
                );
                if let Err(err) = subscriber.enable().await {
                    error!(during = "subscribing to tag changes", %err);
                    return;
                };

                // Get back the event loop handle, disabling the abort-on-drop effect.
                let event_loop_handle = loop_abort_handle.detach();

                let opcua_session = Self {
                    session,
                    current_values,
                    event_loop_handle,
                };

                session_registry
                    .lock_arc()
                    .insert(partner_id, opcua_session);
            }
            .instrument(task_span),
        );
    }

    /// Ask the session to stop and wait for the operation completeness.
    ///
    /// This function takes ownership of the [`OpcUaSession`].
    pub(super) async fn stop(self) -> Result<(), SessionStopError> {
        self.session
            .disconnect()
            .await
            .map_err(SessionStopError::Disconnect)?;
        self.event_loop_handle
            .await
            .map_err(SessionStopError::JoinLoopTask)?;

        Ok(())
    }

    /// Get the current values of tags encoded in JSON.
    pub(super) fn current_values_json(&self) -> io::Result<Vec<u8>> {
        // Clone the storage to keep the lock as shortly as possible.
        let current_values = self.current_values.lock_arc().clone();

        let changes = current_values
            .iter()
            .map(|(tag_name, (value, timestamp))| (tag_name, value, timestamp))
            .collect::<Vec<_>>();

        let ctx = self.session.encoding_context().read();

        encode_tag_changes(&changes, &ctx.context())
    }
}

/// Create parameters to connect a client and establish a session.
fn connection_params(config: &PartnerConfig) -> (EndpointDescription, IdentityToken) {
    let endpoint: EndpointDescription = EndpointDescription::from((
        config.server_url.as_str(),
        config.security_policy.as_str(),
        MessageSecurityMode::from(config.security_mode.as_str()),
    ));

    let user_identity_token =
        if let Some((user, pass)) = config.user.as_ref().zip(config.password.as_ref()) {
            IdentityToken::new_user_name(user, pass)
        } else {
            IdentityToken::new_anonymous()
        };

    (endpoint, user_identity_token)
}
