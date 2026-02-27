# Multi-stage build: compile in full Go image, run in tiny alpine
# Final image ~12MB instead of ~800MB

# Stage 1: Build
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o gateway main.go

# Stage 2: Run
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/gateway .
EXPOSE 8080
# Exec form so the binary is PID 1 and receives SIGTERM directly.
# Shell form (CMD sh -c ./gateway) would swallow the signal.
CMD ["./gateway"]
