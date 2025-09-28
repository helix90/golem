# Learn and Learnf Tags in AIML

The `<learn>` and `<learnf>` tags enable AIML bots to dynamically acquire new knowledge during conversations, making them more adaptive and intelligent over time.

## Overview

- **`<learn>` Tag**: Enables session-specific learning. New categories are added to the bot's knowledge base but are temporary and only last for the current session.
- **`<learnf>` Tag**: Enables persistent learning. New categories are permanently added to the bot's knowledge base and persist across all sessions.

## Basic Syntax

### Learn Tag (Session-Specific)
```xml
<learn>
  <category>
    <pattern>PATTERN TO LEARN</pattern>
    <template>Response template</template>
  </category>
</learn>
```

### Learnf Tag (Persistent)
```xml
<learnf>
  <category>
    <pattern>PATTERN TO LEARN</pattern>
    <template>Response template</template>
  </category>
</learnf>
```

## Key Features

### 1. Dynamic Category Addition
Both tags allow the bot to learn new patterns and responses at runtime:

```xml
<category>
  <pattern>TEACH ME *</pattern>
  <template>
    <learn>
      <category>
        <pattern>I KNOW *</pattern>
        <template>Yes, I know about <star/>.</template>
      </category>
    </learn>
    I've learned that pattern!
  </template>
</category>
```

### 2. Multiple Category Learning
Learn multiple categories in a single operation:

```xml
<learn>
  <category>
    <pattern>GOOD MORNING</pattern>
    <template>Good morning! How are you today?</template>
  </category>
  <category>
    <pattern>GOOD EVENING</pattern>
    <template>Good evening! How was your day?</template>
  </category>
</learn>
```

### 3. Wildcard Support
Learn patterns with wildcards:

```xml
<learn>
  <category>
    <pattern>I LIKE *</pattern>
    <template>I'm glad you like <star/>!</template>
  </category>
</learn>
```

### 4. Complex Template Learning
Learn categories with sophisticated templates:

```xml
<learn>
  <category>
    <pattern>CALCULATE * PLUS *</pattern>
    <template>
      <think>
        <set name="num1"><star/></set>
        <set name="num2"><star index="2"/></set>
      </think>
      <star/> plus <star index="2"/> equals <get name="num1"/> + <get name="num2"/>.
    </template>
  </category>
</learn>
```

## Advanced Features

### 1. Learning with SRAI
```xml
<learn>
  <category>
    <pattern>RESPONSE TO *</pattern>
    <template>
      <srai>HELLO</srai> I remember you asked about <star/>.
    </template>
  </category>
</learn>
```

### 2. Learning with Conditions
```xml
<learn>
  <category>
    <pattern>CONDITIONAL *</pattern>
    <template>
      <condition name="mood">
        <li value="happy">I'm happy about <star/>!</li>
        <li value="sad">I'm sad about <star/>.</li>
        <li>I feel neutral about <star/>.</li>
      </condition>
    </template>
  </category>
</learn>
```

### 3. Learning with Random Responses
```xml
<learn>
  <category>
    <pattern>RANDOM *</pattern>
    <template>
      <random>
        <li>I like <star/>!</li>
        <li><star/> is interesting!</li>
        <li>Tell me more about <star/>!</li>
      </random>
    </template>
  </category>
</learn>
```

### 4. Learning with Date/Time
```xml
<learn>
  <category>
    <pattern>WHAT TIME IS IT</pattern>
    <template>
      The current time is <time format="short"/> on <date format="long"/>.
    </template>
  </category>
</learn>
```

### 5. Learning with Properties
```xml
<learn>
  <category>
    <pattern>WHO ARE YOU</pattern>
    <template>
      I am <get name="name"/>, your AI assistant.
    </template>
  </category>
</learn>
```

### 6. Learning with External Services (SRAIX)
```xml
<learn>
  <category>
    <pattern>WEATHER IN *</pattern>
    <template>
      <sraix service="weather_service">What's the weather like in <star/>?</sraix>
    </template>
  </category>
</learn>
```

## Implementation Details

### How It Works

1. **Pattern Recognition**: The bot recognizes `<learn>` and `<learnf>` tags in templates
2. **Content Parsing**: The content within the tags is parsed as AIML categories
3. **Validation**: Categories are validated for proper structure and content
4. **Integration**: Valid categories are added to the knowledge base
5. **Indexing**: Patterns are normalized and indexed for fast lookup
6. **Cleanup**: The learn tags are removed from the final response

### Pattern Normalization

Learned patterns undergo the same normalization as regular patterns:
- Case conversion to uppercase
- Whitespace normalization
- Punctuation removal
- Wildcard preservation

### Error Handling

- **Invalid AIML**: Malformed content within learn tags is ignored
- **Validation Errors**: Categories with missing patterns or templates are rejected
- **Parsing Errors**: Unparseable content results in graceful degradation
- **Empty Content**: Empty learn tags are processed without errors

### Memory Management

- **Session Learning**: Categories learned with `<learn>` are stored in memory and lost when the session ends
- **Persistent Learning**: Categories learned with `<learnf>` are stored in memory (persistent storage implementation pending)
- **Pattern Updates**: Existing patterns are updated rather than duplicated
- **Memory Limits**: Consider implementing memory limits for production use

## Use Cases

