package golem

import (
	"fmt"
	"strings"
	"time"
)

// ConsolidatedTemplateProcessor manages the consolidated template processing pipeline
type ConsolidatedTemplateProcessor struct {
	registry *ProcessorRegistry
	golem    *Golem
}

// NewConsolidatedTemplateProcessor creates a new consolidated template processor
func NewConsolidatedTemplateProcessor(g *Golem) *ConsolidatedTemplateProcessor {
	registry := NewProcessorRegistry()

	// Register comprehensive processors in the correct order
	// Cast to TemplateProcessor interface since they implement all required methods
	registry.RegisterProcessor(TemplateProcessor(&ComprehensiveWildcardProcessor{golem: g}))
	registry.RegisterProcessor(TemplateProcessor(&ComprehensiveVariableProcessor{golem: g}))
	registry.RegisterProcessor(TemplateProcessor(&ComprehensiveRecursiveProcessor{golem: g}))
	registry.RegisterProcessor(TemplateProcessor(&ComprehensiveDataProcessor{golem: g}))
	registry.RegisterProcessor(TemplateProcessor(&ComprehensiveTextProcessor{golem: g}))
	registry.RegisterProcessor(TemplateProcessor(&ComprehensiveFormatProcessor{golem: g}))

	return &ConsolidatedTemplateProcessor{
		registry: registry,
		golem:    g,
	}
}

// ProcessTemplate processes a template using the consolidated pipeline
func (ctp *ConsolidatedTemplateProcessor) ProcessTemplate(template string, wildcards map[string]string, ctx *VariableContext) (string, error) {
	startTime := time.Now()

	// Check cache first if enabled
	hasConditionTags := strings.Contains(template, "<condition ")
	if ctp.golem.templateConfig.EnableCaching && !hasConditionTags {
		cacheKey := ctp.golem.generateTemplateCacheKey(template, wildcards, ctx)
		if cached, found := ctp.golem.getFromTemplateCache(cacheKey); found {
			ctp.golem.templateMetrics.CacheHits++
			ctp.golem.updateCacheHitRate()
			ctp.golem.LogDebug("Template cache hit for key: %s", cacheKey)
			return cached, nil
		}
		ctp.golem.templateMetrics.CacheMisses++
		ctp.golem.updateCacheHitRate()
	}

	// Log initial state
	ctp.golem.LogInfo("Template text: '%s'", template)
	ctp.golem.LogInfo("Wildcards: %v", wildcards)

	// Process using the consolidated pipeline
	response, err := ctp.registry.ProcessTemplate(template, wildcards, ctx)
	if err != nil {
		return template, err
	}

	ctp.golem.LogInfo("Final response: '%s'", response)

	finalResponse := strings.TrimSpace(response)

	// Update metrics
	processingTime := float64(time.Since(startTime).Nanoseconds()) / 1000000.0 // Convert to milliseconds
	ctp.golem.templateMetrics.TotalProcessed++
	ctp.golem.templateMetrics.LastProcessed = time.Now().Format(time.RFC3339)

	// Update average processing time
	if ctp.golem.templateMetrics.TotalProcessed == 1 {
		ctp.golem.templateMetrics.AverageProcessTime = processingTime
	} else {
		ctp.golem.templateMetrics.AverageProcessTime = (ctp.golem.templateMetrics.AverageProcessTime*float64(ctp.golem.templateMetrics.TotalProcessed-1) + processingTime) / float64(ctp.golem.templateMetrics.TotalProcessed)
	}

	// Cache the result if caching is enabled
	if ctp.golem.templateConfig.EnableCaching && !hasConditionTags {
		cacheKey := ctp.golem.generateTemplateCacheKey(template, wildcards, ctx)
		ctp.golem.storeInTemplateCache(cacheKey, finalResponse)
	}

	// Update memory peak
	currentMemory := len([]byte(finalResponse))
	if currentMemory > ctp.golem.templateMetrics.MemoryPeak {
		ctp.golem.templateMetrics.MemoryPeak = currentMemory
	}

	return finalResponse, nil
}

// GetProcessorMetrics returns metrics for all processors
func (ctp *ConsolidatedTemplateProcessor) GetProcessorMetrics() map[string]*ProcessorMetrics {
	return ctp.registry.GetMetrics()
}

// GetProcessor returns a specific processor
func (ctp *ConsolidatedTemplateProcessor) GetProcessor(name string) (TemplateProcessor, bool) {
	return ctp.registry.GetProcessor(name)
}

