# 🔐 Secure API Key Management for SRAIX

This guide shows you how to securely handle API keys for SRAIX configurations without committing them to git.

## 🚨 Security Best Practices

1. **Never commit API keys to git**
2. **Use environment variables when possible**
3. **Keep secrets files in `.gitignore`**
4. **Use different keys for development/production**
5. **Rotate keys regularly**

## 📋 Available Methods

### Method 1: Environment Variables (Recommended)

Set your API keys as environment variables:

```bash
# Set environment variables
export OPENAI_API_KEY="sk-your-openai-key-here"
export WEATHER_API_KEY="your-weather-key-here"

# Or create a .env file (add to .gitignore!)
echo "OPENAI_API_KEY=sk-your-openai-key-here" > .env
echo "WEATHER_API_KEY=your-weather-key-here" >> .env
```

Then use them in your code:

```go
config := &golem.SRAIXConfig{
    Name:    "openai_service",
    BaseURL: "https://api.openai.com/v1/chat/completions",
    Headers: map[string]string{
        "Authorization": "Bearer " + os.Getenv("OPENAI_API_KEY"),
    },
}
```

### Method 2: Template Configuration with Environment Variable Substitution

Create a template configuration file (`sraix_config_template.json`):

```json
[
  {
    "name": "openai_service",
    "base_url": "https://api.openai.com/v1/chat/completions",
    "headers": {
      "Authorization": "Bearer ${OPENAI_API_KEY}",
      "Content-Type": "application/json"
    }
  }
]
```

Load it using the helper function:

```go
err := LoadSRAIXConfigsWithEnvVars(g, "sraix_config_template.json")
```

### Method 3: Secrets File (Not in Git)

Create a `secrets.json` file (add to `.gitignore`):

```json
{
  "openai_api_key": "sk-your-actual-key-here",
  "weather_api_key": "your-weather-key-here"
}
```

Load it using the helper function:

```go
err := LoadSRAIXConfigsFromSecrets(g, "secrets.json")
```

### Method 4: Docker Secrets (For Containerized Deployments)

Create a `docker-compose.yml`:

```yaml
version: '3.8'
services:
  golem:
    build: .
    environment:
      - OPENAI_API_KEY_FILE=/run/secrets/openai_key
    secrets:
      - openai_key

secrets:
  openai_key:
    file: ./secrets/openai_key.txt
```

## 🛠️ Setup Instructions

### 1. Copy the example files

```bash
# Copy the template configuration
cp sraix_config_template.json my_sraix_config.json

# Copy the secrets example
cp secrets.json.example secrets.json
```

### 2. Add to .gitignore

Make sure these files are in your `.gitignore`:

```gitignore
# API Keys and Secrets
secrets.json
*.key
*.pem
.env
.env.local
.env.production

# Configuration files with sensitive data
*_config.json
*_secrets.json
```

### 3. Set up your API keys

Choose one of these methods:

#### Option A: Environment Variables
```bash
export OPENAI_API_KEY="sk-your-key-here"
export WEATHER_API_KEY="your-weather-key"
```

#### Option B: Secrets File
Edit `secrets.json` with your actual API keys.

#### Option C: Template Configuration
Edit `sraix_config_template.json` and use environment variable substitution.

### 4. Run the demo

```bash
go run secure_sraix_demo.go
```

## 🔧 Helper Functions

The `sraix_config_loader.go` file provides helper functions:

- `LoadSRAIXConfigsWithEnvVars()` - Load from template with env var substitution
- `LoadSRAIXConfigsFromSecrets()` - Load from secrets file
- `substituteEnvVars()` - Replace `${VAR_NAME}` with environment variables

## 🚀 Production Deployment

### Environment Variables
```bash
# Set production environment variables
export OPENAI_API_KEY="sk-prod-key-here"
export WEATHER_API_KEY="prod-weather-key"
```

### Docker with Secrets
```bash
# Create secrets directory
mkdir -p secrets

# Add your keys to files
echo "sk-prod-key-here" > secrets/openai_key.txt
echo "prod-weather-key" > secrets/weather_key.txt

# Run with docker-compose
docker-compose up
```

### Kubernetes Secrets
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: golem-secrets
type: Opaque
data:
  openai-key: c2stcHJvZC1rZXktaGVyZQ==  # base64 encoded
  weather-key: cHJvZC13ZWF0aGVyLWtleQ==
```

## ⚠️ Security Warnings

1. **Never log API keys** - The helper functions avoid printing headers
2. **Use HTTPS only** - All external service URLs should use HTTPS
3. **Validate inputs** - Sanitize user inputs before sending to external services
4. **Monitor usage** - Keep track of API usage and costs
5. **Rotate keys** - Change API keys regularly
6. **Use least privilege** - Only give necessary permissions to API keys

## 🧪 Testing

Test your configuration:

```bash
# Test with environment variables
OPENAI_API_KEY="test-key" go run secure_sraix_demo.go

# Test with secrets file
go run secure_sraix_demo.go
```

## 📝 Example Services

### OpenAI ChatGPT
```json
{
  "name": "openai_service",
  "base_url": "https://api.openai.com/v1/chat/completions",
  "method": "POST",
  "headers": {
    "Authorization": "Bearer ${OPENAI_API_KEY}",
    "Content-Type": "application/json"
  },
  "response_format": "json",
  "response_path": "choices.0.message.content"
}
```

### Weather Service
```json
{
  "name": "weather_service",
  "base_url": "https://api.openweathermap.org/data/2.5/weather",
  "method": "GET",
  "response_format": "json",
  "response_path": "weather.0.description"
}
```

### Translation Service
```json
{
  "name": "translation_service",
  "base_url": "https://api.mymemory.translated.net/get",
  "method": "GET",
  "response_format": "json",
  "response_path": "responseData.translatedText"
}
```

## 🆘 Troubleshooting

### Common Issues

1. **"Environment variable not set"** - Make sure you've exported the variable
2. **"Service not found"** - Check the service name in your AIML templates
3. **"Authentication failed"** - Verify your API key is correct
4. **"Timeout"** - Increase the timeout value in your configuration

### Debug Mode

Enable verbose logging to see what's happening:

```go
g := golem.New(true) // Enable verbose mode
```

This will show:
- SRAIX requests being made
- Response data received
- Error conditions encountered

## 📚 Additional Resources

- [OpenAI API Documentation](https://platform.openai.com/docs)
- [OpenWeatherMap API](https://openweathermap.org/api)
- [Docker Secrets](https://docs.docker.com/engine/swarm/secrets/)
- [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)
