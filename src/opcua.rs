use std::collections::HashMap;
use std::fmt::Display;
use std::sync::Arc;

use anyhow::{anyhow, Context as _};
use clap::Args;
use opcua::client::prelude::*;
pub(crate) use opcua::client::prelude::{Session, SessionCommand};
use opcua::sync::RwLock;
use serde::Deserialize;
use tokio::sync::mpsc;
use tracing::{error, info, info_span, instrument, warn};

use opcua_proxy::OPCUA_HEALTH_INTERVAL;

use crate::model::{NodeIdentifier, TagChange, TagsConfigGroup};

#[derive(Args)]
pub(crate) struct Config {
    /// Path of the PKI directory
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

impl From<NodeIdentifier> for Identifier {
    fn from(identifier: NodeIdentifier) -> Self {
        match identifier {
            NodeIdentifier::Numeric(n) => Identifier::Numeric(n),
            NodeIdentifier::String(s) => Identifier::String(s.into()),
        }
    }
}

#[derive(Debug, Clone, Deserialize)]
struct Tag {
    name: String,
    node_id: NodeId,
}

#[derive(Debug)]
pub(crate) struct TagSet(Vec<Tag>);

impl TagSet {
    #[instrument(name = "tag_set_from_config", skip_all)]
    pub(crate) fn from_config(
        config: Vec<TagsConfigGroup>,
        namespaces: &Namespaces,
        session: &impl ViewService,
    ) -> anyhow::Result<Self> {
        let mut tag_set: Vec<Tag> = Vec::new();

        for config_group in config {
            match config_group {
                TagsConfigGroup::Container {
                    namespace_uri,
                    node_identifier,
                } => {
                    let namespace = namespaces
                        .get(&namespace_uri)
                        .with_context(|| format!("namespace `{namespace_uri}` not found"))?;
                    let node_id = NodeId::new(*namespace, node_identifier);
                    let browse_description = BrowseDescription {
                        node_id,
                        browse_direction: BrowseDirection::Forward,
                        reference_type_id: ReferenceTypeId::HasComponent.into(),
                        include_subtypes: false,
                        node_class_mask: NodeClassMask::VARIABLE.bits(),
                        result_mask: BrowseDescriptionResultMask::RESULT_MASK_DISPLAY_NAME.bits(),
                    };
                    let browse_result = session
                        .browse(&[browse_description])
                        .context("Browse error")?
                        .unwrap()
                        .pop()
                        .context("empty Browse results")?;
                    if !browse_result.status_code.is_good() {
                        return Err(anyhow!("BrowseResult error: {}", browse_result.status_code));
                    }
                    if !browse_result.continuation_point.is_null() {
                        return Err(anyhow!(
                            "got a ContinuationPoint, handling it is unimplemented"
                        ));
                    }
                    for ReferenceDescription {
                        node_id,
                        display_name,
                        ..
                    } in browse_result.references.unwrap()
                    {
                        tag_set.push(Tag {
                            name: display_name.to_string(),
                            node_id: node_id.node_id,
                        });
                    }
                }
                TagsConfigGroup::Tag {
                    name,
                    namespace_uri,
                    node_identifier,
                } => {
                    let namespace = namespaces
                        .get(&namespace_uri)
                        .with_context(|| format!("namespace `{namespace_uri}` not found"))?;
                    let node_id = NodeId::new(*namespace, node_identifier);
                    tag_set.push(Tag { name, node_id })
                }
            }
        }

        info!(status = "success");
        Ok(Self(tag_set))
    }

    pub(crate) fn check_contains_tags<T>(&self, tags_names: &[T]) -> anyhow::Result<()>
    where
        T: AsRef<str> + Display,
    {
        for tag_name in tags_names {
            if !self.0.iter().any(|tag| tag.name == tag_name.as_ref()) {
                return Err(anyhow!("tag `{tag_name}` was not found in the tag set"));
            }
        }
        Ok(())
    }

