package golem

import (
	"fmt"
	"log"
)

// DemonstrateThatPatternConflictDetection demonstrates the conflict detection system
func DemonstrateThatPatternConflictDetection() {
	fmt.Println("=== That Pattern Conflict Detection Demo ===")

	// Create a golem instance
	golem := New(false)

	// Example patterns that are likely to have conflicts
	patterns := []string{
		"HELLO",
		"HELLO WORLD",
		"* HELLO",
		"GOOD MORNING",
		"GOOD *",
		"WHAT IS YOUR NAME",
		"WHO ARE YOU",
		"* ARE YOU",
		"* * * * * WORLD",
		"VERY SPECIFIC PATTERN",
		"HELLO THERE",
		"* THERE",
		"GOOD AFTERNOON",
		"GOOD EVENING",
		"* EVENING",
	}

	fmt.Printf("Analyzing %d patterns for conflicts...\n", len(patterns))
	fmt.Println("Patterns:")
	for i, pattern := range patterns {
		fmt.Printf("  %d. %s\n", i+1, pattern)
	}
	fmt.Println()

	// Create conflict detector
	detector := NewThatPatternConflictDetector(patterns)

	// Detect conflicts
	conflicts := detector.DetectConflicts(golem)

	fmt.Printf("Found %d conflicts:\n", len(conflicts))
	fmt.Println()

	// Group conflicts by type
	conflictsByType := make(map[string][]ThatPatternConflict)
	for _, conflict := range conflicts {
		conflictsByType[conflict.Type] = append(conflictsByType[conflict.Type], conflict)
	}

	// Display conflicts by type
	for conflictType, typeConflicts := range conflictsByType {
		fmt.Printf("=== %s Conflicts (%d) ===\n", conflictType, len(typeConflicts))
		for i, conflict := range typeConflicts {
			fmt.Printf("%d. %s\n", i+1, conflict.Description)
			fmt.Printf("   Severity: %s\n", conflict.Severity)
			fmt.Printf("   Pattern 1: %s\n", conflict.Pattern1)
			fmt.Printf("   Pattern 2: %s\n", conflict.Pattern2)
			fmt.Printf("   Suggestions:\n")
			for j, suggestion := range conflict.Suggestions {
				fmt.Printf("     %d. %s\n", j+1, suggestion)
			}
			fmt.Printf("   Examples:\n")
			for j, example := range conflict.Examples {
				fmt.Printf("     %d. %s\n", j+1, example)
			}
			fmt.Println()
		}
	}

	// Display summary statistics
	fmt.Println("=== Conflict Summary ===")
	totalConflicts := len(conflicts)
	criticalConflicts := 0
	highConflicts := 0
	mediumConflicts := 0
	lowConflicts := 0

	for _, conflict := range conflicts {
		switch conflict.Severity {
		case "critical":
			criticalConflicts++
		case "high":
			highConflicts++
		case "medium":
			mediumConflicts++
		case "low":
			lowConflicts++
		}
	}

	fmt.Printf("Total conflicts: %d\n", totalConflicts)
	fmt.Printf("Critical: %d\n", criticalConflicts)
	fmt.Printf("High: %d\n", highConflicts)
	fmt.Printf("Medium: %d\n", mediumConflicts)
	fmt.Printf("Low: %d\n", lowConflicts)
	fmt.Println()

	// Display recommendations
	fmt.Println("=== Recommendations ===")
	if criticalConflicts > 0 {
		fmt.Println("⚠️  CRITICAL: Address critical conflicts immediately")
	}
	if highConflicts > 0 {
		fmt.Println("🔴 HIGH: Address high-priority conflicts soon")
	}
	if mediumConflicts > 0 {
		fmt.Println("🟡 MEDIUM: Consider addressing medium-priority conflicts")
	}
	if lowConflicts > 0 {
		fmt.Println("🟢 LOW: Low-priority conflicts can be addressed when convenient")
	}

	if totalConflicts == 0 {
		fmt.Println("✅ No conflicts detected! Your patterns are well-structured.")
	}

	fmt.Println()
	fmt.Println("=== Pattern Analysis ===")

	// Analyze pattern specificity
	fmt.Println("Pattern Specificity Analysis:")
	for i, pattern := range patterns {
		specificity := detector.calculatePatternSpecificity(pattern)
		wildcardCount := detector.countWildcards(pattern)
		fmt.Printf("  %d. %s (specificity: %.2f, wildcards: %d)\n", i+1, pattern, specificity, wildcardCount)
	}

	fmt.Println()
	fmt.Println("=== Conflict Resolution Tips ===")
	fmt.Println("1. Reorder patterns to ensure more specific patterns come first")
	fmt.Println("2. Use different wildcard types for different purposes")
	fmt.Println("3. Add more specific words to differentiate patterns")
	fmt.Println("4. Consider combining patterns if they serve the same purpose")
	fmt.Println("5. Use consistent wildcard strategies across patterns")
	fmt.Println("6. Review pattern intent and matching scope")

	fmt.Println()
	fmt.Println("Demo completed successfully!")
}

