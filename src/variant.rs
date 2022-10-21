use std::fmt;

use opcua::types::{Array, Variant as OpcUaVariant};
use serde::ser::{self, Serialize, Serializer};

struct Bytes(Vec<u8>);

impl Serialize for Bytes {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: Serializer,
    {
        serializer.serialize_bytes(&self.0)
    }
}

/// Wraps [opcua::types::Variant] to provide custom, seamless serializing.
#[derive(Clone)]
pub(crate) struct Variant(OpcUaVariant);

impl fmt::Display for Variant {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        self.0.fmt(f)
    }
}

impl From<OpcUaVariant> for Variant {
    fn from(v: OpcUaVariant) -> Self {
        Self(v)
    }
}

impl Serialize for Variant {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: Serializer,
    {
        match self.0 {
            OpcUaVariant::Empty => serializer.serialize_unit(),
            OpcUaVariant::Boolean(v) => serializer.serialize_bool(v),
            OpcUaVariant::SByte(v) => serializer.serialize_i8(v),
            OpcUaVariant::Byte(v) => serializer.serialize_u8(v),
            OpcUaVariant::Int16(v) => serializer.serialize_i16(v),
            OpcUaVariant::UInt16(v) => serializer.serialize_u16(v),
            OpcUaVariant::Int32(v) => serializer.serialize_i32(v),
            OpcUaVariant::UInt32(v) => serializer.serialize_u32(v),
            OpcUaVariant::Int64(v) => serializer.serialize_i64(v),
            OpcUaVariant::UInt64(v) => serializer.serialize_u64(v),
            OpcUaVariant::Float(v) => serializer.serialize_f32(v),
            OpcUaVariant::Double(v) => serializer.serialize_f64(v),
            OpcUaVariant::String(ref v) => v.value().serialize(serializer),
            OpcUaVariant::DateTime(ref v) => v.as_chrono().serialize(serializer),
            OpcUaVariant::Guid(ref v) => v.serialize(serializer),
            OpcUaVariant::StatusCode(v) => v.serialize(serializer),
            OpcUaVariant::ByteString(ref v) => v
                .value
                .as_ref()
                .map(|v| Bytes(v.to_owned()))
                .serialize(serializer),
            OpcUaVariant::Array(ref v) => serialize_array(v, serializer),
            _ => Err(ser::Error::custom(format!(
                "serialization unimplemented for {:?}",
                &self.0.type_id()
            ))),
        }
    }
}

fn serialize_array<S>(array: &Array, serializer: S) -> Result<S::Ok, S::Error>
where
    S: Serializer,
{
    if array.has_dimensions() {
        Err(ser::Error::custom(
            "serialization unimplemented for multi-dimensional arrays",
        ))
    } else {
        array
            .values
            .iter()
            .cloned()
            .map(Variant::from)
            .collect::<Vec<_>>()
            .serialize(serializer)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    mod serialize {
        use opcua::types::{
            Array, ByteString, DateTime, Guid, StatusCode, UAString, VariantTypeId,
        };
        use serde_test::{assert_ser_tokens, Token};

        use super::*;

        #[test]
        fn empty() {
            let s = Variant::from(OpcUaVariant::Empty);
            assert_ser_tokens(&s, &[Token::Unit]);
        }

        #[test]
        fn boolean() {
            let s = Variant::from(OpcUaVariant::Boolean(true));
            assert_ser_tokens(&s, &[Token::Bool(true)]);
        }

        #[test]
        fn sbyte() {
            let s = Variant::from(OpcUaVariant::SByte(i8::MIN));
            assert_ser_tokens(&s, &[Token::I8(i8::MIN)]);
        }

        #[test]
        fn byte() {
            let s = Variant::from(OpcUaVariant::Byte(u8::MAX));
            assert_ser_tokens(&s, &[Token::U8(u8::MAX)]);
        }

        #[test]
        fn int16() {
            let s = Variant::from(OpcUaVariant::Int16(i16::MIN));
            assert_ser_tokens(&s, &[Token::I16(i16::MIN)]);
        }

        #[test]
        fn uint16() {
            let s = Variant::from(OpcUaVariant::UInt16(u16::MAX));
            assert_ser_tokens(&s, &[Token::U16(u16::MAX)]);
        }

        #[test]
        fn int32() {
            let s = Variant::from(OpcUaVariant::Int32(i32::MIN));
            assert_ser_tokens(&s, &[Token::I32(i32::MIN)]);
        }

        #[test]
        fn uint32() {
            let s = Variant::from(OpcUaVariant::UInt32(u32::MAX));
            assert_ser_tokens(&s, &[Token::U32(u32::MAX)]);
        }

        #[test]
        fn int64() {
            let s = Variant::from(OpcUaVariant::Int64(i64::MIN));
            assert_ser_tokens(&s, &[Token::I64(i64::MIN)]);
        }

        #[test]
        fn uint64() {
            let s = Variant::from(OpcUaVariant::UInt64(u64::MAX));
            assert_ser_tokens(&s, &[Token::U64(u64::MAX)]);
        }

        #[test]
        fn float() {
            let s = Variant::from(OpcUaVariant::Float(42.0));
            assert_ser_tokens(&s, &[Token::F32(42.0)]);
        }

        #[test]
        fn double() {
            let s = Variant::from(OpcUaVariant::Double(456.54));
            assert_ser_tokens(&s, &[Token::F64(456.54)]);
        }

        #[test]
        fn null_string() {
            let s = Variant::from(OpcUaVariant::String(UAString::null()));
            assert_ser_tokens(&s, &[Token::None]);
        }

        #[test]
        fn string() {
            let s = Variant::from(OpcUaVariant::String("test string".into()));
            assert_ser_tokens(&s, &[Token::Some, Token::String("test string")]);
        }

        #[test]
        fn datetime() {
            let s = Variant::from(OpcUaVariant::DateTime(Box::new(DateTime::epoch())));
            assert_ser_tokens(&s, &[Token::Str("1601-01-01T00:00:00Z")]);
        }

        #[test]
        fn guid() {
            let s = Variant::from(OpcUaVariant::Guid(Box::new(Guid::null())));
            assert_ser_tokens(&s, &[Token::Str("00000000-0000-0000-0000-000000000000")]);
        }

        #[test]
        fn statuscode() {
            let s = Variant::from(OpcUaVariant::StatusCode(StatusCode::BadUnexpectedError));
            assert_ser_tokens(&s, &[Token::U32(0x8001_0000)]);
        }

        #[test]
        fn null_bytestring() {
            let s = Variant::from(OpcUaVariant::ByteString(ByteString::null()));
            assert_ser_tokens(&s, &[Token::None]);
        }

        #[test]
        fn bytestring() {
            let s = Variant::from(OpcUaVariant::ByteString((&[1, 2, 3, 4]).into()));
            assert_ser_tokens(&s, &[Token::Some, Token::Bytes(&[1, 2, 3, 4])]);
        }

        #[test]
        fn one_dimension_array() {
            let s = Variant::from(OpcUaVariant::Array(Box::new(
                Array::new_single(
                    VariantTypeId::Byte,
                    (1u8..=4u8).map(OpcUaVariant::from).collect::<Vec<_>>(),
                )
                .unwrap(),
            )));
            assert_ser_tokens(
                &s,
                &[
                    Token::Seq { len: Some(4) },
                    Token::U8(1),
                    Token::U8(2),
                    Token::U8(3),
                    Token::U8(4),
                    Token::SeqEnd,
                ],
            );
        }
    }
}
