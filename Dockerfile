# Multi-stage build for Golem Telegram Bot
FROM golang:1.21-alpine AS builder

# Install git and ca-certificates for dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the Telegram bot application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o telegram-bot examples/telegram_bot.go

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
