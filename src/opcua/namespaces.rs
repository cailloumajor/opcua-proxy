use std::collections::HashMap;
use std::sync::Arc;

use opcua::client::prelude::*;
use opcua::sync::RwLock;
use tracing::{error, instrument};

use super::session::SESSION_LOCK_TIMEOUT;

pub(super) type Namespaces = HashMap<String, u16>;

#[instrument(skip_all)]
pub(crate) fn get_namespaces<T>(session: Arc<RwLock<T>>) -> Result<Namespaces, ()>
where
    T: AttributeService,
{
    let namespace_array_nodeid: NodeId = VariableId::Server_NamespaceArray.into();
    let read_results = {
        let session = session.try_read_for(SESSION_LOCK_TIMEOUT).ok_or_else(|| {
            error!(kind = "session lock timeout");
        })?;
        session
            .read(
                &[namespace_array_nodeid.into()],
                TimestampsToReturn::Neither,
                0.0,
            )
            .map_err(|err| {
                error!(kind = "reading nodes", %err);
            })?
    };
    let data_value = read_results.first().ok_or_else(|| {
        error!(kind = "missing namespace array");
    })?;
    let result_variant = data_value.value.as_ref().ok_or_else(|| {
        let description = data_value.status().description();
        error!(kind = "value error", description);
    })?;
    let Variant::Array(namespace_array) = result_variant else {
        let type_id = result_variant.type_id();
        error!(kind = "bad value type", expected = "array", ?type_id);
        return Err(());
    };
    let namespaces = namespace_array
        .values
        .iter()
        .zip(0..)
        .map(|(variant, namespace_index)| {
            if let Variant::String(uastring) = variant {
                Ok((uastring.to_string(), namespace_index))
            } else {
                let type_id = result_variant.type_id();
                error!(kind = "bad member type", expected = "string", ?type_id);
                Err(())
            }
        })
        .collect::<Result<Vec<_>, _>>()?
        .into_iter()
        .collect();

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
                    TestCases::Success => Ok(vec![
                        Variant::from(vec!["urn:ns:ns1".to_string(), "urn:ns:ns2".to_string()])
                            .into(),
                    ]),
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
            let mock = Arc::new(RwLock::new(AttributeServiceMock(TestCases::ReadError)));
            let result = get_namespaces(mock);
            assert!(result.is_err());
        }

        #[test]
        fn missing_data_value() {
            let mock = Arc::new(RwLock::new(AttributeServiceMock(TestCases::EmptyResult)));
            let result = get_namespaces(mock);
            assert!(result.is_err());
        }

        #[test]
        fn missing_value() {
            let mock = Arc::new(RwLock::new(AttributeServiceMock(TestCases::NoValue)));
            let result = get_namespaces(mock);
            assert!(result.is_err());
        }

        #[test]
        fn not_an_array() {
            let mock = Arc::new(RwLock::new(AttributeServiceMock(TestCases::NotAnArray)));
            let result = get_namespaces(mock);
            assert!(result.is_err());
        }

        #[test]
        fn bad_member_type() {
            let mock = Arc::new(RwLock::new(AttributeServiceMock(TestCases::BadMemberType)));
            let result = get_namespaces(mock);
            assert!(result.is_err());
        }

        #[test]
        fn success() {
            let mock = Arc::new(RwLock::new(AttributeServiceMock(TestCases::Success)));
            let result = get_namespaces(mock);
            let expected = HashMap::from([
                ("urn:ns:ns1".to_string(), 0u16),
                ("urn:ns:ns2".to_string(), 1u16),
            ]);
            assert_eq!(result.unwrap(), expected);
        }
    }
}
