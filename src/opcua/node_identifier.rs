use opcua::types::Identifier;
use serde::Deserialize;

#[derive(Debug, Clone, Deserialize)]
#[serde(untagged)]
pub(crate) enum NodeIdentifier {
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
