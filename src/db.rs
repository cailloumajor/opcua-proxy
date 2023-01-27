use std::collections::HashMap;
use std::fmt;
use std::time::Duration;

use actix::prelude::*;
use anyhow::{Context as _, Result};
use futures_util::FutureExt;
use mongodb::bson::{self, doc, DateTime, Document};
use mongodb::options::{ClientOptions, UpdateOptions};
use mongodb::results::DeleteResult;
use mongodb::{Client, Database};
use tracing::{debug, debug_span, error, info, info_span, Instrument};

use opcua_proxy::{DATABASE, OPCUA_DATA_COLL, OPCUA_HEALTH_COLL};

use crate::variant::Variant;

const VALUES_KEY: &str = "val";
const TIMESTAMPS_KEY: &str = "ts";

pub(crate) type DatabaseActorAddress = Addr<DatabaseActor>;

#[tracing::instrument(skip_all)]
pub(crate) async fn create_client(uri: &str, partner_id: &str) -> Result<Client> {
    let mut options = ClientOptions::parse(uri)
        .await
        .context("error parsing connection string URI")?;
    let app_name = format!("OPC-UA proxy ({partner_id})");
    options.app_name = app_name.into();
    options.server_selection_timeout = Duration::from_secs(2).into();
    let client = Client::with_options(options).context("error creating the client")?;

    info!(status = "success");
    Ok(client)
}

pub(crate) struct DatabaseActor {
    partner_id: String,
    db: Database,
}

impl DatabaseActor {
    pub(crate) fn new(partner_id: String, client: Client) -> Self {
        Self {
            partner_id,
            db: client.database(DATABASE),
        }
    }
}

impl Actor for DatabaseActor {
    type Context = Context<Self>;

    fn started(&mut self, ctx: &mut Self::Context) {
        let collection = self.db.collection::<Document>(OPCUA_DATA_COLL);
        let document_id = self.partner_id.to_owned();
        let query = doc! { "_id": &document_id };
        ctx.wait(
            async move {
                match collection.delete_one(query, None).await {
                    Ok(DeleteResult { deleted_count, .. }) => {
                        info!(status = "deleted", document_id, deleted_count);
                    }
                    Err(err) => {
                        error!(when = "deleting document", document_id, %err);
                    }
                }
            }
            .instrument(info_span!("database_actor_started_hook"))
            .into_actor(self),
        )
    }
}

#[derive(Debug)]
struct DataValue {
    value: Variant,
    source_timestamp: DateTime,
}

#[derive(Debug, Message)]
#[rtype(result = "()")]
pub(crate) struct DataChangeMessage(HashMap<String, DataValue>);

impl DataChangeMessage {
    pub(crate) fn with_capacity(cap: usize) -> Self {
        Self(HashMap::with_capacity(cap))
    }

    pub(crate) fn insert(&mut self, tag_name: String, value: Variant, source_millis: i64) {
        let source_timestamp = DateTime::from_millis(source_millis);
        let data_value = DataValue {
            value,
            source_timestamp,
        };
        self.0.insert(tag_name, data_value);
    }
}

impl Handler<DataChangeMessage> for DatabaseActor {
    type Result = ResponseFuture<()>;

    fn handle(&mut self, msg: DataChangeMessage, _ctx: &mut Self::Context) -> Self::Result {
        let collection = self.db.collection::<Document>(OPCUA_DATA_COLL);
        let query = doc! { "_id": &self.partner_id };
        let options = UpdateOptions::builder().upsert(true).build();
        async move {
            debug!(event = "message received", ?msg);
            let mut values_map = HashMap::with_capacity(msg.0.len());
            let mut timestamps_map = HashMap::with_capacity(msg.0.len());
            for (tag_name, data_value) in msg.0 {
                values_map.insert(format!("{VALUES_KEY}.{tag_name}"), data_value.value);
                timestamps_map.insert(
                    format!("{TIMESTAMPS_KEY}.{tag_name}"),
                    data_value.source_timestamp,
                );
            }
            let values_doc = match bson::to_document(&values_map) {
                Ok(doc) => doc,
                Err(err) => {
                    error!(when = "encoding values document", %err);
                    return;
                }
            };
            let timestamps_doc = match bson::to_document(&timestamps_map) {
                Ok(doc) => doc,
                Err(err) => {
                    error!(when = "encoding timestamps document", %err);
                    return;
                }
            };
            let update = vec![
                doc! { "$addFields": { "updatedAt": "$$NOW" } },
                doc! { "$addFields": values_doc },
                doc! { "$addFields": timestamps_doc },
            ];
            if let Err(err) = collection.update_one(query, update, options).await {
                error!(when = "updating document", %err);
            }
        }
        .instrument(debug_span!("handle data change message"))
        .boxed()
    }
}

#[derive(Message)]
#[rtype(result = "()")]
pub(crate) struct HealthMessage(DateTime);

impl fmt::Display for HealthMessage {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        self.0.fmt(f)
    }
}

impl From<i64> for HealthMessage {
    fn from(millis: i64) -> Self {
        Self(DateTime::from_millis(millis))
    }
}

impl Handler<HealthMessage> for DatabaseActor {
    type Result = ResponseFuture<()>;

    fn handle(&mut self, msg: HealthMessage, _ctx: &mut Self::Context) -> Self::Result {
        let collection = self.db.collection::<Document>(OPCUA_HEALTH_COLL);
        let query = doc! { "_id": &self.partner_id };
        let update = doc! {
            "$set": { "serverDateTime": msg.0 },
            "$currentDate": { "updatedAt": true },
        };
        let options = UpdateOptions::builder().upsert(true).build();
        async move {
            debug!(event="message received", %msg);
            match collection.update_one(query, update, options).await {
                Ok(_) => (),
                Err(err) => {
                    error!(when="updating document", %err)
                }
            }
        }
        .instrument(debug_span!("handle health message"))
        .boxed()
    }
}
