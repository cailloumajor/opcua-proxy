use std::path::PathBuf;
use std::time::Duration;

use anyhow::anyhow;
use opcua::client::{Client, ClientBuilder};

/// The session timeout to request on the OPC-UA server.
const SESSION_TIMEOUT: Duration = Duration::from_mins(20);

/// Create an OPC-UA [`Client`].
pub(crate) fn create_client<P>(pki_dir: P) -> anyhow::Result<Client>
where
    P: Into<PathBuf>,
{
    let session_timeout_millis = u32::try_from(SESSION_TIMEOUT.as_millis())
        .expect("session timeout milliseconds should fit in an u32");

    ClientBuilder::new()
        .application_name(env!("CARGO_PKG_DESCRIPTION"))
        .product_uri(concat!("urn:", env!("CARGO_PKG_NAME")))
        .application_uri(concat!("urn:", env!("CARGO_PKG_NAME")))
        .pki_dir(pki_dir)
        .certificate_path(concat!("own/", env!("CARGO_PKG_NAME"), "-cert.der"))
        .private_key_path(concat!("private/", env!("CARGO_PKG_NAME"), "-key.pem"))
        .session_retry_limit(-1)
        .session_timeout(session_timeout_millis)
        .client()
        .map_err(|_| anyhow!("See logs for client builder errors"))
}
