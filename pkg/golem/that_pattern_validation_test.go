package golem

import (
	"testing"
)

func TestValidateThatPatternDetailed(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected struct {
			isValid    bool
			errorCount int
			warnCount  int
			suggCount  int
		}
	}{
		{
			name:    "Valid simple pattern",
			pattern: "HELLO WORLD",
			expected: struct {
				isValid    bool
				errorCount int
				warnCount  int
				suggCount  int
			}{
				isValid:    true,
				errorCount: 0,
				warnCount:  0,
				suggCount:  1, // Suggestion to add wildcards
			},
		},
		{
			name:    "Valid pattern with wildcards",
			pattern: "HELLO * WORLD",
			expected: struct {
				isValid    bool
				errorCount int
				warnCount  int
				suggCount  int
			}{
				isValid:    true,
				errorCount: 0,
				warnCount:  0,
				suggCount:  0,
			},
		},
		{
			name:    "Empty pattern",
			pattern: "",
			expected: struct {
				isValid    bool
				errorCount int
				warnCount  int
				suggCount  int
			}{
				isValid:    false,
				errorCount: 1,
				warnCount:  0,
				suggCount:  0,
			},
		},
		{
			name:    "Too many wildcards",
			pattern: "* * * * * * * * * * *",
			expected: struct {
				isValid    bool
				errorCount int
				warnCount  int
				suggCount  int
			}{
				isValid:    false,
				errorCount: 1,
				warnCount:  4, // Multiple warnings for wildcard pattern
				suggCount:  1, // Suggestion to reduce wildcards
			},
		},
		{
			name:    "Invalid characters",
			pattern: "HELLO @ WORLD",
			expected: struct {
				isValid    bool
				errorCount int
				warnCount  int
				suggCount  int
			}{
				isValid:    false,
				errorCount: 1,
				warnCount:  0,
				suggCount:  1, // Still gets suggestion to add wildcards
			},
		},
		{
			name:    "Unbalanced set tags",
			pattern: "HELLO <set>WORLD",
			expected: struct {
				isValid    bool
				errorCount int
				warnCount  int
				suggCount  int
			}{
				isValid:    false,
				errorCount: 2, // Invalid characters + unbalanced tags
				warnCount:  0,
				suggCount:  1, // Still gets suggestion to add wildcards
			},
		},
		{
			name:    "Consecutive wildcards",
			pattern: "HELLO ** WORLD",
			expected: struct {
				isValid    bool
				errorCount int
				warnCount  int
				suggCount  int
			}{
				isValid:    true,
				errorCount: 0,
				warnCount:  1, // Consecutive wildcards warning
				suggCount:  0,
			},
		},
		{
			name:    "Pattern starts with wildcard",
			pattern: "* HELLO WORLD",
			expected: struct {
				isValid    bool
				errorCount int
				warnCount  int
				suggCount  int
			}{
				isValid:    true,
				errorCount: 0,
				warnCount:  1, // Starts with wildcard warning
				suggCount:  0,
			},
		},
		{
			name:    "Very short pattern",
			pattern: "HI",
			expected: struct {
				isValid    bool
				errorCount int
				warnCount  int
				suggCount  int
			}{
				isValid:    true,
				errorCount: 0,
				warnCount:  1, // Very short pattern warning
				suggCount:  4, // Multiple suggestions for short pattern
			},
		},
		{
			name:    "Very long pattern",
			pattern: "THIS IS A VERY LONG PATTERN THAT CONTAINS MANY WORDS AND SHOULD TRIGGER A WARNING BECAUSE IT IS OVER TWO HUNDRED CHARACTERS LONG AND SHOULD BE BROKEN DOWN INTO SMALLER MORE SPECIFIC PATTERNS FOR BETTER PERFORMANCE AND MAINTAINABILITY",
			expected: struct {
				isValid    bool
				errorCount int
				warnCount  int
				suggCount  int
			}{
				isValid:    true,
				errorCount: 0,
				warnCount:  1, // Very long pattern warning
				suggCount:  2, // Multiple suggestions for long pattern
			},
		},
		{
			name:    "Deeply nested alternation",
			pattern: "HELLO (WORLD (AND (UNIVERSE)))",
			expected: struct {
				isValid    bool
				errorCount int
				warnCount  int
				suggCount  int
			}{
				isValid:    true,
				errorCount: 0,
				warnCount:  0, // No warning for this depth
				suggCount:  1, // Gets suggestion to add wildcards
			},
		},
		{
			name:    "Repeated words",
			pattern: "HELLO HELLO HELLO HELLO WORLD",
			expected: struct {
				isValid    bool
				errorCount int
				warnCount  int
				suggCount  int
			}{
				isValid:    true,
				errorCount: 0,
				warnCount:  1, // Repeated word warning
				suggCount:  1, // Gets suggestion to add wildcards
			},
		},
		{
			name:    "Multiple spaces",
			pattern: "HELLO    WORLD",
			expected: struct {
				isValid    bool
				errorCount int
				warnCount  int
				suggCount  int
			}{
				isValid:    true,
				errorCount: 0,
				warnCount:  0,
				suggCount:  2, // Multiple suggestions including whitespace
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateThatPatternDetailed(tt.pattern)

			if result.IsValid != tt.expected.isValid {
				t.Errorf("Expected IsValid=%v, got %v", tt.expected.isValid, result.IsValid)
			}

			if len(result.Errors) != tt.expected.errorCount {
				t.Errorf("Expected %d errors, got %d: %v", tt.expected.errorCount, len(result.Errors), result.Errors)
			}

			if len(result.Warnings) != tt.expected.warnCount {
				t.Errorf("Expected %d warnings, got %d: %v", tt.expected.warnCount, len(result.Warnings), result.Warnings)
			}

			if len(result.Suggestions) != tt.expected.suggCount {
				t.Errorf("Expected %d suggestions, got %d: %v", tt.expected.suggCount, len(result.Suggestions), result.Suggestions)
			}

			// Verify stats are populated (only for non-empty patterns)
			if tt.pattern != "" {
				if result.Stats["length"] == nil {
					t.Error("Expected length in stats")
				}
				if result.Stats["word_count"] == nil {
					t.Error("Expected word_count in stats")
				}
				if result.Stats["wildcard_count"] == nil {
					t.Error("Expected wildcard_count in stats")
				}
			}
		})
	}
}

