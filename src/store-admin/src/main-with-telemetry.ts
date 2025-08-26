// src/main.ts
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import './assets/styles.scss'
import { initTelemetry } from './telemetry'
import App from './App.vue'
import router from './router'

// Initialize OpenTelemetry
const { tracer } = initTelemetry();

// Add tracer to the global window object for debugging/reference
(window as any).__otel = { tracer };

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')
