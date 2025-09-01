import { createApp } from 'vue'
import { createPinia } from 'pinia'
import './assets/styles.scss'

import App from './App.vue'
import router from './router'

// Initialize OpenTelemetry first
import { initTelemetry } from './telemetry'

let telemetryInitialized = false;
try {
  const { tracer } = initTelemetry();
  // Add tracer to the global window object for debugging/reference
  (window as any).__otel = { tracer };
  telemetryInitialized = true;
  console.log('OpenTelemetry initialized successfully');
} catch (error) {
  console.warn('OpenTelemetry initialization failed:', error);
  console.log('Application will continue without OpenTelemetry');
}

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.mount('#app')

// Log initialization status
console.log('Store-admin mounted with OpenTelemetry:', telemetryInitialized);