// DemonstrateConflictDetectionWithRealPatterns demonstrates conflict detection with realistic patterns
func DemonstrateConflictDetectionWithRealPatterns() {
	fmt.Println("=== Realistic Pattern Conflict Detection Demo ===")

	// Create a golem instance
	golem := New(false)

	// Realistic AIML patterns that might have conflicts
	patterns := []string{
		"HELLO",
		"HELLO WORLD",
		"HELLO THERE",
		"GOOD MORNING",
		"GOOD AFTERNOON",
		"GOOD EVENING",
		"GOOD NIGHT",
		"WHAT IS YOUR NAME",
		"WHO ARE YOU",
		"WHERE ARE YOU FROM",
		"WHAT DO YOU DO",
		"TELL ME ABOUT YOURSELF",
		"WHAT CAN YOU DO",
		"HELP ME",
		"THANK YOU",
		"GOODBYE",
		"SEE YOU LATER",
		"HAVE A NICE DAY",
		"TAKE CARE",
		"* HELLO",
		"* WORLD",
		"* MORNING",
		"* AFTERNOON",
		"* EVENING",
		"* NIGHT",
		"* NAME",
		"* YOU",
		"* FROM",
		"* DO",
		"* YOURSELF",
		"* CAN",
		"* HELP",
		"* THANK",
		"* GOODBYE",
		"* LATER",
		"* DAY",
		"* CARE",
		"* * * * * WORLD",
		"* * * * * MORNING",
		"* * * * * AFTERNOON",
		"* * * * * EVENING",
		"* * * * * NIGHT",
		"* * * * * NAME",
		"* * * * * YOU",
		"* * * * * FROM",
		"* * * * * DO",
		"* * * * * YOURSELF",
		"* * * * * CAN",
		"* * * * * HELP",
		"* * * * * THANK",
		"* * * * * GOODBYE",
		"* * * * * LATER",
		"* * * * * DAY",
		"* * * * * CARE",
	}

	fmt.Printf("Analyzing %d realistic patterns for conflicts...\n", len(patterns))

	// Create conflict detector
	detector := NewThatPatternConflictDetector(patterns)

	// Detect conflicts
	conflicts := detector.DetectConflicts(golem)

	fmt.Printf("Found %d conflicts in realistic patterns\n", len(conflicts))

	// Display most critical conflicts
	fmt.Println("\n=== Most Critical Conflicts ===")
	criticalConflicts := 0
	for _, conflict := range conflicts {
		if conflict.Severity == "critical" {
			criticalConflicts++
			fmt.Printf("CRITICAL: %s\n", conflict.Description)
			fmt.Printf("  Pattern 1: %s\n", conflict.Pattern1)
			fmt.Printf("  Pattern 2: %s\n", conflict.Pattern2)
			fmt.Printf("  Suggestions:\n")
			for i, suggestion := range conflict.Suggestions {
				fmt.Printf("    %d. %s\n", i+1, suggestion)
			}
			fmt.Println()
		}
	}

	if criticalConflicts == 0 {
		fmt.Println("No critical conflicts found.")
	}

	// Display high-priority conflicts
	fmt.Println("=== High-Priority Conflicts ===")
	highConflicts := 0
	for _, conflict := range conflicts {
		if conflict.Severity == "high" {
			highConflicts++
			fmt.Printf("HIGH: %s\n", conflict.Description)
			fmt.Printf("  Pattern 1: %s\n", conflict.Pattern1)
			fmt.Printf("  Pattern 2: %s\n", conflict.Pattern2)
			fmt.Println()
		}
	}

	if highConflicts == 0 {
		fmt.Println("No high-priority conflicts found.")
	}

	// Display conflict statistics
	fmt.Println("=== Conflict Statistics ===")
	conflictTypes := make(map[string]int)
	severityLevels := make(map[string]int)

	for _, conflict := range conflicts {
		conflictTypes[conflict.Type]++
		severityLevels[conflict.Severity]++
	}

	fmt.Println("Conflicts by type:")
	for conflictType, count := range conflictTypes {
		fmt.Printf("  %s: %d\n", conflictType, count)
	}

	fmt.Println("\nConflicts by severity:")
	for severity, count := range severityLevels {
		fmt.Printf("  %s: %d\n", severity, count)
	}

	fmt.Println("\nDemo completed successfully!")
}

// RunConflictDetectionDemo runs the conflict detection demonstration
func RunConflictDetectionDemo() {
	fmt.Println("That Pattern Conflict Detection System")
	fmt.Println("=====================================")
	fmt.Println()

	// Run basic demo
	DemonstrateThatPatternConflictDetection()
	fmt.Println()

	// Run realistic patterns demo
	DemonstrateConflictDetectionWithRealPatterns()
}

func init() {
	// Register the demo function
	log.Println("That Pattern Conflict Detection Demo Available")
}
