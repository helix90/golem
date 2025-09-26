# Golem Telegram Bot - Docker Setup

This guide explains how to run the Golem Telegram Bot using Docker Compose.

## Quick Start

1. **Get a Telegram Bot Token**:
   - Message [@BotFather](https://t.me/botfather) on Telegram
   - Use `/newbot` command to create a new bot
   - Copy the token you receive

2. **Configure Environment**:
   ```bash
   # Copy the example environment file
   cp env.example .env
   
   # Edit .env and add your bot token
   nano .env
   ```

3. **Run the Bot**:
   ```bash
   # Start the bot
   docker-compose up -d
   
   # View logs
   docker-compose logs -f telegram-bot
   
   # Stop the bot
   docker-compose down
   ```

## Files Overview

- `Dockerfile` - Multi-stage build for the Telegram bot
- `docker-compose.yml` - Service orchestration
- `env.example` - Environment configuration template
- `.dockerignore` - Files to exclude from Docker build context

## Configuration

### Required Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `TELEGRAM_BOT_TOKEN` | Your bot token from @BotFather | `1234567890:ABCdefGHIjklMNOpqrsTUVwxyz` |

### Optional Environment Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `VERBOSE` | Enable verbose logging | `false` | `true` |
| `TZ` | Container timezone | `UTC` | `America/New_York` |

## Usage

### Basic Commands

```bash
# Start the bot in background
docker-compose up -d

# Start with verbose logging
VERBOSE=true docker-compose up -d

# View logs
docker-compose logs -f

# View logs for specific service
docker-compose logs -f telegram-bot

# Stop the bot
docker-compose down

# Stop and remove volumes
docker-compose down -v

# Rebuild and start
docker-compose up --build -d

# Check service status
docker-compose ps

# Execute commands in running container
docker-compose exec telegram-bot sh
```

### Bot Commands

Once running, users can interact with the bot using these commands:

- `/start` - Welcome message and introduction
- `/help` - Show help information
- `/status` - Display bot status and statistics
- `/reload` - Reload the AIML knowledge base
- `/session` - Show current session information
- `/clear` - Clear conversation history

## Customization

### Using Custom AIML Files

1. **Replace the default testdata**:
   ```bash
   # Remove the default volume mount
   # Edit docker-compose.yml and change:
   - ./testdata:/app/aiml-data:ro
   # to:
   - ./your-custom-aiml:/app/aiml-data:ro
   ```

2. **Add custom AIML files**:
   ```bash
   # Create your AIML directory
   mkdir custom-aiml
   
   # Add your .aiml, .map, .set files
   cp your-files/* custom-aiml/
   
   # Update docker-compose.yml
   # Change the volume mount to point to your directory
   ```

### Environment-Specific Configuration

Create different environment files for different deployments:

```bash
# Development
cp env.example .env.development
# Edit .env.development with dev settings

# Production
cp env.example .env.production
# Edit .env.production with prod settings

# Use specific environment file
docker-compose --env-file .env.production up -d
```

## Monitoring (Optional)

The Docker Compose file includes optional monitoring services:

### Enable Monitoring

```bash
# Start with monitoring services
docker-compose --profile monitoring up -d
```

### Available Monitoring Services

- **Prometheus** (port 9090) - Metrics collection
- **Loki** (port 3100) - Log aggregation

### Access Monitoring

- Prometheus: http://localhost:9090
- Loki: http://localhost:3100

## Troubleshooting

### Common Issues

1. **Bot Token Invalid**:
   ```bash
   # Check if token is set correctly
   docker-compose exec telegram-bot env | grep TELEGRAM_BOT_TOKEN
   
   # Verify token format (should be: number:hash)
   echo $TELEGRAM_BOT_TOKEN
   ```

2. **AIML Files Not Loading**:
   ```bash
   # Check if testdata directory exists
   ls -la testdata/
   
   # Check container logs for AIML loading errors
   docker-compose logs telegram-bot | grep -i aiml
   ```

3. **Container Won't Start**:
   ```bash
   # Check container status
   docker-compose ps
   
   # Check logs for errors
   docker-compose logs telegram-bot
   
   # Check if port is already in use
   netstat -tulpn | grep :8080
   ```

4. **Permission Issues**:
   ```bash
   # Fix file permissions
   sudo chown -R $USER:$USER testdata/
   
   # Rebuild container
   docker-compose up --build -d
   ```

### Debug Mode

Enable verbose logging for detailed debugging:

```bash
# Set verbose mode
export VERBOSE=true

# Start with verbose logging
docker-compose up -d

# Follow logs
docker-compose logs -f telegram-bot
```

### Health Checks

The container includes health checks:

```bash
# Check health status
docker-compose ps

# Manual health check
docker-compose exec telegram-bot pgrep telegram-bot
```

## Development

### Building from Source

```bash
# Build the image locally
docker build -t golem-telegram-bot .

# Run the built image
docker run -e TELEGRAM_BOT_TOKEN=your_token golem-telegram-bot
```

### Development with Live Reload

For development, you might want to mount the source code:

```yaml
# Add to docker-compose.yml under volumes
volumes:
  - ./examples:/app/examples:ro
  - ./pkg:/app/pkg:ro
```

## Security Considerations

1. **Environment Variables**: Never commit `.env` files with real tokens
2. **Non-root User**: The container runs as a non-root user (appuser)
3. **Read-only Volumes**: AIML data is mounted as read-only
4. **Resource Limits**: Container has memory and CPU limits set
5. **Network Isolation**: Uses a dedicated Docker network

## Production Deployment

### Recommendations

1. **Use Secrets Management**:
   ```bash
   # Use Docker secrets instead of environment variables
   echo "your_token" | docker secret create telegram_bot_token -
   ```

2. **Set Resource Limits**:
   ```yaml
   # Adjust limits in docker-compose.yml
   deploy:
     resources:
       limits:
         memory: 1G
         cpus: '1.0'
   ```

3. **Enable Logging**:
   ```yaml
   # Configure log rotation
   logging:
     driver: "json-file"
     options:
       max-size: "50m"
       max-file: "5"
   ```

4. **Use Health Checks**:
   ```yaml
   # Add custom health check endpoint
   healthcheck:
     test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
   ```

## Support

For issues and questions:

1. Check the troubleshooting section above
2. Review the container logs: `docker-compose logs telegram-bot`
3. Check the main project documentation
4. Open an issue on the repository

---

**Happy Bot Building with Docker! 🐳🤖**
