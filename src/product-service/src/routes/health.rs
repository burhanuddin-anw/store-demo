use actix_web::{Error, HttpResponse};
use serde_json::json;
use tracing::{info, instrument};

#[instrument(name = "health_check", skip_all)]
pub async fn health() -> Result<HttpResponse, Error> {
    let version = std::env::var("APP_VERSION").unwrap_or_else(|_| "0.1.0".to_string());
    let health = json!({"status": "ok", "version": version});
    
    // Log health check with tracing
    info!("Health check request processed successfully");
    
    Ok(HttpResponse::Ok().json(health))
}
