use std::collections::{HashMap, HashSet};
use std::sync::{Arc, Mutex};
use std::time::Duration;

use futures_util::future::join_all;
use opcua::client::prelude::{Client, Session, SessionCommand};
use opcua::sync::RwLock;
use tokio::sync::oneshot;
use tokio::task::spawn_blocking;
use tokio::time::{interval, MissedTickBehavior};
use tokio_util::sync::CancellationToken;
use tracing::{error, info, instrument, warn};
use url::Url;

use crate::db::{
    DataChangeChannel, DataChangeMessage, HealthChannel, HealthCommand, HealthMessage,
};

use super::namespaces::get_namespaces;
use super::session::{create_session, fetch_partners_config, PartnerConfig};
use super::subscription::{subscribe_to_health, subscribe_to_tags};
use super::tag_set::tag_set_from_config_groups;

const CHANNEL_SEND_TIMEOUT: Duration = Duration::from_millis(100);

#[derive(Clone)]
pub(crate) struct SessionManager {
    config_api_url: Url,
    opcua_client: Arc<Mutex<Client>>,
    shutdown_token: CancellationToken,
    data_change_channel: DataChangeChannel,
    health_channel: HealthChannel,
    partners_config: Arc<Mutex<Option<HashSet<Arc<PartnerConfig>>>>>,
    session_senders: Arc<Mutex<HashMap<Arc<PartnerConfig>, oneshot::Sender<SessionCommand>>>>,
}

impl SessionManager {
    pub(crate) fn new(
        config_api_url: Url,
        opcua_client: Client,
        shutdown_token: CancellationToken,
        data_change_channel: DataChangeChannel,
        health_channel: HealthChannel,
    ) -> Self {
        let opcua_client = Arc::new(Mutex::new(opcua_client));
        let sessions_config = Default::default();
        let session_senders = Default::default();
        Self {
            config_api_url,
            opcua_client,
            shutdown_token,
            data_change_channel,
            health_channel,
            partners_config: sessions_config,
            session_senders,
        }
    }

    #[instrument(skip_all, name = "session_manager_run")]
    pub(crate) async fn run(self) {
        info!(status = "started");
        let mut config_refresh_timer = interval(Duration::from_secs(60));
        config_refresh_timer.set_missed_tick_behavior(MissedTickBehavior::Delay);
        let mut ensure_running_timer = interval(Duration::from_secs(1));
        ensure_running_timer.set_missed_tick_behavior(MissedTickBehavior::Delay);
        loop {
            tokio::select! {
                biased;
                _ = self.shutdown_token.cancelled() => {
                    info!(msg = "shutdown signal received", reaction = "exiting");
                    self.stop();
                    break;
                }
                _ = config_refresh_timer.tick() => {
                    if self.refresh_config().await.is_ok() {
                        self.cleanup_sessions().await;
                    } else {
                        config_refresh_timer.reset_after(Duration::from_secs(5));
                    }
                }
                _ = ensure_running_timer.tick() => {
                    self.start_missing_sessions().await;
                    self.restart_stopped().await;
                }
            }
        }
    }

    #[instrument(skip_all)]
    fn stop(&self) {
        self.session_senders
            .lock()
            .unwrap()
            .drain()
            .for_each(|(config, sender)| {
                if sender.send(SessionCommand::Stop).is_err() {
                    let partner_id = &config.partner_id;
                    error!(kind = "sending session stop command", partner_id);
                }
            });
    }

    #[instrument(skip_all)]
    async fn refresh_config(&self) -> Result<(), ()> {
        let fetched_config = fetch_partners_config(self.config_api_url.clone())
            .await
            .map_err(|err| {
                error!(during = "fetching partners config", %err);
            })?
            .into_iter()
            .map(Arc::new)
            .collect::<HashSet<_>>();
        self.partners_config.lock().unwrap().replace(fetched_config);
        Ok(())
    }

