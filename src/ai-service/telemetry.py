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
from opentelemetry.instrumentation.aiohttp_client import AioHttpClientInstrumentor
from opentelemetry.semconv.resource import ResourceAttributes


def configure_telemetry(app):
    """Configure OpenTelemetry auto-instrumentation for the FastAPI application"""
    
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
    AioHttpClientInstrumentor().instrument()
    
    print("OpenTelemetry auto-instrumentation initialized for ai-service")
    
    return tracer_provider
