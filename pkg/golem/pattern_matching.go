package golem

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// PatternMatching provides pattern matching and validation utilities
type PatternMatching struct {
	golem *Golem
}

// NewPatternMatching creates a new pattern matching instance
func NewPatternMatching(golem *Golem) *PatternMatching {
	return &PatternMatching{golem: golem}
}

// ComparePatternPriorities compares two pattern priorities
func (pm *PatternMatching) ComparePatternPriorities(p1, p2 int) bool {
	// Lower priority number = higher priority
	return p1 < p2
}

// CalculatePatternPriority calculates the priority of a pattern
func (pm *PatternMatching) CalculatePatternPriority(pattern string) PatternPriorityInfo {
	return pm.CalculatePatternPriorityInternal(pattern)
}

// CalculatePatternPriorityCached calculates pattern priority with caching
func (pm *PatternMatching) CalculatePatternPriorityCached(pattern string) PatternPriorityInfo {
	// For now, just calculate without caching
	// TODO: Implement caching when patternPriorityCache is available
	return pm.CalculatePatternPriorityInternal(pattern)
}

// CalculatePatternPriorityInternal calculates pattern priority without caching
func (pm *PatternMatching) CalculatePatternPriorityInternal(pattern string) PatternPriorityInfo {
	priority := 0
	wildcardCount := 0
	hasUnderscore := false
	wildcardPosition := 0

	// Count wildcards and determine position
	for i, char := range pattern {
		if char == '*' {
			wildcardCount++
			if wildcardPosition == 0 {
				wildcardPosition = i
			}
		} else if char == '_' {
			hasUnderscore = true
		}
	}

	// Calculate priority based on various factors
	priority = wildcardCount * 1000 // More wildcards = lower priority

	if hasUnderscore {
		priority += 100 // Underscore patterns have lower priority
	}

	priority += wildcardPosition // Earlier wildcards = lower priority

	// Exact matches have highest priority
	if wildcardCount == 0 {
		priority = 0
	}

	return PatternPriorityInfo{
		Priority:         priority,
		WildcardCount:    wildcardCount,
		HasUnderscore:    hasUnderscore,
		WildcardPosition: wildcardPosition,
	}
}

// SortPatternsByPriority sorts patterns by their priority
func (pm *PatternMatching) SortPatternsByPriority(patterns []PatternPriority) {
	// Use the existing sort function from aiml_native.go
	sortPatternsByPriority(patterns)
}

// MatchPatternWithWildcards matches a pattern with wildcards
func (pm *PatternMatching) MatchPatternWithWildcards(input, pattern string) (bool, map[string]string) {
	wildcards := make(map[string]string)

	// Convert pattern to regex
	regexPattern := pm.PatternToRegex(pattern)
	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return false, wildcards
	}

	// Check if input matches
	matches := re.FindStringSubmatch(input)
	if matches == nil {
		return false, wildcards
	}

	// Extract wildcard values
	subexpNames := re.SubexpNames()
	for i, match := range matches[1:] { // Skip the full match
		if i < len(subexpNames) && subexpNames[i+1] != "" {
			wildcards[subexpNames[i+1]] = match
		} else {
			wildcards[fmt.Sprintf("star%d", i+1)] = match
		}
	}

	return true, wildcards
}

// PatternToRegex converts an AIML pattern to a regex
func (pm *PatternMatching) PatternToRegex(pattern string) string {
	// Escape special regex characters except wildcards
	escaped := regexp.QuoteMeta(pattern)

	// Replace escaped wildcards with regex wildcards
	escaped = strings.ReplaceAll(escaped, "\\*", "(.*?)")
	escaped = strings.ReplaceAll(escaped, "\\_", "(.*?)")

	// Add anchors
	return "^" + escaped + "$"
}

// PatternToRegexWithSets converts pattern to regex with set support
func (pm *PatternMatching) PatternToRegexWithSets(pattern string, kb *AIMLKnowledgeBase) string {
	// For now, just convert without caching
	// TODO: Implement caching when patternRegexCache is available
	return pm.PatternToRegexWithSetsInternal(pattern, kb)
}

// PatternToRegexWithSetsCached converts pattern with caching
func (pm *PatternMatching) PatternToRegexWithSetsCached(pattern string, kb *AIMLKnowledgeBase) string {
	// For now, just convert without caching
	// TODO: Implement caching when patternRegexCache is available
	return pm.PatternToRegexWithSetsInternal(pattern, kb)
}

// PatternToRegexWithSetsInternal converts pattern without caching
func (pm *PatternMatching) PatternToRegexWithSetsInternal(pattern string, kb *AIMLKnowledgeBase) string {
	// Basic pattern to regex conversion
	// This is a simplified version - the full implementation would handle sets
	escaped := regexp.QuoteMeta(pattern)
	escaped = strings.ReplaceAll(escaped, "\\*", "(.*?)")
	escaped = strings.ReplaceAll(escaped, "\\_", "(.*?)")

	return "^" + escaped + "$"
}

// FindMatchingParen finds the matching parenthesis
func (pm *PatternMatching) FindMatchingParen(pattern string, openPos int) int {
	if openPos >= len(pattern) || pattern[openPos] != '(' {
		return -1
	}

	depth := 1
	for i := openPos + 1; i < len(pattern); i++ {
		if pattern[i] == '(' {
			depth++
		} else if pattern[i] == ')' {
			depth--
			if depth == 0 {
				return i
			}
		}
	}

	return -1
}

// ValidatePattern validates an AIML pattern
func (pm *PatternMatching) ValidatePattern(pattern string) error {
	// Check for balanced parentheses
	openCount := 0
	for _, char := range pattern {
		if char == '(' {
			openCount++
		} else if char == ')' {
			openCount--
			if openCount < 0 {
				return fmt.Errorf("unbalanced parentheses in pattern: %s", pattern)
			}
		}
	}

	if openCount != 0 {
		return fmt.Errorf("unbalanced parentheses in pattern: %s", pattern)
	}

	// Check for valid characters
	for _, char := range pattern {
		if char == '*' || char == '_' {
			continue
		}
		if !isValidPatternChar(char) {
			return fmt.Errorf("invalid character in pattern: %c", char)
		}
	}

	return nil
}

// isValidPatternChar checks if a character is valid in a pattern
func isValidPatternChar(char rune) bool {
	return unicode.IsLetter(char) || unicode.IsDigit(char) ||
		char == ' ' || char == '(' || char == ')' ||
		char == '*' || char == '_' || char == '^' ||
		char == '$' || char == '.' || char == '?' ||
		char == '+' || char == '|' || char == '[' ||
		char == ']' || char == '{' || char == '}'
}
