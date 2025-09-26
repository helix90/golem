# SRAIX (Substitute, Resubstitute, and Input eXternal) Support

SRAIX allows your AIML bot to communicate with external HTTP/HTTPS services, enabling dynamic responses based on real-time data from APIs, other bots, or web services.

## Overview

SRAIX extends the concept of SRAI (Substitute, Resubstitute, and Input) by allowing your bot to send input to external services and process their responses. This enables your AIML bot to:

- Get real-time weather information
- Translate text using external translation services
- Query AI services like ChatGPT or Claude
- Access any HTTP/HTTPS API
- Integrate with other chatbots or services

## Basic Usage

### AIML Template Syntax

```xml
<sraix service="service_name">Input to send to external service</sraix>
```

### Example

```xml
<category>
  <pattern>WHAT IS THE WEATHER IN *</pattern>
  <template>
    <sraix service="weather_service">What is the weather in <star/>?</sraix>
  </template>
</category>
```

## Configuration

SRAIX services are configured using JSON configuration files. Each service has the following properties:

### Required Properties

- `name`: Unique identifier for the service
- `base_url`: The HTTP/HTTPS endpoint URL

### Optional Properties

- `method`: HTTP method (default: "POST")
- `headers`: Custom headers to include in requests
- `timeout`: Request timeout in seconds (default: 30)
- `response_format`: Response format - "json", "xml", or "text" (default: "text")
- `response_path`: JSON path to extract specific data (e.g., "data.message")
- `fallback_response`: Response when service is unavailable
- `include_wildcards`: Whether to include wildcard data in requests (default: false)

### Example Configuration

```json
[
  {
    "name": "weather_service",
    "base_url": "https://api.weather.com/v1/current",
    "method": "POST",
    "headers": {
      "Authorization": "Bearer YOUR_API_KEY",
      "Content-Type": "application/json"
    },
    "timeout": 10,
    "response_format": "json",
    "response_path": "data.conditions",
    "fallback_response": "Sorry, I couldn't get the weather information right now.",
    "include_wildcards": true
  }
]
```

## Request Format

### POST Requests

For POST requests, the input is sent as JSON:

```json
{
  "input": "What is the weather in New York?",
  "wildcards": {
    "star1": "New York",
    "star2": "today"
  }
}
```

### GET Requests

For GET requests, the input is sent as a query parameter:

```
https://api.example.com/echo?input=What+is+the+weather+in+New+York%3F
```

## Response Processing

### JSON Responses

When `response_format` is "json", the response is parsed and optionally filtered using `response_path`:

```json
{
  "data": {
    "message": "The weather is sunny and 75°F"
  }
}
```

With `response_path: "data.message"`, the result would be: "The weather is sunny and 75°F"

### Text Responses

When `response_format` is "text", the raw response body is used directly.

### XML Responses

When `response_format` is "xml", the raw XML response is used (future enhancement for XML parsing).

## Error Handling

SRAIX includes robust error handling:

1. **Service Not Found**: If the service name doesn't exist, the SRAIX tag is left unchanged
2. **HTTP Errors**: If the service returns an error status (4xx, 5xx), the fallback response is used
3. **Timeout**: If the request times out, the fallback response is used
4. **Network Errors**: If the request fails due to network issues, the fallback response is used

## Wildcard Support

When `include_wildcards` is true, wildcard data from the matched pattern is included in the request:

```xml
<category>
  <pattern>WEATHER IN * AND *</pattern>
  <template>
    <sraix service="weather_service">Weather in <star1/> and <star2/></sraix>
  </template>
</category>
```

This would send:
```json
{
  "input": "Weather in New York and London",
  "wildcards": {
    "star1": "New York",
    "star2": "London"
  }
}
```

## Integration with Other AIML Tags

SRAIX works seamlessly with other AIML tags:

```xml
<category>
  <pattern>COMPLEX REQUEST *</pattern>
  <template>
    <think>
      <set name="user_request"><star/></set>
    </think>
    I'll help you with that. <sraix service="ai_service"><get name="user_request"/></sraix>
  </template>
</category>
```

## Loading Configurations

### From File

```go
g := golem.New(true)
err := g.LoadSRAIXConfigsFromFile("sraix_config.json")
```

### From Directory

```go
err := g.LoadSRAIXConfigsFromDirectory("./configs")
```

### Programmatically

```go
config := &golem.SRAIXConfig{
    Name:           "my_service",
    BaseURL:        "https://api.example.com",
    Method:         "POST",
    ResponseFormat: "json",
    ResponsePath:   "data.message",
}
err := g.AddSRAIXConfig(config)
```

## Security Considerations

1. **API Keys**: Store API keys securely and never commit them to version control
2. **Input Validation**: External services should validate input to prevent injection attacks
3. **Rate Limiting**: Be aware of API rate limits and implement appropriate delays
4. **HTTPS**: Always use HTTPS for external service communication
5. **Timeout**: Set appropriate timeouts to prevent hanging requests

## Best Practices

1. **Fallback Responses**: Always provide meaningful fallback responses
2. **Error Logging**: Monitor SRAIX requests for failures and errors
3. **Caching**: Consider caching responses for frequently requested data
4. **Testing**: Test SRAIX configurations thoroughly before deployment
5. **Documentation**: Document your SRAIX services and their expected inputs/outputs

## Example Services

### Weather Service
```json
{
  "name": "weather",
  "base_url": "https://api.openweathermap.org/data/2.5/weather",
  "method": "GET",
  "response_format": "json",
  "response_path": "weather.0.description"
}
```

### Translation Service
```json
{
  "name": "translate",
  "base_url": "https://api.mymemory.translated.net/get",
  "method": "GET",
  "response_format": "json",
  "response_path": "responseData.translatedText"
}
```

### AI Chat Service
```json
{
  "name": "ai_chat",
  "base_url": "https://api.openai.com/v1/chat/completions",
  "method": "POST",
  "headers": {
    "Authorization": "Bearer YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  "response_format": "json",
  "response_path": "choices.0.message.content"
}
```

## Troubleshooting

### Common Issues

1. **Service Not Found**: Check that the service name in the SRAIX tag matches the configuration
2. **Authentication Errors**: Verify API keys and headers are correct
3. **Timeout Errors**: Increase the timeout value or check network connectivity
4. **JSON Path Errors**: Verify the response_path matches the actual JSON structure
5. **CORS Issues**: Some services may not allow requests from certain origins

### Debug Mode

Enable verbose logging to see SRAIX request details:

```go
g := golem.New(true) // Enable verbose mode
```

This will log:
- SRAIX requests being made
- Response data received
- Error conditions encountered
