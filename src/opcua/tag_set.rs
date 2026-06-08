use std::pin::pin;

use futures_util::TryStreamExt;
use opcua::client::Session;
use opcua::client::browser::BrowseFilter;
use opcua::types::{
    BrowseResultMaskFlags, Error as OpcUaError, NodeClassMask, NodeId, ReferenceTypeId,
};

use crate::opcua::config::TagsConfigGroup;

#[derive(Debug, Clone)]
pub(super) struct Tag {
    pub(super) name: String,
    pub(super) node_id: NodeId,
}

pub(super) async fn tag_set_from_config_groups(
    session: &Session,
    config: &[TagsConfigGroup],
) -> Result<Vec<Tag>, OpcUaError> {
    let mut tag_set: Vec<Tag> = Vec::new();

    for config_group in config {
        match config_group {
            TagsConfigGroup::Container {
                namespace_uri,
                node_identifier,
            } => {
                let browse_filter = BrowseFilter::new_hierarchical()
                    .max_depth(1)
                    .node_class_mask(NodeClassMask::VARIABLE)
                    .result_mask(
                        BrowseResultMaskFlags::DisplayName | BrowseResultMaskFlags::ReferenceTypeId,
                    );
                let namespace = session.get_namespace_index(namespace_uri).await?;
                let node_id = NodeId::new(namespace, node_identifier.clone());
                let browse_description = browse_filter.new_description_from_node(node_id);
                let browser = session.browser().handler(browse_filter);

                let mut browser_stream = pin!(browser.run(vec![browse_description]));

                while let Some(browse_result) = browser_stream.try_next().await? {
                    let (_, ref_descriptions) = browse_result.into_results();
                    for ref_description in ref_descriptions.into_iter().filter(|r| {
                        matches!(
                            r.reference_type_id.as_reference_type_id(),
                            Ok(ReferenceTypeId::HasComponent | ReferenceTypeId::Organizes)
                        )
                    }) {
                        tag_set.push(Tag {
                            name: ref_description.display_name.to_string(),
                            node_id: ref_description.node_id.node_id,
                        });
                    }
                }
            }

            TagsConfigGroup::Tag {
                name,
                namespace_uri,
                node_identifier,
            } => {
                let namespace = session.get_namespace_index(namespace_uri).await?;
                let node_id = NodeId::new(namespace, node_identifier.clone());

                tag_set.push(Tag {
                    name: name.clone(),
                    node_id,
                })
            }
        }
    }

    Ok(tag_set)
}
