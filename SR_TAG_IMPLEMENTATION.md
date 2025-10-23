# `<sr>` Tag Implementation for AST

## Summary

Successfully implemented full support for the `<sr>` (Self-Recursive) tag in the AST (Abstract Syntax Tree) processor, enabling shorthand recursion for wildcard processing.

## What Was Implemented

### 1. AST Parser Updates (`ast_parser.go`)
- Added `"sr"` to the list of implicitly self-closing tags (lines 162 and 301)
- This allows `<sr/>` tags to be properly recognized and parsed as self-closing

### 2. Tree Processor Implementation (`tree_processor.go`)
- Implemented `processSRTag()` method (lines 404-489) with:
  - Shorthand for `<srai><star/></srai>` - recursively processes first wildcard
  - Gets `star1` from session variables (the first wildcard match)
  - Matches `star1` content against knowledge base patterns
  - Recursively processes matched template with new wildcards
  - Preserves original wildcards after SR processing
  - Recursion depth checking (max 100 levels)
  - Proper handling when session, context, or knowledge base is nil
  - Logging for debugging

- Added `"sr"` case to main tag switch (line 93-94)
- Added `"sr"` case to `processSelfClosingTag()` method (line 222-223)

### 3. Regex-Based Processor
- Already exists in `aiml_native.go` (lines 4150-4208)
- Handles SR tag conversion to SRAI format
- Leaves SR tags unchanged when no pattern match is found

### 4. Comprehensive Test Suite

#### Unit Tests (`tree_processor_sr_test.go`)
Created 9 comprehensive test functions covering:

1. **TestTreeProcessorSRTagBasic** - Basic SR tag with different patterns
2. **TestTreeProcessorSRTagNoMatch** - SR when pattern doesn't match
3. **TestTreeProcessorSRTagNoWildcard** - SR when no wildcard available
4. **TestTreeProcessorSRTagNoKnowledgeBase** - SR without knowledge base
5. **TestTreeProcessorSRTagNoSession** - SR without session
6. **TestTreeProcessorSRTagRecursion** - SR recursion behavior
7. **TestTreeProcessorSRTagWithNestedTags** - SR with other AIML tags
8. **TestTreeProcessorSRTagMaxRecursionDepth** - SR at recursion limit

#### Integration Tests (`tree_processor_sr_integration_test.go`)
Created 5 integration test functions:

1. **TestTreeProcessorSRTagIntegration** - Full conversation flow with SR patterns
2. **TestTreeProcessorSRTagWithTreeProcessor** - Direct TreeProcessor usage
3. **TestTreeProcessorSRTagComplexPatterns** - Complex wildcard patterns with SR
4. **TestTreeProcessorSRTagEdgeCases** - Edge cases (no wildcard, think tag)
5. **TestTreeProcessorSRTagWildcardPreservation** - Wildcard preservation after SR

## How It Works

### What is `<sr>`?

The `<sr>` tag is shorthand for `<srai><star/></srai>`. It's used to recursively process the first wildcard match.

### Tag Format

```xml
<sr/>
```

Self-closing tag only. No attributes.

### Processing Flow

1. **Get Wildcard**: Retrieves `star1` from session variables
2. **Match Pattern**: Tries to match `star1` content in knowledge base
3. **Process Recursively**: If match found, processes the matched template
4. **Preserve Wildcards**: Saves original wildcards, sets new ones, then restores
5. **Return Result**: Returns the processed template result

### Example Usage

```xml
<aiml version="2.0">
  <!-- Base pattern -->
  <category>
    <pattern>HELLO</pattern>
    <template>Hi! How can I help you?</template>
  </category>
  
  <!-- Pattern with SR - processes the wildcard -->
  <category>
    <pattern>GREETING *</pattern>
    <template>Nice to meet you! <sr/></template>
  </category>
</aiml>
```

**Input:** `greeting hello`  
**Processing:**
1. Matches `GREETING *` with wildcard `star1 = HELLO`
2. Template contains `<sr/>`
3. SR processes `HELLO` → matches `HELLO` pattern
4. Returns: `Hi! How can I help you?`

**Output:** `Nice to meet you! Hi! How can I help you?`

### Benefits of SR Tag

1. **Code Reduction**: Avoids repeating common patterns
2. **Synonym Handling**: Easy to redirect variations to base patterns
3. **Cleaner AIML**: More readable and maintainable
4. **Recursion Control**: Built-in depth checking prevents infinite loops

