# Missing AIML Tags Analysis

## Overview
Analysis of Pandorabots AIML files to identify tags that are not currently supported in our Golem implementation.

## Currently Supported Tags
Our implementation supports the following tags:
- `<person>`, `<person2>`, `<gender>`
- `<sentence>`, `<word>`
- `<uppercase>`, `<lowercase>`, `<formal>`, `<capitalize>`
- `<explode>`, `<reverse>`, `<acronym>`
- `<trim>`, `<substring>`, `<replace>`, `<pluralize>`
- `<shuffle>`, `<length>`, `<count>`
- `<split>`, `<join>`, `<unique>`
- `<indent>`, `<dedent>`
- `<normalize>`, `<denormalize>`
- `<srai>`, `<sraix>`, `<learn>`, `<unlearn>`
- `<think>`, `<condition>`, `<set>`, `<get>`
- `<bot>`, `<size>`, `<version>`, `<id>`
- `<that>`, `<topic>`, `<map>`, `<list>`, `<array>`
- `<star>`, `<input>`, `<random>`, `<li>`

## Missing Tags Found in Pandorabots Files

### 1. **`<uniq>`** - Unique Predicate Tag
**Usage Found:**
```xml
<uniq><subj>?singular</subj><pred>hasPlural</pred><obj><star/></obj></uniq>
<uniq><subj><star/></subj><pred>sound</pred><obj>?sound</obj></uniq>
```
**Purpose:** Used for predicate-subject-object relationships, similar to RDF triples.
**Files:** `animal.aiml`

### 2. **`<subj>`, `<pred>`, `<obj>`** - Subject, Predicate, Object Tags
**Usage Found:**
```xml
<subj>?singular</subj>
<pred>hasPlural</pred>
<obj><star/></obj>
```
**Purpose:** Used within `<uniq>` tags to define RDF-like triples.
**Files:** `animal.aiml`

### 3. **`<first>`** - First Element Tag
**Usage Found:**
```xml
<first><get var="list"/></first>
<first><normalize><star/></normalize></first>
```
**Purpose:** Returns the first element of a list or string.
**Files:** `utilities.aiml`

### 4. **`<rest>`** - Rest of List Tag
**Usage Found:**
```xml
<rest><get var="list"/></rest>
```
**Purpose:** Returns all elements except the first from a list.
**Files:** `utilities.aiml`

### 5. **`<loop>`** - Loop Control Tag
**Usage Found:**
```xml
<loop/>
```
**Purpose:** Used for loop control in conditionals, similar to `continue` in programming.
**Files:** `utilities.aiml`

### 6. **`<eval>`** - Evaluation Tag
**Usage Found:**
```xml
<eval><get var="pattern1"/></eval>
<eval><star/></eval>
```
**Purpose:** Evaluates the content as AIML code, allowing dynamic pattern generation.
**Files:** `train.aiml`, `utilities.aiml`

### 7. **`<learn>`** - Learning Tag (Different from our implementation)
**Usage Found:**
```xml
<learn>
<category>
<pattern><eval><get var="pattern1"/></eval></pattern>
<template><eval><get var="response"/></eval></template>
</category>
</learn>
```
**Purpose:** Creates new categories dynamically at runtime.
**Files:** `train.aiml`, `utilities.aiml`

### 8. **`<javascript>`** - JavaScript Execution Tag
**Usage Found:**
```xml
<javascript><get var="formula"/></javascript>
<javascript>
function foo() {
if (0 < 1) return "true";
return "false";
}
</javascript>
```
**Purpose:** Executes JavaScript code for calculations and logic.
**Files:** `udc.aiml`, `utilities.aiml`

### 9. **`<random>`** - Random Selection Tag
**Usage Found:**
```xml
<random>
<li>I don't know the answer.</li>
<li>I used my lifeline to ask another robot, but he didn't know.</li>
<li>I asked another robot, but he didn't know.</li>
</random>
```
**Purpose:** Randomly selects one of the `<li>` elements.
**Files:** `zpand-webbot.aiml`, `udc.aiml`

