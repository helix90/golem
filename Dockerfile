# Multi-stage build for Golem Telegram Bot
FROM golang:1.21-alpine AS builder

# Install git and ca-certificates for dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Set Go module mode
ENV GO111MODULE=on
ENV CGO_ENABLED=0

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application as a static binary
RUN go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o telegram-bot examples/telegram_bot.go

# Final stage - minimal image with just the binary
FROM scratch

# Copy ca-certificates from builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the static binary from builder stage
COPY --from=builder /app/telegram-bot /telegram-bot

# Copy AIML test data
COPY --from=builder /app/testdata /testdata

# Set working directory
WORKDIR /

# Run the application
CMD ["/telegram-bot"]
