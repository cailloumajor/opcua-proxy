mod db;
mod model;
mod opcua;
mod variant;

use std::sync::Arc;

use anyhow::{anyhow, Context as _};
use clap::Parser;
use clap_verbosity_flag::{InfoLevel, LogLevel, Verbosity};
use futures_util::StreamExt;
use opcua_proxy::CommonArgs;
use signal_hook::consts::TERM_SIGNALS;
use signal_hook::low_level::signal_name;
use signal_hook_tokio::Signals;
use tokio::sync::oneshot;
use tokio::task::spawn_blocking;
use tracing::{error, info, instrument};
use tracing_log::log::LevelFilter;
use tracing_log::LogTracer;

use db::MongoDBDatabase;

#[derive(Parser)]
struct Args {
    #[command(flatten)]
    common: CommonArgs,

    /// Path of JSON file to get tag set from
    #[arg(env, long)]
    tag_set_config_path: String,

    #[command(flatten)]
    opcua: opcua::Config,

    #[command(flatten)]
    verbose: Verbosity<InfoLevel>,
}

fn filter_from_verbosity<T>(verbosity: &Verbosity<T>) -> tracing::level_filters::LevelFilter
where
    T: LogLevel,
{
    match verbosity.log_level_filter() {
        LevelFilter::Off => tracing::level_filters::LevelFilter::OFF,
        LevelFilter::Error => tracing::level_filters::LevelFilter::ERROR,
        LevelFilter::Warn => tracing::level_filters::LevelFilter::WARN,
        LevelFilter::Info => tracing::level_filters::LevelFilter::INFO,
        LevelFilter::Debug => tracing::level_filters::LevelFilter::DEBUG,
        LevelFilter::Trace => tracing::level_filters::LevelFilter::TRACE,
    }
}

#[instrument(skip_all)]
async fn handle_signals(
    signals: Signals,
    session_stop_tx: oneshot::Sender<opcua::SessionCommand>,
) -> anyhow::Result<()> {
    let mut signals_stream = signals.map(|signal| signal_name(signal).unwrap_or("unknown"));
    let mut session_stop_tx = Some(session_stop_tx);
    let mut graceful = false;
    info!(status = "starting");
    while let Some(signal) = signals_stream.next().await {
        info!(msg = "received signal", reaction = "shutting down", signal);
        if let Some(tx) = session_stop_tx.take() {
            tx.send(opcua::SessionCommand::Stop).unwrap();
            graceful = true;
        } else {
            error!(err = "stop command has already been sent", signal);
        }
    }
    info!(status = "terminating");
    if graceful {
        Ok(())
    } else {
        Err(anyhow!("unexpected shutdown"))
    }
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let args = Args::parse();

    tracing_subscriber::fmt()
        .with_max_level(filter_from_verbosity(&args.verbose))
        .init();

    LogTracer::init_with_filter(LevelFilter::Warn).context("error initializing log tracer")?;

    let database =
        MongoDBDatabase::create(&args.common.mongodb_uri, &args.common.partner_id).await?;
    database.delete_data_collection().await;

    let opcua_session =
        spawn_blocking(move || opcua::create_session(&args.opcua, &args.common.partner_id))
            .await
            .context("error joining opcua::create_session blocking task")?
            .context("error creating OPC-UA session")?;

    let namespaces = {
        let session = opcua_session.read();
        opcua::get_namespaces(&*session).context("error getting namespaces")?
    };

    let tag_set = spawn_blocking(move || opcua::TagSet::from_file(&args.tag_set_config_path))
        .await
        .context("error joining tag set creation blocking task")?
        .context("error getting tag set")?;

    let tags_receiver = {
        let session = opcua_session.read();
        opcua::subscribe_to_tags(&*session, &namespaces, Arc::new(tag_set))
            .context("error subscribing to tags")?
    };
    let data_change_task = database.handle_data_change(tags_receiver);

    let health_receiver = {
        let session = opcua_session.read();
        opcua::subscribe_to_health(&*session).context("error subscribing to health")?
    };
    let health_task = database.handle_health(health_receiver);

    let session_stop_tx = opcua::Session::run_async(opcua_session);

    let signals = Signals::new(TERM_SIGNALS).context("error registering termination signals")?;
    let signals_handle = signals.handle();
    let signals_task = tokio::spawn(handle_signals(signals, session_stop_tx));

    tokio::try_join!(data_change_task, health_task)
        .context("error joining data change and/or health tasks")?;

    signals_handle.close();

    signals_task.await.context("error joining signals task")??;

    Ok(())
}
