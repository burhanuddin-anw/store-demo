// OpenTelemetry auto-instrumentation setup file
'use strict';

// This file should be loaded first via NODE_OPTIONS env variable
// Set NODE_OPTIONS=--require=./instrumentation.js

const { NodeSDK } = require('@opentelemetry/sdk-node');
const { Resource } = require('@opentelemetry/resources');
const { SemanticResourceAttributes } = require('@opentelemetry/semantic-conventions');
const { ConsoleSpanExporter } = require('@opentelemetry/sdk-trace-node');
const { OTLPTraceExporter } = require('@opentelemetry/exporter-trace-otlp-proto');
const { getNodeAutoInstrumentations } = require('@opentelemetry/auto-instrumentations-node');

const serviceName = 'order-service';
const serviceVersion = process.env.APP_VERSION || '0.1.0';
const environment = process.env.ENVIRONMENT || 'development';

// Configure the SDK to export telemetry data to the console
// Enable all auto-instrumentations from the meta package
const sdk = new NodeSDK({
  resource: new Resource({
    [SemanticResourceAttributes.SERVICE_NAME]: serviceName,
    [SemanticResourceAttributes.SERVICE_VERSION]: serviceVersion,
    'environment': environment,
  }),
  traceExporter: new OTLPTraceExporter({
    // Use the endpoint from environment variable or default
    url: process.env.OTEL_EXPORTER_OTLP_ENDPOINT || 'http://localhost:4318/v1/traces',
  }),
  instrumentations: [
    getNodeAutoInstrumentations({
      // Add specific instrumentation configurations here
      '@opentelemetry/instrumentation-http': {
        enabled: true,
      },
      '@opentelemetry/instrumentation-fastify': {
        enabled: true,
      },
      '@opentelemetry/instrumentation-mongodb': {
        enabled: true,
      },
    }),
  ],
});

// initialize the SDK and register with the OpenTelemetry API
// this enables the API to record telemetry
sdk.start();

// gracefully shut down the SDK on process exit
process.on('SIGTERM', () => {
  sdk.shutdown()
    .then(() => console.log('Tracing terminated'))
    .catch((error) => console.log('Error terminating tracing', error))
    .finally(() => process.exit(0));
});

console.log('OpenTelemetry auto-instrumentation initialized for order-service');
