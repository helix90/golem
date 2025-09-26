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
- Adds local package with `go get github.com/helix/golem/pkg/golem`
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

## The `go get` Solution

When Go suggests adding `go get github.com/helix/golem/pkg/golem`, this is the correct approach for Docker builds. The updated Dockerfiles now include:

```dockerfile
RUN go get github.com/helix/golem/pkg/golem && \
    go mod tidy
```

This explicitly tells Go to treat the local package as a dependency, which resolves the module resolution issue in Docker containers.

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
