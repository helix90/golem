package golem

import (
	"fmt"
	"log"
)

// DemoEnhancedThatPatternValidation demonstrates the enhanced validation features
func DemoEnhancedThatPatternValidation() {
	fmt.Println("=== Enhanced That Pattern Validation Demo ===")

	// Test various patterns
	patterns := []string{
		"HELLO WORLD",                   // Valid simple pattern
		"HELLO * WORLD",                 // Valid with wildcard
		"* * * * * * * * * * *",         // Too many wildcards
		"HELLO @ WORLD",                 // Invalid character
		"HELLO <set>WORLD",              // Unbalanced tags
		"HELLO ** WORLD",                // Consecutive wildcards
		"* HELLO WORLD",                 // Starts with wildcard
		"HI",                            // Very short
		"HELLO HELLO HELLO HELLO WORLD", // Repeated words
		"HELLO    WORLD",                // Multiple spaces
	}

	for i, pattern := range patterns {
		fmt.Printf("\n--- Test %d: %s ---\n", i+1, pattern)
		result := ValidateThatPatternDetailed(pattern)

		fmt.Printf("Valid: %v\n", result.IsValid)

		if len(result.Errors) > 0 {
			fmt.Printf("Errors: %v\n", result.Errors)
		}

		if len(result.Warnings) > 0 {
			fmt.Printf("Warnings: %v\n", result.Warnings)
		}

		if len(result.Suggestions) > 0 {
			fmt.Printf("Suggestions: %v\n", result.Suggestions)
		}

		fmt.Printf("Stats: Length=%v, Words=%v, Wildcards=%v\n",
			result.Stats["length"],
			result.Stats["word_count"],
			result.Stats["wildcard_count"])
	}
}

func init() {
	// Run demo when package is imported
	log.Println("Enhanced That Pattern Validation Demo Available")
}
