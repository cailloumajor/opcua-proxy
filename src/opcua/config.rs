use std::time::Duration;

use serde::Deserialize;
use url::Url;

use crate::opcua::node_identifier::NodeIdentifier;

/// The maximum time allowed for HTTP requests.
const HTTP_TIMEOUT: Duration = Duration::from_secs(5);

/// Represents the configuration of an OPC-UA partner.
#[derive(Clone, Debug, Deserialize, Eq, Hash, PartialEq)]
#[serde(rename_all = "camelCase")]
pub(super) struct PartnerConfig {
    #[serde(rename = "_id")]
    pub(super) partner_id: String,
    pub(super) server_url: String,
    pub(super) security_policy: String,
    pub(super) security_mode: String,
    pub(super) user: Option<String>,
    pub(super) password: Option<String>,
    pub(super) tags: Vec<TagsConfigGroup>,
}

#[derive(Clone, Debug, Deserialize, Eq, Hash, PartialEq)]
#[serde(rename_all = "camelCase", tag = "type")]
pub(super) enum TagsConfigGroup {
    #[serde(rename_all = "camelCase")]
    Container {
        namespace_uri: String,
        node_identifier: NodeIdentifier,
    },
    #[serde(rename_all = "camelCase")]
    Tag {
        name: String,
        namespace_uri: String,
        node_identifier: NodeIdentifier,
    },
}

/// Represents a wrapped HTTP client, allowing fetching the OPC-UA partners configuration.
pub(super) struct ConfigFetchService {
    client: reqwest::Client,
    url: Url,
}

impl ConfigFetchService {
    /// Create a [`ConfigFetchService`].
    pub(super) fn new(url: Url) -> reqwest::Result<Self> {
        let client = reqwest::Client::builder().timeout(HTTP_TIMEOUT).build()?;

        Ok(Self { client, url })
    }

    /// Fetch the OPC-UA partners configuration.
    pub(super) async fn fetch(&self) -> reqwest::Result<Vec<PartnerConfig>> {
        self.client
            .get(self.url.clone())
            .send()
            .await?
            .error_for_status()?
            .json()
            .await
    }
}
