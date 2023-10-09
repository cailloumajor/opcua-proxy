use std::sync::Arc;

use opcua::client::prelude::*;
use opcua::sync::RwLock;
use serde::Deserialize;
use tracing::{error, instrument};

use super::namespaces::Namespaces;
use super::node_identifier::NodeIdentifier;
use super::session::SESSION_LOCK_TIMEOUT;

#[derive(Debug, Clone, Deserialize)]
pub(super) struct Tag {
    pub(super) name: String,
    pub(super) node_id: NodeId,
}

#[derive(Debug, Deserialize, Eq, Hash, PartialEq)]
#[serde(rename_all = "camelCase", tag = "type")]
pub(super) enum TagsConfigGroup {
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

#[instrument(skip_all)]
pub(super) fn tag_set_from_config_groups<T>(
    config: &[TagsConfigGroup],
    namespaces: &Namespaces,
    session: Arc<RwLock<T>>,
) -> Result<Vec<Tag>, ()>
where
    T: ViewService,
{
    let mut tag_set: Vec<Tag> = Vec::new();

    for config_group in config {
        match config_group {
            TagsConfigGroup::Container {
                namespace_uri,
                node_identifier,
            } => {
                let namespace = namespaces.get(namespace_uri).ok_or_else(|| {
                    error!(kind = "namespace not found", namespace_uri);
                })?;
                let node_id = NodeId::new(*namespace, node_identifier.clone());
                let result_mask = (BrowseDescriptionResultMask::RESULT_MASK_DISPLAY_NAME
                    | BrowseDescriptionResultMask::RESULT_MASK_REFERENCE_TYPE)
                    .bits();
                let browse_description = BrowseDescription {
                    node_id: node_id.clone(),
                    browse_direction: BrowseDirection::Forward,
                    reference_type_id: ReferenceTypeId::HierarchicalReferences.into(),
                    include_subtypes: true,
                    node_class_mask: NodeClassMask::VARIABLE.bits(),
                    result_mask,
                };
                let browse_result = {
                    let session = session.try_read_for(SESSION_LOCK_TIMEOUT).ok_or_else(|| {
                        error!(kind = "session lock timeout");
                    })?;
                    session
                        .browse(&[browse_description])
                        .map_err(|err| {
                            error!(kind="Browse request", %err);
                        })?
                        .unwrap()
                        .pop()
                        .ok_or_else(|| {
                            error!(kind = "empty Browse results");
                        })?
                };
                if !browse_result.status_code.is_good() {
                    let status_code = browse_result.status_code;
                    error!(kind = "BrowseResult", %status_code);
                    return Err(());
                }
                if !browse_result.continuation_point.is_null() {
                    error!(kind = "unimplemented ContinuationPoint");
                    return Err(());
                }
                let references = browse_result.references.ok_or_else(|| {
                    error!(kind = "NodeId is missing forward reference", %node_id);
                })?;
                for ReferenceDescription {
                    node_id,
                    display_name,
                    ..
                } in references.into_iter().filter(|ref_description| {
                    use ReferenceTypeId::*;
                    matches!(
                        ref_description.reference_type_id.as_reference_type_id(),
                        Ok(HasComponent | Organizes)
                    )
                }) {
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
                let namespace = namespaces.get(namespace_uri).ok_or_else(|| {
                    error!(kind = "namespace not found", namespace_uri);
                })?;
                let node_id = NodeId::new(*namespace, node_identifier.clone());
                tag_set.push(Tag {
                    name: name.clone(),
                    node_id,
                })
            }
        }
    }

    Ok(tag_set)
}

#[cfg(test)]
mod tests {
    use super::*;

    mod tag_set_from_config_groups {
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
            let config = &[TagsConfigGroup::Container {
                namespace_uri: "nonexistent".to_string(),
                node_identifier: NodeIdentifier::Numeric(0),
            }];
            let namespaces = HashMap::new();
            let session = Arc::new(RwLock::new(ViewServiceMock {
                browse_outcome: Ok(Some(vec![])),
            }));

            let result = tag_set_from_config_groups(config, &namespaces, session);

            assert!(result.is_err());
        }

        #[test]
        fn browse_error() {
            let config = &[TagsConfigGroup::Container {
                namespace_uri: "urn:ns".to_string(),
                node_identifier: NodeIdentifier::Numeric(0),
            }];
            let namespaces = HashMap::from([("urn:ns".to_string(), 1)]);
            let session = Arc::new(RwLock::new(ViewServiceMock {
                browse_outcome: Err(StatusCode::BadInternalError),
            }));

            let result = tag_set_from_config_groups(config, &namespaces, session);

            assert!(result.is_err());
        }

        #[test]
        fn empty_browse_results() {
            let config = &[TagsConfigGroup::Container {
                namespace_uri: "urn:ns".to_string(),
                node_identifier: NodeIdentifier::Numeric(0),
            }];
            let namespaces = HashMap::from([("urn:ns".to_string(), 1)]);
            let session = Arc::new(RwLock::new(ViewServiceMock {
                browse_outcome: Ok(Some(vec![])),
            }));

            let result = tag_set_from_config_groups(config, &namespaces, session);

            assert!(result.is_err());
        }

        #[test]
        fn continuation_point() {
            let config = &[TagsConfigGroup::Container {
                namespace_uri: "urn:ns".to_string(),
                node_identifier: NodeIdentifier::Numeric(0),
            }];
            let namespaces = HashMap::from([("urn:ns".to_string(), 1)]);
            let session = Arc::new(RwLock::new(ViewServiceMock {
                browse_outcome: Ok(Some(vec![BrowseResult {
                    status_code: StatusCode::Good,
                    continuation_point: "deadbeef".into(),
                    references: Some(vec![]),
                }])),
            }));

            let result = tag_set_from_config_groups(config, &namespaces, session);

            assert!(result.is_err());
        }

        #[test]
        fn tag_namespace_not_found() {
            let config = &[TagsConfigGroup::Tag {
                name: "somename".to_string(),
                namespace_uri: "nonexistent".to_string(),
                node_identifier: NodeIdentifier::Numeric(0),
            }];
            let namespaces = HashMap::new();
            let session = Arc::new(RwLock::new(ViewServiceMock {
                browse_outcome: Ok(Some(vec![])),
            }));

            let result = tag_set_from_config_groups(config, &namespaces, session);

            assert!(result.is_err());
        }

        #[test]
        fn success() {
            let config = &[
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
            let references = [
                (VariableId::LocalTime, ReferenceTypeId::HasComponent),
                (VariableId::Server_ServiceLevel, ReferenceTypeId::Organizes),
                (
                    VariableId::Server_ServerStatus,
                    ReferenceTypeId::HasEventSource,
                ),
            ]
            .iter()
            .map(|(variable_id, reference_type)| ReferenceDescription {
                reference_type_id: reference_type.into(),
                is_forward: true,
                node_id: NodeId::from(variable_id).into(),
                browse_name: QualifiedName::null(),
                display_name: format!("{variable_id:?}").into(),
                node_class: NodeClass::Unspecified,
                type_definition: ExpandedNodeId::null(),
            })
            .collect();
            let session = Arc::new(RwLock::new(ViewServiceMock {
                browse_outcome: Ok(Some(vec![BrowseResult {
                    status_code: StatusCode::Good,
                    continuation_point: ByteString::null(),
                    references: Some(references),
                }])),
            }));

            let result = tag_set_from_config_groups(config, &namespaces, session)
                .expect("result should not be an error");

            assert_eq!(result.len(), 3);
            assert_eq!(result[0].name, "LocalTime");
            assert_eq!(result[0].node_id, VariableId::LocalTime.into());
            assert_eq!(result[1].name, "Server_ServiceLevel");
            assert_eq!(result[1].node_id, VariableId::Server_ServiceLevel.into());
            assert_eq!(result[2].name, "somename");
            assert_eq!(
                result[2].node_id,
                NodeId::new(2, NodeIdentifier::String("some_node_id".to_string()))
            )
        }
    }
}
