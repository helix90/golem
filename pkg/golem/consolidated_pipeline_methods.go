package golem

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
	if g.consolidatedProcessor == nil {
		return make(map[string]interface{})
	}
	return g.consolidatedProcessor.GetProcessorStats()
}

// ResetProcessorMetrics resets metrics for all processors
func (g *Golem) ResetProcessorMetrics() {
	if g.consolidatedProcessor != nil {
		g.consolidatedProcessor.ResetMetrics()
	}
}

// GetProcessingOrder returns the current processing order
func (g *Golem) GetProcessingOrder() []string {
	return g.GetConsolidatedProcessor().GetProcessingOrder()
}

// SetProcessingOrder allows reordering of processors
func (g *Golem) SetProcessingOrder(order []string) error {
	if g.consolidatedProcessor == nil {
		g.consolidatedProcessor = NewConsolidatedTemplateProcessor(g)
	}
	return g.consolidatedProcessor.SetProcessingOrder(order)
}

// GetProcessor returns a specific processor by name
func (g *Golem) GetProcessor(name string) (TemplateProcessor, bool) {
	if g.consolidatedProcessor == nil {
		return nil, false
	}
	return g.consolidatedProcessor.GetProcessor(name)
}

// GetProcessorsByType returns processors of a specific type
func (g *Golem) GetProcessorsByType(processorType ProcessorType) []TemplateProcessor {
	if g.consolidatedProcessor == nil {
		return []TemplateProcessor{}
	}
	return g.consolidatedProcessor.GetProcessorsByType(processorType)
}
