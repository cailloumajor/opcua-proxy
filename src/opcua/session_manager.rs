use std::collections::HashMap;
use std::sync::Arc;
use std::sync::atomic::{AtomicBool, Ordering};
use std::time::Duration;

use anyhow::Context as _;
use futures_util::FutureExt;
use futures_util::future::join_all;
use opcua::client::Client;
use parking_lot::Mutex;
use tokio::sync::{mpsc, oneshot};
use tokio::task::JoinHandle;
use tokio::time::{MissedTickBehavior, interval};
use tracing::{Instrument, debug, error, info, info_span, instrument};
use url::Url;

use crate::centrifugo::{CurrentTagValuesChannel, TagChangeMessage};
use crate::channel::roundtrip_channel;
use crate::opcua::config::{ConfigFetchService, PartnerConfig};
use crate::opcua::session::OpcUaSession;

/// The time interval for fetching the OPC-UA partners configuration.
const CONFIG_REFRESH_PERIOD: Duration = Duration::from_mins(1);

/// The capacity of the current partner data message channel.
const MESSAGE_QUEUE_CAPACITY: usize = 20;

/// Represents a manager of OPC-UA sessions.
pub(crate) struct SessionManager {
    /// Shareable OPC-UA client.
    opcua_client: Arc<Client>,
    data_change_channel: mpsc::Sender<TagChangeMessage>,
    config_fetcher: ConfigFetchService,
    /// A mapping of partner ID to session object.
    sessions: Arc<Mutex<HashMap<String, OpcUaSession>>>,
}

impl SessionManager {
    /// Create a new [`SessionManager`].
    pub(crate) fn new(
        config_api_url: Url,
        opcua_client: Arc<Client>,
        data_change_channel: mpsc::Sender<TagChangeMessage>,
    ) -> anyhow::Result<Self> {
        let config_fetcher =
            ConfigFetchService::new(config_api_url).context("Failed to build fetch service")?;

        Ok(Self {
            opcua_client,
            data_change_channel,
            config_fetcher,
            sessions: Default::default(),
        })
    }

    /// Spawn the manager main loop, returning a channel to request current data, and a handle to the task.
    ///
    /// This method consumes the [`SessionManager`].
    pub(crate) fn spawn(
        mut self,
        healthy: Arc<AtomicBool>,
    ) -> (CurrentTagValuesChannel, JoinHandle<()>) {
        let (current_data_tx, mut current_data_rx) =
            roundtrip_channel::<String, _>(MESSAGE_QUEUE_CAPACITY);

        let task = tokio::spawn(
            async move {
                info!(status = "started");

                let mut config_refresh_timer = interval(CONFIG_REFRESH_PERIOD);
                config_refresh_timer.set_missed_tick_behavior(MissedTickBehavior::Delay);

                loop {
                    tokio::select! {
                        biased;

                        _ = config_refresh_timer.tick() => {
                            if self.refresh_config().await {
                                healthy.store(true, Ordering::Relaxed);
                            } else {
                                healthy.store(false, Ordering::Relaxed);
                                // Retry configuration refresh sooner if it failed.
                                config_refresh_timer.reset_after(Duration::from_secs(5));
                            }

                        }

                        received = current_data_rx.recv() => {
                            if let Some((partner_id,tx)) = received {
                                self.send_partner_current_values(&partner_id, tx);
                            } else {
                                info!(msg = "current values channel closed");
                                self.stop().await;
                                break;
                            }
                        }
                    }
                }

                info!(status = "terminating");
            }
            .instrument(info_span!("session_manager")),
        );

        (current_data_tx, task)
    }

    /// Stop this [`SessionManager`], asking all managed sessions to stop and waiting for the operation to complete.
    async fn stop(&self) {
        // Pull out all managed sessions.
        let to_stop = self.sessions.lock_arc().drain().collect();

        self.stop_sessions(to_stop).await;
    }

    /// Fetch the partners configuration, start sessions that do not already exist and stop sessions
    /// for which there is no configuration anymore.
    ///
    /// Return a boolean indicating if configuration fetching succeeded.
    #[instrument(skip_all)]
    async fn refresh_config(&self) -> bool {
        let partners_configs = match self.config_fetcher.fetch().await {
            Ok(c) => c,
            Err(err) => {
                error!(during = "OPC-UA partners configuration fetching", %err);
                return false;
            }
        };

        debug!(?partners_configs);

        // Clone the keys to keep the lock as shortly as possible.
        let already_spawned = self.sessions.lock_arc().keys().cloned().collect::<Vec<_>>();
        let (to_retain, to_spawn): (Vec<PartnerConfig>, Vec<PartnerConfig>) = partners_configs
            .into_iter()
            .partition(|p| already_spawned.contains(&p.partner_id));

        let to_stop = self
            .sessions
            .lock_arc()
            .extract_if(|partner_id, _| !to_retain.iter().any(|p| *partner_id == p.partner_id))
            .collect();

        for config in to_spawn {
            OpcUaSession::spawn(
                Arc::clone(&self.opcua_client),
                config,
                self.data_change_channel.clone(),
                Arc::clone(&self.sessions),
            );
        }

        self.stop_sessions(to_stop).await;

        true
    }

    /// Stop sessions from provided collection of partner ID and [`OpcUaSession`].
    #[instrument(skip_all)]
    async fn stop_sessions(&self, to_stop: Vec<(String, OpcUaSession)>) {
        let session_stop_handles = to_stop.into_iter().map(|(partner_id, session)| async move {
            info!(msg = "stopping session", partner_id);

            session.stop().map(|result| (partner_id, result)).await
        });

        for (partner_id, result) in join_all(session_stop_handles).await {
            if let Err(status_code) = result {
                error!(during = "stopping OPC-UA session", partner_id, %status_code);
            }
        }
    }

    /// Get the current tag values, encoded in JSON, for the provided partner ID, and send it via the provided channel.
    ///
    /// Return whether sending was successful.
    #[instrument(skip(self, tx))]
    fn send_partner_current_values(
        &mut self,
        partner_id: &str,
        tx: oneshot::Sender<Option<Vec<u8>>>,
    ) {
        let data = match self
            .sessions
            .lock_arc()
            .get(partner_id)
            .map(|s| s.current_values_json())
            .transpose()
        {
            Ok(d) => d,
            Err(err) => {
                error!(during = "encoding current values to JSON", %err);
                // Channel sender will be dropped, thus aborting the request.
                return;
            }
        };

        if tx.send(data).is_err() {
            error!(during = "sending current values to proxy server");
        }
    }
}
