# Golem Architecture: State Management Patterns

## Critical Architectural Issue: State Persistence

This document explains the state management patterns in Golem and how to avoid common pitfalls.

## The Problem

Golem maintains state across operations:
- **AIML Knowledge Base**: Loaded patterns and categories
- **Chat Sessions**: Active user sessions with history and variables
- **Bot Properties**: Configuration values

However, the CLI creates a **new Golem instance for each command**, causing state to be lost between commands.

## Three Usage Patterns

### 1. Single-Command Mode (Current CLI Default)
```bash
golem load testdata/sample.aiml    # Creates instance A, loads AIML, exits
golem chat hello                   # Creates instance B (no AIML), fails
```

**Problem**: State is lost between commands.

### 2. Interactive Mode (New Solution)
```bash
golem interactive
golem> load testdata/sample.aiml   # Same instance, state preserved
golem> chat hello                  # Same instance, state preserved
golem> quit
```

**Solution**: Single persistent instance across commands.

### 3. Library Mode (User Controlled)
```go
g := golem.New(false)
g.Execute("load", []string{"testdata/sample.aiml"})
g.Execute("chat", []string{"hello"})  // State preserved
```

**Solution**: User controls instance lifecycle.

## Implementation Details

### State-Bearing Fields
```go
type Golem struct {
    aimlKB    *AIMLKnowledgeBase  // Loaded AIML patterns
    sessions  map[string]*ChatSession  // Active chat sessions
    currentID string              // Currently active session
    // ... other fields
}
```

### Critical Methods
- `Execute()`: Operates on current instance state
- `SetKnowledgeBase()`: Sets AIML knowledge base
- `GetKnowledgeBase()`: Retrieves AIML knowledge base

## Prevention of Regression

### Code Comments
Critical sections are documented with:
```go
// CRITICAL ARCHITECTURAL NOTE:
// This struct maintains state across multiple operations
// DO NOT modify the state management without understanding the implications
```

### Clear Usage Patterns
- Single-command mode: New instance per command (state lost)
- Interactive mode: Persistent instance (state preserved)
- Library mode: User controls instance lifecycle

### Help Documentation
```bash
golem -help
# Shows clear distinction between modes
# Warns about state loss in single-command mode
```

## Best Practices

### For CLI Users
- Use `golem interactive` for persistent state
- Use single commands only for stateless operations
- Load AIML once in interactive mode, then chat

### For Library Users
- Create one Golem instance per application lifecycle
- Share knowledge base between instances if needed
- Use `GetKnowledgeBase()` and `SetKnowledgeBase()` for state sharing

### For Developers
- **NEVER** modify state management without understanding all three patterns
- Always test both single-command and interactive modes
- Document state-bearing operations clearly
- Consider state implications when adding new features

## Testing State Persistence

```bash
# Test single-command mode (should show state loss)
golem load testdata/sample.aiml
golem chat hello  # Should fail

# Test interactive mode (should show state preservation)
golem interactive
golem> load testdata/sample.aiml
golem> chat hello  # Should succeed
golem> quit
```

## Conclusion

The state persistence issue is now properly documented and solved with:
1. **Clear architectural documentation** in code comments
2. **Interactive mode** for persistent CLI usage
3. **Library mode** for user-controlled state management
4. **Comprehensive testing** to prevent regression

This prevents the repeated occurrence of the state loss issue and provides clear guidance for future development.
