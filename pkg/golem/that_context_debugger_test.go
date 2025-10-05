package golem

import (
	"fmt"
	"strings"
	"testing"
)

func TestThatContextDebugger(t *testing.T) {
	// Create a test session
	session := &ChatSession{
		ThatHistory: []string{"Hello", "How are you?", "Good morning"},
		Topic:       "greetings",
	}

	// Create debugger
	debugger := NewThatContextDebugger(session)

	// Test initial state
	if debugger.EnableTracing {
		t.Error("Expected tracing to be disabled by default")
	}
	if debugger.EnableProfiling {
		t.Error("Expected profiling to be disabled by default")
	}
	if len(debugger.TraceLog) != 0 {
		t.Error("Expected empty trace log")
	}
	if len(debugger.PerformanceLog) != 0 {
		t.Error("Expected empty performance log")
	}
}

func TestThatContextDebuggerEnableDisable(t *testing.T) {
	session := &ChatSession{}
	debugger := NewThatContextDebugger(session)

	// Test enabling
	debugger.EnableDebugging()
	if !debugger.EnableTracing {
		t.Error("Expected tracing to be enabled")
	}
	if !debugger.EnableProfiling {
		t.Error("Expected profiling to be enabled")
	}

	// Test disabling
	debugger.DisableDebugging()
	if debugger.EnableTracing {
		t.Error("Expected tracing to be disabled")
	}
	if debugger.EnableProfiling {
		t.Error("Expected profiling to be disabled")
	}
}

func TestTraceThatMatching(t *testing.T) {
	session := &ChatSession{}
	debugger := NewThatContextDebugger(session)
	debugger.EnableDebugging()

	// Test tracing
	debugger.TraceThatMatching("HELLO *", "HELLO WORLD", true, "Hello there!", 1000000, nil)

	if len(debugger.TraceLog) != 1 {
		t.Errorf("Expected 1 trace entry, got %d", len(debugger.TraceLog))
	}

	entry := debugger.TraceLog[0]
	if entry.Operation != "that_matching" {
		t.Errorf("Expected operation 'that_matching', got '%s'", entry.Operation)
	}
	if entry.Pattern != "HELLO *" {
		t.Errorf("Expected pattern 'HELLO *', got '%s'", entry.Pattern)
	}
	if entry.Input != "HELLO WORLD" {
		t.Errorf("Expected input 'HELLO WORLD', got '%s'", entry.Input)
	}
	if !entry.Matched {
		t.Error("Expected matched to be true")
	}
	if entry.Result != "Hello there!" {
		t.Errorf("Expected result 'Hello there!', got '%s'", entry.Result)
	}
	if entry.Duration != 1000000 {
		t.Errorf("Expected duration 1000000, got %d", entry.Duration)
	}
}

func TestTraceThatMatchingWithError(t *testing.T) {
	session := &ChatSession{}
	debugger := NewThatContextDebugger(session)
	debugger.EnableDebugging()

	// Test tracing with error
	err := fmt.Errorf("pattern matching failed")
	debugger.TraceThatMatching("INVALID", "INPUT", false, "", 500000, err)

	if len(debugger.TraceLog) != 1 {
		t.Errorf("Expected 1 trace entry, got %d", len(debugger.TraceLog))
	}

	entry := debugger.TraceLog[0]
	if entry.Error != "pattern matching failed" {
		t.Errorf("Expected error 'pattern matching failed', got '%s'", entry.Error)
	}
	if entry.Matched {
		t.Error("Expected matched to be false")
	}
}

