package golem

import "time"

// Note: The consolidated pipeline is now always enabled and cannot be disabled

// GetConsolidatedProcessor returns the consolidated template processor
func (g *Golem) GetConsolidatedProcessor() *ConsolidatedTemplateProcessor {
	if g.consolidatedProcessor == nil {
		g.consolidatedProcessor = NewConsolidatedTemplateProcessor(g)
	}
	return g.consolidatedProcessor
}

// GetProcessorMetrics returns metrics for all processors in the consolidated pipeline
func (g *Golem) GetProcessorMetrics() map[string]*ProcessorMetrics {
	if g.consolidatedProcessor == nil {
		return make(map[string]*ProcessorMetrics)
	}
	return g.consolidatedProcessor.GetProcessorMetrics()
}

// GetProcessorStats returns detailed statistics for all processors
func (g *Golem) GetProcessorStats() map[string]interface{} {
	// Ensure TreeProcessor is initialized
	if g.treeProcessor == nil {
		g.treeProcessor = NewTreeProcessor(g)
	}

	// Return TreeProcessor metrics (tree-based AST processing is now the only method)
	if g.treeProcessor.metrics != nil {
		stats := make(map[string]interface{})

		for name, metrics := range g.treeProcessor.metrics.GetMetrics() {
			stats[name] = map[string]interface{}{
				"total_calls":     metrics.TotalCalls,
				"total_time_ms":   metrics.TotalTime.Milliseconds(),
				"average_time_ms": metrics.AverageTime.Milliseconds(),
				"last_call_time":  metrics.LastCallTime.Format(time.RFC3339),
				"error_count":     metrics.ErrorCount,
				"cache_hits":      metrics.CacheHits,
				"cache_misses":    metrics.CacheMisses,
			}
			// Avoid divide by zero
			if metrics.CacheHits+metrics.CacheMisses > 0 {
				hitRate := float64(metrics.CacheHits) / float64(metrics.CacheHits+metrics.CacheMisses)
				stats[name].(map[string]interface{})["hit_rate"] = hitRate
			} else {
				stats[name].(map[string]interface{})["hit_rate"] = 0.0
			}
		}

		return stats
	}

	// Return empty map if metrics not initialized
	return make(map[string]interface{})
}

// ResetProcessorMetrics resets metrics for all processors
func (g *Golem) ResetProcessorMetrics() {
	// Ensure TreeProcessor is initialized
	if g.treeProcessor == nil {
		g.treeProcessor = NewTreeProcessor(g)
	}

	// Reset TreeProcessor metrics (tree-based AST processing is now the only method)
	if g.treeProcessor.metrics != nil {
		g.treeProcessor.metrics.ResetMetrics()
	}
}

// GetProcessingOrder returns the current processing order
func (g *Golem) GetProcessingOrder() []string {
	// Ensure TreeProcessor is initialized
	if g.treeProcessor == nil {
		g.treeProcessor = NewTreeProcessor(g)
	}

	// Return TreeProcessor's logical processor order (tree-based AST processing is now the only method)
	if g.treeProcessor.metrics != nil {
		return g.treeProcessor.metrics.order
	}

	// Return empty slice if metrics not initialized
	return []string{}
}

// SetProcessingOrder allows reordering of processors
// Note: This method is deprecated as TreeProcessor has a fixed processing order
func (g *Golem) SetProcessingOrder(order []string) error {
	// TreeProcessor has a fixed processing order and cannot be reordered
	// This method is kept for backward compatibility but has no effect
	g.LogWarn("SetProcessingOrder is deprecated: TreeProcessor uses a fixed processing order")
	return nil
}

// GetProcessor returns a specific processor by name
// Note: This method returns TreeProcessor's logical sub-processors
func (g *Golem) GetProcessor(name string) (TemplateProcessor, bool) {
	// Ensure TreeProcessor is initialized
	if g.treeProcessor == nil {
		g.treeProcessor = NewTreeProcessor(g)
	}

	// Return TreeProcessor's logical sub-processors
	if g.treeProcessor.metrics != nil {
		processor, ok := g.treeProcessor.metrics.processors[name]
		return processor, ok
	}
	return nil, false
}

// GetProcessorsByType returns processors of a specific type
// Note: This method returns TreeProcessor's logical sub-processors by type
func (g *Golem) GetProcessorsByType(processorType ProcessorType) []TemplateProcessor {
	// Ensure TreeProcessor is initialized
	if g.treeProcessor == nil {
		g.treeProcessor = NewTreeProcessor(g)
	}

	// Return TreeProcessor's logical sub-processors filtered by type
	if g.treeProcessor.metrics != nil {
		var processors []TemplateProcessor
		for _, processor := range g.treeProcessor.metrics.processors {
			if processor.Type() == processorType {
				processors = append(processors, processor)
			}
		}
		return processors
	}
	return []TemplateProcessor{}
}