### 10. **`<input>`** - Input Tag
**Usage Found:**
```xml
<input/>
```
**Purpose:** References the current user input.
**Files:** `zpand-webbot.aiml`, `train.aiml`

### 11. **`<sraix>`** - External Service Tag (Different from our implementation)
**Usage Found:**
```xml
<sraix service="pannous">WHAT IS <star/></sraix>
<sraix bot="drwallace/wndef" botid="f038d2f99e345a95" host="callmom.pandorabots.com">WNDEF <get var="word"/></sraix>
```
**Purpose:** Calls external services or other bots.
**Files:** `sraix.aiml`, `utilities.aiml`

### 12. **`<think>`** - Think Tag (Different from our implementation)
**Usage Found:**
```xml
<think>
<set var="name"><star/></set>
<set var="verb"><star index="2"/></set>
</think>
```
**Purpose:** Executes code without outputting it to the user.
**Files:** `train.aiml`, `udc.aiml`

### 13. **`<oob>`** - Out of Band Tag
**Usage Found:**
```xml
<oob><email><to><get name="email"/></to><subject>Transcript of <bot name="name"/> with <get name="name"/> on <date/></subject><body>
<oob><schedule><title><star/></title><description><lowercase><star index="2"/></lowercase></description><get name="sraix"/></schedule></oob>
```
**Purpose:** Handles out-of-band operations like email, scheduling, etc.
**Files:** `oob.aiml`, `contactaction.aiml`, `sraix.aiml`

### 14. **`<br/>`** - Line Break Tag
**Usage Found:**
```xml
<br/>
```
**Purpose:** Inserts line breaks in output.
**Files:** `update.aiml`

### 15. **`<date/>`** - Date Tag
**Usage Found:**
```xml
<date/>
```
**Purpose:** Returns current date/time.
**Files:** `oob.aiml`, `contactaction.aiml`

## Priority for Implementation

### High Priority (Core AIML2 Features)
1. **`<uniq>`, `<subj>`, `<pred>`, `<obj>`** - RDF-like predicate system
2. **`<first>`, `<rest>`** - List manipulation
3. **`<loop>`** - Loop control
4. **`<eval>`** - Dynamic evaluation
5. **`<random>`** - Random selection
6. **`<input>`** - Input reference

### Medium Priority (Enhanced Features)
7. **`<learn>`** - Dynamic learning (enhanced version)
8. **`<think>`** - Silent execution (enhanced version)
9. **`<sraix>`** - External services (enhanced version)
10. **`<javascript>`** - JavaScript execution

### Low Priority (Specialized Features)
11. **`<oob>`** - Out of band operations
12. **`<br/>`** - Line breaks
13. **`<date/>`** - Date/time

## Implementation Notes

1. **`<uniq>` system** is particularly important as it's used extensively in the animal.aiml file for knowledge representation.

2. **`<first>` and `<rest>`** are fundamental list operations that should be implemented early.

3. **`<loop>`** is used for iteration control and is essential for complex list processing.

4. **`<eval>`** allows dynamic AIML generation, which is powerful but needs careful security considerations.

5. **`<random>`** is a simple but important feature for varied responses.

6. **`<input>`** is basic but essential for referencing user input in templates.

7. **`<learn>`** in the Pandorabots context is more sophisticated than our current implementation, allowing dynamic category creation.

8. **`<javascript>`** requires a JavaScript engine integration, which adds complexity but enables powerful calculations.

9. **`<oob>`** tags are for platform-specific operations and may not be relevant for all use cases.

## Next Steps

1. Implement high-priority tags first
2. Create comprehensive tests for each new tag
3. Ensure proper integration with existing template processing pipeline
4. Consider security implications for `<eval>` and `<javascript>` tags
5. Document new tags in AIML2 specification compliance
