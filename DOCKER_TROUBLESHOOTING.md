# Docker Build Troubleshooting Guide

## Issue: Module Resolution Error

If you encounter the error:
```
examples/telegram_bot.go:13:2: no required module provides package github.com/helix/golem/pkg/golem; to add it
```

## Debugging Steps

### 1. Test Local Build
```bash
# Ensure local build works
go mod tidy
go build -o telegram-bot examples/telegram_bot.go
```

### 2. Check Module Structure
```bash
# Verify module information
go list -m all
go list -f '{{.ImportPath}}' ./pkg/golem
```

### 3. Test Docker Build with Debug
```bash
# Use the minimal Dockerfile for debugging
docker build -f Dockerfile.minimal -t golem-debug .

# Or use the debug script
./build-debug.sh
```

### 4. Check Build Context
```bash
# Verify all files are included in build context
docker build --no-cache -t golem-test .
```

## Common Solutions

### Solution 1: Use Updated Dockerfile (RECOMMENDED)
The main `Dockerfile` has been updated with better module resolution:
- Sets `GO111MODULE=on` early
- Uses `replace` directive to handle local package without Git authentication
- Runs `go mod tidy` after copying source
- Includes debug output to verify module structure

### Solution 2: Alternative Dockerfile
If the main Dockerfile doesn't work, try `Dockerfile.alternative`:
```bash
docker build -f Dockerfile.alternative -t golem-telegram-bot .
```

### Solution 3: Minimal Test
Use `Dockerfile.minimal` for debugging:
```bash
docker build -f Dockerfile.minimal -t golem-minimal .
```

## Root Causes

1. **Local package not recognized**: Go doesn't recognize local packages in Docker context
2. **Module not initialized**: Go module might not be properly initialized in Docker context
3. **Build context issues**: Files might not be included in Docker build context
4. **Timing issues**: Dependencies downloaded before source code copied
5. **Path resolution**: Module path not resolved correctly in container

## The Proper Multi-Stage Solution (RECOMMENDED)

The correct approach is to use a proper multi-stage Docker build that compiles the Go binary in one stage and copies only the executable to the final stage:

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

# Install dependencies
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

# Run the application
CMD ["/telegram-bot"]
```

## Why This Works

Most Go programs deployed in Docker work because they:
1. **Build the entire module together** - no external module resolution needed
2. **Use proper module structure** - local packages are part of the same module
3. **Don't try to resolve local packages as external dependencies**

The key insight is that `github.com/helix/golem/pkg/golem` is a **local package** within the same module, not an external dependency that needs to be fetched from GitHub.

## Why Replace Directives Don't Work

Replace directives are meant for:
- Replacing external dependencies with local versions
- Using forks of external packages
- Testing with local versions of dependencies

They are **not needed** for packages within the same module. When you import `github.com/helix/golem/pkg/golem` from within the `github.com/helix/golem` module, Go automatically resolves it as a local package.

This approach:
- ✅ **No Git authentication issues** - dependencies downloaded in build stage only
- ✅ **Minimal final image** - only the compiled binary and data files
- ✅ **Static binary** - no runtime dependencies needed
- ✅ **Fast builds** - Docker layer caching for dependencies
- ✅ **Secure** - minimal attack surface with scratch base image
- ✅ **Standard Go practice** - how most production Go applications are built
- ✅ **No module resolution issues** - everything compiled in build stage

## What NOT to Do

❌ **Don't use replace directives for local packages** - they're meant for external dependencies
❌ **Don't try to fetch local packages from GitHub** - they're already in your module
❌ **Don't use complex Git configuration** - it's unnecessary for local packages
❌ **Don't copy go.mod separately** - copy the entire project at once

## The Root Cause

The fundamental issue was **misunderstanding Go modules**:
- We treated local packages as external dependencies
- We tried to use replace directives for packages within the same module
- We overcomplicated a simple build process

The solution is to understand that `github.com/helix/golem/pkg/golem` is a **local package** within the `github.com/helix/golem` module, not an external dependency.

## Verification

After successful build, verify the container works:
```bash
# Test the built container
docker run --rm -e TELEGRAM_BOT_TOKEN=test -e AIML_PATH=testdata golem-telegram-bot
```

## Environment Variables

Make sure to set required environment variables:
```bash
export TELEGRAM_BOT_TOKEN="your_bot_token_here"
export AIML_PATH="testdata"
export VERBOSE="true"
```

## Docker Compose

Use the provided docker-compose.yml:
```bash
# Copy environment file
cp env.example .env
# Edit .env with your bot token
# Run with docker compose
docker compose up -d
```
