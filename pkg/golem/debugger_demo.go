package golem

import (
	"fmt"
	"log"
)

// DemoThatContextDebugging demonstrates the debugging tools
func DemoThatContextDebugging() {
	fmt.Println("=== That Context Debugging Tools Demo ===")

	// Create a session with some history
	session := &ChatSession{
		ThatHistory: []string{
			"Hello there!",
			"How are you doing?",
			"Good morning!",
			"Hello there!", // Duplicate
			"Nice to meet you",
		},
		Topic: "greetings",
	}

	// Create debugger
	debugger := NewThatContextDebugger(session)
	debugger.EnableDebugging()

	fmt.Println("\n1. Tracing Pattern Matching Operations")
	fmt.Println("=====================================")

	// Simulate pattern matching operations
	patterns := []struct {
		pattern  string
		input    string
		matched  bool
		result   string
		duration int64
		err      error
	}{
		{"HELLO *", "HELLO WORLD", true, "Hello there!", 1000000, nil},
		{"HOW *", "HOW ARE YOU", true, "I'm fine, thanks!", 500000, nil},
		{"INVALID *", "SOME INPUT", false, "", 200000, fmt.Errorf("no match found")},
		{"GOOD *", "GOOD MORNING", true, "Good morning to you!", 300000, nil},
		{"NICE *", "NICE TO MEET YOU", true, "Nice to meet you too!", 400000, nil},
	}

	for i, p := range patterns {
		debugger.TraceThatMatching(p.pattern, p.input, p.matched, p.result, p.duration, p.err)
		fmt.Printf("Operation %d: Pattern='%s', Input='%s', Matched=%v\n",
			i+1, p.pattern, p.input, p.matched)
	}

	fmt.Println("\n2. Tracing History Operations")
	fmt.Println("=============================")

	// Simulate history operations
	historyOps := []struct {
		operation string
		input     string
		duration  int64
		err       error
	}{
		{"add_to_history", "New response added", 100000, nil},
		{"compress_history", "Compressed 5 items to 3", 200000, nil},
		{"validate_history", "Validated history integrity", 50000, nil},
		{"clear_history", "Cleared old entries", 30000, nil},
	}

	for i, op := range historyOps {
		debugger.TraceThatHistoryOperation(op.operation, op.input, op.duration, op.err)
		fmt.Printf("History Op %d: %s - %s\n", i+1, op.operation, op.input)
	}

	fmt.Println("\n3. Recording Performance Metrics")
	fmt.Println("==============================")

	// Simulate performance recording
	perfOps := []struct {
		operation    string
		duration     int64
		memory       int64
		patternCount int
		historySize  int
		cacheHits    int
		cacheMisses  int
	}{
		{"pattern_matching", 1000000, 1024, 5, 10, 3, 2},
		{"history_compression", 500000, 512, 2, 5, 1, 1},
		{"cache_operations", 200000, 256, 1, 3, 2, 0},
		{"validation", 100000, 128, 0, 2, 0, 1},
	}

	for i, perf := range perfOps {
		debugger.RecordPerformance(perf.operation, perf.duration, perf.memory,
			perf.patternCount, perf.historySize, perf.cacheHits, perf.cacheMisses)
		fmt.Printf("Performance %d: %s - %dns, %d bytes\n",
			i+1, perf.operation, perf.duration, perf.memory)
	}

	fmt.Println("\n4. Trace Summary")
	fmt.Println("===============")

	traceSummary := debugger.GetTraceSummary()
	fmt.Printf("Total Operations: %v\n", traceSummary["total_operations"])
	fmt.Printf("Operations: %v\n", traceSummary["operations"])
	fmt.Printf("Errors: %v\n", traceSummary["errors"])
	fmt.Printf("Match Rate: %.2f%%\n", traceSummary["match_rate"].(float64)*100)
	fmt.Printf("Average Duration: %.0f ns\n", traceSummary["avg_duration_ns"].(float64))

	fmt.Println("\n5. Performance Summary")
	fmt.Println("=====================")

	perfSummary := debugger.GetPerformanceSummary()
	fmt.Printf("Total Operations: %v\n", perfSummary["total_operations"])
	fmt.Printf("Average Duration: %.0f ns\n", perfSummary["avg_duration_ns"].(float64))
	fmt.Printf("Average Memory: %.0f bytes\n", perfSummary["avg_memory_bytes"].(float64))

	operationStats := perfSummary["operation_stats"].(map[string]map[string]interface{})
	for op, stats := range operationStats {
		fmt.Printf("  %s: count=%v, avg=%.0fns, min=%v, max=%v\n",
			op, stats["count"], stats["avg_ns"], stats["min_ns"], stats["max_ns"])
	}

	fmt.Println("\n6. Comprehensive Analysis")
	fmt.Println("========================")

	analysis := debugger.AnalyzeThatPatterns()

	// History analysis
	historyAnalysis := analysis["history_analysis"].(map[string]interface{})
	fmt.Printf("History Analysis:\n")
	fmt.Printf("  Total Responses: %v\n", historyAnalysis["total_responses"])
	fmt.Printf("  Unique Responses: %v\n", historyAnalysis["unique_responses"])
	fmt.Printf("  Average Length: %.1f\n", historyAnalysis["avg_length"])
	fmt.Printf("  Repetition Rate: %.2f%%\n", historyAnalysis["repetition_rate"].(float64)*100)

	// Pattern analysis
	patternAnalysis := analysis["pattern_analysis"].(map[string]interface{})
	fmt.Printf("\nPattern Analysis:\n")
	fmt.Printf("  Total Patterns: %v\n", patternAnalysis["total_patterns"])
	fmt.Printf("  Total Attempts: %v\n", patternAnalysis["total_attempts"])
	fmt.Printf("  Total Matches: %v\n", patternAnalysis["total_matches"])
	fmt.Printf("  Overall Effectiveness: %.2f%%\n",
		patternAnalysis["overall_effectiveness"].(float64)*100)

	// Performance analysis
	performanceAnalysis := analysis["performance_analysis"].(map[string]interface{})
	fmt.Printf("\nPerformance Analysis:\n")
	fmt.Printf("  Average Duration: %.0f ns\n", performanceAnalysis["avg_duration_ns"].(float64))
	fmt.Printf("  Min Duration: %v ns\n", performanceAnalysis["min_duration_ns"])
	fmt.Printf("  Max Duration: %v ns\n", performanceAnalysis["max_duration_ns"])
	fmt.Printf("  Average Memory: %.0f bytes\n", performanceAnalysis["avg_memory_bytes"].(float64))

	// Recommendations
	recommendations := analysis["recommendations"].([]string)
	fmt.Printf("\nRecommendations:\n")
	for i, rec := range recommendations {
		fmt.Printf("  %d. %s\n", i+1, rec)
	}

	fmt.Println("\n7. Export Debug Data")
	fmt.Println("===================")

	exported := debugger.ExportDebugData()
	fmt.Printf("Exported data contains:\n")
	fmt.Printf("  Trace Log Entries: %d\n", len(exported["trace_log"].([]ThatTraceEntry)))
	fmt.Printf("  Performance Log Entries: %d\n", len(exported["performance_log"].([]ThatPerformanceEntry)))
	fmt.Printf("  Summary Data: Available\n")

	fmt.Println("\n8. Clear Debug Data")
	fmt.Println("===================")

	fmt.Printf("Before clear - Trace entries: %d, Performance entries: %d\n",
		len(debugger.TraceLog), len(debugger.PerformanceLog))

	debugger.ClearDebugData()

	fmt.Printf("After clear - Trace entries: %d, Performance entries: %d\n",
		len(debugger.TraceLog), len(debugger.PerformanceLog))

	fmt.Println("\n=== Debugging Demo Complete ===")
}

func init() {
	// Run demo when package is imported
	log.Println("That Context Debugging Tools Demo Available")
}
