# OpenTelemetry Instrumentation for AKS Store Demo

This document describes the OpenTelemetry instrumentation implemented across the microservices of the AKS Store Demo application.

## 1. Overview

The AKS Store Demo application has been instrumented with OpenTelemetry to provide observability across its microservices. The implementation includes:

- **Auto-Instrumentation** for JavaScript (Node.js) and Python services
- **Manual Instrumentation** for Go and Rust services
- **Centralized Collector** deployment on AKS
- **Exporters** for Jaeger, Prometheus, and OpenSearch

## 2. Service Instrumentation

### 2.1 Go (makeline-service)

The Go service uses manual instrumentation with the OpenTelemetry Go SDK:

- Added dependencies to `go.mod`
- Added tracer provider initialization in `main.go`
- Added middleware for HTTP request tracking
- Instrumented key functions for queue processing and database operations

### 2.2 JavaScript/Node.js (order-service)

The Node.js service uses auto-instrumentation:

- Added OpenTelemetry dependencies to `package.json`
- Created `instrumentation.js` file for SDK configuration
- Set up auto-instrumentation for HTTP, Fastify, and MongoDB
- Configured through environment variables in the Kubernetes deployment

### 2.3 Python (ai-service)

The Python service uses auto-instrumentation:

- Added OpenTelemetry dependencies to `requirements.txt`
- Created `telemetry.py` module for initializing instrumentation
- Set up auto-instrumentation for FastAPI, requests, and HTTP clients
- Uses `opentelemetry-instrument` wrapper in the Docker container

### 2.4 Rust (product-service, virtual-customer, virtual-worker)

The Rust services use manual instrumentation:

- Added OpenTelemetry dependencies to `Cargo.toml`
- Created a `telemetry.rs` module for tracer initialization
- Integrated with the Rust tracing ecosystem using `tracing-opentelemetry`
- Added span creation to key service endpoints using the `#[instrument]` attribute

### 2.5 Vue.js (store-admin, store-front)

The Vue.js frontend applications use browser-based instrumentation:

- Added OpenTelemetry Web SDK dependencies to `package.json`
- Created a `telemetry.ts` module for initialization
- Set up auto-instrumentation for document loads, user interactions, and HTTP requests
- Configured propagation of trace context to backend services
- Enabled through environment variables at build time

## 3. OpenTelemetry Collector

The OpenTelemetry Collector is deployed as a centralized service on AKS to receive, process, and export telemetry data:

- Receives telemetry via OTLP (gRPC and HTTP protocols)
- Processes data using batch, memory limiter, and resource processors
- Exports data to multiple backends (Jaeger, Prometheus, OpenSearch)
- Configured with appropriate resource attributes for environment identification

## 4. Observability Stack

The following observability tools are deployed alongside the application:

### 4.1 Jaeger (Distributed Tracing)

- Receives trace data from the OpenTelemetry Collector
- Provides a UI for visualizing trace data at `http://<jaeger-ui-service-ip>:16686`
- Stores trace data in-memory (for production, configure persistent storage)

### 4.2 Prometheus (Metrics)

- Scrapes metrics from the OpenTelemetry Collector
- Provides a UI for querying metrics at `http://<prometheus-service-ip>:9090`
- Configured to discover and scrape Kubernetes pods with appropriate annotations

### 4.3 OpenSearch (Optional)

- Can be configured to receive logs and traces from the OpenTelemetry Collector
- Provides powerful search and visualization capabilities
- For production use, deploy with proper storage configuration

## 5. Usage

### 5.1 Viewing Traces

1. Access the Jaeger UI at `http://<jaeger-ui-service-ip>:16686`
2. Select a service from the dropdown menu
3. Click "Find Traces" to see distributed traces across services

### 5.2 Viewing Metrics

1. Access the Prometheus UI at `http://<prometheus-service-ip>:9090`
2. Use the query interface to explore metrics
3. Example queries:
   - `rate(http_server_duration_seconds_count[5m])`: HTTP request rate
   - `histogram_quantile(0.95, rate(http_server_duration_seconds_bucket[5m]))`: 95th percentile latency

## 6. Configuration

Each service can be configured through environment variables:

- `OTEL_EXPORTER_OTLP_ENDPOINT`: The endpoint for the OpenTelemetry Collector
- `ENVIRONMENT`: The environment name (production, staging, development)
- Service-specific variables as documented in each service's deployment file

## 7. Next Steps

- Add additional spans to capture more business logic
- Configure alert rules based on metrics
- Set up log correlation with traces
- Add custom metrics for business KPIs
- Consider implementing sampling strategies for production
