use mongodb::bson::spec::BinarySubtype;
use mongodb::bson::{self, Bson};
use opcua::types::Variant as OpcUaVariant;
use tracing::{instrument, warn};

/// Wraps [opcua::types::Variant] to provide custom, seamless serializing.
#[derive(Debug, Clone)]
pub(crate) struct Variant(OpcUaVariant);

impl From<OpcUaVariant> for Variant {
    fn from(v: OpcUaVariant) -> Self {
        Self(v)
    }
}

impl From<Variant> for Bson {
    #[instrument(name = "variant_to_bson")]
    fn from(value: Variant) -> Self {
        match value.0 {
            OpcUaVariant::Empty => Bson::Null,
            OpcUaVariant::Boolean(v) => Bson::Boolean(v),
            OpcUaVariant::SByte(v) => Bson::Int32(v.into()),
            OpcUaVariant::Byte(v) => Bson::Int32(v.into()),
            OpcUaVariant::Int16(v) => Bson::Int32(v.into()),
            OpcUaVariant::UInt16(v) => Bson::Int32(v.into()),
            OpcUaVariant::Int32(v) => Bson::Int32(v),
            OpcUaVariant::UInt32(v) => Bson::Int64(v.into()),
            OpcUaVariant::Int64(v) => Bson::Int64(v),
            OpcUaVariant::UInt64(v) => match i64::try_from(v) {
                Ok(val) => Bson::Int64(val),
                Err(err) => {
                    warn!(kind = "conversion", %err);
                    Bson::Null
                }
            },
            OpcUaVariant::Float(v) => Bson::Double(v.into()),
            OpcUaVariant::Double(v) => Bson::Double(v),
            OpcUaVariant::String(v) => Bson::String(v.into()),
            OpcUaVariant::LocalizedText(v) => Bson::String(v.text.into()),
            OpcUaVariant::DateTime(v) => Bson::String(format!("{:?}", v.as_chrono())),
            OpcUaVariant::Guid(v) => Bson::String(v.to_string()),
            OpcUaVariant::StatusCode(v) => Bson::Int64(v.bits().into()),
            OpcUaVariant::ByteString(v) => Bson::Binary(bson::Binary {
                subtype: BinarySubtype::Generic,
                bytes: v.value.unwrap_or_default(),
            }),
            OpcUaVariant::Array(v) => {
                if v.dimensions.is_some_and(|d| !d.is_empty()) {
                    warn!(kind = "unimplemented serialization for multi-dimensional arrays");
                    Bson::Null
                } else {
                    Bson::Array(
                        v.values
                            .into_iter()
                            .map(|val| Variant(val).into())
                            .collect(),
                    )
                }
            }
            _ => {
                let type_id = value.0.type_id();
                warn!(kind = "unimplemented serialization", ?type_id);
                Bson::Null
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    mod variant_into_bson {
        use opcua::types::{
            Array, ByteString, DateTime, DiagnosticInfo, Guid, LocalizedText, StatusCode, UAString,
            VariantTypeId,
        };

        use super::*;

        #[test]
        fn empty() {
            let variant = Variant(OpcUaVariant::Empty);
            let bson = Bson::from(variant);
            assert_eq!(bson, Bson::Null);
        }

        #[test]
        fn boolean() {
            let variant = Variant(OpcUaVariant::Boolean(true));
            let bson = Bson::from(variant);
            assert_eq!(bson, Bson::Boolean(true));
        }

        #[test]
        fn sbyte() {
            let variant = Variant(OpcUaVariant::SByte(i8::MIN));
            let bson = Bson::from(variant);
            assert_eq!(bson, Bson::Int32(-128));
        }

        #[test]
        fn byte() {
            let variant = Variant(OpcUaVariant::Byte(u8::MAX));
            let bson = Bson::from(variant);
            assert_eq!(bson, Bson::Int32(255));
        }

        #[test]
        fn int16() {
            let variant = Variant(OpcUaVariant::Int16(i16::MIN));
            let bson = Bson::from(variant);
            assert_eq!(bson, Bson::Int32(-32768));
        }

        #[test]
        fn uint16() {
            let variant = Variant(OpcUaVariant::UInt16(u16::MAX));
            let bson = Bson::from(variant);
            assert_eq!(bson, Bson::Int32(65535));
        }

        #[test]
        fn int32() {
            let variant = Variant(OpcUaVariant::Int32(i32::MIN));
            let bson = Bson::from(variant);
            assert_eq!(bson, Bson::Int32(-2147483648));
        }

        #[test]
        fn uint32() {
            let variant = Variant(OpcUaVariant::UInt32(u32::MAX));
            let bson = Bson::from(variant);
            assert_eq!(bson, Bson::Int64(4294967295));
        }

        #[test]
        fn int64() {
            let variant = Variant(OpcUaVariant::Int64(i64::MIN));
            let bson = Bson::from(variant);
            assert_eq!(bson, Bson::Int64(-9223372036854775808));
        }

        #[test]
        fn uint64() {
            let variant = Variant(OpcUaVariant::UInt64(u32::MAX.into()));
            let bson = Bson::from(variant);
            assert_eq!(bson, Bson::Int64(4294967295));
        }

        #[test]
        fn uint64_overflowing() {
            let variant = Variant(OpcUaVariant::UInt64(u64::MAX));
            let bson = Bson::from(variant);
            assert_eq!(bson, Bson::Null);
        }

        #[test]
        fn float() {
            let variant = Variant(OpcUaVariant::Float(42.0));
            let bson = Bson::from(variant);
            assert_eq!(bson, Bson::Double(42.0));
        }

        #[test]
        fn double() {
            let variant = Variant(OpcUaVariant::Double(456.54));
            let bson = Bson::from(variant);
            assert_eq!(bson, Bson::Double(456.54));
        }

        #[test]
        fn null_string() {
            let variant = Variant(OpcUaVariant::String(UAString::null()));
            let bson = Bson::from(variant);
            assert_eq!(bson, Bson::String("".to_string()));
        }

        #[test]
        fn string() {
            let variant = Variant(OpcUaVariant::String("test string".into()));
            let bson = Bson::from(variant);
            assert_eq!(bson, Bson::String("test string".to_string()));
        }

        #[test]
        fn null_localized_text() {
            let variant = Variant(OpcUaVariant::from(LocalizedText::null()));
            let bson = Bson::from(variant);
            assert_eq!(bson, Bson::String("".to_string()));
        }

        #[test]
        fn localized_text() {
            let variant = Variant(OpcUaVariant::from(LocalizedText::new(
                "somelocale",
                "some text",
            )));
            let bson = Bson::from(variant);
            assert_eq!(bson, Bson::String("some text".to_string()));
        }

        #[test]
        fn datetime() {
            let variant = Variant(OpcUaVariant::DateTime(Box::new(DateTime::epoch())));
            let bson = Bson::from(variant);
            assert_eq!(bson, Bson::String("1601-01-01T00:00:00Z".to_string()));
        }

        #[test]
        fn guid() {
            let variant = Variant(OpcUaVariant::Guid(Box::new(Guid::null())));
            let bson = Bson::from(variant);
            assert_eq!(
                bson,
                Bson::String("00000000-0000-0000-0000-000000000000".to_string())
            );
        }

        #[test]
        fn statuscode() {
            let variant = Variant(OpcUaVariant::StatusCode(StatusCode::BadUnexpectedError));
            let bson = Bson::from(variant);
            assert_eq!(bson, Bson::Int64(2147549184));
        }

        #[test]
        fn null_bytestring() {
            let variant = Variant(OpcUaVariant::ByteString(ByteString::null()));
            let bson = Bson::from(variant);
            assert_eq!(
                bson,
                Bson::Binary(bson::Binary {
                    subtype: BinarySubtype::Generic,
                    bytes: vec![]
                })
            );
        }

        #[test]
        fn bytestring() {
            let variant = Variant(OpcUaVariant::ByteString((&[1, 2, 3, 4]).into()));
            let bson = Bson::from(variant);
            assert_eq!(
                bson,
                Bson::Binary(bson::Binary {
                    subtype: BinarySubtype::Generic,
                    bytes: vec![1, 2, 3, 4]
                })
            );
        }

        #[test]
        fn one_dimension_array() {
            let variant = Variant(OpcUaVariant::Array(Box::new(
                Array::new(
                    VariantTypeId::Byte,
                    (1u8..=4u8).map(OpcUaVariant::from).collect::<Vec<_>>(),
                )
                .unwrap(),
            )));
            let bson = Bson::from(variant);
            assert_eq!(
                bson,
                Bson::Array(vec![
                    Bson::Int32(1),
                    Bson::Int32(2),
                    Bson::Int32(3),
                    Bson::Int32(4),
                ])
            );
        }

        #[test]
        fn multi_dimension_array() {
            let variant = Variant(OpcUaVariant::Array(Box::new(
                Array::new_multi(
                    VariantTypeId::Byte,
                    (1u8..=4u8).map(OpcUaVariant::from).collect::<Vec<_>>(),
                    [2, 2],
                )
                .unwrap(),
            )));
            let bson = Bson::from(variant);
            assert_eq!(bson, Bson::Null);
        }

        #[test]
        fn unimplemented() {
            let variant = Variant(OpcUaVariant::from(DiagnosticInfo::null()));
            let bson = Bson::from(variant);
            assert_eq!(bson, Bson::Null);
        }
    }
}
