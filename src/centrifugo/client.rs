use std::fmt;
use std::time::Duration;

use centrifugo_grpc::api::PublishRequest;
use centrifugo_grpc::api::centrifugo_api_client::CentrifugoApiClient;
use tokio::sync::mpsc;
use tonic::transport::{Channel, Endpoint};
use tracing::{Instrument, error, info, info_span};

use super::config::{CentrifugoApiKey, CentrifugoConfig};

/// The capacity of the tag change message channel.
const MESSAGE_QUEUE_CAPACITY: usize = 20;

/// The maximum time allowed for gRPC requests to last.
const REQUEST_TIMEOUT: Duration = Duration::from_secs(1);

/// Represents a tag change to be handled by Centrifugo client.
pub(crate) struct TagChangeMessage {
    /// The identifier of the OPC-UA partner from which the change originates.
    pub(crate) partner_id: String,
    /// The data to publish.
    pub(crate) data: Vec<u8>,
}

/// A client for the Centrifugo server gRPC API.
#[derive(Clone)]
pub(crate) struct CentrifugoClient {
    /// The client to send commands to Centrifugo.
    client: CentrifugoApiClient<Channel>,
    /// Centrifugo channel namespace to use in messages.
    namespace: String,
    /// Common metadata to use for each request.
    api_key: CentrifugoApiKey,
}

impl CentrifugoClient {
    /// Create a new [`CentrifugoClient`].
    pub(crate) fn new(config: &CentrifugoConfig) -> Self {
        let endpoint = Endpoint::from_shared(config.centrifugo_server_uri.to_string())
            .expect("creating an endpoint from an URI should not fail");
        let channel = endpoint.timeout(REQUEST_TIMEOUT).connect_lazy();
        let client = CentrifugoApiClient::new(channel);
        let namespace = config.centrifugo_namespace.clone();
        let api_key = config.centrifugo_api_key.clone();

        Self {
            client,
            namespace,
            api_key,
        }
    }

    /// Start a task to handle [`TagChangeMessage`], returning the sending side of the message channel.
    ///
    /// This method consumes the [`CentrifugoClient`].
    pub(crate) fn handle_tag_changes(self) -> mpsc::Sender<TagChangeMessage> {
        let (tx, mut rx) = mpsc::channel::<TagChangeMessage>(MESSAGE_QUEUE_CAPACITY);

        tokio::spawn(
            async move {
                info!(status = "starting");

                while let Some(message) = rx.recv().await {
                    if let Err(err) = self.publish(message).await {
                        error!(during = "publishing change to Centrifugo", %err);
                    }
                }
            }
            .instrument(info_span!("centrifugo_client_tag_change_handler")),
        );

        tx
    }

    /// Publish provided tag change.
    async fn publish(&self, msg: TagChangeMessage) -> Result<(), PublishError> {
        let mut request = tonic::Request::new(PublishRequest {
            channel: format!("{}:{}", self.namespace, msg.partner_id),
            data: msg.data,
            ..Default::default()
        });
        request
            .metadata_mut()
            .insert("authorization", self.api_key.clone().into());

        let response = self
            .client
            .clone()
            .publish(request)
            .await
            .map(|resp| resp.into_inner())
            .map_err(PublishError::Grpc)?;

        if let Some(error) = response.error {
            return Err(PublishError::Centrifugo(error));
        }

        Ok(())
    }
}

/// Represents errors that can occur when publishing.
#[derive(Debug)]
pub(super) enum PublishError {
    /// A gRPC error.
    Grpc(tonic::Status),
    /// An error returned by Centrifugo server.
    Centrifugo(centrifugo_grpc::api::Error),
}

impl fmt::Display for PublishError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Grpc(status) => write!(f, "gRPC error: {status}"),
            Self::Centrifugo(err) => write!(
                f,
                "Centrifugo server error, code: {}, message: {}",
                err.code, err.message
            ),
        }
    }
}

impl std::error::Error for PublishError {}
