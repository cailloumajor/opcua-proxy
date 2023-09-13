use clap::Args;
pub(crate) use opcua::client::prelude::{Session, SessionCommand};

mod namespaces;
mod node_identifier;
mod session;
mod subscription;
mod tag_set;
mod variant;

pub(crate) use namespaces::get_namespaces;
pub(crate) use session::create_session;
pub(crate) use subscription::{subscribe_to_health, subscribe_to_tags, TagChange};
pub(crate) use tag_set::{TagSet, TagsConfigGroup};

#[derive(Args)]
pub(crate) struct Config {
    /// Path of the PKI directory
    #[arg(env, long)]
    pki_dir: String,

    /// URL of OPC-UA server to connect to
    #[arg(env, long)]
    opcua_server_url: String,

    /// OPC-UA security policy
    #[arg(env, long, default_value = "Basic256Sha256")]
    opcua_security_policy: String,

    /// OPC-UA security mode
    #[arg(env, long, default_value = "SignAndEncrypt")]
    opcua_security_mode: String,

    /// OPC-UA authentication username (optional)
    #[arg(env, long)]
    opcua_user: Option<String>,

    /// OPC-UA authentication password (optional)
    #[arg(env, long)]
    opcua_password: Option<String>,
}
