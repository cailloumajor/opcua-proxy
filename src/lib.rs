use clap::Args;

pub const OPCUA_DATA_COLL: &str = "data";
pub const OPCUA_HEALTH_COLL: &str = "health";
pub const OPCUA_HEALTH_INTERVAL: u16 = 5000;

#[derive(Args)]
pub struct CommonArgs {
    /// URL of MongoDB server
    #[arg(env, long, default_value = "mongodb://mongodb")]
    pub mongodb_uri: String,

    /// Name of the MongoDB database to use
    #[arg(env, long)]
    pub mongodb_database: String,
}
