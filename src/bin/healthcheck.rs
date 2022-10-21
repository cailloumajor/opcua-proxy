use std::time::Duration;

use anyhow::{anyhow, Result};
use clap::Parser;
use mongodb::{
    bson::{doc, DateTime},
    options::{ClientOptions, FindOneOptions},
    Client,
};
use serde::Deserialize;

use opcua_proxy::{CommonArgs, DATABASE, OPCUA_HEALTH_COLL, OPCUA_HEALTH_INTERVAL};

const SERVER_SELECTION_TIMEOUT: Duration = Duration::from_secs(1);

#[derive(Parser)]
struct Args {
    #[command(flatten)]
    common: CommonArgs,
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase")]
struct Health {
    updated_at: DateTime,
    updated_since: i64,
}

#[tokio::main]
async fn main() -> Result<()> {
    let args = Args::parse();

    let app_name = format!("OPC-UA proxy healtheck ({})", args.common.partner_id);

    let mut options = ClientOptions::parse(args.common.mongodb_uri).await?;
    options.app_name = app_name.into();
    options.server_selection_timeout = SERVER_SELECTION_TIMEOUT.into();

    let client = Client::with_options(options)?;
    let db = client.database(DATABASE);
    let collection = db.collection::<Health>(OPCUA_HEALTH_COLL);
    let query = doc! { "_id": &args.common.partner_id };
    let projection = doc! {
        "updatedAt": true,
        "updatedSince" : {
            "$dateDiff": {
                "startDate": "$updatedAt",
                "endDate": "$$NOW",
                "unit": "millisecond",
            },
        },
    };
    let options = FindOneOptions::builder().projection(projection).build();

    let health = collection
        .find_one(query, options)
        .await?
        .ok_or_else(|| anyhow!("document was not found"))?;

    if health.updated_since > OPCUA_HEALTH_INTERVAL.into() {
        return Err(anyhow!("outdated health data: {}", health.updated_at));
    }

    Ok(())
}