func TestTraceThatHistoryOperation(t *testing.T) {
	session := &ChatSession{
		ThatHistory: []string{"Hello", "How are you?"},
		Topic:       "greetings",
	}
	debugger := NewThatContextDebugger(session)
	debugger.EnableDebugging()

	// Test tracing history operation
	debugger.TraceThatHistoryOperation("add_to_history", "New response", 200000, nil)

	if len(debugger.TraceLog) != 1 {
		t.Errorf("Expected 1 trace entry, got %d", len(debugger.TraceLog))
	}

	entry := debugger.TraceLog[0]
	if entry.Operation != "add_to_history" {
		t.Errorf("Expected operation 'add_to_history', got '%s'", entry.Operation)
	}
	if entry.Input != "New response" {
		t.Errorf("Expected input 'New response', got '%s'", entry.Input)
	}
	if !entry.Matched {
		t.Error("Expected matched to be true (no error)")
	}

	// Check context
	if entry.Context["history_size"] != 2 {
		t.Errorf("Expected history_size 2, got %v", entry.Context["history_size"])
	}
	if entry.Context["topic"] != "greetings" {
		t.Errorf("Expected topic 'greetings', got '%v'", entry.Context["topic"])
	}
}

func TestRecordPerformance(t *testing.T) {
	session := &ChatSession{}
	debugger := NewThatContextDebugger(session)
	debugger.EnableDebugging()

	// Test recording performance
	debugger.RecordPerformance("pattern_matching", 1000000, 1024, 5, 10, 3, 2)

	if len(debugger.PerformanceLog) != 1 {
		t.Errorf("Expected 1 performance entry, got %d", len(debugger.PerformanceLog))
	}

	entry := debugger.PerformanceLog[0]
	if entry.Operation != "pattern_matching" {
		t.Errorf("Expected operation 'pattern_matching', got '%s'", entry.Operation)
	}
	if entry.Duration != 1000000 {
		t.Errorf("Expected duration 1000000, got %d", entry.Duration)
	}
	if entry.MemoryUsage != 1024 {
		t.Errorf("Expected memory usage 1024, got %d", entry.MemoryUsage)
	}
	if entry.PatternCount != 5 {
		t.Errorf("Expected pattern count 5, got %d", entry.PatternCount)
	}
	if entry.HistorySize != 10 {
		t.Errorf("Expected history size 10, got %d", entry.HistorySize)
	}
	if entry.CacheHits != 3 {
		t.Errorf("Expected cache hits 3, got %d", entry.CacheHits)
	}
	if entry.CacheMisses != 2 {
		t.Errorf("Expected cache misses 2, got %d", entry.CacheMisses)
	}
}

func TestGetTraceSummary(t *testing.T) {
	session := &ChatSession{}
	debugger := NewThatContextDebugger(session)
	debugger.EnableDebugging()

	// Test empty trace log
	summary := debugger.GetTraceSummary()
	if summary["total_operations"] != 0 {
		t.Errorf("Expected 0 operations, got %v", summary["total_operations"])
	}

	// Add some trace entries
	debugger.TraceThatMatching("PATTERN1", "INPUT1", true, "RESULT1", 1000000, nil)
	debugger.TraceThatMatching("PATTERN2", "INPUT2", false, "", 500000, fmt.Errorf("no match"))
	debugger.TraceThatHistoryOperation("add", "NEW", 200000, nil)

	summary = debugger.GetTraceSummary()
	if summary["total_operations"] != 3 {
		t.Errorf("Expected 3 operations, got %v", summary["total_operations"])
	}

	operations := summary["operations"].(map[string]int)
	if operations["that_matching"] != 2 {
		t.Errorf("Expected 2 that_matching operations, got %d", operations["that_matching"])
	}
	if operations["add"] != 1 {
		t.Errorf("Expected 1 add operation, got %d", operations["add"])
	}

	if summary["errors"] != 1 {
		t.Errorf("Expected 1 error, got %v", summary["errors"])
	}
	if summary["matched_count"] != 2 {
		t.Errorf("Expected 2 matched operations, got %v", summary["matched_count"])
	}

	matchRate := summary["match_rate"].(float64)
	if matchRate != 2.0/3.0 {
		t.Errorf("Expected match rate 2/3, got %f", matchRate)
	}
}

