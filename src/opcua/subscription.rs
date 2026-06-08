use std::collections::HashMap;
use std::sync::Arc;
use std::time::Duration;

use opcua::client::{MonitoredItem, OnSubscriptionNotification, Session};
use opcua::types::{
    DataValue, DateTime, Error as OpcUaError, MonitoredItemCreateRequest, MonitoringMode,
    MonitoringParameters, TimestampsToReturn, Variant,
};
use parking_lot::Mutex;
use tokio::sync::mpsc;
use tracing::{error, info_span, warn};

use crate::centrifugo::TagChangeMessage;
use crate::opcua::utils::encode_tag_changes;

use super::tag_set::Tag;

/// A subscriber for OPC-UA changes notification that signals changes to Centrifugo.
pub(super) struct CentrifugoSubscriber {
    /// The OPC-UA session.
    session: Arc<Session>,
    /// The ID of the OPC-UA remote partner.
    partner_id: String,
    /// The set of tags to monitor.
    tag_set: Vec<Tag>,
    /// Storage for current tag values.
    values_map: Arc<Mutex<HashMap<String, (Variant, DateTime)>>>,
    /// The channel to send tag changes.
    tag_change_channel: mpsc::Sender<TagChangeMessage>,
}

impl CentrifugoSubscriber {
    const PUBLISH_INTERVAL: Duration = Duration::from_secs(1);

    /// Create a new [`CentrifugoSubscriber`].
    pub(super) fn new(
        session: Arc<Session>,
        partner_id: String,
        tag_set: Vec<Tag>,
        values_map: Arc<Mutex<HashMap<String, (Variant, DateTime)>>>,
        tag_change_channel: mpsc::Sender<TagChangeMessage>,
    ) -> Self {
        Self {
            session,
            partner_id,
            tag_set,
            values_map,
            tag_change_channel,
        }
    }

    /// Enable this subscriber in the OPC-UA session, i.e. create the subscription and monitored items.
    pub(super) async fn enable(self) -> Result<(), OpcUaError> {
        let items_to_create = self
            .tag_set
            .iter()
            // Client handles start at 1.
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
            .collect();

        let cloned_session = Arc::clone(&self.session);
        let subscription_id = cloned_session
            .create_subscription(Self::PUBLISH_INTERVAL, 50, 10, 0, 0, true, self)
            .await
            .map_err(|status| OpcUaError::new(status, "error creating subscription"))?;

        let results = cloned_session
            .create_monitored_items(subscription_id, TimestampsToReturn::Source, items_to_create)
            .await
            .map_err(|status| OpcUaError::new(status, "error creating monitored items"))?;

        for item in results {
            let status_code = item.result.status_code;
            if !status_code.is_good() {
                return Err(OpcUaError::new(
                    status_code,
                    format!("error on monitored item {}", item.item_to_monitor.node_id),
                ));
            }
        }

        Ok(())
    }
}

impl OnSubscriptionNotification for CentrifugoSubscriber {
    fn on_data_value(&mut self, notification: DataValue, item: &MonitoredItem) {
        let node_id = &item.item_to_monitor().node_id;

        let _entered = info_span!("tags_values_change_handler", %node_id).entered();

        let client_handle = item.client_handle();
        // Client handle starts at 1.
        let tag_index = match client_handle.checked_sub(1) {
            Some(idx) => usize::try_from(idx).expect("u32 should fit in usize"),
            None => {
                error!(err = "client handle is zero");
                return;
            }
        };
        let Some(tag) = self.tag_set.get(tag_index) else {
            error!(err = "tag not found for client handle", client_handle);
            return;
        };
        let Some(value) = notification.value else {
            warn!(msg = "missing value");
            return;
        };
        let Some(source_timestamp) = notification.source_timestamp else {
            error!(err = "missing source timestamp");
            return;
        };

        let data = {
            let ctx = self.session.encoding_context().read();
            match encode_tag_changes(&[(&tag.name, &value, &source_timestamp)], &ctx.context()) {
                Ok(d) => d,
                Err(err) => {
                    error!(during = "encoding tag change to JSON", %err);
                    return;
                }
            }
        };

        // Record the current value.
        self.values_map
            .lock_arc()
            .insert(tag.name.clone(), (value, source_timestamp));

        let message = TagChangeMessage {
            partner_id: self.partner_id.clone(),
            data,
        };

        if let Err(err) = self.tag_change_channel.try_send(message) {
            error!(when = "sending message to channel", %err);
        }
    }
}
