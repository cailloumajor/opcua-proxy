use std::collections::HashMap;
use std::fs;
use std::sync::Arc;

use anyhow::{anyhow, Context as _};
use clap::Args;
use opcua::client::prelude::*;
pub(crate) use opcua::client::prelude::{Session, SessionCommand};
use opcua::sync::RwLock;
use serde::Deserialize;
use tokio::sync::mpsc;
use tracing::info;
use tracing::{error, info_span};

use opcua_proxy::OPCUA_HEALTH_INTERVAL;

use crate::model::TagChange;

#[derive(Args)]
pub(crate) struct Config {
    // Path of the PKI directory
    #[arg(env, long)]
    pki_dir: String,

    /// URL of OPC-UA server to connect to
    #[arg(env, long)]
    opcua_server_url: String,

    /// OPC-UA security policy
    #[arg(env, long, default_value = "Basic256Sha256")]
    opcua_security_policy: String,

    /// OPC-UA security mode
    #[arg(env, long, default_value = "SignAndEncrypt")]
    opcua_security_mode: String,

    /// OPC-UA authentication username (optional)
    #[arg(env, long)]
    opcua_user: Option<String>,

    /// OPC-UA authentication password (optional)
    #[arg(env, long)]
    opcua_password: Option<String>,
}

type Namespaces = HashMap<String, u16>;

#[derive(Clone, Deserialize)]
#[serde(untagged)]
enum NodeIdentifier {
    Numeric(u32),
    String(String),
}

impl From<NodeIdentifier> for Identifier {
    fn from(identifier: NodeIdentifier) -> Self {
        match identifier {
            NodeIdentifier::Numeric(n) => Identifier::Numeric(n),
            NodeIdentifier::String(s) => Identifier::String(s.into()),
        }
    }
}

#[derive(Clone, Deserialize)]
struct Tag {
    name: String,
    nsu: String,
    nid: NodeIdentifier,
}

pub(crate) struct TagSet(Vec<Tag>);

impl TagSet {
    pub fn from_file(path: &str) -> anyhow::Result<Self> {
        let contents =
            fs::read_to_string(path).with_context(|| format!("error reading file {path}"))?;
        let json = serde_json::from_str(&contents)
            .with_context(|| format!("error deserializing contents of file {path}"))?;

        Ok(Self(json))
    }

    fn monitored_items(
        &self,
        namespaces: &Namespaces,
    ) -> anyhow::Result<Vec<MonitoredItemCreateRequest>> {
        self.0
            .iter()
            .zip(1..)
            .map(|(tag, client_handle)| {
                let ns = namespaces
                    .get(&tag.nsu)
                    .with_context(|| format!("namespace not found: {}", tag.nsu))?;
                let create_request = MonitoredItemCreateRequest::new(
                    NodeId::new(*ns, tag.nid.clone()).into(),
                    MonitoringMode::Reporting,
                    MonitoringParameters {
                        client_handle,
                        ..Default::default()
                    },
                );
                Ok(create_request)
            })
            .collect()
    }
}

#[tracing::instrument(skip_all)]
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

    let mut client_builder = ClientBuilder::new()
        .application_name(env!("CARGO_PKG_DESCRIPTION"))
        .product_uri(PRODUCT_URI)
        .application_uri(format!("{PRODUCT_URI}:{partner_id}"))
        .pki_dir(config.pki_dir.clone())
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

#[tracing::instrument(skip_all)]
pub(crate) fn get_namespaces(session: &impl AttributeService) -> anyhow::Result<Namespaces> {
    let namespace_array_nodeid: NodeId = VariableId::Server_NamespaceArray.into();
    let read_result = session.read(
        &[namespace_array_nodeid.into()],
        TimestampsToReturn::Neither,
        0.0,
    )?;
    let data_value = read_result
        .get(0)
        .ok_or_else(|| anyhow!("missing namespace array"))?;
    let result_variant = data_value
        .value
        .as_ref()
        .ok_or_else(|| anyhow!("value error: {}", data_value.status().description()))?;
    let namespace_variants = match result_variant {
        Variant::Array(array) => Ok(&array.values),
        _ => Err(anyhow!(
            "bad value type: {:?} (expected an array)",
            result_variant.type_id()
        )),
    }?;
    let namespaces = namespace_variants
        .iter()
        .zip(0..)
        .map(|(variant, namespace_index)| match variant {
            Variant::String(uastring) => Ok((uastring.to_string(), namespace_index)),
            _ => Err(anyhow!(
                "bad member type: {:?} (expected a string)",
                variant.type_id()
            )),
        })
        .collect::<anyhow::Result<Vec<_>>>()?
        .into_iter()
        .collect();

    info!(status = "success");
    Ok(namespaces)
}

