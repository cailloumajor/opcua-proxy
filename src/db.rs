use std::collections::HashMap;
use std::time::Duration;

use anyhow::Context as _;
use mongodb::bson::{self, doc, Document};
use mongodb::options::{ClientOptions, UpdateOptions};
use mongodb::results::DeleteResult;
use mongodb::{Client, Database};
use tokio::sync::mpsc;
use tokio::task::JoinHandle;
use tracing::{debug, error, info, info_span, Instrument};

use opcua_proxy::{DATABASE, OPCUA_DATA_COLL, OPCUA_HEALTH_COLL};

use crate::model::{DataChangeMessage, HealthMessage};

const VALUES_KEY: &str = "val";
const TIMESTAMPS_KEY: &str = "ts";

pub(crate) struct MongoDBDatabase {
    partner_id: String,
    db: Database,
}

impl MongoDBDatabase {
    #[tracing::instrument(name = "create_mongodb_database", skip_all)]
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
            partner_id: partner_id.to_string(),
            db: client.database(DATABASE),
        })
    }

    #[tracing::instrument(skip_all)]
    pub(crate) async fn delete_data_collection(&self) {
        let collection = self.db.collection::<Document>(OPCUA_DATA_COLL);
        let document_id = self.partner_id.to_owned();
        let query = doc! { "_id": &document_id };
        match collection.delete_one(query, None).await {
            Ok(DeleteResult { deleted_count, .. }) => {
                info!(status = "deleted", document_id, deleted_count);
            }
            Err(err) => {
                error!(when = "deleting document", document_id, %err);
            }
        }
    }

    pub(crate) fn handle_data_change(
        &self,
        mut messages: mpsc::Receiver<DataChangeMessage>,
    ) -> JoinHandle<()> {
        let collection = self.db.collection::<Document>(OPCUA_DATA_COLL);
        let query = doc! { "_id": &self.partner_id };
        let options = UpdateOptions::builder().upsert(true).build();

        tokio::spawn(
            async move {
                info!(status = "starting");

                while let Some(message) = messages.recv().await {
                    debug!(event = "message received", ?message);
                    let mut values_map = HashMap::with_capacity(message.len());
                    let mut timestamps_map = HashMap::with_capacity(message.len());
                    for (tag_name, value, source_timestamp) in message {
                        values_map.insert(format!("{VALUES_KEY}.{tag_name}"), value);
                        timestamps_map
                            .insert(format!("{TIMESTAMPS_KEY}.{tag_name}"), source_timestamp);
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
                        .update_one(query.to_owned(), update, options.to_owned())
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
        mut messages: mpsc::Receiver<HealthMessage>,
    ) -> JoinHandle<()> {
        let collection = self.db.collection::<Document>(OPCUA_HEALTH_COLL);
        let query = doc! { "_id": &self.partner_id };
        let options = UpdateOptions::builder().upsert(true).build();

        tokio::spawn(
            async move {
                info!(status = "starting");

                while let Some(message) = messages.recv().await {
                    debug!(event="message received", %message);
                    let update = doc! {
                        "$set": { "serverDateTime": message.date_time() },
                        "$currentDate": { "updatedAt": true },
                    };
                    match collection
                        .update_one(query.to_owned(), update, options.to_owned())
                        .await
                    {
                        Ok(_) => (),
                        Err(err) => {
                            error!(when="updating document", %err)
                        }
                    }
                }

                info!(status = "terminating");
            }
            .instrument(info_span!("mongodb_health_handler")),
        )
    }
}
