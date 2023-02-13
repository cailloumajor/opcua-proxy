use std::collections::{hash_map, HashMap};
use std::fmt;

use mongodb::bson::DateTime;

use crate::variant::Variant;

#[derive(Debug)]
struct DataValue {
    value: Variant,
    source_timestamp: DateTime,
}

#[derive(Debug)]
pub(crate) struct DataChangeMessage(HashMap<String, DataValue>);

pub(crate) struct DataChangeMessageIter {
    inner: hash_map::IntoIter<String, DataValue>,
}

impl Iterator for DataChangeMessageIter {
    type Item = (String, Variant, DateTime);

    fn next(&mut self) -> Option<Self::Item> {
        self.inner.next().map(
            |(
                name,
                DataValue {
                    value,
                    source_timestamp,
                },
            )| (name, value, source_timestamp),
        )
    }

    fn size_hint(&self) -> (usize, Option<usize>) {
        self.inner.size_hint()
    }
}

impl IntoIterator for DataChangeMessage {
    type Item = (String, Variant, DateTime);
    type IntoIter = DataChangeMessageIter;

    fn into_iter(self) -> Self::IntoIter {
        DataChangeMessageIter {
            inner: self.0.into_iter(),
        }
    }
}

impl DataChangeMessage {
    pub fn with_capacity(cap: usize) -> Self {
        Self(HashMap::with_capacity(cap))
    }

    pub fn insert(&mut self, tag_name: String, value: Variant, source_millis: i64) {
        let source_timestamp = DateTime::from_millis(source_millis);
        let data_value = DataValue {
            value,
            source_timestamp,
        };
        self.0.insert(tag_name, data_value);
    }

    pub fn len(&self) -> usize {
        self.0.len()
    }
}

pub(crate) struct HealthMessage(DateTime);

impl HealthMessage {
    pub(crate) fn date_time(&self) -> DateTime {
        self.0
    }
}

impl fmt::Display for HealthMessage {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        self.0.fmt(f)
    }
}

impl From<i64> for HealthMessage {
    fn from(millis: i64) -> Self {
        Self(DateTime::from_millis(millis))
    }
}
