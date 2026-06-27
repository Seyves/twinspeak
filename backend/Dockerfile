# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 for static binary, -ldflags for optimization
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o /app/bin/server \
    ./cmd/twinspeek/main.go

# Final stage
FROM alpine:latest

# Install CA certificates and timezone data (wget is already in busybox)
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy the binary
COPY --from=builder /app/bin/server /app/server

# Copy config example (user should mount actual config)
COPY --from=builder /app/config-example.yaml /app/config-example.yaml

# Create non-root user
RUN addgroup -g 1001 -S appuser && \
    adduser -S appuser -u 1001 -G appuser

# Expose port
EXPOSE 8080

# Run as non-root user
USER appuser

# Run the application
ENTRYPOINT ["/app/server"]
CMD ["-c", "/config/config.yaml"]
