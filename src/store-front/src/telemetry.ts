import { WebTracerProvider } from '@opentelemetry/sdk-trace-web';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';
import { Resource } from '@opentelemetry/resources';
import { ATTR_SERVICE_NAME, ATTR_SERVICE_VERSION } from '@opentelemetry/semantic-conventions';
import { SimpleSpanProcessor, BatchSpanProcessor } from '@opentelemetry/sdk-trace-base';
import { ZoneContextManager } from '@opentelemetry/context-zone';

// Initialize the OpenTelemetry provider
export function initTelemetry() {
  try {
    // Use environment variables if available
    const serviceName = 'store-front';
    const serviceVersion = import.meta.env.VITE_APP_VERSION || '0.1.0';
    const environment = import.meta.env.VITE_ENVIRONMENT || 'development';
    const collectorUrl = import.meta.env.VITE_OTEL_EXPORTER_OTLP_ENDPOINT || 'http://localhost:4318/v1/traces';

  // Create a resource that identifies your application
  const resource = new Resource({
    [ATTR_SERVICE_NAME]: serviceName,
    [ATTR_SERVICE_VERSION]: serviceVersion,
    'environment': environment
  });

  // Create a provider to manage tracing
  const provider = new WebTracerProvider({ resource });

  // Create exporter to send traces to collector
  const exporter = new OTLPTraceExporter({
    url: collectorUrl,
    headers: {}
  });

  // Use simple span processor for development, batch for production
  const isProduction = environment === 'production';
  if (isProduction) {
    provider.addSpanProcessor(new BatchSpanProcessor(exporter));
  } else {
    provider.addSpanProcessor(new SimpleSpanProcessor(exporter));
  }

  // Register the provider
  provider.register({
    contextManager: new ZoneContextManager()
  });

  // Create a tracer for manual instrumentation if needed
  const tracer = provider.getTracer(serviceName);
  
  console.log(`OpenTelemetry instrumentation initialized for ${serviceName}`);
  
    return { tracer, provider };
  } catch (error) {
    console.error('Failed to initialize OpenTelemetry:', error);
    return { tracer: undefined, provider: undefined };
  }
}

// Cleanup function to be called when the application is unmounting
export function shutdownTelemetry(provider?: WebTracerProvider) {
  if (provider) {
    try {
      provider.shutdown()
        .then(() => console.log('OpenTelemetry provider shut down successfully'))
        .catch(error => console.error('Error shutting down OpenTelemetry provider:', error));
    } catch (error) {
      console.error('Error during OpenTelemetry shutdown:', error);
    }
  }
}
