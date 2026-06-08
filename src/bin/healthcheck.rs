use std::time::Duration;

use anyhow::{Context, anyhow};
use clap::Parser;
use opcua_proxy::CommonArgs;
use tonic::Request;
use tonic::transport::Endpoint;
use tonic_health::pb::HealthCheckRequest;
use tonic_health::pb::health_check_response::ServingStatus;
use tonic_health::pb::health_client::HealthClient;

#[derive(Parser)]
struct Args {
    #[command(flatten)]
    common: CommonArgs,
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let args = Args::try_parse().context("Failed to parse arguments")?;

    let addr = args.common.centrifugo_proxy_listen_address;

    let endpoint = Endpoint::from_shared(format!("http://{}:{}", addr.ip(), addr.port()))
        .context("Failed to create endpoint")?
        .connect_timeout(Duration::from_secs(1))
        .connect()
        .await
        .context("Failed to connect health client")?;

    let mut client = HealthClient::new(endpoint);

    let req = Request::new(HealthCheckRequest::default());

    let resp = client
        .check(req)
        .await
        .context("Failed to check for health status")?;

    if resp.into_inner().status() != ServingStatus::Serving {
        return Err(anyhow!("Status is unhealthy"));
    }

    Ok(())
}
