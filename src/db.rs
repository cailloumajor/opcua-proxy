use std::collections::HashMap;
use std::time::Duration;

use anyhow::Context as _;
use arcstr::ArcStr;
use mongodb::bson::{self, doc, DateTime, Document};
use mongodb::options::{ClientOptions, ReplaceOptions, UpdateOptions};
use mongodb::{Client, Database};
use tokio::runtime::Runtime;
use tokio::sync::mpsc;
use tokio::task::JoinHandle;
use tracing::{debug, error, info, info_span, instrument, Instrument};

use opcua_proxy::{DATABASE, OPCUA_DATA_COLL, OPCUA_HEALTH_COLL};

use crate::opcua::TagChange;

const VALUES_KEY: &str = "val";
const TIMESTAMPS_KEY: &str = "ts";

#[derive(Clone)]
pub(crate) struct MongoDBDatabase {
    partner_id: ArcStr,
    db: Database,
}

impl MongoDBDatabase {
    #[instrument(skip_all)]
    async fn delete_health_collection(&self) {
        let query = doc! { "_id": self.partner_id.as_str() };
        match self
            .db
            .collection::<Document>(OPCUA_HEALTH_COLL)
            .delete_one(query, None)
            .await
        {
            Ok(delete_result) if delete_result.deleted_count > 0 => {
                info!(msg = "deleted health collection");
            }
            Ok(_) => {}
            Err(err) => {
                error!(when = "deleting health collection", %err);
            }
        }
    }

    #[instrument(name = "create_mongodb_database", skip_all)]
    pub(crate) async fn create(uri: &str, partner_id: &str) -> anyhow::Result<Self> {
        let mut options = ClientOptions::parse(uri)
            .await
            .context("error parsing connection string URI")?;
        let app_name = format!("OPC-UA proxy ({partner_id})");
        options.app_name = app_name.into();
        options.server_selection_timeout = Duration::from_secs(2).into();
        let client = Client::with_options(options).context("error creating the client")?;

        info!(status = "success");
        Ok(Self {
            partner_id: ArcStr::from(partner_id),
            db: client.database(DATABASE),
        })
    }

    #[instrument(skip_all)]
    pub(crate) async fn initialize_data_collection(&self) -> anyhow::Result<()> {
        let collection = self.db.collection::<Document>(OPCUA_DATA_COLL);
        let query = doc! { "_id": self.partner_id.as_str() };
        let options = ReplaceOptions::builder().upsert(true).build();
        let replacement = doc! {
            "val": {},
            "ts": {},
            "updatedAt": DateTime::from_millis(-1000),
        };
        collection.replace_one(query, replacement, options).await?;

        info!(status = "success");
        Ok(())
    }

    pub(crate) fn handle_data_change(
        &self,
        runtime: &Runtime,
        mut messages: mpsc::Receiver<Vec<TagChange>>,
    ) -> JoinHandle<()> {
        let collection = self.db.collection::<Document>(OPCUA_DATA_COLL);
        let query = doc! { "_id": self.partner_id.as_str() };
        let options = UpdateOptions::builder().upsert(true).build();

        runtime.spawn(
            async move {
                info!(status = "starting");

                while let Some(message) = messages.recv().await {
                    debug!(event = "message received", ?message);
                    let mut values_map = HashMap::with_capacity(message.len());
                    let mut timestamps_map = HashMap::with_capacity(message.len());
                    for TagChange {
                        tag_name,
                        value,
                        source_timestamp,
                    } in message
                    {
                        values_map.insert(format!("{VALUES_KEY}.{tag_name}"), value);
                        timestamps_map.insert(
                            format!("{TIMESTAMPS_KEY}.{tag_name}"),
                            DateTime::from_millis(source_timestamp),
                        );
                    }
                    let values_doc = match bson::to_document(&values_map) {
                        Ok(doc) => doc,
                        Err(err) => {
                            error!(when = "encoding values document", %err);
                            continue;
                        }
                    };
                    let timestamps_doc = match bson::to_document(&timestamps_map) {
                        Ok(doc) => doc,
                        Err(err) => {
                            error!(when = "encoding timestamps document", %err);
                            continue;
                        }
                    };
                    let update = vec![
                        doc! { "$addFields": { "updatedAt": "$$NOW" } },
                        doc! { "$addFields": values_doc },
                        doc! { "$addFields": timestamps_doc },
                    ];
                    if let Err(err) = collection
                        .update_one(query.clone(), update, options.clone())
                        .await
                    {
                        error!(when = "updating document", %err);
                    }
                }

                info!(status = "terminating");
            }
            .instrument(info_span!("mongodb_data_change_handler")),
        )
    }

    pub(crate) fn handle_health(
        &self,
        runtime: &Runtime,
        mut messages: mpsc::Receiver<i64>,
    ) -> JoinHandle<()> {
        let collection = self.db.collection::<Document>(OPCUA_HEALTH_COLL);
        let query = doc! { "_id": self.partner_id.as_str() };
        let options = UpdateOptions::builder().upsert(true).build();
        let cloned_self = self.clone();

        runtime.spawn(
            async move {
                info!(status = "starting");

                cloned_self.delete_health_collection().await;

                while let Some(message) = messages.recv().await {
                    debug!(event="message received", %message);
                    let update = doc! {
                        "$set": { "serverDateTime": message },
                        "$currentDate": { "updatedAt": true },
                    };
                    if let Err(err) = collection
                        .update_one(query.clone(), update, options.clone())
                        .await
                    {
                        error!(when="updating document", %err);
                    }
                }

                cloned_self.delete_health_collection().await;

                info!(status = "terminating");
            }
            .instrument(info_span!("mongodb_health_handler")),
        )
    }
}
