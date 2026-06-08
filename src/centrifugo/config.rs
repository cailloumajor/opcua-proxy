use std::str::FromStr;

use clap::Args;
use tonic::metadata::errors::InvalidMetadataValue;
use tonic::metadata::{Ascii, MetadataValue};
use tonic::transport::Uri;

#[derive(Args)]
pub(crate) struct CentrifugoConfig {
    /// URL of the Centrifugo server.
    #[arg(env, long)]
    pub(super) centrifugo_server_uri: Uri,

    /// The Centrifugo channel namespace to use in publish messages.
    #[arg(env, long)]
    pub(crate) centrifugo_namespace: String,

    /// The API key to use for Centrifugo gRPC API.
    #[arg(env, long)]
    pub(super) centrifugo_api_key: CentrifugoApiKey,
}

#[derive(Clone)]
pub(super) struct CentrifugoApiKey(MetadataValue<Ascii>);

impl From<CentrifugoApiKey> for MetadataValue<Ascii> {
    fn from(value: CentrifugoApiKey) -> Self {
        value.0
    }
}

impl FromStr for CentrifugoApiKey {
    type Err = InvalidMetadataValue;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        let val_str = format!("apikey {s}");
        let inner = val_str.parse()?;
        Ok(Self(inner))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn parse() {
        let api_key = "blablabla"
            .parse::<CentrifugoApiKey>()
            .expect("parsing Centrifugo API key should not fail");

        let metadata_value = api_key
            .0
            .to_str()
            .expect("metadata value to str should not fail");

        assert_eq!(metadata_value, "apikey blablabla");
    }
}
