use clap::Args;

pub const DATABASE: &str = "opcua";
pub const OPCUA_DATA_COLL: &str = "data";
pub const OPCUA_HEALTH_COLL: &str = "health";
pub const OPCUA_HEALTH_INTERVAL: u16 = 5000;

#[derive(Args)]
pub struct CommonArgs {
    /// OPC-UA partner device ID
    #[arg(env, long)]
    pub partner_id: String,

    /// URL of MongoDB database
    #[arg(env, long, default_value = "mongodb://mongo")]
    pub mongodb_uri: String,
}
