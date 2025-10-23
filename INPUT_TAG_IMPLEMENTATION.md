# Input Tag Implementation

## Overview

This document describes the implementation of the `<input/>` tag in the AIML AST-based template processor. The `<input/>` tag returns the most recent user input from the conversation history.

## Tag Syntax

```xml
<input/>
```

The `<input/>` tag is a **self-closing** tag with **no attributes**.

## Behavior

The `<input/>` tag returns the most recent user input stored in the `RequestHistory`. This is equivalent to `<request index="1"/>`.

### Important Note on Timing

When using `ProcessInput()` to process a conversation:
- The **current** input being processed has not yet been added to `RequestHistory`
- Therefore, `<input/>` returns the **previous** user input during template processing
- The current input is added to `RequestHistory` **after** the template is fully processed

This means:
```
User: "hello"
Bot processes: <input/> → "" (no previous input)
Bot adds "hello" to RequestHistory

User: "how are you"
Bot processes: <input/> → "hello" (previous input)
Bot adds "how are you" to RequestHistory
```

### Comparison with Request Tag

| Tag | Behavior |
|-----|----------|
| `<input/>` | Always returns the most recent user input from `RequestHistory` (last item) |
| `<request index="1"/>` | Same as `<input/>` - returns most recent |
| `<request index="2"/>` | Returns 2nd most recent user input |
| `<request index="N"/>` | Returns Nth most recent user input |

## Implementation Details

### File: `tree_processor.go`

The `processInputTag()` method implements the `<input/>` tag processing:

```go
func (tp *TreeProcessor) processInputTag(node *ASTNode, content string) string {
	// Process input tag - returns the most recent user input
	// <input/> always returns the current/most recent user input (last item in RequestHistory)
	// This is different from <request> which can take an index attribute
	
	if tp.ctx == nil || tp.ctx.Session == nil {
		tp.golem.LogDebug("Input tag: no context or session available")
		return ""
	}
	
	// Get the most recent user input from request history
	if len(tp.ctx.Session.RequestHistory) == 0 {
		tp.golem.LogDebug("Input tag: no request history available")
		return ""
	}
	
	// Return the last (most recent) item from RequestHistory
	currentInput := tp.ctx.Session.RequestHistory[len(tp.ctx.Session.RequestHistory)-1]
	
	tp.golem.LogDebug("Input tag: returning '%s'", currentInput)
	
	return currentInput
}
```

### File: `ast_parser.go`

The `<input>` tag is registered as an implicitly self-closing tag:

```go
implicitlySelfClosing := map[string]bool{
	// ... other tags ...
	"input":    true,
	// ... other tags ...
}
```

### Switch Statement

The `input` case is handled in both `processTag()` and `processSelfClosingTag()`:

```go
case "input":
	return tp.processInputTag(node, content)
```

## Edge Cases

### 1. Empty Request History

If `RequestHistory` is empty:
```xml
Template: You said: <input/>
Result:   You said: 
```

### 2. No Session or Context

If the context or session is nil:
```xml
Template: You said: <input/>
Result:   You said: 
```

### 3. Special Characters

The `<input/>` tag preserves all special characters, whitespace, and Unicode:
```xml
Input:    "Hello, world! 世界 🌍"
Template: <input/>
Result:   Hello, world! 世界 🌍
```

### 4. Nested Tags

The `<input/>` tag can be nested inside other tags:
```xml
Template: <uppercase><input/></uppercase>
Input:    hello world
Result:   HELLO WORLD
```

## Examples

### Example 1: Echo the Previous Input

```xml
<category>
	<pattern>ECHO</pattern>
	<template>You previously said: <input/></template>
</category>
```

Conversation:
```
User: hello
Bot:  You previously said:          (no previous input)

User: echo
Bot:  You previously said: hello    (echoes previous input)
```

### Example 2: Save Input to Variable

