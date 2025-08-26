# Grafana Configuration

## Access Information
- **URL**: http://localhost:3003
- **Username**: admin
- **Password**: admin

## Setup Steps

1. **Add Prometheus Data Source**:
   - Go to Configuration → Data Sources
   - Click "Add data source"
   - Select "Prometheus"
   - URL: `http://prometheus:9090`
   - Click "Save & Test"

2. **Import Dashboard**:
   You can import pre-built dashboards for:
   - OpenTelemetry Collector: Dashboard ID `15983`
   - Node.js Applications: Dashboard ID `11159`
   - Go Applications: Dashboard ID `10826`
   - RabbitMQ: Dashboard ID `4279`

## Available Metrics

### Application Metrics (via OpenTelemetry)
- `http_server_duration_milliseconds` - HTTP request duration
- `http_server_active_requests` - Active HTTP requests
- `http_client_duration_milliseconds` - HTTP client request duration

### OpenTelemetry Collector Metrics
- `otelcol_receiver_accepted_spans_total` - Spans received
- `otelcol_receiver_refused_spans_total` - Spans refused
- `otelcol_exporter_sent_spans_total` - Spans exported
- `otelcol_processor_batch_send_size_sum` - Batch sizes

### Infrastructure Metrics
- `process_cpu_seconds_total` - Process CPU usage
- `process_resident_memory_bytes` - Process memory usage
- `go_memstats_*` - Go runtime metrics (for Go services)

## Example Queries for Grafana

### Request Rate
```
rate(http_server_duration_milliseconds_count[5m])
```

### Average Response Time
```
rate(http_server_duration_milliseconds_sum[5m]) / rate(http_server_duration_milliseconds_count[5m])
```

### Error Rate
```
rate(http_server_duration_milliseconds_count{http_status_code=~"5.."}[5m]) / rate(http_server_duration_milliseconds_count[5m])
```

### OpenTelemetry Spans Rate
```
rate(otelcol_receiver_accepted_spans_total[5m])
```