func TestFindInvalidCharacters(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected []string
	}{
		{
			name:     "Valid pattern",
			pattern:  "HELLO WORLD",
			expected: []string{},
		},
		{
			name:     "Invalid character at position 5",
			pattern:  "HELLO@WORLD",
			expected: []string{"'@' at position 5"},
		},
		{
			name:     "Multiple invalid characters",
			pattern:  "HELLO@WORLD#TEST",
			expected: []string{"'@' at position 5"},
		},
		{
			name:     "Invalid character at start",
			pattern:  "@HELLO WORLD",
			expected: []string{"'@' at position 0"},
		},
		{
			name:     "Invalid character at end",
			pattern:  "HELLO WORLD@",
			expected: []string{"'@' at position 11"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findInvalidCharacters(tt.pattern)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d invalid characters, got %d", len(tt.expected), len(result))
			}
			for i, expected := range tt.expected {
				if i < len(result) && result[i] != expected {
					t.Errorf("Expected invalid character %s, got %s", expected, result[i])
				}
			}
		})
	}
}

func TestValidateBalancedTags(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected []string
	}{
		{
			name:     "Balanced tags",
			pattern:  "HELLO <set>WORLD</set>",
			expected: []string{},
		},
		{
			name:     "Unbalanced set tags",
			pattern:  "HELLO <set>WORLD",
			expected: []string{"Unbalanced set tags: 1 opening, 0 closing"},
		},
		{
			name:     "Unbalanced topic tags",
			pattern:  "HELLO <topic>WORLD",
			expected: []string{"Unbalanced topic tags: 1 opening, 0 closing"},
		},
		{
			name:     "Unbalanced alternation groups",
			pattern:  "HELLO (WORLD",
			expected: []string{"Unbalanced alternation groups: 1 opening, 0 closing"},
		},
		{
			name:     "Multiple unbalanced tags",
			pattern:  "HELLO <set>WORLD <topic>TEST",
			expected: []string{"Unbalanced set tags: 1 opening, 0 closing", "Unbalanced topic tags: 1 opening, 0 closing"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateBalancedTags(tt.pattern)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d tag errors, got %d", len(tt.expected), len(result))
			}
			for i, expected := range tt.expected {
				if i < len(result) && result[i] != expected {
					t.Errorf("Expected tag error %s, got %s", expected, result[i])
				}
			}
		})
	}
}

