# Multi-stage build: compile in full Go image, run in tiny scratch image
# This keeps the final image ~10MB instead of ~800MB

# Stage 1: Build
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY main.go .
# CGO_ENABLED=0 produces a static binary — no libc dependency
# This lets us use "scratch" (empty) as the runtime image
RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

# Stage 2: Run
FROM alpine:3.19
# We use alpine instead of scratch so we have basic tools for debugging
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
# IMPORTANT: Run the binary directly, NOT through a shell.
# If you use "CMD sh -c ./server", the shell (PID 1) receives SIGTERM,
# but doesn't forward it to the Go process. The Go process never
# gets SIGTERM and can't shut down gracefully!
CMD ["./server"]
