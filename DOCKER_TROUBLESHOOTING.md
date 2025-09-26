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

## The Pre-configured `go.mod` Solution (RECOMMENDED)

Instead of using `go get` (which requires Git authentication), the updated Dockerfiles use a pre-configured `go.mod` file with the replace directive:

```dockerfile
# Copy the pre-configured go.mod with replace directive
COPY go.mod.docker ./go.mod
COPY go.sum ./

# Download dependencies (now with replace directive in place)
RUN go mod download

# Copy the rest of the source code
COPY . .

# Ensure all dependencies are resolved
RUN go mod tidy
```

The `go.mod.docker` file contains:
```
module github.com/helix/golem

go 1.21

require github.com/go-telegram/bot v1.17.0

replace github.com/helix/golem => ./
replace github.com/helix/golem/pkg/golem => ./pkg/golem
```

**Important**: The second replace directive is crucial because Go treats `github.com/helix/golem/pkg/golem` as a separate module path that needs its own replacement.

## Why Two Replace Directives?

Go modules work hierarchically:
- `github.com/helix/golem` - The main module
- `github.com/helix/golem/pkg/golem` - A subpackage that Go treats as a separate module path

Even though they're in the same repository, Go's module resolution treats them as distinct paths that each need their own replace directive.

This approach:
- ✅ **Avoids Git authentication issues** - no need for tokens or credentials
- ✅ **Pre-configured replace directive** - no runtime modification needed
- ✅ **Works perfectly in Docker containers** - no terminal access required
- ✅ **Standard Go approach** - uses official replace directive syntax
- ✅ **No Git fetch attempts** - replace directive is in place from the start
- ✅ **Uses correct path format** - `./` satisfies Go's requirements
- ✅ **Simpler and more reliable** - no complex Git configuration needed

## Alternative: Git Configuration Solution

If the replace directive still causes issues, you can configure Git to use anonymous access:

```dockerfile
# Configure Git for anonymous access to public repositories
RUN git config --global url."https://github.com/".insteadOf "git@github.com:" && \
    git config --global --add url."https://github.com/".insteadOf "ssh://git@github.com/" && \
    git config --global --add url."https://github.com/".insteadOf "git://github.com/" && \
    git config --global credential.helper store && \
    echo "https://anonymous:anonymous@github.com" > ~/.git-credentials
```

This approach:
- Forces Git to use HTTPS instead of SSH
- Provides anonymous credentials for public repositories
- Works with both `go get` and `replace` directive approaches
- Handles the "terminal prompts disabled" error

## Alternative: `go get` Solution

If you prefer the `go get` approach, you can use it with Git configuration:

```dockerfile
RUN git config --global url."https://github.com/".insteadOf "git@github.com:" && \
    go get github.com/helix/golem/pkg/golem && \
    go mod tidy
```

However, the `replace` directive with Git configuration is the most reliable approach for Docker builds.

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

## Why Public Repos Need Authentication

Even though GitHub repositories are public, Git still requires authentication for:
- **Rate limiting**: GitHub limits anonymous requests
- **Protocol detection**: Git tries SSH first, then falls back to HTTPS
- **Credential caching**: Git expects credentials to be available
- **Docker context**: No terminal available for interactive authentication

The Git configuration forces HTTPS with anonymous credentials, bypassing these issues.

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
