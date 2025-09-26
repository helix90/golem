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

### Solution 1: Use Updated Dockerfile
The main `Dockerfile` has been updated with better module resolution:
- Sets `GO111MODULE=on` early
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

1. **Module not initialized**: Go module might not be properly initialized in Docker context
2. **Build context issues**: Files might not be included in Docker build context
3. **Timing issues**: Dependencies downloaded before source code copied
4. **Path resolution**: Module path not resolved correctly in container

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
