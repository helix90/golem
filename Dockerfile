# Multi-stage build for Golem Telegram Bot
FROM golang:1.21-alpine AS builder

# Install git and ca-certificates for dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Set Go module mode to ensure local packages are found
ENV GO111MODULE=on
ENV CGO_ENABLED=0

# Copy go mod files first
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire source code
COPY . .

# Use replace directive to handle local package without Git authentication
RUN echo "replace github.com/helix/golem => ." >> go.mod && \
    go mod tidy && \
    echo "=== Module information ===" && \
    go list -m all && \
    echo "=== Directory structure ===" && \
    ls -la && \
    echo "=== pkg/golem directory ===" && \
    ls -la pkg/golem/ && \
    echo "=== Testing import resolution ===" && \
    go list -f '{{.ImportPath}}' ./pkg/golem

# Build the Telegram bot application
RUN go build -a -installsuffix cgo -o telegram-bot examples/telegram_bot.go

# Final stage - minimal image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/telegram-bot .

# Copy AIML test data
COPY --from=builder /app/testdata ./testdata

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port (if needed for health checks)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD pgrep telegram-bot || exit 1

# Run the application
CMD ["./telegram-bot"]
