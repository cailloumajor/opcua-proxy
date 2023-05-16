use std::sync::Arc;

use anyhow::{anyhow, Context as _};
use clap::Parser;
use clap_verbosity_flag::{InfoLevel, Verbosity};
use env_logger::Env;
use futures_util::StreamExt;
use opcua_proxy::CommonArgs;
use signal_hook::consts::TERM_SIGNALS;
use signal_hook::low_level::signal_name;
use signal_hook_tokio::Signals;
use tokio::sync::oneshot;
use tracing::{error, info, instrument};
use tracing_log::format_trace;
use url::Url;

mod config;
mod db;
mod level_filter;
mod model;
mod opcua;
mod variant;

use config::fetch_config;
use db::MongoDBDatabase;
use level_filter::VerbosityLevelFilter;

#[derive(Parser)]
struct Args {
    #[command(flatten)]
    common: CommonArgs,

    /// Base API URL to get configuration from
    #[arg(env, long)]
    config_api_url: Url,

    #[command(flatten)]
    opcua: opcua::Config,

    #[command(flatten)]
    verbose: Verbosity<InfoLevel>,
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

fn main() -> anyhow::Result<()> {
    let args = Args::parse();

    tracing_subscriber::fmt()
        .with_max_level(VerbosityLevelFilter::from(&args.verbose))
        .init();
    env_logger::Builder::from_env(Env::default().default_filter_or("info,opcua=warn"))
        .format(|_, record| format_trace(record))
        .init();

    let rt = tokio::runtime::Builder::new_multi_thread()
        .enable_all()
        .build()
        .context("error building async runtime")?;

    let config_from_api = rt
        .block_on(fetch_config(&args.config_api_url, &args.common.partner_id))
        .context("error fetching configuration")?;

    let opcua_session = opcua::create_session(&args.opcua, &args.common.partner_id)
        .context("error creating OPC-UA session")?;

    let tag_set = {
        let session = opcua_session.read();
        let namespaces = opcua::get_namespaces(&*session).context("error getting namespaces")?;
        opcua::TagSet::from_config(config_from_api.tags, &namespaces, &*session)
            .context("error converting config to tag set")?
    };

    tag_set
        .check_contains_tags(&config_from_api.record_age_for_tags)
        .context("error checking age-recorded tags")?;

    let database = rt.block_on(async {
        let db = MongoDBDatabase::create(&args.common.mongodb_uri, &args.common.partner_id)
            .await
            .context("error creating MongoDB database handle")?;
        db.initialize_data_collection(&config_from_api.record_age_for_tags)
            .await
            .context("error initializing MongoDB data collection")?;
        Ok::<_, anyhow::Error>(db)
    })?;
    let database = Arc::new(database);

    let tags_receiver = {
        let session = opcua_session.read();
        opcua::subscribe_to_tags(&*session, Arc::new(tag_set))
            .context("error subscribing to tags")?
    };
    let data_change_task = database.handle_data_change(&rt, tags_receiver);

    let health_receiver = {
        let session = opcua_session.read();
        opcua::subscribe_to_health(&*session).context("error subscribing to health")?
    };
    let health_task = database.handle_health(&rt, health_receiver);

    let session_stop_tx = opcua::Session::run_async(opcua_session);

    let signals = rt
        .block_on(async { Signals::new(TERM_SIGNALS) })
        .context("error registering termination signals")?;
    let signals_handle = signals.handle();
    let signals_task = rt.spawn(handle_signals(signals, session_stop_tx));

    rt.block_on(async { tokio::try_join!(data_change_task, health_task) })
        .context("error joining data change and/or health tasks")?;

    signals_handle.close();

    rt.block_on(signals_task)
        .context("error joining signals task")??;

    Ok(())
}