### 1. User-Specific Learning
```xml
<category>
  <pattern>MY NAME IS *</pattern>
  <template>
    <learn>
      <category>
        <pattern>WHAT IS MY NAME</pattern>
        <template>Your name is <star/>.</template>
      </category>
    </learn>
    Nice to meet you, <star/>!
  </template>
</category>
```

### 2. Domain-Specific Knowledge
```xml
<category>
  <pattern>TEACH ME ABOUT *</pattern>
  <template>
    <learn>
      <category>
        <pattern>TELL ME ABOUT *</pattern>
        <template>I can tell you about <star/> based on what you taught me.</template>
      </category>
    </learn>
    I've learned about <star/>!
  </template>
</category>
```

### 3. Dynamic Response Generation
```xml
<category>
  <pattern>CREATE RESPONSE FOR *</pattern>
  <template>
    <learn>
      <category>
        <pattern>RESPONSE FOR *</pattern>
        <template>
          <random>
            <li>Here's a response about <star/>!</li>
            <li>I have something to say about <star/>!</li>
            <li>Let me respond to <star/>!</li>
          </random>
        </template>
      </category>
    </learn>
    I've created a response pattern for <star/>!
  </template>
</category>
```

### 4. Recursive Learning
```xml
<category>
  <pattern>LEARN TO LEARN</pattern>
  <template>
    <learn>
      <category>
        <pattern>TEACH ME SOMETHING NEW</pattern>
        <template>
          <learn>
            <category>
              <pattern>I LEARNED *</pattern>
              <template>Great! You learned <star/>. I'm proud of you!</template>
            </category>
          </learn>
          I've learned how to learn!
        </template>
      </category>
    </learn>
    I've learned recursive learning!
  </template>
</category>
```

## Best Practices

### 1. Validation
Always validate learned content to prevent malicious or inappropriate patterns:

```xml
<learn>
  <category>
    <pattern>SAFE *</pattern>
    <template>
      <think>
        <set name="safe_content"><star/></set>
      </think>
      I've safely learned about <get name="safe_content"/>.
    </template>
  </category>
</learn>
```

### 2. Error Handling
Include fallback responses for learning failures:

```xml
<category>
  <pattern>LEARN *</pattern>
  <template>
    <learn>
      <category>
        <pattern>LEARNED *</pattern>
        <template>I learned about <star/>!</template>
      </category>
    </learn>
    <random>
      <li>I've learned that pattern!</li>
      <li>Learning complete!</li>
      <li>I now know how to respond to that!</li>
    </random>
  </template>
</category>
```

### 3. Memory Management
Consider implementing memory limits and cleanup:

```xml
<category>
  <pattern>CLEAR LEARNING</pattern>
  <template>
    <think>
      <set name="learning_enabled">false</set>
    </think>
    I've disabled learning for now.
  </template>
</category>
```

### 4. Learning Logs
Track what the bot has learned:

```xml
<category>
  <pattern>WHAT HAVE YOU LEARNED</pattern>
  <template>
    <think>
      <set name="learned_count">0</set>
    </think>
    I've learned many new patterns! I'm constantly growing and adapting.
  </template>
</category>
```

## Security Considerations

### 1. Input Validation
- Validate all learned patterns for appropriate content
- Implement length limits for patterns and templates
- Sanitize user input before learning

### 2. Access Control
- Consider implementing learning permissions
- Monitor learning frequency and patterns
- Implement learning quotas

### 3. Content Filtering
- Filter inappropriate or harmful content
- Implement content moderation
- Regular review of learned patterns

## Performance Considerations

### 1. Memory Usage
- Monitor memory consumption with large numbers of learned patterns
- Implement pattern cleanup and archiving
- Consider persistent storage for learnf patterns

### 2. Pattern Matching
- Learned patterns are indexed for fast lookup
- Consider pattern complexity impact on matching performance
- Implement pattern optimization

### 3. Learning Frequency
- Monitor learning frequency to prevent abuse
- Implement rate limiting for learning operations
- Consider learning cooldown periods

## Future Enhancements

### 1. Persistent Storage
- Implement file-based storage for learnf patterns
- Database integration for large-scale learning
- Pattern versioning and migration

### 2. Learning Analytics
- Track learning patterns and effectiveness
- Implement learning metrics and reporting
- Pattern usage statistics

### 3. Advanced Learning
- Machine learning integration
- Pattern similarity detection
- Automatic pattern optimization

### 4. Learning Management
- Pattern editing and deletion
- Learning rollback capabilities
- Pattern import/export functionality

## Troubleshooting

### Common Issues

1. **Patterns Not Learning**: Check for valid AIML syntax within learn tags
2. **Memory Issues**: Monitor memory usage with large numbers of learned patterns
3. **Performance Degradation**: Consider pattern complexity and quantity
4. **Learning Failures**: Check error logs for validation failures

### Debug Mode

Enable verbose logging to see learning operations:

```go
g := golem.New(true) // Enable verbose mode
```

This will log:
- Learning operations being performed
- Pattern validation results
- Category addition confirmations
- Error conditions encountered

## Conclusion

The `<learn>` and `<learnf>` tags provide powerful capabilities for creating adaptive and intelligent AIML bots. By enabling dynamic learning, bots can grow and adapt to user needs, creating more engaging and personalized conversations.

Remember to implement proper validation, security measures, and performance monitoring when using these features in production environments.