func TestValidatePatternStructure(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected []string
	}{
		{
			name:     "No issues",
			pattern:  "HELLO WORLD",
			expected: []string{},
		},
		{
			name:     "Consecutive wildcards",
			pattern:  "HELLO ** WORLD",
			expected: []string{"Consecutive wildcards detected. This may cause matching issues."},
		},
		{
			name:     "Starts with wildcard",
			pattern:  "* HELLO WORLD",
			expected: []string{"Pattern starts with wildcard. Consider if this is intentional."},
		},
		{
			name:     "Ends with wildcard",
			pattern:  "HELLO WORLD *",
			expected: []string{"Pattern ends with wildcard. Consider if this is intentional."},
		},
		{
			name:     "Very short pattern",
			pattern:  "HI",
			expected: []string{"Very short pattern. Consider if this provides enough specificity."},
		},
		{
			name:     "Very long pattern",
			pattern:  "THIS IS A VERY LONG PATTERN THAT CONTAINS MANY WORDS AND SHOULD TRIGGER A WARNING BECAUSE IT IS OVER TWO HUNDRED CHARACTERS LONG AND SHOULD BE BROKEN DOWN INTO SMALLER MORE SPECIFIC PATTERNS FOR BETTER PERFORMANCE AND MAINTAINABILITY",
			expected: []string{"Very long pattern. Consider breaking into smaller, more specific patterns."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validatePatternStructure(tt.pattern)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d structure warnings, got %d", len(tt.expected), len(result))
			}
			for i, expected := range tt.expected {
				if i < len(result) && result[i] != expected {
					t.Errorf("Expected structure warning %s, got %s", expected, result[i])
				}
			}
		})
	}
}

func TestValidatePatternPerformance(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected []string
	}{
		{
			name:     "No performance issues",
			pattern:  "HELLO WORLD",
			expected: []string{},
		},
		{
			name:     "High wildcard count",
			pattern:  "* * * * * * WORLD",
			expected: []string{"High wildcard count may impact matching performance.", "Word '*' appears 6 times. Consider if this is intentional."},
		},
		{
			name:     "Deeply nested alternation",
			pattern:  "HELLO (WORLD (AND (UNIVERSE)))",
			expected: []string{}, // No performance warning for this depth
		},
		{
			name:     "Repeated words",
			pattern:  "HELLO HELLO HELLO HELLO WORLD",
			expected: []string{"Word 'HELLO' appears 4 times. Consider if this is intentional."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validatePatternPerformance(tt.pattern)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d performance warnings, got %d", len(tt.expected), len(result))
			}
			for i, expected := range tt.expected {
				if i < len(result) && result[i] != expected {
					t.Errorf("Expected performance warning %s, got %s", expected, result[i])
				}
			}
		})
	}
}

