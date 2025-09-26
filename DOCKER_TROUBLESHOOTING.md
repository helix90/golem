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

## The `replace` Directive Solution (RECOMMENDED)

Instead of using `go get` (which requires Git authentication), the updated Dockerfiles use a `replace` directive in the correct order:

```dockerfile
# Copy go.mod and go.sum first
COPY go.mod go.sum ./

# Add replace directive BEFORE downloading dependencies
RUN echo "replace github.com/helix/golem => ./" >> go.mod

# Download dependencies (now with replace directive in place)
RUN go mod download

# Copy the rest of the source code
COPY . .

# Ensure all dependencies are resolved
RUN go mod tidy
```

This approach:
- Avoids Git authentication issues
- Tells Go to use the local directory instead of fetching from GitHub
- Works perfectly in Docker containers without terminal access
- Is the standard Go way to handle local packages in modules
- **Critical**: Add replace directive BEFORE `go mod download` to prevent Git fetch attempts
- **Important**: Uses `./` (not `.`) to satisfy Go's path format requirements

## Alternative: `go get` Solution

If you prefer the `go get` approach, you can use it with Git configuration:

```dockerfile
RUN git config --global url."https://github.com/".insteadOf "git@github.com:" && \
    go get github.com/helix/golem/pkg/golem && \
    go mod tidy
```

However, the `replace` directive is simpler and more reliable for Docker builds.

## Why Order Matters

The critical issue was the **order of operations**:

❌ **Wrong Order** (causes Git authentication error):
```dockerfile
COPY go.mod go.sum ./
RUN go mod download          # Tries to fetch from GitHub
COPY . .
RUN echo "replace ..." >> go.mod  # Too late!
```

✅ **Correct Order** (avoids Git authentication):
```dockerfile
COPY go.mod go.sum ./
RUN echo "replace ..." >> go.mod  # Add replace FIRST
RUN go mod download          # Now uses local replacement
COPY . .
```

When `go mod download` runs without the replace directive, Go tries to fetch the module from GitHub, causing the authentication error.

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
