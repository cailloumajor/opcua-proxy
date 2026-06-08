use std::net::SocketAddr;
use std::sync::Arc;
use std::sync::atomic::{AtomicBool, Ordering};
use std::time::Duration;

use centrifugo_grpc::proxy::centrifugo_proxy_server::{CentrifugoProxy, CentrifugoProxyServer};
use centrifugo_grpc::proxy::{
    Error as ProxyError, SubscribeRequest, SubscribeResponse, SubscribeResult,
};
use tonic::transport::Server;
use tonic::{Request, Response, Status};
use tonic_health::pb::health_check_response::ServingStatus;
use tonic_health::pb::health_server::{Health, HealthServer};
use tonic_health::pb::{HealthCheckRequest, HealthCheckResponse};
use tonic_health::server::WatchStream;
use tracing::{error, info, instrument};

use crate::channel::RoundtripSender;

/// The sending side of the channel to request current tag values.
pub(crate) type CurrentTagValuesChannel = RoundtripSender<String, Option<Vec<u8>>>;

/// Maximum time allowed for sending current values request.
const CHANNEL_SEND_TIMEOUT: Duration = Duration::from_millis(50);

/// Maximum time allowed for receiving current values.
const CHANNEL_RECEIVE_TIMEOUT: Duration = Duration::from_millis(100);

/// Centrifugo gRPC proxy service.
struct ProxyService {
    /// The expected Centrifugo namespace.
    centrifugo_ns_prefix: String,
    /// The sender side of the channel to request current data for a partner.
    current_data_channel: CurrentTagValuesChannel,
}

#[tonic::async_trait]
impl CentrifugoProxy for ProxyService {
    async fn subscribe(
        &self,
        req: Request<SubscribeRequest>,
    ) -> Result<Response<SubscribeResponse>, Status> {
        let SubscribeRequest { channel, .. } = req.get_ref();

        let Some(channel_name) = channel.strip_prefix(&self.centrifugo_ns_prefix) else {
            return Ok(Response::new(SubscribeResponse {
                error: Some(ProxyError {
                    code: 1000,
                    message: "unknown channel namespace".into(),
                    temporary: false,
                }),
                ..Default::default()
            }));
        };

        let data = match self
            .current_data_channel
            .roundtrip(
                channel_name.into(),
                CHANNEL_SEND_TIMEOUT,
                CHANNEL_RECEIVE_TIMEOUT,
            )
            .await
        {
            Ok(d) => d,
            Err(err) => {
                error!(during = "current values channel roundtrip", %err);
                return Ok(subscribe_response_internal_error());
            }
        };

        let result = data.map(|d| SubscribeResult {
            data: d,
            ..Default::default()
        });
        let subscribe_response = SubscribeResponse {
            result,
            ..Default::default()
        };

        Ok(Response::new(subscribe_response))
    }
}

/// Health service, wrapping a shareable indicator of health status.
struct HealthService(Arc<AtomicBool>);

#[tonic::async_trait]
impl Health for HealthService {
    type WatchStream = WatchStream;

    async fn check(
        &self,
        _: Request<HealthCheckRequest>,
    ) -> Result<Response<HealthCheckResponse>, Status> {
        let status = if self.0.load(Ordering::Relaxed) {
            ServingStatus::Serving
        } else {
            ServingStatus::NotServing
        };

        Ok(Response::new(HealthCheckResponse {
            status: status as i32,
        }))
    }

    async fn watch(
        &self,
        _: Request<HealthCheckRequest>,
    ) -> Result<Response<Self::WatchStream>, Status> {
        Err(Status::unimplemented("Not implemented"))
    }
}

/// Run the Centrifugo proxy server, provided the address to listen on, a shutdown future,
/// and a channel to send current tag values requests.
#[instrument(skip_all)]
pub(crate) async fn run_centrifugo_proxy_server<F>(
    listen_addr: SocketAddr,
    centrifugo_ns: &str,
    shutdown: F,
    current_data_channel: CurrentTagValuesChannel,
    healthy: Arc<AtomicBool>,
) -> Result<(), tonic::transport::Error>
where
    F: Future<Output = ()>,
{
    info!(status = "started");

    let health_svc = HealthServer::new(HealthService(healthy));

    let mut centrifugo_ns_prefix = String::from(centrifugo_ns);
    centrifugo_ns_prefix.push(':');

    let proxy_svc = CentrifugoProxyServer::new(ProxyService {
        current_data_channel,
        centrifugo_ns_prefix,
    });

    Server::builder()
        .add_service(health_svc)
        .add_service(proxy_svc)
        .serve_with_shutdown(listen_addr, shutdown)
        .await?;

    info!(status = "terminating");

    Ok(())
}

/// Build a proxy subscribe response with internal error content.
fn subscribe_response_internal_error() -> Response<SubscribeResponse> {
    Response::new(SubscribeResponse {
        error: Some(ProxyError {
            code: 1999,
            message: "internal server error".into(),
            temporary: false,
        }),
        ..Default::default()
    })
}