// GetProcessorsByType returns processors of a specific type
func (ctp *ConsolidatedTemplateProcessor) GetProcessorsByType(processorType ProcessorType) []TemplateProcessor {
	return ctp.registry.GetProcessorsByType(processorType)
}

// ResetMetrics resets all processor metrics
func (ctp *ConsolidatedTemplateProcessor) ResetMetrics() {
	ctp.registry.ResetMetrics()
}

// GetProcessingOrder returns the current processing order
func (ctp *ConsolidatedTemplateProcessor) GetProcessingOrder() []string {
	var order []string
	for _, processor := range ctp.registry.GetAllProcessors() {
		order = append(order, processor.Name())
	}
	return order
}

// SetProcessingOrder allows reordering of processors (for advanced use cases)
func (ctp *ConsolidatedTemplateProcessor) SetProcessingOrder(order []string) error {
	// Validate that all processors exist
	for _, name := range order {
		if _, exists := ctp.registry.processors[name]; !exists {
			return fmt.Errorf("processor %s not found", name)
		}
	}

	ctp.registry.order = order
	return nil
}

// EnableProcessor enables a specific processor
func (ctp *ConsolidatedTemplateProcessor) EnableProcessor(name string) error {
	processor, exists := ctp.registry.GetProcessor(name)
	if !exists {
		return fmt.Errorf("processor %s not found", name)
	}

	// For now, we don't have a disabled state, but this could be extended
	// to support enabling/disabling processors dynamically
	_ = processor
	return nil
}

// DisableProcessor disables a specific processor
func (ctp *ConsolidatedTemplateProcessor) DisableProcessor(name string) error {
	// For now, we don't have a disabled state, but this could be extended
	// to support enabling/disabling processors dynamically
	return fmt.Errorf("processor disabling not yet implemented")
}

// GetProcessorStats returns detailed statistics for all processors
func (ctp *ConsolidatedTemplateProcessor) GetProcessorStats() map[string]interface{} {
	stats := make(map[string]interface{})

	for name, metrics := range ctp.registry.GetMetrics() {
		stats[name] = map[string]interface{}{
			"total_calls":     metrics.TotalCalls,
			"total_time_ms":   metrics.TotalTime.Milliseconds(),
			"average_time_ms": metrics.AverageTime.Milliseconds(),
			"last_call_time":  metrics.LastCallTime.Format(time.RFC3339),
			"error_count":     metrics.ErrorCount,
			"cache_hits":      metrics.CacheHits,
			"cache_misses":    metrics.CacheMisses,
			"hit_rate":        float64(metrics.CacheHits) / float64(metrics.CacheHits+metrics.CacheMisses),
		}
	}

	return stats
}

// ComprehensiveWildcardProcessor handles all wildcard processing
type ComprehensiveWildcardProcessor struct {
	*BaseProcessor
	golem *Golem
}

