mod db;
mod model;
mod opcua;
mod variant;

use std::process;
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;

use anyhow::Context as _;
use clap::Parser;
use clap_verbosity_flag::{InfoLevel, LogLevel, Verbosity};
use futures_util::StreamExt;
use opcua_proxy::CommonArgs;
use signal_hook::consts::TERM_SIGNALS;
use signal_hook::low_level::signal_name;
use signal_hook_tokio::Signals;
use tokio::runtime::Runtime;
use tokio::sync::oneshot;
use tracing::{error, info, instrument};
use tracing_log::log::LevelFilter;
use tracing_log::LogTracer;

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
    raw_exit: Arc<AtomicBool>,
    session_stop_tx: oneshot::Sender<opcua::SessionCommand>,
) {
    let mut signals_stream =
        signals.map(|signal| (signal, signal_name(signal).unwrap_or("unknown")));
    let mut session_stop_tx = Some(session_stop_tx);
    while let Some((signum, signal)) = signals_stream.next().await {
        if raw_exit.load(Ordering::Relaxed) {
            info!(msg = "received signal", reaction = "exiting", signal);
            process::exit(signum);
        } else {
            info!(msg = "received signal", reaction = "shutting down", signal);
            if let Some(tx) = session_stop_tx.take() {
                tx.send(opcua::SessionCommand::Stop).unwrap();
            } else {
                error!(err = "stop command has already been sent", signal);
            }
        }
    }
}

fn main() -> anyhow::Result<()> {
    let rt = Runtime::new().context("error creating async runtime")?;

    let args = Args::parse();

    tracing_subscriber::fmt()
        .with_max_level(filter_from_verbosity(&args.verbose))
        .init();

    LogTracer::init_with_filter(LevelFilter::Warn).context("error initializing log tracer")?;

    let should_raw_exit = Arc::new(AtomicBool::new(true));
    let (session_tx, session_rx) = oneshot::channel();

    let signals = rt
        .block_on(async { Signals::new(TERM_SIGNALS) })
        .context("error registering termination signals")?;
    let signals_handle = signals.handle();
    let signals_task = rt.spawn(handle_signals(
        signals,
        Arc::clone(&should_raw_exit),
        session_tx,
    ));

    let database = rt.block_on(db::create_database(
        &args.common.mongodb_uri,
        &args.common.partner_id,
    ))?;
    let database_actor = Arc::new(db::DatabaseActor::new(
        args.common.partner_id.clone(),
        database,
    ));
    rt.block_on(database_actor.delete_data_collection());

    let opcua_session = opcua::create_session(&args.opcua, &args.common.partner_id)
        .context("error creating session")?;
    should_raw_exit.store(false, Ordering::Relaxed);

    let namespaces = {
        let session = opcua_session.read();
        opcua::get_namespaces(&*session).context("error getting namespaces")?
    };

    let tag_set = opcua::TagSet::from_file(&args.tag_set_config_path) //
        .context("error getting tag set")?;

    let tags_receiver = {
        let session = opcua_session.read();
        opcua::subscribe_to_tags(&*session, &namespaces, Arc::new(tag_set))
            .context("error subscribing to tags")?
    };
    let actor = Arc::clone(&database_actor);
    let data_change_task = rt.spawn(async move { actor.handle_data_change(tags_receiver).await });

    let health_receiver = {
        let session = opcua_session.read();
        opcua::subscribe_to_health(&*session).context("error subscribing to health")?
    };
    let actor = Arc::clone(&database_actor);
    let health_task = rt.spawn(async move { actor.handle_health(health_receiver).await });

    opcua::Session::run_loop(opcua_session, 10, session_rx);

    signals_handle.close();

    rt.block_on(async { tokio::try_join!(signals_task, data_change_task, health_task) })
        .context("error joining tasks")?;

    Ok(())
}
