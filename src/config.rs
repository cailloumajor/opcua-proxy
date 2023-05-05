use anyhow::{anyhow, Context as _};
use awc::Client;
use tracing::{debug, info, instrument};
use url::Url;

use crate::model::ConfigFromApi;

#[instrument(skip_all)]
pub(crate) async fn fetch_config(api_url: &Url, partner_id: &str) -> anyhow::Result<ConfigFromApi> {
    let config_url = api_url
        .join(partner_id)
        .context("error joining config API URL and partner ID")?;
    let client = Client::default();
    let mut response = client
        .get(config_url.as_str())
        .send()
        .await
        .map_err(|err| anyhow!(err.to_string()))
        .context("tags configuration request error")?;
    let response_status = response.status();
    if !response_status.is_success() {
        return Err(anyhow!("bad response status: {}", response_status));
    }
    let tags_config = response
        .json()
        .await
        .context("tags configuration deserialization error")?;

    debug!(?tags_config);

    info!(status = "success");
    Ok(tags_config)
}
