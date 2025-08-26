FROM rust:1.82 AS builder

RUN USER=root cargo new --bin virtual-worker

# Create a new directory for our application
WORKDIR /virtual-worker

# Copy the Cargo.toml and Cargo.lock files to the container
COPY Cargo.toml Cargo.lock ./

# Build the dependencies and cache them
RUN cargo build --release

# Remove the .rs files created by cargo
RUN rm src/*.rs

# Copy the source files to the container
ADD . ./

# Remove the dummy target directory created by cargo
RUN rm ./target/release/deps/virtual_worker*

# Build the application
RUN cargo build --release

# Create a new stage and copy the binary from the builder stage
FROM debian:bookworm-slim AS runner
WORKDIR /app

# Set the build argument for the app version number
ARG APP_VERSION=0.1.0
ARG ENVIRONMENT=production
ARG OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317

# Install OpenSSL
RUN apt-get update && apt-get install -y libssl-dev && rm -rf /var/lib/apt/lists/*

# Copy the binary from the builder stage
COPY --from=builder /virtual-worker/target/release/virtual-worker /app

# Set the environment variables
ENV APP_VERSION=$APP_VERSION
ENV ENVIRONMENT=$ENVIRONMENT
ENV OTEL_EXPORTER_OTLP_ENDPOINT=$OTEL_EXPORTER_OTLP_ENDPOINT
ENV RUST_LOG=info,virtual_worker=debug,tracing=debug

# Run the application
CMD ["./virtual-worker"]