```xml
<category>
	<pattern>REMEMBER *</pattern>
	<template>
		<think><set name="saved"><input/></set></think>
		I'll remember that you said: <input/>
	</template>
</category>
```

Conversation:
```
User: how are you
Bot:  I'll remember that you said: how are you

User: remember this
Bot:  I'll remember that you said: how are you
      (Variable "saved" now contains "how are you")
```

### Example 3: Format Input

```xml
<category>
	<pattern>FORMAT</pattern>
	<template>
		Original: <input/>
		Uppercase: <uppercase><input/></uppercase>
		Formal: <formal><input/></formal>
	</template>
</category>
```

Conversation:
```
User: hello world
Bot:  Original: hello world
      Uppercase: HELLO WORLD
      Formal: Hello World
```

### Example 4: Compare with Request Tag

```xml
<category>
	<pattern>COMPARE</pattern>
	<template>
		Input: <input/>
		Request 1: <request index="1"/>
		Request 2: <request index="2"/>
		Request 3: <request index="3"/>
	</template>
</category>
```

Conversation (with history ["first", "second", "third"]):
```
User: compare
Bot:  Input: third
      Request 1: third
      Request 2: second
      Request 3: first
```

## Testing

### Unit Tests (`tree_processor_input_test.go`)

Comprehensive unit tests cover:
- Basic input tag processing
- Edge cases (empty history, no session, no context)
- Special characters and Unicode
- Nested structures
- Comparison with request tag
- Maximum history size

### Integration Tests (`tree_processor_input_integration_test.go`)

Integration tests cover:
- Full conversation flow with ProcessInput
- Interaction with variables
- Nested tags and formatting
- Empty history scenarios
- Comparison between input and request tags

## Related Tags

- `<request index="N"/>` - Returns the Nth most recent user input
- `<response index="N"/>` - Returns the Nth most recent bot response
- `<that/>` - Returns the most recent bot response

## AIML Specification Reference

According to the AIML specification:
- `<input/>` returns the **current** user input being processed
- However, in the context of `ProcessInput()`, the "current" input hasn't been added to history yet
- Therefore, `<input/>` returns the last item in `RequestHistory`, which is the **previous** input

This implementation aligns with the AIML standard while accounting for the timing of when inputs are added to the request history.

## Bug Fixes and Improvements

### Request Tag Index Fix

During implementation, a bug was discovered and fixed in the `processRequestTag()` method:
- **Before**: Used direct array indexing: `RequestHistory[index-1]`
- **After**: Uses `GetRequestByIndex(index)` which properly handles reverse indexing
- This ensures that `<request index="1"/>` correctly returns the most recent request

## Performance Considerations

The `<input/>` tag is highly efficient:
- O(1) time complexity (direct array access)
- No regex matching required
- No string manipulation
- Minimal memory overhead

## Debugging

When debugging `<input/>` tag issues:

1. **Check Request History**: Verify that `RequestHistory` is being populated correctly
2. **Check Timing**: Remember that `<input/>` returns the **previous** input during `ProcessInput()`
3. **Check Context**: Ensure the session and context are not nil
4. **Enable Logging**: Use debug logging to see what value `<input/>` returns

Example debug output:
```
[DEBUG] Input tag: returning 'hello world'
```

## Future Enhancements

Potential future enhancements:
1. **Index Attribute**: Consider adding `<input index="N"/>` support (though `<request>` already provides this)
2. **Sentence Attribute**: Add support for `<input index="N,M"/>` to get specific sentences
3. **Transform Attribute**: Add built-in transformations like `<input transform="uppercase"/>`

However, these features may be unnecessary given that:
- `<request>` already provides indexed access
- Transformations can be achieved with nested tags like `<uppercase><input/></uppercase>`

## Conclusion

The `<input/>` tag implementation provides a simple, efficient way to reference the most recent user input in AIML templates. It integrates seamlessly with the AST-based template processor and maintains compatibility with the existing regex-based processor.

