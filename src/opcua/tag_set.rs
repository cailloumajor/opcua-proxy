use anyhow::{anyhow, Context as _};
use opcua::client::prelude::*;
use serde::Deserialize;
use tracing::{info, instrument};

use super::namespaces::Namespaces;
use super::node_identifier::NodeIdentifier;

#[derive(Debug, Clone, Deserialize)]
pub(super) struct Tag {
    pub(super) name: String,
    pub(super) node_id: NodeId,
}

#[derive(Clone, Debug)]
pub(crate) struct TagSet(Vec<Tag>);

impl TagSet {
    pub(super) fn into_inner(self) -> Vec<Tag> {
        self.0
    }
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase", tag = "type")]
pub(crate) enum TagsConfigGroup {
    #[serde(rename_all = "camelCase")]
    Container {
        namespace_uri: String,
        node_identifier: NodeIdentifier,
    },
    #[serde(rename_all = "camelCase")]
    Tag {
        name: String,
        namespace_uri: String,
        node_identifier: NodeIdentifier,
    },
}

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
                        node_id: node_id.clone(),
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
                    let references = browse_result
                        .references
                        .with_context(|| format!("NodeId `{node_id}` has no forward reference"))?;
                    for ReferenceDescription {
                        node_id,
                        display_name,
                        ..
                    } in references
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

    pub(super) fn monitored_items(&self) -> anyhow::Result<Vec<MonitoredItemCreateRequest>> {
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

#[cfg(test)]
mod tests {
    use super::*;

    mod from_config {
        use std::collections::HashMap;

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

            fn unregister_nodes(&self, nodes_to_unregister: &[NodeId]) -> Result<(), StatusCode> {
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
            let namespaces = HashMap::from([("urn:ns".to_string(), 1), ("urn:ns2".to_string(), 2)]);
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
}
