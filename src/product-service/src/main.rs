use product_service::{configuration::Settings, startup::run, telemetry};
use std::io::Error;
use tracing::info;

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    // Initialize OpenTelemetry tracer
    let _tracer = telemetry::init_tracer()
        .map_err(|e| Error::new(std::io::ErrorKind::Other, e.to_string()))?;
    
    // Initialize application
    let settings = Settings::new().set_wasm_rules_engine(false);
    
    // Log startup
    info!("Starting product-service with OpenTelemetry tracing");
    
    // Run the application
    let server = run(settings)?;
    server.await?;
    
    // Shutdown tracer on application exit
    telemetry::shutdown_tracer();
    
    Ok(())
}