func TestGeneratePatternSuggestions(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		stats    map[string]interface{}
		expected []string
	}{
		{
			name:     "No wildcards",
			pattern:  "HELLO WORLD",
			stats:    map[string]interface{}{"wildcard_count": 0, "length": 11, "word_count": 2},
			expected: []string{"Consider adding wildcards (*, _, ^, #) for more flexible matching."},
		},
		{
			name:     "Too many wildcards",
			pattern:  "* * * * * * *",
			stats:    map[string]interface{}{"wildcard_count": 7, "length": 13, "word_count": 7},
			expected: []string{"Consider reducing wildcards for more specific matching."},
		},
		{
			name:     "Short pattern",
			pattern:  "HI",
			stats:    map[string]interface{}{"wildcard_count": 0, "length": 2, "word_count": 1},
			expected: []string{"Consider adding wildcards (*, _, ^, #) for more flexible matching.", "Short patterns may match too broadly. Consider adding more context.", "Single-word patterns are very broad. Consider adding context words.", "Consider adding spaces between words for better readability."},
		},
		{
			name:     "Long pattern",
			pattern:  "THIS IS A VERY LONG PATTERN WITH MANY WORDS",
			stats:    map[string]interface{}{"wildcard_count": 0, "length": 120, "word_count": 8},
			expected: []string{"Consider adding wildcards (*, _, ^, #) for more flexible matching.", "Long patterns may be too specific. Consider using wildcards for flexibility."},
		},
		{
			name:     "Multiple spaces",
			pattern:  "HELLO    WORLD",
			stats:    map[string]interface{}{"wildcard_count": 0, "length": 13, "word_count": 2},
			expected: []string{"Consider adding wildcards (*, _, ^, #) for more flexible matching.", "Multiple consecutive spaces detected. Consider normalizing whitespace."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generatePatternSuggestions(tt.pattern, tt.stats)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d suggestions, got %d", len(tt.expected), len(result))
			}
			for i, expected := range tt.expected {
				if i < len(result) && result[i] != expected {
					t.Errorf("Expected suggestion %s, got %s", expected, result[i])
				}
			}
		})
	}
}

func TestThatPatternValidationIntegration(t *testing.T) {
	// Test complex pattern with multiple issues
	pattern := "HELLO WORLD"
	result := ValidateThatPatternDetailed(pattern)

	// Should have suggestions
	if len(result.Suggestions) < 1 {
		t.Errorf("Expected suggestions, got %d", len(result.Suggestions))
	}

	// Verify stats are comprehensive
	expectedStats := []string{"length", "word_count", "wildcard_count", "wildcard_types"}
	for _, stat := range expectedStats {
		if result.Stats[stat] == nil {
			t.Errorf("Expected stat %s to be present", stat)
		}
	}

	// Verify wildcard types breakdown
	wildcardTypes, ok := result.Stats["wildcard_types"].(map[string]int)
	if !ok {
		t.Error("Expected wildcard_types to be a map")
	} else {
		expectedTypes := []string{"star", "underscore", "caret", "hash", "dollar"}
		for _, wcType := range expectedTypes {
			if _, exists := wildcardTypes[wcType]; !exists {
				t.Errorf("Expected wildcard type %s to be present", wcType)
			}
		}
	}
}

func TestThatPatternValidationEdgeCases(t *testing.T) {
	// Test empty pattern
	result := ValidateThatPatternDetailed("")
	if result.IsValid {
		t.Error("Expected empty pattern to be invalid")
	}
	if len(result.Errors) == 0 {
		t.Error("Expected error for empty pattern")
	}

	// Test pattern with only spaces
	result = ValidateThatPatternDetailed("   ")
	if !result.IsValid {
		t.Error("Expected whitespace-only pattern to be valid")
	}

	// Test pattern with only wildcards
	result = ValidateThatPatternDetailed("* * *")
	if !result.IsValid {
		t.Error("Expected wildcard-only pattern to be valid")
	}
	if len(result.Warnings) == 0 {
		t.Error("Expected warnings for wildcard-only pattern")
	}

	// Test pattern with maximum wildcards
	result = ValidateThatPatternDetailed("* * * * * * * * *")
	if !result.IsValid {
		t.Error("Expected pattern with 9 wildcards to be valid")
	}

	// Test pattern with too many wildcards
	result = ValidateThatPatternDetailed("* * * * * * * * * * *")
	if result.IsValid {
		t.Error("Expected pattern with 11 wildcards to be invalid")
	}
}
