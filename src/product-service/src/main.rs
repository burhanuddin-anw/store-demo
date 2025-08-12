use env_logger::Env;
use product_service::{configuration::Settings, startup::run, telemetry};
use std::io::Error;
use tracing::{info, instrument};

#[actix_web::main]
#[instrument]
async fn main() -> std::io::Result<()> {
    // Initialize OpenTelemetry tracer
    let _tracer = telemetry::init_tracer()
        .map_err(|e| Error::new(std::io::ErrorKind::Other, e.to_string()))?;
    
    // Initialize application
    let settings = Settings::new().set_wasm_rules_engine(false);
    
    // Initialize env logger
    env_logger::init_from_env(Env::default().default_filter_or(&settings.log_level));
    
    // Log startup
    info!("Starting product-service with OpenTelemetry tracing");
    
    // Run the application
    let result = run(settings)?.await;
    
    // Shutdown tracer on application exit
    telemetry::shutdown_tracer();
    
    Ok(result)
}
