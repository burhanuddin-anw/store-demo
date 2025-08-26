import { createApp } from 'vue'
import { createPinia } from 'pinia'
import './assets/styles.scss'

import App from './App.vue'
import router from './router'

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.mount('#app')

// Initialize OpenTelemetry asynchronously after the app is mounted
import('./telemetry').then(({ initTelemetry }) => {
  const { tracer } = initTelemetry();
  // Add tracer to the global window object for debugging/reference
  (window as any).__otel = { tracer };
}).catch(error => {
  console.warn('OpenTelemetry initialization failed:', error);
  console.log('Application will continue without OpenTelemetry');
});
