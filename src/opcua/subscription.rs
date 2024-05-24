use std::sync::Arc;

use opcua::client::prelude::*;
use opcua::sync::RwLock;
use tracing::{error, info_span, instrument, warn};

use opcua_proxy::OPCUA_HEALTH_INTERVAL;

use crate::db::{
    DataChangeChannel, DataChangeMessage, HealthChannel, HealthCommand, HealthMessage,
};

use super::session::SESSION_LOCK_TIMEOUT;
use super::tag_set::Tag;

#[derive(Debug)]
pub(crate) struct TagChange {
    pub(crate) tag_name: String,
    pub(crate) value: super::variant::Variant,
    pub(crate) source_timestamp: i64,
}

#[instrument(skip_all, fields(partner_id))]
pub(super) fn subscribe_to_tags<T>(
    session: Arc<RwLock<T>>,
    partner_id: Arc<str>,
    tag_set: Arc<Vec<Tag>>,
    data_change_channel: DataChangeChannel,
) -> Result<(), ()>
where
    T: SubscriptionService + MonitoredItemService,
{
    let shared_tag_set = Arc::clone(&tag_set);
    let data_change_callback = DataChangeCallback::new(move |monitored_items| {
        let _entered = info_span!("tags_values_change_handler", %partner_id).entered();
        let mut changes = Vec::with_capacity(monitored_items.len());
        for item in monitored_items {
            let node_id = &item.item_to_monitor().node_id;
            let client_handle = item.client_handle();
            let index = usize::try_from(client_handle).unwrap() - 1;
            let Some(tag) = shared_tag_set.get(index) else {
                error!(%node_id, client_handle, err="tag not found for client handle");
                continue;
            };
            let Some(last_value) = &item.last_value().value else {
                warn!(%node_id, msg = "missing value");
                continue;
            };
            let Some(source_timestamp) = item
                .last_value()
                .source_timestamp
                .map(|dt| dt.as_chrono().timestamp_millis())
            else {
                error!(%node_id, err = "missing source timestamp");
                continue;
            };
            changes.push(TagChange {
                tag_name: tag.name.clone(),
                value: last_value.clone().into(),
                source_timestamp,
            })
        }
        if changes.is_empty() {
            warn!(msg = "discarded empty tags changes message");
            return;
        }
        let message = DataChangeMessage {
            partner_id: partner_id.to_string(),
            changes,
        };
        if let Err(err) = data_change_channel.try_send(message) {
            error!(when = "sending message to channel", %err);
        }
    });

    let subscription_id = {
        let session = session
            .try_read_for(SESSION_LOCK_TIMEOUT)
            .ok_or_else(|| error!(kind = "session lock timeout"))?;
        session
            .create_subscription(1000.0, 50, 10, 0, 0, true, data_change_callback)
            .map_err(|err| {
                error!(kind = "subscription creation", %err);
            })?
    };

    let items_to_create = tag_set
        .iter()
        .zip(1..)
        .map(|(tag, client_handle)| {
            MonitoredItemCreateRequest::new(
                tag.node_id.clone().into(),
                MonitoringMode::Reporting,
                MonitoringParameters {
                    client_handle,
                    ..Default::default()
                },
            )
        })
        .collect::<Vec<_>>();

    let results = {
        let session = session
            .try_read_for(SESSION_LOCK_TIMEOUT)
            .ok_or_else(|| error!(kind = "session lock timeout"))?;
        session
            .create_monitored_items(
                subscription_id,
                TimestampsToReturn::Source,
                &items_to_create,
            )
            .map_err(|err| {
                error!(kind = "monitored items creation", %err);
            })?
    };

    for (i, MonitoredItemCreateResult { status_code, .. }) in results.iter().enumerate() {
        if !status_code.is_good() {
            let node_id = &items_to_create[i].item_to_monitor.node_id;
            error!(kind = "monitored item status", %node_id, %status_code);
            return Err(());
        }
    }

    Ok(())
}

#[instrument(skip_all, fields(partner_id))]
pub(super) fn subscribe_to_health<T>(
    session: Arc<RwLock<T>>,
    partner_id: Arc<str>,
    health_channel: HealthChannel,
) -> Result<(), ()>
where
    T: SubscriptionService + MonitoredItemService,
{
    let data_change_callback = DataChangeCallback::new(move |monitored_items| {
        let _entered = info_span!("health_value_change_handler", %partner_id).entered();
        let Some(Variant::DateTime(server_time)) = monitored_items
            .first()
            .and_then(|item| item.last_value().value.as_ref())
        else {
            error!(?monitored_items, err = "unexpected monitored items");
            return;
        };
        let server_timestamp = server_time.as_chrono().timestamp_millis();
        let message = HealthMessage {
            partner_id: partner_id.to_string(),
            command: HealthCommand::Update(server_timestamp),
        };
        if let Err(err) = health_channel.try_send(message) {
            error!(when = "sending message to channel", %err);
        }
    });

    let subscription_id = {
        let session = session
            .try_read_for(SESSION_LOCK_TIMEOUT)
            .ok_or_else(|| error!(kind = "session lock timeout"))?;
        session
            .create_subscription(
                OPCUA_HEALTH_INTERVAL.into(),
                50,
                10,
                1,
                0,
                true,
                data_change_callback,
            )
            .map_err(|err| {
                error!(kind = "subscription creation", %err);
            })?
    };

    let server_time_node: NodeId = VariableId::Server_ServerStatus_CurrentTime.into();

    let results = {
        let session = session
            .try_read_for(SESSION_LOCK_TIMEOUT)
            .ok_or_else(|| error!(kind = "session lock timeout"))?;
        session
            .create_monitored_items(
                subscription_id,
                TimestampsToReturn::Neither,
                &[server_time_node.into()],
            )
            .map_err(|err| {
                error!(kind = "monitored items creation", %err);
            })?
    };

    let status_code = results
        .first()
        .map(|result| result.status_code)
        .ok_or_else(|| {
            error!(kind = "misssing result for monitored item creation");
        })?;

    if !status_code.is_good() {
        error!(kind = "monitored item status", %status_code);
        return Err(());
    }

    Ok(())
}