func (p *ComprehensiveWildcardProcessor) Name() string                { return "wildcard" }
func (p *ComprehensiveWildcardProcessor) Type() ProcessorType         { return ProcessorTypeWildcard }
func (p *ComprehensiveWildcardProcessor) Priority() ProcessorPriority { return PriorityEarly }
func (p *ComprehensiveWildcardProcessor) Condition() ProcessorCondition {
	return ProcessorCondition{SkipIfEmpty: true}
}
func (p *ComprehensiveWildcardProcessor) Process(template string, wildcards map[string]string, ctx *VariableContext) (string, error) {
	response := template

	// Store wildcards in context for that wildcard processing
	if ctx != nil {
		for key, value := range wildcards {
			ctx.LocalVars[key] = value
		}
	}

	// Replace indexed star tags first
	for key, value := range wildcards {
		switch key {
		case "star1":
			response = strings.ReplaceAll(response, "<star index=\"1\"/>", value)
			response = strings.ReplaceAll(response, "<star1/>", value)
		case "star2":
			response = strings.ReplaceAll(response, "<star index=\"2\"/>", value)
			response = strings.ReplaceAll(response, "<star2/>", value)
		case "star3":
			response = strings.ReplaceAll(response, "<star index=\"3\"/>", value)
			response = strings.ReplaceAll(response, "<star3/>", value)
		case "star4":
			response = strings.ReplaceAll(response, "<star index=\"4\"/>", value)
			response = strings.ReplaceAll(response, "<star4/>", value)
		case "star5":
			response = strings.ReplaceAll(response, "<star index=\"5\"/>", value)
			response = strings.ReplaceAll(response, "<star5/>", value)
		case "star6":
			response = strings.ReplaceAll(response, "<star index=\"6\"/>", value)
			response = strings.ReplaceAll(response, "<star6/>", value)
		case "star7":
			response = strings.ReplaceAll(response, "<star index=\"7\"/>", value)
			response = strings.ReplaceAll(response, "<star7/>", value)
		case "star8":
			response = strings.ReplaceAll(response, "<star index=\"8\"/>", value)
			response = strings.ReplaceAll(response, "<star8/>", value)
		case "star9":
			response = strings.ReplaceAll(response, "<star index=\"9\"/>", value)
			response = strings.ReplaceAll(response, "<star9/>", value)
		}
	}

	// Then replace generic <star/> tags sequentially
	starIndex := 1
	for strings.Contains(response, "<star/>") && starIndex <= 9 {
		key := fmt.Sprintf("star%d", starIndex)
		if value, exists := wildcards[key]; exists {
			// Replace only the first occurrence
			response = strings.Replace(response, "<star/>", value, 1)
		} else if len(wildcards) == 1 {
			// If there's only one wildcard captured, use it for all remaining <star/> tags
			for _, value := range wildcards {
				response = strings.Replace(response, "<star/>", value, 1)
				break
			}
		} else {
			// If no wildcard value exists, replace with empty string
			response = strings.Replace(response, "<star/>", "", 1)
		}
		starIndex++
	}

	return response, nil
}
func (p *ComprehensiveWildcardProcessor) ShouldProcess(template string, ctx *VariableContext) bool {
	return strings.Contains(template, "<star")
}
func (p *ComprehensiveWildcardProcessor) GetMetrics() *ProcessorMetrics { return &ProcessorMetrics{} }
func (p *ComprehensiveWildcardProcessor) ResetMetrics()                 {}

// ComprehensiveVariableProcessor handles all variable-related processing
type ComprehensiveVariableProcessor struct {
	*BaseProcessor
	golem *Golem
}

func (p *ComprehensiveVariableProcessor) Name() string                { return "variable" }
func (p *ComprehensiveVariableProcessor) Type() ProcessorType         { return ProcessorTypeVariable }
func (p *ComprehensiveVariableProcessor) Priority() ProcessorPriority { return PriorityEarly }
func (p *ComprehensiveVariableProcessor) Condition() ProcessorCondition {
	return ProcessorCondition{RequiresContext: true, SkipIfEmpty: true}
}
func (p *ComprehensiveVariableProcessor) Process(template string, wildcards map[string]string, ctx *VariableContext) (string, error) {
	response := template

	// Replace property tags
	response = p.golem.replacePropertyTags(response)

	// Process bot tags (short form of property access)
	response = p.golem.processBotTagsWithContext(response, ctx)

	// Process think tags FIRST (internal processing, no output)
	// This allows local variables to be set before variable replacement
	response = p.golem.processThinkTagsWithContext(response, ctx)

	// Process topic setting tags first (special handling for topic)
	response = p.golem.processTopicSettingTagsWithContext(response, ctx)

	// Process set tags (before session variable replacement)
	response = p.golem.processSetTagsWithContext(response, ctx)

	// Replace session variable tags using context
	response = p.golem.replaceSessionVariableTagsWithContext(response, ctx)

	return response, nil
}
func (p *ComprehensiveVariableProcessor) ShouldProcess(template string, ctx *VariableContext) bool {
	// Check if template contains any variable-related tags
	variableTags := []string{
		"<get", "</get>",
		"<set", "</set>",
		"<bot", "</bot>",
		"<think", "</think>",
		"<topic", "</topic>",
		"<name", "</name>",
		"<value", "</value>",
	}

	for _, tag := range variableTags {
		if strings.Contains(template, tag) {
			return true
		}
	}

	return false
}
func (p *ComprehensiveVariableProcessor) GetMetrics() *ProcessorMetrics { return &ProcessorMetrics{} }
func (p *ComprehensiveVariableProcessor) ResetMetrics()                 {}

// ComprehensiveRecursiveProcessor handles all recursive processing (SRAI, SRAIX)
type ComprehensiveRecursiveProcessor struct {
	*BaseProcessor
	golem *Golem
}