func TestGetPerformanceSummary(t *testing.T) {
	session := &ChatSession{}
	debugger := NewThatContextDebugger(session)
	debugger.EnableDebugging()

	// Test empty performance log
	summary := debugger.GetPerformanceSummary()
	if summary["total_operations"] != 0 {
		t.Errorf("Expected 0 operations, got %v", summary["total_operations"])
	}

	// Add some performance entries
	debugger.RecordPerformance("operation1", 1000000, 1024, 5, 10, 3, 2)
	debugger.RecordPerformance("operation2", 2000000, 2048, 10, 20, 6, 4)
	debugger.RecordPerformance("operation1", 1500000, 1536, 7, 15, 4, 3)

	summary = debugger.GetPerformanceSummary()
	if summary["total_operations"] != 3 {
		t.Errorf("Expected 3 operations, got %v", summary["total_operations"])
	}

	operationStats := summary["operation_stats"].(map[string]map[string]interface{})

	// Check operation1 stats
	op1Stats := operationStats["operation1"]
	if op1Stats["count"] != 2 {
		t.Errorf("Expected operation1 count 2, got %v", op1Stats["count"])
	}

	// Check operation2 stats
	op2Stats := operationStats["operation2"]
	if op2Stats["count"] != 1 {
		t.Errorf("Expected operation2 count 1, got %v", op2Stats["count"])
	}

	// Check averages
	avgDuration := summary["avg_duration_ns"].(float64)
	expectedAvg := (1000000.0 + 2000000.0 + 1500000.0) / 3.0
	if avgDuration != expectedAvg {
		t.Errorf("Expected avg duration %f, got %f", expectedAvg, avgDuration)
	}
}

func TestAnalyzeThatHistory(t *testing.T) {
	session := &ChatSession{
		ThatHistory: []string{
			"Hello",
			"How are you?",
			"Good morning",
			"Hello",        // Duplicate
			"Good morning", // Duplicate
			"Hello",        // Another duplicate
		},
	}
	debugger := NewThatContextDebugger(session)

	analysis := debugger.analyzeThatHistory()

	if analysis["total_responses"] != 6 {
		t.Errorf("Expected 6 total responses, got %v", analysis["total_responses"])
	}
	if analysis["unique_responses"] != 3 {
		t.Errorf("Expected 3 unique responses, got %v", analysis["unique_responses"])
	}

	// Check most common response
	if analysis["most_common_response"] != "Hello" {
		t.Errorf("Expected most common response 'Hello', got '%v'", analysis["most_common_response"])
	}
	if analysis["most_common_count"] != 3 {
		t.Errorf("Expected most common count 3, got %v", analysis["most_common_count"])
	}

	// Check repetition rate
	repetitionRate := analysis["repetition_rate"].(float64)
	expectedRate := 3.0 / 6.0
	if repetitionRate != expectedRate {
		t.Errorf("Expected repetition rate %f, got %f", expectedRate, repetitionRate)
	}
}

func TestAnalyzePatternUsage(t *testing.T) {
	session := &ChatSession{}
	debugger := NewThatContextDebugger(session)
	debugger.EnableDebugging()

	// Add some trace entries
	debugger.TraceThatMatching("PATTERN1", "INPUT1", true, "RESULT1", 1000000, nil)
	debugger.TraceThatMatching("PATTERN1", "INPUT2", true, "RESULT2", 500000, nil)
	debugger.TraceThatMatching("PATTERN2", "INPUT3", false, "", 200000, fmt.Errorf("no match"))
	debugger.TraceThatMatching("PATTERN2", "INPUT4", true, "RESULT3", 300000, nil)

	analysis := debugger.analyzePatternUsage()

	if analysis["total_patterns"] != 2 {
		t.Errorf("Expected 2 total patterns, got %v", analysis["total_patterns"])
	}
	if analysis["total_attempts"] != 4 {
		t.Errorf("Expected 4 total attempts, got %v", analysis["total_attempts"])
	}
	if analysis["total_matches"] != 3 {
		t.Errorf("Expected 3 total matches, got %v", analysis["total_matches"])
	}

	// Check overall effectiveness
	effectiveness := analysis["overall_effectiveness"].(float64)
	expectedEffectiveness := 3.0 / 4.0
	if effectiveness != expectedEffectiveness {
		t.Errorf("Expected effectiveness %f, got %f", expectedEffectiveness, effectiveness)
	}

	// Check most effective pattern
	if analysis["most_effective_pattern"] != "PATTERN1" {
		t.Errorf("Expected most effective pattern 'PATTERN1', got '%v'", analysis["most_effective_pattern"])
	}

	// Check least effective pattern
	if analysis["least_effective_pattern"] != "PATTERN2" {
		t.Errorf("Expected least effective pattern 'PATTERN2', got '%v'", analysis["least_effective_pattern"])
	}
}

