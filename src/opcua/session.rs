use std::sync::Arc;

use anyhow::Context as _;
use opcua::client::prelude::*;
use opcua::sync::RwLock;
use tracing::{info, info_span, instrument};

use super::Config;

#[instrument(skip_all)]
pub(crate) fn create_session(
    config: &Config,
    partner_id: &str,
) -> anyhow::Result<Arc<RwLock<Session>>> {
    const PRODUCT_URI: &str = concat!("urn:", env!("CARGO_PKG_NAME"));

    let (user_token_id, user_identity_token) =
        if let (Some(user), Some(pass)) = (&config.opcua_user, &config.opcua_password) {
            ("default", Some(ClientUserToken::user_pass(user, pass)))
        } else {
            (ANONYMOUS_USER_TOKEN_ID, None)
        };

    let default_endpoint = ClientEndpoint {
        url: config.opcua_server_url.clone(),
        security_policy: config.opcua_security_policy.clone(),
        security_mode: config.opcua_security_mode.clone(),
        user_token_id: user_token_id.to_owned(),
    };

    let cert_key_prefix = format!("{}-{}", env!("CARGO_PKG_NAME"), partner_id);

    let mut client_builder = ClientBuilder::new()
        .application_name(env!("CARGO_PKG_DESCRIPTION"))
        .product_uri(PRODUCT_URI)
        .application_uri(format!("{PRODUCT_URI}:{partner_id}"))
        .pki_dir(config.pki_dir.clone())
        .certificate_path(format!("own/{cert_key_prefix}-cert.der"))
        .private_key_path(format!("private/{cert_key_prefix}-key.pem"))
        .endpoint("default", default_endpoint)
        .default_endpoint("default")
        .session_retry_interval(2000)
        .session_retry_limit(10)
        .session_timeout(1_200_000)
        .multi_threaded_executor();

    if let Some(token) = user_identity_token {
        client_builder = client_builder.user_token(user_token_id, token);
    }

    let mut client = client_builder
        .client()
        .context("error building the client")?;

    let session = client
        .connect_to_endpoint_id(None)
        .context("error establishing session")?;

    {
        let mut session = session.write();
        session.set_connection_status_callback(ConnectionStatusCallback::new(|connected| {
            let _entered = info_span!("connection status callback").entered();
            info!(msg = "connection status changed", connected);
        }));
        session.set_session_closed_callback(SessionClosedCallback::new(|status_code| {
            let _entered = info_span!("session closed callback").entered();
            info!(msg = "session has been closed", %status_code);
        }))
    }

    info!(status = "success");
    Ok(session)
}
