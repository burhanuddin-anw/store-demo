"""
OpenTelemetry auto-instrumentation wrapper for ai-service
"""

import os
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.instrumentation.requests import RequestsInstrumentor
# Using aiohttp-client-instrumentation package
# Note: aiohttp instrumentation requires opentelemetry-instrumentation-aiohttp-client package
try:
    from opentelemetry.instrumentation.aiohttp_client import AioHttpClientInstrumentor
    AIOHTTP_AVAILABLE = True
except ImportError:
    AIOHTTP_AVAILABLE = False
from opentelemetry.semconv.resource import ResourceAttributes


def configure_telemetry(app):
    """Configure OpenTelemetry auto-instrumentation for the FastAPI application"""
    try:
        # Get collector endpoint from environment or use default
        otlp_endpoint = os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")
        
        # Get service details from environment
        service_name = "ai-service"
        service_version = os.getenv("APP_VERSION", "0.1.0")
        environment = os.getenv("ENVIRONMENT", "development")
        
        # Configure resource attributes
        resource = Resource.create({
            ResourceAttributes.SERVICE_NAME: service_name,
            ResourceAttributes.SERVICE_VERSION: service_version,
            "environment": environment,
        })
        
        # Create a TracerProvider with the resource
        tracer_provider = TracerProvider(resource=resource)
        
        # Create and add span processor with OTLP exporter
        span_processor = BatchSpanProcessor(
            OTLPSpanExporter(endpoint=otlp_endpoint, insecure=True)
        )
        tracer_provider.add_span_processor(span_processor)
        
        # Set the tracer provider
        trace.set_tracer_provider(tracer_provider)
        
        # Instrument FastAPI
        FastAPIInstrumentor.instrument_app(app)
        
        # Instrument other libraries
        RequestsInstrumentor().instrument()
        if AIOHTTP_AVAILABLE:
            AioHttpClientInstrumentor().instrument()
        
        print("OpenTelemetry instrumentation initialized for ai-service")
        return tracer_provider
    except Exception as e:
        print(f"Failed to initialize OpenTelemetry instrumentation: {e}")
        return None

def shutdown_telemetry(tracer_provider):
    """Shutdown OpenTelemetry tracer provider gracefully"""
    if tracer_provider:
        try:
            tracer_provider.shutdown()
            print("OpenTelemetry tracer provider shut down successfully")
        except Exception as e:
            print(f"Error shutting down tracer provider: {e}")
