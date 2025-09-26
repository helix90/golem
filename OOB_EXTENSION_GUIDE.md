# Golem OOB (Out-of-Band) Extension Guide

## Overview

Golem now supports Out-of-Band (OOB) messages through a plugin-based architecture that makes it highly extensible for future enhancements.

## Architecture

### Core Components

1. **OOBManager**: Central registry and dispatcher for OOB handlers
2. **OOBHandler Interface**: Contract that all OOB handlers must implement
3. **Built-in Handlers**: System info, session info, and properties handlers
4. **Message Parser**: Supports both `<oob>...</oob>` and `[OOB]...[/OOB]` formats

### Handler Interface

```go
type OOBHandler interface {
    CanHandle(message string) bool
    Process(message string, session *ChatSession) (string, error)
    GetName() string
    GetDescription() string
}
```

## Built-in Handlers

### 1. System Info Handler
- **Name**: `system_info`
- **Triggers**: `SYSTEM INFO`, `SYSTEMINFO`
- **Commands**:
  - `SYSTEM INFO` - General system information
  - `SYSTEM INFO VERSION` - Version information
  - `SYSTEM INFO STATUS` - System status
  - `SYSTEM INFO HANDLERS` - List available handlers

### 2. Session Info Handler
- **Name**: `session_info`
- **Triggers**: `SESSION INFO`, `SESSIONINFO`
- **Commands**:
  - `SESSION INFO` - Basic session information
  - `SESSION INFO DETAILS` - Detailed session information with variables

### 3. Properties Handler
- **Name**: `properties`
- **Triggers**: `PROPERTIES`, `GET PROPERTY`, `SET PROPERTY`
- **Commands**:
  - `PROPERTIES` - List all properties
  - `PROPERTIES GET <key>` - Get specific property
  - `PROPERTIES SET <key> <value>` - Set property value

## Usage Examples

### CLI Commands

```bash
# List OOB handlers
golem oob list

# Test OOB handler
golem oob test SYSTEM INFO

# Register custom handler
golem oob register CUSTOM "Custom handler description"

# Send OOB message in chat
golem chat "<oob>SYSTEM INFO</oob>"
golem chat "[OOB]PROPERTIES GET name[/OOB]"
```

### Interactive Mode

```bash
golem interactive
golem> oob list
golem> oob test SYSTEM INFO VERSION
golem> chat <oob>SESSION INFO</oob>
golem> quit
```

### Library Usage

```go
g := golem.New(false)

// Register custom handler
customHandler := &MyCustomHandler{}
g.oobMgr.RegisterHandler(customHandler)

// Process OOB message
response, err := g.oobMgr.ProcessOOB("CUSTOM MESSAGE", session)
```

## Creating Custom OOB Handlers

### Step 1: Implement the Interface

```go
type MyCustomHandler struct {
    // Add any fields needed for your handler
}

func (h *MyCustomHandler) CanHandle(message string) bool {
    return strings.HasPrefix(strings.ToUpper(message), "MYCOMMAND")
}

func (h *MyCustomHandler) Process(message string, session *ChatSession) (string, error) {
    // Process the OOB message
    return "Custom response", nil
}

func (h *MyCustomHandler) GetName() string {
    return "my_custom"
}

func (h *MyCustomHandler) GetDescription() string {
    return "Handles custom OOB messages"
}
```

### Step 2: Register the Handler

```go
// In your application
g := golem.New(false)
g.oobMgr.RegisterHandler(&MyCustomHandler{})
```

### Step 3: Test the Handler

```bash
golem oob test MYCOMMAND hello world
golem chat "<oob>MYCOMMAND test</oob>"
```

## Advanced Handler Examples

### Database Handler

```go
type DatabaseHandler struct {
    db *sql.DB
}

func (h *DatabaseHandler) CanHandle(message string) bool {
    return strings.HasPrefix(strings.ToUpper(message), "DATABASE")
}

func (h *DatabaseHandler) Process(message string, session *ChatSession) (string, error) {
    parts := strings.Fields(strings.ToUpper(message))
    if len(parts) < 2 {
        return "Usage: DATABASE [QUERY|INSERT|UPDATE] ...", nil
    }
    
    switch parts[1] {
    case "QUERY":
        return h.handleQuery(parts[2:])
    case "INSERT":
        return h.handleInsert(parts[2:])
    default:
        return "Unknown database command", nil
    }
}
```

### File System Handler

```go
type FileSystemHandler struct{}

func (h *FileSystemHandler) CanHandle(message string) bool {
    return strings.HasPrefix(strings.ToUpper(message), "FILE")
}

func (h *FileSystemHandler) Process(message string, session *ChatSession) (string, error) {
    parts := strings.Fields(strings.ToUpper(message))
    if len(parts) < 2 {
        return "Usage: FILE [READ|WRITE|LIST] ...", nil
    }
    
    switch parts[1] {
    case "READ":
        return h.readFile(parts[2:])
    case "WRITE":
        return h.writeFile(parts[2:])
    case "LIST":
        return h.listFiles(parts[2:])
    default:
        return "Unknown file command", nil
    }
}
```

### API Handler

```go
type APIHandler struct {
    client *http.Client
    baseURL string
}

func (h *APIHandler) CanHandle(message string) bool {
    return strings.HasPrefix(strings.ToUpper(message), "API")
}

func (h *APIHandler) Process(message string, session *ChatSession) (string, error) {
    parts := strings.Fields(strings.ToUpper(message))
    if len(parts) < 2 {
        return "Usage: API [GET|POST|PUT|DELETE] <endpoint>", nil
    }
    
    method := parts[1]
    endpoint := parts[2]
    
    // Make HTTP request
    resp, err := h.client.Get(h.baseURL + endpoint)
    if err != nil {
        return fmt.Sprintf("API Error: %v", err), nil
    }
    defer resp.Body.Close()
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return fmt.Sprintf("Read Error: %v", err), nil
    }
    
    return string(body), nil
}
```

## Message Format Support

### XML Style
```
<oob>COMMAND ARG1 ARG2</oob>
```

### Bracket Style
```
[OOB]COMMAND ARG1 ARG2[/OOB]
```

### Case Insensitive
Both formats are case-insensitive and will be converted to uppercase for processing.

## Integration Points

### Chat Command Integration
OOB messages are automatically detected and processed in chat commands before normal AIML pattern matching.

### Session Context
All OOB handlers receive the current chat session, allowing them to:
- Access session variables
- Modify session state
- Read chat history
- Create new sessions

### Error Handling
- Handlers can return errors for proper error reporting
- Unknown OOB messages are handled gracefully
- Verbose logging shows OOB processing steps

## Future Extension Ideas

1. **WebSocket Handler**: Real-time communication
2. **Plugin System**: Dynamic handler loading
3. **Authentication Handler**: User management
4. **Logging Handler**: Advanced logging and monitoring
5. **Configuration Handler**: Runtime configuration changes
6. **Metrics Handler**: Performance monitoring
7. **Notification Handler**: Push notifications
8. **Integration Handlers**: Third-party service integration

## Best Practices

1. **Handler Naming**: Use descriptive, lowercase names with underscores
2. **Error Handling**: Always return meaningful error messages
3. **Session Safety**: Be careful when modifying session state
4. **Performance**: Keep handlers lightweight and fast
5. **Documentation**: Provide clear descriptions for handlers
6. **Testing**: Test handlers thoroughly before deployment
7. **Logging**: Use verbose logging for debugging

## Conclusion

The OOB system provides a powerful, extensible foundation for adding new capabilities to Golem without modifying the core codebase. The plugin-based architecture makes it easy to add new handlers and integrate with external systems.
