use std::collections::HashMap;

use anyhow::anyhow;
use opcua::client::prelude::*;
use tracing::{info, instrument};

pub(super) type Namespaces = HashMap<String, u16>;

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
