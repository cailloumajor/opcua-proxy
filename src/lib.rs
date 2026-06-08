use std::net::SocketAddr;

use clap::Args;

#[derive(Args)]
pub struct CommonArgs {
    /// The address for the Centrifugo proxy server (gRPC) to listen on.
    #[arg(env, long, default_value = "0.0.0.0:50051")]
    pub centrifugo_proxy_listen_address: SocketAddr,
}
