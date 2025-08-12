use opentelemetry::global;
use opentelemetry::sdk::{resource::Resource, trace};
use opentelemetry_otlp::WithExportConfig;
use opentelemetry_sdk::trace::Tracer;
use std::env;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt, EnvFilter};
use opentelemetry::{KeyValue, Value};

pub fn init_tracer() -> Result<Tracer, anyhow::Error> {
    // Get the collector endpoint from environment or use default
    let otlp_endpoint = env::var("OTEL_EXPORTER_OTLP_ENDPOINT")
        .unwrap_or_else(|_| "http://localhost:4317".to_string());

    // Set up resource attributes
    let service_name = "virtual-worker";
    let service_version = env::var("APP_VERSION").unwrap_or_else(|_| "0.1.0".to_string());
    let environment = env::var("ENVIRONMENT").unwrap_or_else(|_| "development".to_string());

    let resource = Resource::new(vec![
        KeyValue::new("service.name", service_name),
        KeyValue::new("service.version", service_version),
        KeyValue::new("environment", Value::String(environment)),
    ]);

    // Configure OpenTelemetry with the OTLP exporter
    let tracer_provider = opentelemetry_otlp::new_pipeline()
        .tracing()
        .with_exporter(
            opentelemetry_otlp::new_exporter()
                .tonic()
                .with_endpoint(otlp_endpoint),
        )
        .with_trace_config(
            trace::config()
                .with_resource(resource)
                .with_sampler(trace::Sampler::AlwaysOn),
        )
        .install_batch(opentelemetry::runtime::Tokio)?;

    // Create a tracer
    let tracer = tracer_provider.tracer(service_name);

    // Configure tracing subscriber with OpenTelemetry
    tracing_subscriber::registry()
        .with(EnvFilter::try_from_default_env().unwrap_or_else(|_| "info".into()))
        .with(tracing_opentelemetry::layer().with_tracer(tracer.clone()))
        .init();

    println!("OpenTelemetry tracing initialized for virtual-worker");

    Ok(tracer)
}

pub fn shutdown_tracer() {
    global::shutdown_tracer_provider();
}
