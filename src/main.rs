use anyhow::Context as _;
use clap::Parser;
use clap_verbosity_flag::{InfoLevel, Verbosity};
use env_logger::Env;
use futures_util::StreamExt;
use opcua::SessionManager;
use opcua_proxy::CommonArgs;
use signal_hook::consts::TERM_SIGNALS;
use signal_hook::low_level::signal_name;
use signal_hook_tokio::Signals;
use tokio_util::sync::CancellationToken;
use tracing::{info, instrument};
use tracing_log::format_trace;
use url::Url;

mod db;
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
    verbose: Verbosity<InfoLevel>,
}

#[instrument(skip_all)]
async fn handle_signals(signals: Signals, shutdown_token: CancellationToken) {
    let mut signals_stream = signals.map(|signal| signal_name(signal).unwrap_or("unknown"));
    info!(status = "started");
    while let Some(signal) = signals_stream.next().await {
        info!(msg = "received signal", reaction = "shutting down", signal);
        shutdown_token.cancel();
    }
}

fn main() -> anyhow::Result<()> {
    let args = Args::parse();

    tracing_subscriber::fmt()
        .with_max_level(args.verbose.tracing_level())
        .init();
    env_logger::Builder::from_env(Env::default().default_filter_or("info,opcua=warn"))
        .format(|_, record| format_trace(record))
        .init();

    let rt = tokio::runtime::Builder::new_multi_thread()
        .enable_all()
        .build()
        .context("error building async runtime")?;

    let shutdown_token = CancellationToken::new();

    let database = rt
        .block_on(db::MongoDB::create(&args.common.mongodb_uri, &args.common))
        .context("error creating MongoDB database handle")?;
    let (data_change_channel, data_change_task) = database.handle_data_change(&rt);
    let (health_channel, health_task) = database.handle_health(&rt);

    let opcua_client = opcua::ClientBuilder::new()
        .application_name(env!("CARGO_PKG_DESCRIPTION"))
        .product_uri(concat!("urn:", env!("CARGO_PKG_NAME")))
        .application_uri(concat!("urn:", env!("CARGO_PKG_NAME")))
        .pki_dir(args.pki_dir)
        .certificate_path(concat!("own/", env!("CARGO_PKG_NAME"), "-cert.der"))
        .private_key_path(concat!("private/", env!("CARGO_PKG_NAME"), "-key.pem"))
        .session_retry_limit(0)
        .session_timeout(1_200_000)
        .multi_threaded_executor()
        .client()
        .context("error building OPC-UA client")?;

    let signals = rt
        .block_on(async { Signals::new(TERM_SIGNALS) })
        .context("error registering termination signals")?;
    let signals_handle = signals.handle();
    let signals_task = rt.spawn(handle_signals(signals, shutdown_token.clone()));

    let manager = SessionManager::new(
        args.config_api_url,
        opcua_client,
        shutdown_token,
        data_change_channel,
        health_channel,
    );
    rt.block_on(manager.run());

    signals_handle.close();

    rt.block_on(async { tokio::try_join!(data_change_task, health_task, signals_task) })
        .context("error joining data change and/or health tasks")?;

    Ok(())
}
