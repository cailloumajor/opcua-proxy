use std::thread;

use actix::{Actor, System};
use anyhow::{Context, Result};

use clap::Parser;
use opcua_proxy::CommonArgs;
use signal_hook::consts::TERM_SIGNALS;
use signal_hook::iterator::Signals;
use signal_hook::low_level::signal_name;
use tokio::sync::oneshot::Sender;
use tracing::{info, info_span};
use tracing_log::log::LevelFilter;
use tracing_log::LogTracer;

mod db;
mod opcua;
mod variant;

#[derive(Parser)]
struct Args {
    #[command(flatten)]
    common: CommonArgs,

    /// Path of JSON file to get tag set from
    #[arg(env, long)]
    tag_set_config_path: String,

    #[command(flatten)]
    opcua: opcua::Config,
}

fn handle_signals(
    mut signals: Signals,
    session_stop_tx: Sender<opcua::SessionCommand>,
    system: System,
) {
    let mut tx = Some(session_stop_tx);
    while let Some(signal) = signals.forever().next() {
        let _entered = info_span!("signals handler").entered();
        let signal_name = signal_name(signal).unwrap_or("unknown");
        info!(msg = "received signal", signal_name);
        if let Some(tx) = tx.take() {
            info!(msg = "sending session stop command");
            tx.send(opcua::SessionCommand::Stop).unwrap();
        }
        info!(msg = "stopping system");
        system.stop();
        signals.handle().close();
    }
    info!("exiting signals handler");
}

fn main() -> Result<()> {
    let args = Args::parse();

    tracing_subscriber::fmt()
        .with_max_level(tracing::Level::DEBUG)
        .init();

    LogTracer::init_with_filter(LevelFilter::Warn).context("error initializing log tracer")?;

    let system = System::new();

    let db_client = system.block_on(db::create_client(
        &args.common.mongodb_uri,
        &args.common.partner_id,
    ))?;
    let addr = system.block_on(async {
        db::DatabaseActor::new(args.common.partner_id.clone(), db_client).start()
    });

    let opcua_session = opcua::create_session(&args.opcua, &args.common.partner_id) //
        .context("error creating session")?;
    let namespaces = opcua::get_namespaces(&*opcua_session.read()) //
        .context("error getting namespaces")?;
    let tag_set = opcua::TagSet::from_file(&args.tag_set_config_path) //
        .context("error getting tag set")?;
    opcua::subscribe_to_tags(&*opcua_session.read(), &namespaces, tag_set, addr.clone())
        .context("error subscribing to tags")?;
    opcua::subscribe_to_health(&*opcua_session.read(), addr)
        .context("error subscribing to health")?;

    let session_stop_tx = opcua::Session::run_async(opcua_session);

    let signals = Signals::new(TERM_SIGNALS)?;
    let current_system = System::current();
    thread::spawn(move || handle_signals(signals, session_stop_tx, current_system));

    system.run().context("error running system")
}
