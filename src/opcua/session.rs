use std::sync::Arc;
use std::time::Duration;

use opcua::client::prelude::*;
use opcua::sync::RwLock;
use serde::Deserialize;
use tracing::{debug, error, info, info_span, instrument};
use url::Url;

use super::tag_set::TagsConfigGroup;

pub(super) const SESSION_LOCK_TIMEOUT: Duration = Duration::from_millis(50);

#[derive(Debug, Deserialize, Eq, Hash, PartialEq)]
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

#[instrument(skip_all)]
pub(super) async fn fetch_partners_config(
    api_url: Url,
) -> Result<Vec<PartnerConfig>, reqwest::Error> {
    let config = reqwest::get(api_url)
        .await?
        .error_for_status()?
        .json()
        .await?;

    debug!(?config);

    Ok(config)
}

#[instrument(skip_all)]
pub(super) fn create_session(
    client: &mut Client,
    config: &PartnerConfig,
) -> Result<Arc<RwLock<Session>>, ()> {
    let user_identity_token = if let (Some(user), Some(pass)) = (&config.user, &config.password) {
        IdentityToken::UserName(user.clone(), pass.clone())
    } else {
        IdentityToken::Anonymous
    };

    let endpoint: EndpointDescription = (
        config.server_url.as_str(),
        config.security_policy.as_str(),
        MessageSecurityMode::from(config.security_mode.as_str()),
    )
        .into();

    let session = client
        .connect_to_endpoint(endpoint, user_identity_token)
        .map_err(|err| {
            error!(kind = "endpoint connection", %err);
        })?;

    {
        let mut session = session.try_write_for(SESSION_LOCK_TIMEOUT).ok_or_else(|| {
            error!(kind = "session lock timeout");
        })?;
        let partner_arc: Arc<str> = Arc::from(config.partner_id.as_str());
        let partner_id = partner_arc.clone();
        session.set_connection_status_callback(ConnectionStatusCallback::new(move |connected| {
            let _entered = info_span!("connection status callback").entered();
            info!(msg = "connection status changed", connected, %partner_id);
        }));
        let partner_id = partner_arc;
        session.set_session_closed_callback(SessionClosedCallback::new(move |status_code| {
            let _entered = info_span!("session closed callback").entered();
            let partner_id = partner_id.clone();
            info!(msg = "session has been closed", %status_code, %partner_id);
        }))
    }

    Ok(session)
}
