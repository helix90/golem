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
- `<uniq>`, `<subj>`, `<pred>`, `<obj>` (RDF operations)
- `<first>`, `<rest>` (List operations)
- `<loop>`, `<eval>`, `<javascript>`, `<system>`, `<gossip>`, `<var>`

## Missing Tags Found in Pandorabots Files

### 1. **`<oob>`** - Out of Band Tag
**Usage Found:**
```xml
<oob><email><to><get name="email"/></to><subject>Transcript of <bot name="name"/> with <get name="name"/> on <date/></subject><body>
<oob><schedule><title><star/></title><description><lowercase><star index="2"/></lowercase></description><get name="sraix"/></schedule></oob>
```
**Purpose:** Handles out-of-band operations like email, scheduling, etc.
**Files:** `oob.aiml`, `contactaction.aiml`, `sraix.aiml`

### 2. **Enhanced `<sraix>`** - External Service Tag with Advanced Attributes ✅ **IMPLEMENTED**
**Usage Found:**
```xml
<sraix service="pannous">WHAT IS <star/></sraix>
<sraix bot="drwallace/wndef" botid="f038d2f99e345a95" host="callmom.pandorabots.com">WNDEF <get var="word"/></sraix>
```
**Purpose:** Calls external services or other bots with bot selection and host specification.
**Files:** `sraix.aiml`, `utilities.aiml`
**Note:** All enhanced SRAIX attributes (`bot`, `botid`, `host`, `default`, `hint`) are now fully implemented.

### 3. **Enhanced `<learn>`** - Dynamic Learning Tag
**Usage Found:**
```xml
<learn>
<category>
<pattern><eval><get var="pattern1"/></eval></pattern>
<template><eval><get var="response"/></eval></template>
</category>
</learn>
```
**Purpose:** Creates new categories dynamically at runtime with `<eval>` support.
**Files:** `train.aiml`, `utilities.aiml`
**Note:** Basic `<learn>` is implemented, but may not support dynamic category creation with `<eval>`.

### 4. **Specialized Tags**
**Usage Found:**
```xml
<search>...</search>
<message>...</message>
<recipient>...</recipient>
<vocabulary/>
<hour>...</hour>
<minute>...</minute>
<description>...</description>
<title>...</title>
<body>...</body>
<from>...</from>
<to>...</to>
<subject>...</subject>
<interval>...</interval>
```
**Purpose:** Various specialized operations for search, messaging, time extraction, etc.
**Files:** Various AIML files

## Priority for Implementation

### High Priority (Platform Integration)
1. **`<oob>`** - Out of band operations for platform integration

### Medium Priority (Enhanced Features)
2. **Enhanced `<learn>`** - Dynamic category creation with `<eval>` support
3. **Specialized Tags** - Search, messaging, time extraction operations

### Low Priority (Advanced Features)
4. **Additional OOB Operations** - Email, scheduling, alarm, dial, SMS, camera, WiFi

## Implementation Notes

1. **RDF System** - The `<uniq>`, `<subj>`, `<pred>`, `<obj>` system is now fully implemented and used extensively in the animal.aiml file for knowledge representation.

2. **List Operations** - The `<first>` and `<rest>` tags are now implemented and provide fundamental list manipulation capabilities.

3. **Control Flow** - The `<loop>` tag is now implemented for iteration control in complex list processing.

4. **Dynamic Evaluation** - The `<eval>` tag is now implemented and allows dynamic AIML generation with proper security considerations.

5. **Advanced Processing** - The `<javascript>`, `<system>`, `<gossip>`, and `<var>` tags are now implemented for advanced processing capabilities.

6. **Enhanced Learning** - The current `<learn>` implementation may need enhancement to support dynamic category creation with `<eval>` as seen in Pandorabots.

7. **OOB Operations** - The `<oob>` tags are for platform-specific operations and are the main missing feature for full AIML2 compliance.

8. **Enhanced SRAIX** - The `<sraix>` tag now has full support for all advanced attributes including `bot`, `botid`, `host`, `default`, and `hint`.

## Next Steps

1. **Implement OOB Operations** - Add `<oob>` tag support for platform integration
2. **Verify Enhanced Learning** - Test and potentially enhance `<learn>` to support dynamic category creation with `<eval>`
3. **Add Specialized Tags** - Implement search, messaging, and time extraction operations as needed
4. **Update Validation** - Add missing tags to the `knownTags` validation list in `aiml_native.go`
