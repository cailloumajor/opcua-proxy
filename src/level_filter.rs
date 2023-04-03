use clap_verbosity_flag::{LogLevel, Verbosity};
use tracing::level_filters::LevelFilter;

pub(crate) struct VerbosityLevelFilter(tracing_log::log::LevelFilter);

impl<T> From<&Verbosity<T>> for VerbosityLevelFilter
where
    T: LogLevel,
{
    fn from(value: &Verbosity<T>) -> Self {
        Self(value.log_level_filter())
    }
}

impl From<VerbosityLevelFilter> for LevelFilter {
    fn from(value: VerbosityLevelFilter) -> Self {
        use tracing_log::log::LevelFilter;
        match value.0 {
            LevelFilter::Off => Self::OFF,
            LevelFilter::Error => Self::ERROR,
            LevelFilter::Warn => Self::WARN,
            LevelFilter::Info => Self::INFO,
            LevelFilter::Debug => Self::DEBUG,
            LevelFilter::Trace => Self::TRACE,
        }
    }
}
