use std::collections::HashMap;
use std::time::Duration;

use anyhow::Context as _;
use mongodb::bson::{self, Bson, DateTime, Document, doc};
use mongodb::options::{ClientOptions, UpdateOptions};
use mongodb::{Client, Database};
use tokio::runtime::Runtime;
use tokio::sync::mpsc;
use tokio::task::JoinHandle;
use tracing::{Instrument, debug, error, info, info_span, instrument};

use opcua_proxy::{CommonArgs, OPCUA_DATA_COLL, OPCUA_HEALTH_COLL};

use crate::opcua::TagChange;

const MESSAGE_QUEUE_CAPACITY: usize = 20;
const VALUES_KEY: &str = "val";
const TIMESTAMPS_KEY: &str = "ts";

#[derive(Debug)]
pub(crate) struct DataChangeMessage {
    pub(crate) partner_id: String,
    pub(crate) changes: Vec<TagChange>,
}

pub(crate) type DataChangeChannel = mpsc::Sender<DataChangeMessage>;

#[derive(Debug)]
pub(crate) enum HealthCommand {
    Update(i64),
    Remove,
}

#[derive(Debug)]
pub(crate) struct HealthMessage {
    pub(crate) partner_id: String,
    pub(crate) command: HealthCommand,
}

pub(crate) type HealthChannel = mpsc::Sender<HealthMessage>;

#[derive(Clone)]
pub(crate) struct MongoDB(Database);

impl MongoDB {
    #[instrument(skip_all)]
    async fn drop_health_collection(&self) {
        if let Err(err) = self
            .0
            .collection::<Document>(OPCUA_HEALTH_COLL)
            .drop()
            .await
        {
            error!(when = "dropping health collection", %err);
        } else {
            info!(msg = "dropped health collection");
        }
    }

    #[instrument(name = "create_mongodb_database", skip_all)]
    pub(crate) async fn create(uri: &str, config: &CommonArgs) -> anyhow::Result<Self> {
        let mut options = ClientOptions::parse(uri)
            .await
            .context("error parsing connection string URI")?;
        options.app_name = env!("CARGO_PKG_NAME").to_string().into();
        options.server_selection_timeout = Duration::from_secs(2).into();
        let client = Client::with_options(options).context("error creating the client")?;

        info!(status = "success");
        Ok(Self(client.database(&config.mongodb_database)))
    }

    pub(crate) fn handle_data_change(
        &self,
        runtime: &Runtime,
    ) -> (DataChangeChannel, JoinHandle<()>) {
        let (tx, mut rx) = mpsc::channel::<DataChangeMessage>(MESSAGE_QUEUE_CAPACITY);
        let collection = self.0.collection::<Document>(OPCUA_DATA_COLL);
        let options = UpdateOptions::builder().upsert(true).build();

        let task = runtime.spawn(
            async move {
                info!(status = "starting");

                while let Some(message) = rx.recv().await {
                    debug!(event = "message received", ?message);
                    let mut updates_map: HashMap<String, Bson> =
                        HashMap::with_capacity(message.changes.len());
                    for TagChange {
                        tag_name,
                        value,
                        source_timestamp,
                    } in message.changes
                    {
                        updates_map.insert(format!("{VALUES_KEY}.{tag_name}"), value.into());
                        updates_map.insert(
                            format!("{TIMESTAMPS_KEY}.{tag_name}"),
                            DateTime::from_millis(source_timestamp).into(),
                        );
                    }
                    let updates_doc = if updates_map.is_empty() {
                        doc! { VALUES_KEY: {}, TIMESTAMPS_KEY: {} }
                    } else {
                        match bson::to_document(&updates_map) {
                            Ok(doc) => doc,
                            Err(err) => {
                                error!(when = "encoding updates document", %err);
                                continue;
                            }
                        }
                    };
                    let updated_at = if updates_map.is_empty() {
                        doc! {
                            "updatedAt": DateTime::from_millis(-1000)
                        }
                    } else {
                        doc! { "updatedAt": "$$NOW" }
                    };
                    let update = vec![
                        doc! { "$addFields": updated_at },
                        doc! { "$addFields": updates_doc },
                    ];
                    let query = doc! { "_id": message.partner_id };
                    if let Err(err) = collection
                        .update_one(query.clone(), update)
                        .with_options(options.clone())
                        .await
                    {
                        error!(when = "updating document", %err);
                    }
                }

                info!(status = "terminating");
            }
            .instrument(info_span!("mongodb_data_change_handler")),
        );

        (tx, task)
    }

    pub(crate) fn handle_health(&self, runtime: &Runtime) -> (HealthChannel, JoinHandle<()>) {
        let (tx, mut rx) = mpsc::channel::<HealthMessage>(MESSAGE_QUEUE_CAPACITY);
        let collection = self.0.collection::<Document>(OPCUA_HEALTH_COLL);
        let options = UpdateOptions::builder().upsert(true).build();
        let cloned_self = self.clone();

        let task = runtime.spawn(
            async move {
                info!(status = "starting");

                cloned_self.drop_health_collection().await;

                while let Some(message) = rx.recv().await {
                    debug!(event = "message received", ?message);
                    let query = doc! { "_id": message.partner_id };
                    match message.command {
                        HealthCommand::Update(server_timestamp) => {
                            let update = doc! {
                                "$set": { "serverDateTime": server_timestamp },
                                "$currentDate": { "updatedAt": true },
                            };
                            if let Err(err) = collection
                                .update_one(query.clone(), update)
                                .with_options(options.clone())
                                .await
                            {
                                error!(when="updating document", %err);
                            }
                        }
                        HealthCommand::Remove => {
                            if let Err(err) = collection.delete_one(query).await {
                                error!(when="deleting document", %err);
                            }
                        }
                    };
                }

                cloned_self.drop_health_collection().await;

                info!(status = "terminating");
            }
            .instrument(info_span!("mongodb_health_handler")),
        );

        (tx, task)
    }
}