#[tracing::instrument(skip_all)]
pub(crate) fn subscribe_to_tags<T>(
    session: &T,
    namespaces: &Namespaces,
    tag_set: Arc<TagSet>,
) -> anyhow::Result<mpsc::Receiver<Vec<TagChange>>>
where
    T: SubscriptionService + MonitoredItemService,
{
    let (sender, receiver) = mpsc::channel(1);
    let cloned_tag_set = Arc::clone(&tag_set);
    let data_change_callback = DataChangeCallback::new(move |monitored_items| {
        let _entered = info_span!("tags_values_change_handler").entered();
        let mut message = Vec::with_capacity(monitored_items.len());
        for item in monitored_items {
            let node_id = &item.item_to_monitor().node_id;
            let client_handle = item.client_handle();
            let index = usize::try_from(client_handle).unwrap() - 1;
            let Some(tag) = cloned_tag_set.0.get(index) else {
                error!(%node_id, client_handle, err="tag not found for client handle");
                continue;
            };
            let Some(last_value) = &item.last_value().value else {
                error!(%node_id, err="missing value");
                continue;
            };
            let Some(source_timestamp) = item
                .last_value()
                .source_timestamp
                .map(|dt| dt.as_chrono().timestamp_millis())
            else {
                error!(%node_id, err="missing source timestamp");
                continue;
            };
            message.push(TagChange {
                tag_name: tag.name.clone(),
                value: last_value.clone().into(),
                source_timestamp,
            })
        }
        if let Err(err) = sender.try_send(message) {
            error!(when = "sending message to channel", %err);
        }
    });

    let items_to_create = tag_set
        .monitored_items(namespaces)
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

#[tracing::instrument(skip_all)]
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

#[cfg(test)]
mod tests {
    use super::*;

    mod get_namespaces {
        use super::*;

        enum TestCases {
            ReadError,
            EmptyResult,
            NoValue,
            NotAnArray,
            BadMemberType,
            Success,
        }

        struct AttributeServiceMock(TestCases);

        impl Service for AttributeServiceMock {
            fn make_request_header(&self) -> RequestHeader {
                RequestHeader::dummy()
            }

            fn send_request<T>(&self, _request: T) -> Result<SupportedMessage, StatusCode>
            where
                T: Into<SupportedMessage>,
            {
                Err(StatusCode::empty())
            }

            fn async_send_request<T>(
                &self,
                _request: T,
                _sender: Option<std::sync::mpsc::SyncSender<SupportedMessage>>,
            ) -> Result<u32, StatusCode>
            where
                T: Into<SupportedMessage>,
            {
                Err(StatusCode::empty())
            }
        }

        impl AttributeService for AttributeServiceMock {
            fn read(
                &self,
                _nodes_to_read: &[ReadValueId],
                _timestamps_to_return: TimestampsToReturn,
                _max_age: f64,
            ) -> Result<Vec<DataValue>, StatusCode> {
                match &self.0 {
                    TestCases::ReadError => Err(StatusCode::empty()),
                    TestCases::EmptyResult => Ok(Vec::new()),
                    TestCases::NoValue => Ok(vec![DataValue::null()]),
                    TestCases::NotAnArray => Ok(vec![Variant::from(false).into()]),
                    TestCases::BadMemberType => Ok(vec![Variant::from(vec![false]).into()]),
                    TestCases::Success => Ok(vec![Variant::from(vec![
                        "urn:ns:ns1".to_string(),
                        "urn:ns:ns2".to_string(),
                    ])
                    .into()]),
                }
            }

            fn history_read(
                &self,
                _history_read_details: HistoryReadAction,
                _timestamps_to_return: TimestampsToReturn,
                _release_continuation_points: bool,
                _nodes_to_read: &[HistoryReadValueId],
            ) -> Result<Vec<HistoryReadResult>, StatusCode> {
                Err(StatusCode::empty())
            }

            fn write(&self, _nodes_to_write: &[WriteValue]) -> Result<Vec<StatusCode>, StatusCode> {
                Err(StatusCode::empty())
            }

            fn history_update(
                &self,
                _history_update_details: &[HistoryUpdateAction],
            ) -> Result<Vec<HistoryUpdateResult>, StatusCode> {
                Err(StatusCode::empty())
            }
        }

        #[test]
        fn missing_read_result() {
            let mock = AttributeServiceMock(TestCases::ReadError);
            let result = get_namespaces(&mock);
            assert!(result.is_err());
        }

        #[test]
        fn missing_data_value() {
            let mock = AttributeServiceMock(TestCases::EmptyResult);
            let result = get_namespaces(&mock);
            assert!(result.is_err());
        }

        #[test]
        fn missing_value() {
            let mock = AttributeServiceMock(TestCases::NoValue);
            let result = get_namespaces(&mock);
            assert!(result.is_err());
        }

        #[test]
        fn not_an_array() {
            let mock = AttributeServiceMock(TestCases::NotAnArray);
            let result = get_namespaces(&mock);
            assert!(result.is_err());
        }

        #[test]
        fn bad_member_type() {
            let mock = AttributeServiceMock(TestCases::BadMemberType);
            let result = get_namespaces(&mock);
            assert!(result.is_err());
        }

        #[test]
        fn success() {
            let mock = AttributeServiceMock(TestCases::Success);
            let result = get_namespaces(&mock);
            let expected = HashMap::from([
                ("urn:ns:ns1".to_string(), 0u16),
                ("urn:ns:ns2".to_string(), 1u16),
            ]);
            assert_eq!(result.unwrap(), expected);
        }
    }
}
