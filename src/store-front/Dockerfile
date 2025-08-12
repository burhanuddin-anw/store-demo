# Use an official Node.js runtime as a parent image
FROM node:22.14-alpine AS builder

WORKDIR /app

# Set build arguments for OpenTelemetry configuration
ARG APP_VERSION=0.1.0
ARG ENVIRONMENT=production
ARG OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318/v1/traces

# Copy package.json and package-lock.json to the container
COPY package*.json ./

# Install app dependencies
RUN npm install

# Copy the rest of the app source code to the container
COPY . .

# Set environment variables for the build
ENV VITE_APP_VERSION=$APP_VERSION
ENV VITE_ENVIRONMENT=$ENVIRONMENT
ENV VITE_OTEL_EXPORTER_OTLP_ENDPOINT=$OTEL_EXPORTER_OTLP_ENDPOINT

# Rename telemetry-enabled main file to use OpenTelemetry
RUN cp src/main-with-telemetry.ts src/main.ts

# Build the app
RUN npm run build

# Run the app on nginx
FROM nginx:stable-alpine-slim AS runner

# Copy the build output to replace the default nginx contents
COPY --from=builder /app/dist /usr/share/nginx/html

# Expose the port the app listens on
EXPOSE 8080

# Set the build argument for the app version number
ARG APP_VERSION=0.1.0

# Set the environment variable for the app version number
ENV APP_VERSION=$APP_VERSION

# Copy the nginx configuration template to the container
COPY nginx.conf /etc/nginx/conf.d/nginx.conf.template

# Update the nginx configuration to use the app version number
# and Copy the nginx configuration template to the container
RUN envsubst '${APP_VERSION}' < /etc/nginx/conf.d/nginx.conf.template > /etc/nginx/conf.d/default.conf

# Start the app
CMD ["nginx", "-g", "daemon off;"]