func (p *ComprehensiveRecursiveProcessor) Name() string                { return "recursive" }
func (p *ComprehensiveRecursiveProcessor) Type() ProcessorType         { return ProcessorTypeRecursive }
func (p *ComprehensiveRecursiveProcessor) Priority() ProcessorPriority { return PriorityNormal }
func (p *ComprehensiveRecursiveProcessor) Condition() ProcessorCondition {
	return ProcessorCondition{RequiresContext: true, RequiresKB: true, SkipIfEmpty: true}
}
func (p *ComprehensiveRecursiveProcessor) Process(template string, wildcards map[string]string, ctx *VariableContext) (string, error) {
	response := template

	// Process SR tags (shorthand for <srai><star/></srai>) AFTER wildcard replacement
	response = p.golem.processSRTagsWithContext(response, wildcards, ctx)

	// Process SRAI tags (recursive)
	response = p.golem.processSRAITagsWithContext(response, ctx)

	// Process SRAIX tags (external services)
	response = p.golem.processSRAIXTagsWithContext(response, ctx)

	return response, nil
}
func (p *ComprehensiveRecursiveProcessor) ShouldProcess(template string, ctx *VariableContext) bool {
	// Check if template contains any recursive tags
	recursiveTags := []string{
		"<srai", "</srai>",
		"<sraix", "</sraix>",
		"<sr", "</sr>",
	}

	for _, tag := range recursiveTags {
		if strings.Contains(template, tag) {
			return true
		}
	}

	return false
}
func (p *ComprehensiveRecursiveProcessor) GetMetrics() *ProcessorMetrics { return &ProcessorMetrics{} }
func (p *ComprehensiveRecursiveProcessor) ResetMetrics()                 {}

// ComprehensiveTextProcessor handles all text processing
type ComprehensiveTextProcessor struct {
	*BaseProcessor
	golem *Golem
}

func (p *ComprehensiveTextProcessor) Name() string                { return "text" }
func (p *ComprehensiveTextProcessor) Type() ProcessorType         { return ProcessorTypeText }
func (p *ComprehensiveTextProcessor) Priority() ProcessorPriority { return PriorityNormal }
func (p *ComprehensiveTextProcessor) Condition() ProcessorCondition {
	return ProcessorCondition{SkipIfEmpty: true}
}
func (p *ComprehensiveTextProcessor) Process(template string, wildcards map[string]string, ctx *VariableContext) (string, error) {
	response := template

	// Process person tags (pronoun substitution)
	response = p.golem.processPersonTagsWithContext(response, ctx)

	// Process gender tags (gender pronoun substitution)
	response = p.golem.processGenderTagsWithContext(response, ctx)

	// Process person2 tags (first-to-third person pronoun substitution)
	response = p.golem.processPerson2TagsWithContext(response, ctx)

	// Process sentence tags (sentence-level processing)
	response = p.golem.processSentenceTagsWithContext(response, ctx)

	// Process word tags (word-level processing)
	response = p.golem.processWordTagsWithContext(response, ctx)

	return response, nil
}
func (p *ComprehensiveTextProcessor) ShouldProcess(template string, ctx *VariableContext) bool {
	// Check if template contains any text processing tags
	textTags := []string{
		"<person", "</person>",
		"<gender", "</gender>",
		"<person2", "</person2>",
		"<sentence", "</sentence>",
		"<word", "</word>",
	}

	for _, tag := range textTags {
		if strings.Contains(template, tag) {
			return true
		}
	}

	return false
}
func (p *ComprehensiveTextProcessor) GetMetrics() *ProcessorMetrics { return &ProcessorMetrics{} }
func (p *ComprehensiveTextProcessor) ResetMetrics()                 {}

// ComprehensiveDataProcessor handles all data processing (date, time, random, etc.)
type ComprehensiveDataProcessor struct {
	*BaseProcessor
	golem *Golem
}

func (p *ComprehensiveDataProcessor) Name() string                { return "data" }
func (p *ComprehensiveDataProcessor) Type() ProcessorType         { return ProcessorTypeData }
func (p *ComprehensiveDataProcessor) Priority() ProcessorPriority { return PriorityNormal }
func (p *ComprehensiveDataProcessor) Condition() ProcessorCondition {
	return ProcessorCondition{SkipIfEmpty: true}
}
func (p *ComprehensiveDataProcessor) Process(template string, wildcards map[string]string, ctx *VariableContext) (string, error) {
	response := template

	// Process date and time tags
	response = p.golem.processDateTimeTags(response)

	// Process random tags
	response = p.golem.processRandomTags(response)

	return response, nil
}
func (p *ComprehensiveDataProcessor) ShouldProcess(template string, ctx *VariableContext) bool {
	// Check if template contains any data tags
	dataTags := []string{
		"<date", "</date>",
		"<time", "</time>",
		"<random", "</random>",
		"<li", "</li>",
	}

	for _, tag := range dataTags {
		if strings.Contains(template, tag) {
			return true
		}
	}

	return false
}
func (p *ComprehensiveDataProcessor) GetMetrics() *ProcessorMetrics { return &ProcessorMetrics{} }
func (p *ComprehensiveDataProcessor) ResetMetrics()                 {}