    #[instrument(skip_all)]
    fn initialize_session(&self, config: Arc<PartnerConfig>) -> Result<Arc<RwLock<Session>>, ()> {
        let session = {
            let mut client = self.opcua_client.lock().unwrap();
            create_session(&mut client, &config)
        }?;
        let namespaces = get_namespaces(session.clone())?;
        let tag_set = tag_set_from_config_groups(&config.tags, &namespaces, session.clone())?;
        let arc_partner_id: Arc<str> = Arc::from(config.partner_id.as_str());
        let arc_tag_set = Arc::new(tag_set);
        subscribe_to_tags(
            session.clone(),
            arc_partner_id.clone(),
            arc_tag_set,
            self.data_change_channel.clone(),
        )?;
        subscribe_to_health(session.clone(), arc_partner_id, self.health_channel.clone())?;
        Ok(session)
    }

    #[instrument(skip_all, fields(partner_id = config.partner_id))]
    async fn start_session(&self, config: Arc<PartnerConfig>) {
        let empty_request = DataChangeMessage {
            changes: vec![],
            partner_id: config.partner_id.clone(),
        };
        if let Err(err) = self
            .data_change_channel
            .send_timeout(empty_request, CHANNEL_SEND_TIMEOUT)
            .await
        {
            error!(kind = "sending initialization data change to channel", %err);
        }
        let cloned_self = self.clone();
        let cloned_config = Arc::clone(&config);
        let Ok(Ok(session)) = spawn_blocking(move || cloned_self.initialize_session(cloned_config))
            .await
            .map_err(|err| error!(kind = "joining session creation task", %err))
        else {
            return;
        };
        let session_sender = Session::run_async(session);
        self.session_senders
            .lock()
            .unwrap()
            .insert(config, session_sender);
    }

    #[instrument(skip_all)]
    async fn cleanup_sessions(&self) {
        let to_stop = {
            let sessions_config_lock = self.partners_config.lock().unwrap();
            let Some(expected_sessions) = sessions_config_lock.as_ref() else {
                warn!(kind = "sessions config not populated");
                return;
            };
            self.session_senders
                .lock()
                .unwrap()
                .keys()
                .filter(|config| !expected_sessions.contains(*config))
                .cloned()
                .collect::<Vec<_>>()
        };
        for config in to_stop {
            let partner_id = &config.partner_id;
            info!(
                msg = "stopping session not expected in the configuration",
                partner_id
            );
            let health_message = HealthMessage {
                partner_id: partner_id.clone(),
                command: HealthCommand::Remove,
            };
            if let Err(err) = self
                .health_channel
                .send_timeout(health_message, CHANNEL_SEND_TIMEOUT)
                .await
            {
                error!(kind = "sending remove command to health channel", %err);
            }
            let session_sender = self
                .session_senders
                .lock()
                .unwrap()
                .remove(&config)
                .unwrap();
            if session_sender.send(SessionCommand::Stop).is_err() {
                error!(kind = "sending session stop command", partner_id);
            }
        }
    }

    #[instrument(skip_all)]
    async fn start_missing_sessions(&self) {
        let to_start = {
            let session_tasks = self.session_senders.lock().unwrap();
            let sessions_config_lock = self.partners_config.lock().unwrap();
            let Some(expected_sessions) = sessions_config_lock.as_ref() else {
                warn!(kind = "sessions config not populated");
                return;
            };
            expected_sessions
                .iter()
                .filter(|&config| !session_tasks.contains_key(config))
                .cloned()
                .collect::<Vec<_>>()
        };
        let start_tasks_iter = to_start.into_iter().map(|session_config| {
            let partner_id: Arc<str> = Arc::from(session_config.partner_id.as_str());
            let cloned_self = self.clone();
            async move {
                info!(msg = "starting required session", %partner_id);
                cloned_self.start_session(session_config).await;
            }
        });
        join_all(start_tasks_iter).await;
    }

    #[instrument(skip_all)]
    async fn restart_stopped(&self) {
        let mut to_restart = Vec::new();
        for (config, session_sender) in self.session_senders.lock().unwrap().iter_mut() {
            if session_sender.is_closed() {
                to_restart.push(Arc::clone(config));
            }
        }
        let start_tasks_iter = to_restart.into_iter().map(|session_config| {
            self.session_senders.lock().unwrap().remove(&session_config);
            let partner_id: Arc<str> = Arc::from(session_config.partner_id.as_str());
            let cloned_self = self.clone();
            async move {
                info!(msg = "restarting failed session", %partner_id);
                cloned_self.start_session(session_config).await;
            }
        });
        join_all(start_tasks_iter).await;
    }
}
