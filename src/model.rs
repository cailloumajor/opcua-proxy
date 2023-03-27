use serde::Deserialize;

use crate::variant::Variant;

#[derive(Debug, Clone, Deserialize)]
#[serde(untagged)]
pub(crate) enum NodeIdentifier {
    Numeric(u32),
    String(String),
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub(crate) struct ConfigFromApi {
    pub(crate) tags: Vec<TagsConfigGroup>,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub(crate) enum TagsConfigGroup {
    #[serde(rename_all = "camelCase")]
    Container {
        namespace_uri: String,
        node_identifier: NodeIdentifier,
    },
}

#[derive(Debug)]
pub(crate) struct TagChange {
    pub(crate) tag_name: String,
    pub(crate) value: Variant,
    pub(crate) source_timestamp: i64,
}
