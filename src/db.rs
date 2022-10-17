use std::collections::HashMap;
use std::fmt;
use std::slice::Iter;
use std::time::Duration;

use actix::prelude::*;
use anyhow::{Context as _, Result};
use futures_util::FutureExt;
use mongodb::{
    bson::{self, doc, Document},
    options::{ClientOptions, UpdateOptions},
    Client, Database,
};
use tracing::{debug, error, info_span, Instrument};

use crate::variant::Variant;

const DATABASE: &str = "opcua";
const OPCUA_DATA_COLL: &str = "data";

pub(crate) type DatabaseActorAddress = Addr<DatabaseActor>;

pub(crate) async fn create_client(uri: impl AsRef<str>) -> Result<Client> {
    let mut options = ClientOptions::parse(uri)
        .await
        .context("error parsing connection string URI")?;
    options.app_name = "OPC-UA proxy".to_string().into();
    options.server_selection_timeout = Duration::from_secs(2).into();
    Client::with_options(options).context("error creating the client")
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

impl DataChangeMessage {
    pub(crate) fn iter(&self) -> Iter<(String, Variant)> {
        self.0.iter()
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
        let collection = self.db.collection::<Document>(OPCUA_DATA_COLL);
        let query = doc! { "_id": &self.partner_id };
        let update_data: HashMap<String, Variant> = msg
            .iter()
            .map(|(k, v)| ("data.".to_owned() + k.as_str(), v.clone()))
            .collect();
        let options = UpdateOptions::builder().upsert(true).build();
        async move {
            debug!(received = "msg", %msg);
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
                    error!(when = "updating document", err = err.to_string());
                }
            }
        }
        .instrument(info_span!("handle data change message"))
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