func TestAnalyzePerformance(t *testing.T) {
	session := &ChatSession{}
	debugger := NewThatContextDebugger(session)
	debugger.EnableDebugging()

	// Add some performance entries
	debugger.RecordPerformance("op1", 1000000, 1024, 5, 10, 3, 2)
	debugger.RecordPerformance("op2", 2000000, 2048, 10, 20, 6, 4)
	debugger.RecordPerformance("op3", 500000, 512, 2, 5, 1, 1)

	analysis := debugger.analyzePerformance()

	if analysis["total_operations"] != 3 {
		t.Errorf("Expected 3 total operations, got %v", analysis["total_operations"])
	}

	// Check duration statistics
	avgDuration := analysis["avg_duration_ns"].(float64)
	expectedAvg := (1000000.0 + 2000000.0 + 500000.0) / 3.0
	if avgDuration != expectedAvg {
		t.Errorf("Expected avg duration %f, got %f", expectedAvg, avgDuration)
	}

	minDuration := analysis["min_duration_ns"].(int64)
	if minDuration != 500000 {
		t.Errorf("Expected min duration 500000, got %d", minDuration)
	}

	maxDuration := analysis["max_duration_ns"].(int64)
	if maxDuration != 2000000 {
		t.Errorf("Expected max duration 2000000, got %d", maxDuration)
	}

	// Check memory statistics
	avgMemory := analysis["avg_memory_bytes"].(float64)
	expectedMemory := (1024.0 + 2048.0 + 512.0) / 3.0
	if avgMemory != expectedMemory {
		t.Errorf("Expected avg memory %f, got %f", expectedMemory, avgMemory)
	}
}

func TestGenerateRecommendations(t *testing.T) {
	session := &ChatSession{
		ThatHistory: make([]string, 60), // Large history
	}
	debugger := NewThatContextDebugger(session)
	debugger.EnableDebugging()

	// Add some performance data with low cache hit rate
	debugger.RecordPerformance("op1", 2000000, 1024, 5, 10, 1, 9) // Low hit rate

	recommendations := debugger.generateRecommendations()

	// Should have recommendations for large history and low cache hit rate
	if len(recommendations) == 0 {
		t.Error("Expected recommendations, got none")
	}

	// Check for specific recommendations
	foundLargeHistory := false
	foundLowCache := false

	for _, rec := range recommendations {
		if strings.Contains(rec, "Large that history") {
			foundLargeHistory = true
		}
		if strings.Contains(rec, "Low cache hit rate") {
			foundLowCache = true
		}
	}

	if !foundLargeHistory {
		t.Error("Expected recommendation for large history")
	}
	if !foundLowCache {
		t.Error("Expected recommendation for low cache hit rate")
	}
}

func TestClearDebugData(t *testing.T) {
	session := &ChatSession{}
	debugger := NewThatContextDebugger(session)
	debugger.EnableDebugging()

	// Add some data
	debugger.TraceThatMatching("PATTERN", "INPUT", true, "RESULT", 1000000, nil)
	debugger.RecordPerformance("op", 1000000, 1024, 5, 10, 3, 2)

	if len(debugger.TraceLog) == 0 {
		t.Error("Expected trace log to have entries")
	}
	if len(debugger.PerformanceLog) == 0 {
		t.Error("Expected performance log to have entries")
	}

	// Clear data
	debugger.ClearDebugData()

	if len(debugger.TraceLog) != 0 {
		t.Error("Expected trace log to be empty after clear")
	}
	if len(debugger.PerformanceLog) != 0 {
		t.Error("Expected performance log to be empty after clear")
	}
}

