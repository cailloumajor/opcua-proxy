use crate::variant::Variant;

#[derive(Debug)]
pub struct TagChange {
    pub(crate) tag_name: String,
    pub(crate) value: Variant,
    pub(crate) source_timestamp: i64,
}
