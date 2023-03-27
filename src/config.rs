use anyhow::{anyhow, Context as _};
use tracing::{debug, info, instrument};
use url::Url;

use crate::model::ConfigFromApi;

type Conn = trillium_client::Conn<'static, trillium_tokio::TcpConnector>;

#[instrument(skip_all)]
pub(crate) async fn fetch_config(api_url: &Url, partner_id: &str) -> anyhow::Result<ConfigFromApi> {
    let config_url = api_url
        .join(partner_id)
        .context("error joining config API URL and partner ID")?;
    let mut response = Conn::get(config_url)
        .await
        .context("tags configuration request error")?;
    let response_status = response.status().expect("missing response status code");
    if !response_status.is_success() {
        return Err(anyhow!("bad response status: {}", response_status));
    }
    let tags_config = response
        .response_json()
        .await
        .context("tags configuration deserialization error")?;

    debug!(?tags_config);

    info!(status = "success");
    Ok(tags_config)
}