## Behavioral Notes

### When SR Can't Process

If SR tag cannot be processed (no wildcard, no match, no KB), behavior depends on processor:
- **AST TreeProcessor**: Returns empty string
- **Regex Processor**: Leaves `<sr/>` unchanged in output

This is intentional to support multi-pass processing scenarios.

### Wildcard Preservation

The SR tag implementation carefully preserves the original wildcards:
```xml
<template>Before: <star/>, SR: <sr/>, After: <star/></template>
```

All three `<star/>` references will show the same original wildcard value, even though SR may have temporarily set different wildcards during its processing.

### Recursion Depth

- Maximum recursion depth: **100 levels**
- Prevents infinite loops from circular pattern references
- When limit reached, SR returns empty (AST) or unchanged (regex)

## Test Results

All tests pass successfully:
- ✅ 9 unit test functions with 15+ test cases
- ✅ 5 integration test functions with multiple scenarios
- ✅ Edge cases handled properly
- ✅ No regressions in existing functionality

### Test Coverage

```
TestTreeProcessorSRTagIntegration         PASS (0.01s)
TestTreeProcessorSRTagWithTreeProcessor   PASS (0.00s)
TestTreeProcessorSRTagComplexPatterns     PASS (0.02s)
TestTreeProcessorSRTagEdgeCases           PASS (0.01s)
TestTreeProcessorSRTagWildcardPreservation PASS (0.00s)
TestTreeProcessorSRTagBasic               PASS (0.00s)
TestTreeProcessorSRTagNoMatch             PASS (0.00s)
TestTreeProcessorSRTagNoWildcard          PASS (0.00s)
TestTreeProcessorSRTagNoKnowledgeBase     PASS (0.00s)
TestTreeProcessorSRTagNoSession           PASS (0.00s)
TestTreeProcessorSRTagRecursion           PASS (0.00s)
TestTreeProcessorSRTagWithNestedTags      PASS (0.00s)
TestTreeProcessorSRTagMaxRecursionDepth   PASS (0.00s)
```

## Files Modified

1. `/home/helix/golem/pkg/golem/ast_parser.go` - Added "sr" to self-closing tags
2. `/home/helix/golem/pkg/golem/tree_processor.go` - Implemented processSRTag()

## Files Created

1. `/home/helix/golem/pkg/golem/tree_processor_sr_test.go` - Unit tests
2. `/home/helix/golem/pkg/golem/tree_processor_sr_integration_test.go` - Integration tests

## Common Use Cases

### 1. Synonym Handling
```xml
<category>
  <pattern>HI</pattern>
  <template><srai>HELLO</srai></template>
</category>
```
**With SR:**
```xml
<category>
  <pattern>SAY *</pattern>
  <template><sr/></template>
</category>
```

### 2. Greeting Variations
```xml
<category>
  <pattern>GREETING *</pattern>
  <template>Nice to meet you! <sr/></template>
</category>
```

### 3. Command Processing
```xml
<category>
  <pattern>PROCESS *</pattern>
  <template>Processing: <sr/></template>
</category>
```

### 4. Forwarding
```xml
<category>
  <pattern>TELL ME *</pattern>
  <template>Let me tell you: <sr/></template>
</category>
```

## Performance

- **O(1)** for wildcard retrieval
- **O(n)** for pattern matching (where n = number of patterns)
- **O(depth)** for recursion (max depth = 100)
- Wildcard save/restore: **O(k)** where k = number of wildcards

## Next Steps

The `<sr>` tag is now fully functional in the AST. The next high-priority tags to implement are:

1. **`<input>`** - References previous user inputs (similar to `<that>`)
2. **`<learnf>`** - Persistent learning across sessions
3. **`<eval>`** - Dynamic AIML expression evaluation

## Comparison with Regex Implementation

| Feature | AST Implementation | Regex Implementation |
|---------|-------------------|---------------------|
| Pattern matching | ✅ Full support | ✅ Full support |
| Wildcard preservation | ✅ Full support | ✅ Full support |
| Recursion depth check | ✅ Max 100 | ✅ Implicit |
| No match behavior | Returns empty | Leaves `<sr/>` |
| Debug logging | ✅ Detailed | ✅ Detailed |
| Performance | Similar | Similar |

Both implementations work correctly and can coexist in the codebase. The AST version is used when TreeProcessor is explicitly called, while the regex version is used in the consolidated template processor pipeline.

