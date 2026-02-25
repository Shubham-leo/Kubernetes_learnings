# Stage 1: Build
FROM rust:1.85-slim AS builder
WORKDIR /app
COPY Cargo.toml Cargo.lock* ./
COPY src ./src
# Pin deps that require Rust 1.88+ to older compatible versions
RUN cargo update actix-http@3.12.0 --precise 3.11.0 && \
    cargo update time@0.3.47 --precise 0.3.41 && \
    cargo build --release

# Stage 2: Run (tiny image - just the binary)
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/target/release/rust-fib-server /usr/local/bin/
EXPOSE 8090
CMD ["rust-fib-server"]
