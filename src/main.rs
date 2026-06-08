use std::sync::Arc;
use std::sync::atomic::AtomicBool;

use anyhow::Context as _;
use clap::Parser;
use clap_verbosity_flag::{InfoLevel, Verbosity};
use futures_util::StreamExt;
use opcua_proxy::CommonArgs;
use signal_hook::consts::TERM_SIGNALS;
use signal_hook::low_level::signal_name;
use signal_hook_tokio::Signals;
use tracing::level_filters::LevelFilter;
use tracing::{info, instrument, warn};
use tracing_subscriber::filter::Targets;
use tracing_subscriber::layer::SubscriberExt;
use tracing_subscriber::util::SubscriberInitExt;
use url::Url;

use self::centrifugo::{CentrifugoClient, CentrifugoConfig, run_centrifugo_proxy_server};
use self::opcua::{SessionManager, create_client};

mod centrifugo;
mod channel;
mod opcua;

#[derive(Parser)]
struct Args {
    #[command(flatten)]
    common: CommonArgs,

    /// Base API URL to get configuration from
    #[arg(env, long)]
    config_api_url: Url,

    /// Path of the PKI directory
    #[arg(env, long)]
    pki_dir: String,

    #[command(flatten)]
    centrifugo_config: CentrifugoConfig,

    /// The logging verbosity of opcua library
    #[arg(env, long, default_value = "warn")]
    opcua_verbosity: LevelFilter,

    #[command(flatten)]
    verbosity: Verbosity<InfoLevel>,
}

#[instrument(skip_all)]
async fn handle_signals(signals: Signals) {
    let mut signals_stream = signals.map(|signal| signal_name(signal).unwrap_or("unknown"));
    info!(status = "started");
    let Some(signal) = signals_stream.next().await else {
        warn!(msg = "signals stream exhausted");
        return;
    };
    info!(msg = "received signal", reaction = "shutting down", signal);
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let args = Args::parse();

    let filter = Targets::new()
        .with_default(args.verbosity)
        .with_target("opcua::", args.opcua_verbosity);

    tracing_subscriber::registry()
        .with(tracing_subscriber::fmt::layer())
        .with(filter)
        .init();

    let tag_change_channel = CentrifugoClient::new(&args.centrifugo_config).handle_tag_changes();

    let opcua_client = create_client(args.pki_dir).context("Failed to create OPC-UA client")?;

    let healthy = Arc::new(AtomicBool::new(false));

    let session_manager = SessionManager::new(
        args.config_api_url,
        Arc::new(opcua_client),
        tag_change_channel,
    )
    .context("Failed to create OPC-UA session manager")?;
    let (current_data_channel, session_manager_task) = session_manager.spawn(Arc::clone(&healthy));

    let signals = Signals::new(TERM_SIGNALS).context("Failed to register termination signals")?;
    let signals_handle = signals.handle();

    run_centrifugo_proxy_server(
        args.common.centrifugo_proxy_listen_address,
        &args.centrifugo_config.centrifugo_namespace,
        handle_signals(signals),
        current_data_channel,
        healthy,
    )
    .await
    .context("Fail to run Centrifugo proxy server")?;

    signals_handle.close();

    session_manager_task
        .await
        .context("Failed to join session manager task")?;

    Ok(())
}
