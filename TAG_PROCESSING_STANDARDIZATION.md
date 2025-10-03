# Tag Processing Pipeline Standardization

## Overview
Standardized the tag processing pipeline to ensure consistent behavior across all template processing functions in the Golem AIML engine.

## Problem Identified
The codebase had two main template processing functions with different processing orders:

1. **`processTemplateWithContext`** - Main template processing with full context support
2. **`processTemplateContentForVariable`** - Variable content processing with limited processing steps

## Issues Found

### Missing Processing Steps in `processTemplateContentForVariable`
The variable content processing function was missing several critical processing steps:
- ❌ SR tags processing (`processSRTagsWithContext`)
- ❌ Bot tags processing (`processBotTagsWithContext`)
- ❌ Think tags processing (`processThinkTagsWithContext`)
- ❌ Topic setting tags (`processTopicSettingTagsWithContext`)
- ❌ Set tags processing (`processSetTagsWithContext`)
- ❌ Person tags processing (`processPersonTagsWithContext`)
- ❌ Gender tags processing (`processGenderTagsWithContext`)
- ❌ Sentence tags processing (`processSentenceTagsWithContext`)
- ❌ Word tags processing (`processWordTagsWithContext`)
- ❌ Request tags processing (`processRequestTags`)
- ❌ Response tags processing (`processResponseTags`)

### Inconsistent Processing Order
The two functions had different processing orders, which could lead to:
- Inconsistent behavior when the same template is processed through different paths
- Unexpected results in complex templates with variable assignments
- Difficult-to-debug issues in production

## Solution Implemented

### Full Standardization Approach
Modified `processTemplateContentForVariable` to use the same processing pipeline as `processTemplateWithContext`:

```go
// processTemplateContentForVariable processes template content for variable assignment without outputting
// This function now uses the same processing pipeline as processTemplateWithContext to ensure consistency
func (g *Golem) processTemplateContentForVariable(template string, wildcards map[string]string, ctx *VariableContext) string {
    if g.verbose {
        g.logger.Printf("Processing variable content: '%s'", template)
        g.logger.Printf("Wildcards: %v", wildcards)
    }

    // Use the main template processing function to ensure consistent processing order
    // This ensures that variable content is processed with the same tag processing pipeline
    // as regular templates, maintaining consistency across the codebase
    result := g.processTemplateWithContext(template, wildcards, ctx)

    if g.verbose {
        g.logger.Printf("Variable content result: '%s'", result)
    }

    return result
}
```

## Benefits

### 1. **Consistency**
- All template processing now follows the same order
- Variable content is processed identically to regular templates
- Eliminates inconsistencies between different processing paths

### 2. **Maintainability**
- Single source of truth for processing order
- Changes to processing pipeline automatically apply to all functions
- Easier to debug and maintain

### 3. **Reliability**
- Variable assignments now have access to all processing features
- Complex templates with variables work consistently
- Reduces edge cases and unexpected behavior

### 4. **Feature Completeness**
- Variable content now supports all AIML tags
- Person, gender, sentence, word processing available in variables
- Request/response history available in variables
- Full SRAI/SR tag support in variables

## Standardized Processing Order

The following processing order is now consistent across all template processing functions:

1. **Wildcard replacement** (indexed star tags)
2. **Wildcard replacement** (generic `<star/>` tags)
3. **SR tags processing** (`processSRTagsWithContext`)
4. **Property tags replacement** (`replacePropertyTags`)
5. **Bot tags processing** (`processBotTagsWithContext`)
6. **Think tags processing** (`processThinkTagsWithContext`) - **EARLY**
7. **Topic setting tags** (`processTopicSettingTagsWithContext`) - **EARLY**
8. **Set tags processing** (`processSetTagsWithContext`) - **EARLY**
9. **Session variable tags replacement** (`replaceSessionVariableTagsWithContext`)
10. **SRAI tags processing** (`processSRAITagsWithContext`)
11. **SRAIX tags processing** (`processSRAIXTagsWithContext`)
12. **Learn tags processing** (`processLearnTagsWithContext`)
13. **Condition tags processing** (`processConditionTagsWithContext`)
14. **Date and time tags processing** (`processDateTimeTags`)
15. **Random tags processing** (`processRandomTags`)
16. **Map tags processing** (`processMapTagsWithContext`)
17. **List tags processing** (`processListTagsWithContext`)
18. **Array tags processing** (`processArrayTagsWithContext`)
19. **Person tags processing** (`processPersonTagsWithContext`)
20. **Gender tags processing** (`processGenderTagsWithContext`)
21. **Sentence tags processing** (`processSentenceTagsWithContext`)
22. **Word tags processing** (`processWordTagsWithContext`)
23. **Request tags processing** (`processRequestTags`)
24. **Response tags processing** (`processResponseTags`)

## Testing

All existing tests continue to pass, confirming that the standardization:
- ✅ Maintains backward compatibility
- ✅ Preserves existing functionality
- ✅ Improves consistency without breaking changes

## Files Modified

- **`/home/helix/golem/pkg/golem/aiml_native.go`**
  - Modified `processTemplateContentForVariable` function
  - Added comprehensive documentation
  - Ensured consistent processing pipeline

## Impact

This standardization ensures that:
- Variable content is processed with the same rigor as regular templates
- All AIML features are available in variable assignments
- The codebase is more maintainable and consistent
- Future changes to the processing pipeline automatically apply to all functions
- Complex templates with variables work reliably and consistently
