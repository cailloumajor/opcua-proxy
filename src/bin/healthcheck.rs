use std::time::Duration;

use anyhow::anyhow;
use clap::Parser;
use futures_util::TryStreamExt;
use mongodb::Client;
use mongodb::bson::doc;
use mongodb::options::{ClientOptions, FindOptions};
use serde::Deserialize;

use opcua_proxy::{CommonArgs, OPCUA_HEALTH_COLL, OPCUA_HEALTH_INTERVAL};

const SERVER_SELECTION_TIMEOUT: Duration = Duration::from_secs(1);

#[derive(Parser)]
struct Args {
    #[command(flatten)]
    common: CommonArgs,
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase")]
struct Health {
    #[serde(rename = "_id")]
    id: String,
    updated_since: i64,
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let args = Args::parse();

    let mut options = ClientOptions::parse(args.common.mongodb_uri).await?;
    options.app_name = env!("CARGO_PKG_NAME").to_string().into();
    options.server_selection_timeout = SERVER_SELECTION_TIMEOUT.into();

    let client = Client::with_options(options)?;
    let db = client.database(&args.common.mongodb_database);
    let collection = db.collection::<Health>(OPCUA_HEALTH_COLL);
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
    let options = FindOptions::builder().projection(projection).build();

    let health_documents = collection
        .find(doc! {})
        .with_options(options)
        .await?
        .try_collect::<Vec<_>>()
        .await?;

    if health_documents.is_empty() {
        return Err(anyhow!("health collection is empty or does not exist"));
    }

    let outdated_ids = health_documents
        .into_iter()
        .filter_map(|doc| {
            if doc.updated_since > OPCUA_HEALTH_INTERVAL.into() {
                Some(doc.id)
            } else {
                None
            }
        })
        .collect::<Vec<_>>();

    if !outdated_ids.is_empty() {
        let ids = outdated_ids.join(", ");
        return Err(anyhow!("outdated health for ids: {ids}"));
    }

    Ok(())
}