    fn monitored_items(&self) -> anyhow::Result<Vec<MonitoredItemCreateRequest>> {
        self.0
            .iter()
            .zip(1..)
            .map(|(tag, client_handle)| {
                let create_request = MonitoredItemCreateRequest::new(
                    tag.node_id.clone().into(),
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

#[instrument(skip_all)]
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

#[instrument(skip_all)]
pub(crate) fn subscribe_to_tags<T>(
    session: &T,
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

#[cfg(test)]
mod tests {
    use super::*;

    mod tag_set {
        use super::*;

        mod from_config {
            use super::*;

            struct ViewServiceMock {
                browse_outcome: Result<Option<Vec<BrowseResult>>, StatusCode>,
            }

            #[allow(unused_variables)]
            impl Service for ViewServiceMock {
                fn make_request_header(&self) -> RequestHeader {
                    unimplemented!();
                }

                fn send_request<T>(&self, request: T) -> Result<SupportedMessage, StatusCode>
                where
                    T: Into<SupportedMessage>,
                {
                    unimplemented!();
                }

                fn async_send_request<T>(
                    &self,
                    request: T,
                    sender: Option<std::sync::mpsc::SyncSender<SupportedMessage>>,
                ) -> Result<u32, StatusCode>
                where
                    T: Into<SupportedMessage>,
                {
                    unimplemented!();
                }
            }

            #[allow(unused_variables)]
            impl ViewService for ViewServiceMock {
                fn browse(
                    &self,
                    nodes_to_browse: &[BrowseDescription],
                ) -> Result<Option<Vec<BrowseResult>>, StatusCode> {
                    self.browse_outcome.clone()
                }

                fn browse_next(
                    &self,
                    release_continuation_points: bool,
                    continuation_points: &[ByteString],
                ) -> Result<Option<Vec<BrowseResult>>, StatusCode> {
                    unimplemented!();
                }

                fn translate_browse_paths_to_node_ids(
                    &self,
                    browse_paths: &[BrowsePath],
                ) -> Result<Vec<BrowsePathResult>, StatusCode> {
                    unimplemented!();
                }

                fn register_nodes(
                    &self,
                    nodes_to_register: &[NodeId],
                ) -> Result<Vec<NodeId>, StatusCode> {
                    unimplemented!();
                }

                fn unregister_nodes(
                    &self,
                    nodes_to_unregister: &[NodeId],
                ) -> Result<(), StatusCode> {
                    unimplemented!();
                }
            }

            #[test]
            fn container_namespace_not_found() {
                let config = vec![TagsConfigGroup::Container {
                    namespace_uri: "nonexistent".to_string(),
                    node_identifier: NodeIdentifier::Numeric(0),
                }];
                let namespaces = HashMap::new();
                let session = ViewServiceMock {
                    browse_outcome: Ok(Some(vec![])),
                };

                let result = TagSet::from_config(config, &namespaces, &session);

                assert!(result.is_err());
            }

            #[test]
            fn browse_error() {
                let config = vec![TagsConfigGroup::Container {
                    namespace_uri: "urn:ns".to_string(),
                    node_identifier: NodeIdentifier::Numeric(0),
                }];
                let namespaces = HashMap::from([("urn:ns".to_string(), 1)]);
                let session = ViewServiceMock {
                    browse_outcome: Err(StatusCode::BadInternalError),
                };

                let result = TagSet::from_config(config, &namespaces, &session);

                assert!(result.is_err());
            }

            #[test]
            fn empty_browse_results() {
                let config = vec![TagsConfigGroup::Container {
                    namespace_uri: "urn:ns".to_string(),
                    node_identifier: NodeIdentifier::Numeric(0),
                }];
                let namespaces = HashMap::from([("urn:ns".to_string(), 1)]);
                let session = ViewServiceMock {
                    browse_outcome: Ok(Some(vec![])),
                };

                let result = TagSet::from_config(config, &namespaces, &session);

                assert!(result.is_err());
            }

            #[test]
            fn continuation_point() {
                let config = vec![TagsConfigGroup::Container {
                    namespace_uri: "urn:ns".to_string(),
                    node_identifier: NodeIdentifier::Numeric(0),
                }];
                let namespaces = HashMap::from([("urn:ns".to_string(), 1)]);
                let session = ViewServiceMock {
                    browse_outcome: Ok(Some(vec![BrowseResult {
                        status_code: StatusCode::Good,
                        continuation_point: "deadbeef".into(),
                        references: Some(vec![]),
                    }])),
                };

                let result = TagSet::from_config(config, &namespaces, &session);

                assert!(result.is_err());
            }

            #[test]
            fn tag_namespace_not_found() {
                let config = vec![TagsConfigGroup::Tag {
                    name: "somename".to_string(),
                    namespace_uri: "nonexistent".to_string(),
                    node_identifier: NodeIdentifier::Numeric(0),
                }];
                let namespaces = HashMap::new();
                let session = ViewServiceMock {
                    browse_outcome: Ok(Some(vec![])),
                };

                let result = TagSet::from_config(config, &namespaces, &session);

                assert!(result.is_err());
            }

            #[test]
            fn success() {
                let config = vec![
                    TagsConfigGroup::Container {
                        namespace_uri: "urn:ns".to_string(),
                        node_identifier: NodeIdentifier::Numeric(0),
                    },
                    TagsConfigGroup::Tag {
                        name: "somename".to_string(),
                        namespace_uri: "urn:ns2".to_string(),
                        node_identifier: NodeIdentifier::String("some_node_id".to_string()),
                    },
                ];
                let namespaces =
                    HashMap::from([("urn:ns".to_string(), 1), ("urn:ns2".to_string(), 2)]);
                let references = [VariableId::LocalTime, VariableId::Server_ServiceLevel]
                    .iter()
                    .map(|var| ReferenceDescription {
                        reference_type_id: NodeId::null(),
                        is_forward: true,
                        node_id: NodeId::from(var).into(),
                        browse_name: QualifiedName::null(),
                        display_name: format!("{var:?}").into(),
                        node_class: NodeClass::Unspecified,
                        type_definition: ExpandedNodeId::null(),
                    })
                    .collect();
                let session = ViewServiceMock {
                    browse_outcome: Ok(Some(vec![BrowseResult {
                        status_code: StatusCode::Good,
                        continuation_point: ByteString::null(),
                        references: Some(references),
                    }])),
                };

                let result = TagSet::from_config(config, &namespaces, &session)
                    .expect("result should not be an error");

                assert_eq!(result.0[0].name, "LocalTime");
                assert_eq!(result.0[0].node_id, VariableId::LocalTime.into());
                assert_eq!(result.0[1].name, "Server_ServiceLevel");
                assert_eq!(result.0[1].node_id, VariableId::Server_ServiceLevel.into());
                assert_eq!(result.0[2].name, "somename");
                assert_eq!(
                    result.0[2].node_id,
                    NodeId::new(2, NodeIdentifier::String("some_node_id".to_string()))
                )
            }
        }

        mod check_contains_tags {
            use super::*;

            fn create_tag_set() -> TagSet {
                TagSet(vec![
                    Tag {
                        name: "firstTag".into(),
                        node_id: (0, "").into(),
                    },
                    Tag {
                        name: "secondTag".into(),
                        node_id: (0, "").into(),
                    },
                    Tag {
                        name: "thirdTag".into(),
                        node_id: (0, "").into(),
                    },
                ])
            }

            #[test]
            fn empty_tags_names() {
                let tag_set = create_tag_set();

                let result = tag_set.check_contains_tags::<&str>(&[]);

                assert!(result.is_ok());
            }

            #[test]
            fn missing_tag_name() {
                let tag_set = create_tag_set();
                let tags_names = &["secondTag", "nonExistent"];

                let result = tag_set.check_contains_tags(tags_names);

                assert!(result.is_err());
            }

            #[test]
            fn success() {
                let tag_set = create_tag_set();
                let tags_names = &["firstTag", "thirdTag"];

                let result = tag_set.check_contains_tags(tags_names);

                assert!(result.is_ok());
            }
        }
    }

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