// ComprehensiveFormatProcessor handles all text formatting operations
type ComprehensiveFormatProcessor struct {
	*BaseProcessor
	golem *Golem
}

func (p *ComprehensiveFormatProcessor) Name() string                { return "format" }
func (p *ComprehensiveFormatProcessor) Type() ProcessorType         { return ProcessorTypeFormat }
func (p *ComprehensiveFormatProcessor) Priority() ProcessorPriority { return PriorityLate }
func (p *ComprehensiveFormatProcessor) Condition() ProcessorCondition {
	return ProcessorCondition{SkipIfEmpty: true}
}
func (p *ComprehensiveFormatProcessor) Process(template string, wildcards map[string]string, ctx *VariableContext) (string, error) {
	response := template

	// Process topic tags (current topic references) - before text processing
	response = p.golem.processTopicTagsWithContext(response, ctx)

	// Process repeat tags first (before text formatting) so they can be processed by other tags
	response = p.golem.processRepeatTagsWithContext(response, ctx)

	// Process all text formatting tags
	response = p.golem.processUppercaseTagsWithContext(response, ctx)
	response = p.golem.processLowercaseTagsWithContext(response, ctx)
	response = p.golem.processFormalTagsWithContext(response, ctx)
	response = p.golem.processExplodeTagsWithContext(response, ctx)
	response = p.golem.processCapitalizeTagsWithContext(response, ctx)
	response = p.golem.processReverseTagsWithContext(response, ctx)
	response = p.golem.processAcronymTagsWithContext(response, ctx)
	response = p.golem.processTrimTagsWithContext(response, ctx)
	response = p.golem.processSubstringTagsWithContext(response, ctx)
	response = p.golem.processReplaceTagsWithContext(response, ctx)
	response = p.golem.processPluralizeTagsWithContext(response, ctx)
	response = p.golem.processShuffleTagsWithContext(response, ctx)
	response = p.golem.processLengthTagsWithContext(response, ctx)
	response = p.golem.processCountTagsWithContext(response, ctx)
	response = p.golem.processSplitTagsWithContext(response, ctx)
	response = p.golem.processJoinTagsWithContext(response, ctx)
	response = p.golem.processIndentTagsWithContext(response, ctx)
	response = p.golem.processDedentTagsWithContext(response, ctx)
	response = p.golem.processUniqueTagsWithContext(response, ctx)

	// Process normalize tags (text normalization)
	response = p.golem.processNormalizeTagsWithContext(response, ctx)

	// Process denormalize tags (text denormalization)
	response = p.golem.processDenormalizeTagsWithContext(response, ctx)

	return response, nil
}
func (p *ComprehensiveFormatProcessor) ShouldProcess(template string, ctx *VariableContext) bool {
	// Check if template contains any formatting tags
	formatTags := []string{
		"<uppercase", "</uppercase>",
		"<lowercase", "</lowercase>",
		"<formal", "</formal>",
		"<explode", "</explode>",
		"<capitalize", "</capitalize>",
		"<reverse", "</reverse>",
		"<acronym", "</acronym>",
		"<trim", "</trim>",
		"<substring", "</substring>",
		"<replace", "</replace>",
		"<pluralize", "</pluralize>",
		"<shuffle", "</shuffle>",
		"<length", "</length>",
		"<count", "</count>",
		"<split", "</split>",
		"<join", "</join>",
		"<indent", "</indent>",
		"<dedent", "</dedent>",
		"<unique", "</unique>",
		"<repeat", "</repeat>",
		"<normalize", "</normalize>",
		"<denormalize", "</denormalize>",
		"<topic", "</topic>",
	}

	for _, tag := range formatTags {
		if strings.Contains(template, tag) {
			return true
		}
	}

	return false
}
func (p *ComprehensiveFormatProcessor) GetMetrics() *ProcessorMetrics { return &ProcessorMetrics{} }
func (p *ComprehensiveFormatProcessor) ResetMetrics()                 {}
