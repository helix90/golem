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

## Thread Safety

One `Golem` instance can serve many goroutines at once, as long as you
finish setting it up before you start serving.

### Setup: one goroutine, before serving

These calls change the engine itself and are **not** guarded against
concurrent chat. Make them before any goroutine starts chatting:

- Loading: `LoadAIML*`, `LoadFile`, `Execute("load", ...)`, `LoadSRAIXConfigs*`, `AddSRAIXConfig`
- Configuration: `SetKnowledgeBase`, `SetLogLevel`, `EnableTreeProcessing`, `DisableTreeProcessing`, `UpdateTemplateProcessingConfig`, `SetPersistentLearningPath`

### Serving: safe from any goroutine

- Chat: `ProcessInput`, `ProcessInputWithThatIndex`, `ProcessTemplate`, `ProcessTemplateWithSession`, `ProcessTemplateWithContext`
- Sessions: `CreateSession`, `GetSessionLearningStats`, `GetSessionLearnedCategories`, `ClearSessionLearning`, `GetLearningSummary`
- Knowledge base data: `GetProperty`, `SetProperty`
- Caches: the `Get...Stats`, `Clear...` and `Invalidate...` methods

Template tags that change shared data (`<set>` on global or topic
variables, `<map>`, `<list>`, `<array>`, set collections, `<learn>`,
`<unlearn>`) are safe while other sessions are matching and processing.
Their changes are shared, so every session sees them.

### Sessions

- **Different sessions run in parallel.** This is the normal pattern: one
  `ChatSession` per user or conversation.
- **Calls on the same session run one at a time.** Each `ChatSession` has a
  lock that the chat methods hold for the whole call. Two goroutines sending
  input to one session won't corrupt it, but the second waits for the first,
  and the order they run in is not defined. If order matters (it usually
  does for `<that>`), send a conversation's input from one goroutine.
- **Don't read session fields directly while a call on that session is in
  progress.** Fields such as `session.Variables` and `session.History` are
  only safe to read from the goroutine that makes the session's calls, between
  calls. From other goroutines, use the Golem methods above.
- **Stats methods return copies.** `GetSessionLearningStats`,
  `GetSessionLearnedCategories` and the cache stats methods return snapshots
  that don't change after they are returned. Changing them doesn't change the
  session or cache.

### Lock rules (for developers)

- `ChatSession.mu`: held for a whole chat call. Exported chat methods take it;
  internal code that already holds it calls the unexported variants
  (`processTemplateForSession`, `processTemplateWithSession`). Never call an
  exported chat method from template processing — the lock is not reentrant.
- `Golem.sessionMutex`: guards the session map, `currentID` and the session ID
  counter. Never wait for a session's lock while holding it: copy the sessions
  with `sessionSnapshot()`, release it, then lock each session.
- `AIMLKnowledgeBase.mu`: guards `Variables`, `TopicVars`, `Properties`, `Maps`,
  `Lists`, `Arrays` and `SetCollections`. Use the locked helpers
  (`lookupVariable`, `setVariable`, `property`, `GetProperty`, `SetProperty`)
  rather than reading the maps directly. Sections that hold it must not call
  back into variable resolution or the caches.
- `AIMLKnowledgeBase.patternsMu`: guards `Categories` and `Patterns`; matching
  takes the read lock, `<learn>`/`<unlearn>` the write lock.
- Each cache has its own lock. Compute anything that reads the session or knowledge
  base (cache keys, scope hashes) *before* taking a cache lock.
- Order, when more than one is held: `ChatSession.mu` → `patternsMu` /
  `AIMLKnowledgeBase.mu` → cache locks.
- Run `make test-race` after touching any of this. The concurrency
  regression tests are in `pkg/golem/aiml_parsing_regression_test.go`.

### Not covered

- The legacy regex processors in `aiml_native.go` (`process*TagsWithContext`)
  write knowledge-base collections without the lock. The tree processor, which
  is the default, doesn't use them.
- `ProcessTemplateWithSession` and `ProcessTemplateWithContext` create an
  empty knowledge base when none is loaded. That isn't safe across
  goroutines, so load one before serving.

## Best Practices

### For CLI Users
- Use `golem interactive` for persistent state
- Use single commands only for stateless operations
- Load AIML once in interactive mode, then chat

### For Library Users
- Create one Golem instance per application lifecycle
- Share knowledge base between instances if needed
- Use `GetKnowledgeBase()` and `SetKnowledgeBase()` for state sharing
- Load and configure before serving, then share the instance across goroutines (see [Thread Safety](#thread-safety))

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
