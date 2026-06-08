use std::io::{self, Write};

use opcua::types::json::{JsonEncodable, JsonStreamWriter, JsonWriter};
use opcua::types::{Context, DateTime, Variant};

type TagChange<'a> = (&'a String, &'a Variant, &'a DateTime);

/// Encode the provided tag changes slice to JSON bytes.
///
/// The JSON output is an object with two properties:
///
/// * the `val` object contains one property with tag name as key and new value as value,
///   for each provided tag change;
/// * the `ts` object contains one property with tag name as key and change timestamp as value,
///   for each provided tag change.
pub(super) fn encode_tag_changes(changes: &[TagChange], ctx: &Context) -> io::Result<Vec<u8>> {
    let mut out = Vec::new();

    let mut stream = JsonStreamWriter::new(&mut out as &mut dyn Write);

    stream.begin_object()?; // Outer object open.
    stream.name("val")?;
    stream.begin_object()?; // `val` object open.
    for (tag_name, value, _) in changes.iter() {
        stream.name(tag_name)?;
        value.serialize_variant_value(&mut stream, ctx)?;
    }
    stream.end_object()?; // `val` object close.
    stream.name("ts")?;
    stream.begin_object()?; // `ts` object open.
    for (tag_name, _, timestamp) in changes.iter() {
        stream.name(tag_name)?;
        timestamp.encode(&mut stream, ctx)?;
    }
    stream.end_object()?; // `ts` object close.
    stream.end_object()?; // Outer object close.

    stream.finish_document()?;

    Ok(out)
}

#[cfg(test)]
mod tests {
    use super::*;

    mod encode_tag_changes {
        use opcua::types::ContextOwned;

        use super::*;

        #[test]
        fn one_element() {
            let ctx = ContextOwned::default();

            let tag_changes = &[(
                &"some_tag".into(),
                &Variant::Float(37.5),
                &DateTime::epoch(),
            )];

            let json = encode_tag_changes(tag_changes, &ctx.context())
                .expect("encoding tag change should not fail");

            assert_eq!(
                json,
                br#"{"val":{"some_tag":37.5},"ts":{"some_tag":"1601-01-01T00:00:00.000Z"}}"#
            );
        }

        #[test]
        fn several_elements() {
            let ctx = ContextOwned::default();

            let tag_changes = &[
                (
                    &"first".into(),
                    &Variant::Boolean(false),
                    &DateTime::ymd(1984, 12, 9),
                ),
                (
                    &"second".into(),
                    &Variant::DateTime(DateTime::epoch().into()),
                    &DateTime::ymd_hms(1970, 1, 1, 13, 54, 23),
                ),
                (&"third".into(), &Variant::Empty, &DateTime::epoch()),
            ];

            let json = encode_tag_changes(tag_changes, &ctx.context())
                .expect("encoding tag change should not fail");

            assert_eq!(
                json,
                br#"{"val":{"first":false,"second":"1601-01-01T00:00:00.000Z","third":null},"ts":{"first":"1984-12-09T00:00:00.000Z","second":"1970-01-01T13:54:23.000Z","third":"1601-01-01T00:00:00.000Z"}}"#
            );
        }
    }
}
