use anyhow::{anyhow, Context as _};
use opcua::client::prelude::*;
use tokio::sync::mpsc;
use tracing::{error, info, info_span, instrument, warn};

use opcua_proxy::OPCUA_HEALTH_INTERVAL;

use super::tag_set::TagSet;

#[derive(Debug)]
pub(crate) struct TagChange {
    pub(crate) tag_name: String,
    pub(crate) value: super::variant::Variant,
    pub(crate) source_timestamp: i64,
}

#[instrument(skip_all)]
pub(crate) fn subscribe_to_tags<T>(
    session: &T,
    tag_set: TagSet,
) -> anyhow::Result<mpsc::Receiver<Vec<TagChange>>>
where
    T: SubscriptionService + MonitoredItemService,
{
    let (sender, receiver) = mpsc::channel(1);
    let cloned_tag_set = tag_set.clone().into_inner();
    let data_change_callback = DataChangeCallback::new(move |monitored_items| {
        let _entered = info_span!("tags_values_change_handler").entered();
        let mut message = Vec::with_capacity(monitored_items.len());
        for item in monitored_items {
            let node_id = &item.item_to_monitor().node_id;
            let client_handle = item.client_handle();
            let index = usize::try_from(client_handle).unwrap() - 1;
            let Some(tag) = cloned_tag_set.get(index) else {
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
            message.push(TagChange {
                tag_name: tag.name.clone(),
                value: last_value.clone().into(),
                source_timestamp,
            })
        }
        if message.is_empty() {
            warn!(msg = "discarded empty tags changes message");
            return;
        }
        if let Err(err) = sender.try_send(message) {
            error!(when = "sending message to channel", %err);
        }
    });

    let items_to_create = tag_set
        .monitored_items()
        .context("error creating monitored items create requests")?;

    let subscription_id = session
        .create_subscription(1000.0, 50, 10, 0, 0, true, data_change_callback)
        .context("error creating subscription")?;

    let results = session
        .create_monitored_items(
            subscription_id,
            TimestampsToReturn::Source,
            &items_to_create,
        )
        .context("error creating monitored items")?;

    for (i, MonitoredItemCreateResult { status_code, .. }) in results.iter().enumerate() {
        if !status_code.is_good() {
            let node_id = &items_to_create[i].item_to_monitor.node_id;
            return Err(anyhow!(
                "error creating monitored item for {}: {}",
                node_id,
                status_code
            ));
        }
    }

    info!(status = "success");
    Ok(receiver)
}

#[instrument(skip_all)]
pub(crate) fn subscribe_to_health<T>(session: &T) -> anyhow::Result<mpsc::Receiver<i64>>
where
    T: SubscriptionService + MonitoredItemService,
{
    let (sender, receiver) = mpsc::channel(1);
    let data_change_callback = DataChangeCallback::new(move |monitored_items| {
        let _entered = info_span!("health_value_change_handler").entered();
        let Some(Variant::DateTime(server_time)) = monitored_items
            .get(0)
            .and_then(|item| item.last_value().value.as_ref())
        else {
            error!(?monitored_items, err = "unexpected monitored items");
            return;
        };
        if let Err(err) = sender.try_send(server_time.as_chrono().timestamp_millis()) {
            error!(when = "sending message to channel", %err);
        }
    });

    let subscription_id = session
        .create_subscription(
            OPCUA_HEALTH_INTERVAL.into(),
            50,
            10,
            1,
            0,
            true,
            data_change_callback,
        )
        .context("error creating subscription")?;

    let server_time_node: NodeId = VariableId::Server_ServerStatus_CurrentTime.into();

    let results = session
        .create_monitored_items(
            subscription_id,
            TimestampsToReturn::Neither,
            &[server_time_node.into()],
        )
        .context("error creating monitored item")?;

    let result = results
        .get(0)
        .ok_or_else(|| anyhow!("missing result for monitored item creation"))?;

    if !result.status_code.is_good() {
        return Err(anyhow!("bad status code : {}", result.status_code));
    }

    info!(status = "success");
    Ok(receiver)
}
