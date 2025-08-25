// Global types for OpenTelemetry
export interface WindowWithOtel extends Window {
  __otel?: {
    tracer?: any;
  };
}

declare global {
  interface Window {
    __otel?: {
      tracer?: any;
    };
  }
}