func TestExportDebugData(t *testing.T) {
	session := &ChatSession{
		ThatHistory: []string{"Hello", "How are you?"},
		Topic:       "greetings",
	}
	debugger := NewThatContextDebugger(session)
	debugger.EnableDebugging()

	// Add some data
	debugger.TraceThatMatching("PATTERN", "INPUT", true, "RESULT", 1000000, nil)
	debugger.RecordPerformance("op", 1000000, 1024, 5, 10, 3, 2)

	// Export data
	exported := debugger.ExportDebugData()

	// Check structure
	if exported["trace_log"] == nil {
		t.Error("Expected trace_log in exported data")
	}
	if exported["performance_log"] == nil {
		t.Error("Expected performance_log in exported data")
	}
	if exported["summary"] == nil {
		t.Error("Expected summary in exported data")
	}

	// Check trace log
	traceLog := exported["trace_log"].([]ThatTraceEntry)
	if len(traceLog) != 1 {
		t.Errorf("Expected 1 trace entry, got %d", len(traceLog))
	}

	// Check performance log
	perfLog := exported["performance_log"].([]ThatPerformanceEntry)
	if len(perfLog) != 1 {
		t.Errorf("Expected 1 performance entry, got %d", len(perfLog))
	}

	// Check summary
	summary := exported["summary"].(map[string]interface{})
	if summary["trace_summary"] == nil {
		t.Error("Expected trace_summary in summary")
	}
	if summary["performance_summary"] == nil {
		t.Error("Expected performance_summary in summary")
	}
	if summary["analysis"] == nil {
		t.Error("Expected analysis in summary")
	}
}

func TestThatContextDebuggerIntegration(t *testing.T) {
	// Test complete debugging workflow
	session := &ChatSession{
		ThatHistory: []string{"Hello", "How are you?", "Good morning"},
		Topic:       "greetings",
	}
	debugger := NewThatContextDebugger(session)
	debugger.EnableDebugging()

	// Simulate some operations
	debugger.TraceThatMatching("HELLO *", "HELLO WORLD", true, "Hello there!", 1000000, nil)
	debugger.TraceThatMatching("HOW *", "HOW ARE YOU", true, "I'm fine", 500000, nil)
	debugger.TraceThatMatching("INVALID", "INPUT", false, "", 200000, fmt.Errorf("no match"))

	debugger.TraceThatHistoryOperation("add", "New response", 100000, nil)
	debugger.TraceThatHistoryOperation("compress", "Compress history", 200000, nil)

	debugger.RecordPerformance("pattern_matching", 1000000, 1024, 5, 10, 3, 2)
	debugger.RecordPerformance("history_management", 500000, 512, 2, 5, 1, 1)

	// Test analysis
	analysis := debugger.AnalyzeThatPatterns()

	if analysis["history_analysis"] == nil {
		t.Error("Expected history_analysis in analysis")
	}
	if analysis["pattern_analysis"] == nil {
		t.Error("Expected pattern_analysis in analysis")
	}
	if analysis["performance_analysis"] == nil {
		t.Error("Expected performance_analysis in analysis")
	}
	if analysis["recommendations"] == nil {
		t.Error("Expected recommendations in analysis")
	}

	// Test summaries
	traceSummary := debugger.GetTraceSummary()
	if traceSummary["total_operations"] != 5 {
		t.Errorf("Expected 5 total operations, got %v", traceSummary["total_operations"])
	}

	perfSummary := debugger.GetPerformanceSummary()
	if perfSummary["total_operations"] != 2 {
		t.Errorf("Expected 2 total operations, got %v", perfSummary["total_operations"])
	}

	// Test export
	exported := debugger.ExportDebugData()
	if exported["trace_log"] == nil {
		t.Error("Expected trace_log in exported data")
	}
	if exported["performance_log"] == nil {
		t.Error("Expected performance_log in exported data")
	}
	if exported["summary"] == nil {
		t.Error("Expected summary in exported data")
	}
}
