# `<condition>` Tag Implementation Plan

## Overview

The `<condition>` tag is one of the most complex AIML2 features, allowing for conditional responses based on variable values. It supports multiple forms and can be nested.

## Implementation Approach

### 1. Core Functions to Add

```go
// In aiml_native.go and golem.go
func (g *Golem) processConditionTags(template string, session *ChatSession) string
func (g *Golem) processConditionContent(content string, varName, actualValue, expectedValue string, session *ChatSession) string
func (g *Golem) processConditionListItems(content string, actualValue string, session *ChatSession) string
func (g *Golem) getVariableValue(varName string, session *ChatSession) string
```

### 2. Processing Order

The condition processing should happen in this order:
1. Wildcard replacement (`<star/>`, `<star index="N"/>`)
2. Property tag replacement (`<get name="property"/>`)
3. Session variable replacement (`<get name="var"/>`)
4. SRAI processing (`<srai>...</srai>`)
5. Think tag processing (`<think>...</think>`)
6. **Condition tag processing** (`<condition>...</condition>`) - **NEW**
7. Date/time tag processing (`<date>...</date>`, `<time>...</time>`)
8. Random tag processing (`<random>...</random>`)

### 3. Condition Types Supported

#### Type 1: Simple Condition with Value
```xml
<condition name="mood" value="happy">I'm glad you're happy!</condition>
```
- If `mood` variable equals "happy", return the content
- Otherwise, return empty string

#### Type 2: Multiple Conditions with List Items
```xml
<condition name="weather">
    <li value="sunny">It's a beautiful sunny day!</li>
    <li value="rainy">Don't forget your umbrella!</li>
    <li value="snowy">Be careful on the roads!</li>
    <li>I hope you have a great day!</li>
</condition>
```
- Check each `<li>` element for matching value
- Return content of first matching condition
- If no match and there's a `<li>` without value, return that as default

#### Type 3: Default Condition
```xml
<condition name="name">Hello <get name="name"/>!</condition>
```
- If `name` variable exists (non-empty), return the content
- Otherwise, return empty string

#### Type 4: Nested Conditions
```xml
<condition name="user_type">
    <li value="admin">Welcome admin! 
        <condition name="time_of_day">
            <li value="morning">Good morning!</li>
            <li value="afternoon">Good afternoon!</li>
            <li>Good day!</li>
        </condition>
    </li>
    <li value="user">Hello user!</li>
    <li>Welcome guest!</li>
</condition>
```
- Support nested condition processing
- Process inner conditions after outer conditions match

### 4. Variable Lookup Priority

The `getVariableValue` function should check in this order:
1. **Session variables** (`session.Variables[varName]`)
2. **Knowledge base variables** (`aimlKB.Variables[varName]`)
3. **Properties** (`aimlKB.Properties[varName]`)
4. **Return empty string** if not found

### 5. Regex Patterns

```go
// Main condition tag regex
conditionRegex := regexp.MustCompile(`(?s)<condition(?: name="([^"]+)"(?: value="([^"]+)")?)?>(.*?)</condition>`)

// List item regex
liRegex := regexp.MustCompile(`<li(?: value="([^"]+)")?>(.*?)</li>`)
```

### 6. Integration Points

#### In ProcessTemplate (aiml_native.go):
```go
// Process condition tags
response = g.processConditionTags(response, nil)
```

#### In ProcessTemplateWithSession (golem.go):
```go
// Process condition tags
response = g.processConditionTags(response, session)
```

### 7. Test Cases Needed

```go
func TestProcessConditionTags(t *testing.T)
func TestProcessConditionContent(t *testing.T)
func TestProcessConditionListItems(t *testing.T)
func TestGetVariableValue(t *testing.T)
func TestConditionWithSession(t *testing.T)
func TestConditionWithProperties(t *testing.T)
func TestConditionWithWildcards(t *testing.T)
func TestConditionWithSRAI(t *testing.T)
func TestConditionWithThink(t *testing.T)
func TestConditionWithRandom(t *testing.T)
func TestNestedConditions(t *testing.T)
func TestConditionDefaultCase(t *testing.T)
```

### 8. Example Usage Scenarios

#### Weather Bot
```xml
<category>
    <pattern>WHAT IS THE WEATHER</pattern>
    <template>
        <condition name="weather">
            <li value="sunny">It's a beautiful sunny day! Perfect for outdoor activities.</li>
            <li value="rainy">It's raining today. Don't forget your umbrella!</li>
            <li value="snowy">It's snowing! Be careful on the roads.</li>
            <li value="cloudy">It's cloudy today. Still a good day to go out.</li>
            <li>I don't have current weather information. You might want to check a weather app.</li>
        </condition>
    </template>
</category>
```

#### User Greeting
```xml
<category>
    <pattern>HELLO</pattern>
    <template>
        <condition name="name">
            <li>Hello <get name="name"/>! Nice to see you again!</li>
            <li>Hello there! What's your name?</li>
        </condition>
    </template>
</category>
```

#### Time-based Response
```xml
<category>
    <pattern>GOOD MORNING</pattern>
    <template>
        <condition name="time_of_day">
            <li value="morning">Good morning! <get name="name"/></li>
            <li value="afternoon">It's actually afternoon now, but good morning to you too!</li>
            <li value="evening">Good evening! Did you mean to say good evening?</li>
            <li>Good morning! <get name="name"/></li>
        </condition>
    </template>
</category>
```

### 9. Implementation Complexity

**High Complexity** - This is one of the most complex AIML2 features because:
- Multiple condition types to support
- Nested condition processing
- Variable lookup across multiple contexts
- Integration with all other AIML tags
- Complex regex patterns for parsing
- Recursive processing for nested conditions

### 10. Estimated Implementation Time

- **Core functionality**: 2-3 hours
- **Comprehensive testing**: 1-2 hours
- **Integration testing**: 1 hour
- **Documentation and examples**: 1 hour
- **Total**: 5-7 hours

### 11. Dependencies

The condition tag implementation depends on:
- Existing variable system (session, knowledge base, properties)
- Template processing pipeline
- Other AIML tag processing (for nested conditions)
- Regex parsing system

### 12. Potential Challenges

1. **Nested condition processing** - Need to handle recursive condition parsing
2. **Variable lookup performance** - Multiple context checks
3. **Regex complexity** - Complex patterns for different condition types
4. **Integration order** - Must work with all other AIML tags
5. **Default case handling** - Proper fallback behavior
6. **Edge cases** - Empty variables, missing conditions, malformed XML

## Conclusion

The `<condition>` tag is a powerful but complex feature that would significantly enhance Golem's conversational capabilities. The implementation would follow the same patterns as other AIML tags but with additional complexity for variable lookup and conditional logic.

Would you like me to proceed with implementing the `<condition>` tag, or would you prefer to see other features first?
