use std::collections::HashMap;
use std::fmt;
use std::time::Duration;
use std::vec::IntoIter;

use actix::prelude::*;
use anyhow::{Context as _, Result};
use futures_util::FutureExt;
use mongodb::{
    bson::{self, doc, DateTime, Document},
    options::{ClientOptions, UpdateOptions},
    Client, Database,
};
use tracing::{debug, debug_span, error, info, Instrument};

use opcua_proxy::{DATABASE, OPCUA_DATA_COLL, OPCUA_HEALTH_COLL};

use crate::variant::Variant;

pub(crate) type DatabaseActorAddress = Addr<DatabaseActor>;

#[tracing::instrument(skip_all)]
pub(crate) async fn create_client(uri: &str, partner_id: &str) -> Result<Client> {
    let mut options = ClientOptions::parse(uri)
        .await
        .context("error parsing connection string URI")?;
    let app_name = format!("OPC-UA proxy ({})", partner_id);
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
}

pub(crate) struct DataChangeMessage(Vec<(String, Variant)>);

impl IntoIterator for DataChangeMessage {
    type Item = (String, Variant);
    type IntoIter = IntoIter<(String, Variant)>;

    fn into_iter(self) -> Self::IntoIter {
        self.0.into_iter()
    }
}

impl FromIterator<(String, Variant)> for DataChangeMessage {
    fn from_iter<T: IntoIterator<Item = (String, Variant)>>(iter: T) -> Self {
        let mut c = Self(Vec::new());
        for i in iter {
            c.0.push(i)
        }
        c
    }
}

impl fmt::Display for DataChangeMessage {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "[ ")?;

        let mut tag_values = self.0.iter().peekable();
        while let Some(tag_value) = tag_values.next() {
            write!(f, "{}={}", tag_value.0, tag_value.1)?;
            if tag_values.peek().is_some() {
                write!(f, ", ")?;
            }
        }

        write!(f, " ]")
    }
}

impl Message for DataChangeMessage {
    type Result = ();
}

impl Handler<DataChangeMessage> for DatabaseActor {
    type Result = ResponseFuture<()>;

    fn handle(&mut self, msg: DataChangeMessage, _ctx: &mut Self::Context) -> Self::Result {
        let message_display = msg.to_string();
        let collection = self.db.collection::<Document>(OPCUA_DATA_COLL);
        let query = doc! { "_id": &self.partner_id };
        let update_data: HashMap<String, Variant> = msg
            .into_iter()
            .map(|(k, v)| ("data.".to_owned() + k.as_str(), v))
            .collect();
        let options = UpdateOptions::builder().upsert(true).build();
        async move {
            debug!(received = "msg", msg = message_display);
            let update_data_doc = match bson::to_document(&update_data) {
                Ok(doc) => doc,
                Err(err) => {
                    error!(
                        when = "encoding data update document",
                        err = err.to_string()
                    );
                    return;
                }
            };
            let update = doc! {
                "$currentDate": { "updatedAt": true },
                "$set": update_data_doc,
            };
            match collection.update_one(query, update, options).await {
                Ok(_) => (),
                Err(err) => {
                    error!(when = "updating document", %err);
                }
            }
        }
        .instrument(debug_span!("handle data change message"))
        .boxed()
    }
}

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

impl Message for HealthMessage {
    type Result = ();
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
            debug!(received = "msg", %msg);
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

#[cfg(test)]
mod tests {
    use super::*;

    mod data_change_message {
        use super::*;
        use opcua::types::Variant as OpcUaVariant;

        #[test]
        fn display() {
            let dcm = DataChangeMessage(vec![
                ("first".into(), OpcUaVariant::from("a value").into()),
                ("second".into(), OpcUaVariant::from(42).into()),
                ("third".into(), OpcUaVariant::from(false).into()),
            ]);
            assert_eq!(dcm.to_string(), "[ first=a value, second=42, third=false ]")
        }
    }
}
