import { WebTracerProvider } from '@opentelemetry/sdk-trace-web';
import { Resource } from '@opentelemetry/resources';
import { SemanticResourceAttributes } from '@opentelemetry/semantic-conventions';
import { BatchSpanProcessor } from '@opentelemetry/sdk-trace-base';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';
import { ZoneContextManager } from '@opentelemetry/context-zone';
import { registerInstrumentations } from '@opentelemetry/instrumentation';
import { getWebAutoInstrumentations } from '@opentelemetry/auto-instrumentations-web';

// Initialize the OpenTelemetry provider
export function initTelemetry() {
  console.log('OpenTelemetry initialization started for store-front');
  
  try {
    // Configure resource
    const resource = new Resource({
      [SemanticResourceAttributes.SERVICE_NAME]: 'store-front',
      [SemanticResourceAttributes.SERVICE_VERSION]: '0.1.0',
      'environment': 'development'
    });

    // Create tracer provider
    const provider = new WebTracerProvider({ resource });

    // Configure OTLP exporter
    const exporter = new OTLPTraceExporter({
      url: 'http://localhost:4318/v1/traces',
    });

    // Add span processor
    provider.addSpanProcessor(new BatchSpanProcessor(exporter));

    // Register the provider
    provider.register({
      contextManager: new ZoneContextManager(),
    });

    // Register auto-instrumentations
    registerInstrumentations({
      instrumentations: [getWebAutoInstrumentations()],
    });

    const tracer = provider.getTracer('store-front');
    
    console.log('OpenTelemetry successfully initialized for store-front');
    
    return { 
      tracer, 
      provider 
    };
  } catch (error) {
    console.error('Failed to initialize OpenTelemetry:', error);
    console.log('Application will continue without OpenTelemetry instrumentation');
    
    return { 
      tracer: undefined, 
      provider: undefined 
    };
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
