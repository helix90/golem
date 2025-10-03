# Logging System Replacement Guide

## Overview

The Golem library now supports structured logging with levels instead of the simple verbose flag. This guide shows how to replace the old verbose logging pattern with the new level-based system.

## New Logging System

### Log Levels
- `LogLevelError` (0): Error messages
- `LogLevelWarn` (1): Warning messages  
- `LogLevelInfo` (2): Informational messages
- `LogLevelDebug` (3): Debug messages (replaces most verbose logging)
- `LogLevelTrace` (4): Very detailed trace messages

### Available Methods
- `g.LogError(format, args...)` - Error messages
- `g.LogWarn(format, args...)` - Warning messages
- `g.LogInfo(format, args...)` - Informational messages
- `g.LogDebug(format, args...)` - Debug messages
- `g.LogTrace(format, args...)` - Trace messages
- `g.LogVerbose(format, args...)` - Backward compatibility (maps to LogDebug when verbose=true)

### Configuration
- `g.SetLogLevel(level)` - Set the logging level
- `g.GetLogLevel()` - Get the current logging level

## Replacement Patterns

### Pattern 1: Simple verbose logging
**OLD:**
```go
if g.verbose {
    g.logger.Printf("Loading AIML from string")
}
```

**NEW:**
```go
g.LogDebug("Loading AIML from string")
```

### Pattern 2: Multiple verbose logs
**OLD:**
```go
if g.verbose {
    g.logger.Printf("Loaded AIML from string successfully")
    g.logger.Printf("Total categories: %d", len(g.aimlKB.Categories))
    g.logger.Printf("Total patterns: %d", len(g.aimlKB.Patterns))
}
```

**NEW:**
```go
g.LogDebug("Loaded AIML from string successfully")
g.LogDebug("Total categories: %d", len(g.aimlKB.Categories))
g.LogDebug("Total patterns: %d", len(g.aimlKB.Patterns))
```

### Pattern 3: Error logging
**OLD:**
```go
if g.verbose {
    g.logger.Printf("Failed to parse learnf content: %v", err)
}
```

**NEW:**
```go
g.LogError("Failed to parse learnf content: %v", err)
```

### Pattern 4: Conditional verbose logging
**OLD:**
```go
if g.verbose {
    g.logger.Printf("Processing learnf: '%s'", learnfContent)
}
```

**NEW:**
```go
g.LogDebug("Processing learnf: '%s'", learnfContent)
```

## Migration Strategy

1. **Replace simple verbose patterns** with `g.LogDebug()`
2. **Replace error logging** with `g.LogError()`
3. **Replace warning messages** with `g.LogWarn()`
4. **Replace informational messages** with `g.LogInfo()`
5. **Use `g.LogTrace()`** for very detailed debugging

## Benefits

- **Structured logging**: Clear separation of log levels
- **Better filtering**: Can filter logs by level
- **Consistent format**: All logs have level prefixes
- **Backward compatibility**: `LogVerbose()` maintains old behavior
- **Performance**: No string formatting when level is disabled

## Example Usage

```go
// Set log level
g.SetLogLevel(LogLevelDebug)

// Log messages
g.LogError("Critical error occurred: %v", err)
g.LogWarn("Deprecated function used")
g.LogInfo("Application started")
g.LogDebug("Processing user input: %s", input)
g.LogTrace("Detailed step-by-step processing")
```

## Search and Replace Commands

To help with migration, here are some regex patterns for search and replace:

1. **Simple verbose logging:**
   - Find: `if g\.verbose \{\s*g\.logger\.Printf\("([^"]+)"\)\s*\}`
   - Replace: `g.LogDebug("$1")`

2. **Verbose logging with args:**
   - Find: `if g\.verbose \{\s*g\.logger\.Printf\("([^"]+)"([^}]+)\)\s*\}`
   - Replace: `g.LogDebug("$1"$2)`

3. **Error logging:**
   - Find: `if g\.verbose \{\s*g\.logger\.Printf\("([^"]*[Ee]rror[^"]*)"([^}]+)\)\s*\}`
   - Replace: `g.LogError("$1"$2)`

