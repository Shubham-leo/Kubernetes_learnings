# Stage 1 -- compile the Go binary in a full SDK image
FROM golang:1.23-alpine AS build
WORKDIR /src
COPY main.go .
# Static linking (CGO_ENABLED=0) so the binary runs on any Linux
# base image, including minimal ones with no C library.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o upstream main.go

# Stage 2 -- copy the binary into a minimal runtime image
# Using alpine instead of scratch so we keep basic debugging tools
# (sh, wget, nslookup) available for troubleshooting in the cluster.
FROM alpine:3.20
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=build /src/upstream .
EXPOSE 8080
# Exec form so the Go process is PID 1 and directly receives SIGTERM.
# Shell form (CMD ./upstream) would make sh PID 1, which swallows the
# signal and prevents graceful shutdown.
CMD ["./upstream"]
