package main

import (
	"fmt"
	"regexp"
	"strings"
)

// Simple Category struct for testing
type Category struct {
	Pattern  string
	Template string
}

// Simple AIMLKnowledgeBase for testing
type AIMLKnowledgeBase struct {
	Categories []Category
	Patterns   map[string]*Category
	Sets       map[string][]string
}

func NewAIMLKnowledgeBase() *AIMLKnowledgeBase {
	return &AIMLKnowledgeBase{
		Categories: []Category{},
		Patterns:   make(map[string]*Category),
		Sets:       make(map[string][]string),
	}
}

// NormalizePattern normalizes input for pattern matching
func NormalizePattern(input string) string {
	// Convert to uppercase and trim whitespace
	normalized := strings.ToUpper(strings.TrimSpace(input))

	// Replace multiple spaces with single space
	spaceRegex := regexp.MustCompile(`\s+`)
	normalized = spaceRegex.ReplaceAllString(normalized, " ")

	return normalized
}

// patternToRegexWithSets converts AIML pattern to regex with proper set matching
func patternToRegexWithSets(pattern string, kb *AIMLKnowledgeBase) string {
	// Handle set matching with proper set validation
	setPattern := regexp.MustCompile(`<set>([^<]+)</set>`)
	pattern = setPattern.ReplaceAllStringFunc(pattern, func(match string) string {
		// Extract set name using regex groups
		matches := setPattern.FindStringSubmatch(match)
		if len(matches) < 2 {
			return "([^\\s]*)"
		}
		setName := strings.ToUpper(strings.TrimSpace(matches[1]))
		if len(kb.Sets[setName]) > 0 {
			// Create regex alternation for set members
			var alternatives []string
			for _, member := range kb.Sets[setName] {
				// Escape only specific regex characters, not the pipe
				upperMember := strings.ToUpper(member)
				// Escape characters that have special meaning in regex, but not |
				escaped := strings.ReplaceAll(upperMember, "(", "\\(")
				escaped = strings.ReplaceAll(escaped, ")", "\\)")
				escaped = strings.ReplaceAll(escaped, "[", "\\[")
				escaped = strings.ReplaceAll(escaped, "]", "\\]")
				escaped = strings.ReplaceAll(escaped, "{", "\\{")
				escaped = strings.ReplaceAll(escaped, "}", "\\}")
				escaped = strings.ReplaceAll(escaped, "^", "\\^")
				escaped = strings.ReplaceAll(escaped, "$", "\\$")
				escaped = strings.ReplaceAll(escaped, ".", "\\.")
				escaped = strings.ReplaceAll(escaped, "+", "\\+")
				escaped = strings.ReplaceAll(escaped, "?", "\\?")
				escaped = strings.ReplaceAll(escaped, "*", "\\*")
				escaped = strings.ReplaceAll(escaped, "-", "\\-")
				escaped = strings.ReplaceAll(escaped, "@", "\\@")
				// Don't escape | as it's needed for alternation
				alternatives = append(alternatives, escaped)
			}
			return "(" + strings.Join(alternatives, "|") + ")"
		}
		// Fallback to wildcard if set not found
		return "([^\\s]*)"
	})

	// Handle topic matching
	topicPattern := regexp.MustCompile(`<topic>([^<]+)</topic>`)
	pattern = topicPattern.ReplaceAllString(pattern, "([^\\s]*)")

	// Build regex pattern by processing each character
	var result strings.Builder
	inAlternationGroup := false
	for i, char := range pattern {
		switch char {
		case '*':
			// Zero+ wildcard: matches zero or more words
			result.WriteString("(.*?)")
		case '_':
			// Single wildcard: matches exactly one word
			result.WriteString("([^\\s]+)")
		case '^':
			// Caret wildcard: matches zero or more words (AIML2)
			result.WriteString("(.*?)")
		case '#':
			// Hash wildcard: matches zero or more words with high priority (AIML2)
			result.WriteString("(.*?)")
		case '$':
			// Dollar wildcard: highest priority exact match (AIML2)
			// For regex purposes, treat as exact match (no wildcard capture)
			// Don't add anything to regex - this will be handled in pattern matching
			continue
		case ' ':
			// Check if this space is followed by a wildcard or preceded by a wildcard
			if (i+1 < len(pattern) && (pattern[i+1] == '*' || pattern[i+1] == '_' || pattern[i+1] == '^' || pattern[i+1] == '#')) ||
				(i > 0 && (pattern[i-1] == '*' || pattern[i-1] == '_' || pattern[i-1] == '^' || pattern[i-1] == '#')) {
				// This space is adjacent to a wildcard, make it optional
				result.WriteString(" ?")
			} else {
				// Regular space
				result.WriteRune(' ')
			}
		case '(':
			// Check if this is the start of an alternation group (contains |)
			// Look ahead to see if there's a | in this group
			groupEnd := findMatchingParen(pattern, i)
			if groupEnd > i && strings.Contains(pattern[i:groupEnd+1], "|") {
				inAlternationGroup = true
				result.WriteRune('(')
			} else {
				// Regular group, escape it
				result.WriteString("\\(")
			}
		case ')':
			if inAlternationGroup {
				inAlternationGroup = false
				result.WriteRune(')')
			} else {
				result.WriteString("\\)")
			}
		case '[', ']', '{', '}', '?', '+', '.':
			// Escape special regex characters
			result.WriteRune('\\')
			result.WriteRune(char)
		case '|':
			// Don't escape pipe character as it's needed for alternation in sets
			result.WriteRune(char)
		default:
			// Regular character
			result.WriteRune(char)
		}
	}

	return "^" + result.String() + "$"
}

// findMatchingParen finds the matching closing parenthesis for an opening parenthesis
func findMatchingParen(pattern string, openPos int) int {
	if openPos >= len(pattern) || pattern[openPos] != '(' {
		return -1
	}

	depth := 1
	for i := openPos + 1; i < len(pattern); i++ {
		switch pattern[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// matchPatternWithWildcardsAndSetsCasePreserving matches input against a pattern with wildcards and sets
func matchPatternWithWildcardsAndSetsCasePreserving(normalizedInput, originalInput, pattern string, kb *AIMLKnowledgeBase) (bool, map[string]string) {
	wildcards := make(map[string]string)

	// Convert pattern to regex with set support
	regexPattern := patternToRegexWithSets(pattern, kb)
	fmt.Printf("DEBUG: pattern='%s' -> regex='%s'\n", pattern, regexPattern)

	re, err := regexp.Compile(regexPattern)
	if err != nil {
		fmt.Printf("DEBUG: Error compiling regex: %v\n", err)
		return false, nil
	}

	matches := re.FindStringSubmatch(normalizedInput)
	fmt.Printf("DEBUG: normalizedInput='%s', matches=%v\n", normalizedInput, matches)

	if matches == nil {
		return false, nil
	}

	// First extract wildcards from normalized input (fallback/default behavior)
	starIndex := 1
	for _, match := range matches[1:] {
		wildcards[fmt.Sprintf("star%d", starIndex)] = match
		starIndex++
	}

	// If original input is different from normalized input, try case-preserving extraction
	if originalInput != normalizedInput {
		fmt.Printf("DEBUG: originalInput != normalizedInput, trying case-preserving extraction\n")
		// Extract wildcard values from the original input for case preservation
		// We need to find the wildcard positions in the original input
		originalNormalized := NormalizePattern(originalInput) // This should be the same as normalizedInput
		fmt.Printf("DEBUG: originalNormalized='%s'\n", originalNormalized)
		// Convert pattern to lowercase for matching against case-preserved input
		lowercasePattern := strings.ToLower(pattern)
		lowercaseRegexPattern := patternToRegexWithSets(lowercasePattern, kb)
		lowercaseRe, err := regexp.Compile(lowercaseRegexPattern)
		if err == nil {
			// Match against the case-preserved input
			casePreservedMatches := lowercaseRe.FindStringSubmatch(originalNormalized)
			fmt.Printf("DEBUG: casePreservedMatches=%v\n", casePreservedMatches)
			if casePreservedMatches != nil && len(casePreservedMatches) > 1 {
				// Overwrite with case-preserved values
				starIndex := 1
				for _, match := range casePreservedMatches[1:] {
					wildcards[fmt.Sprintf("star%d", starIndex)] = match
					starIndex++
				}
			}
		}
	}

	return true, wildcards
}

// MatchPattern attempts to match user input against AIML patterns
func (kb *AIMLKnowledgeBase) MatchPattern(input string) (*Category, map[string]string, error) {
	fmt.Printf("DEBUG: MatchPattern called with input='%s'\n", input)

	// Normalize the input for pattern matching
	normalizedInput := NormalizePattern(input)
	fmt.Printf("DEBUG: normalizedInput='%s'\n", normalizedInput)

	// Try to match against each pattern
	for pattern, category := range kb.Patterns {
		fmt.Printf("DEBUG: Trying pattern '%s'\n", pattern)
		matched, wildcards := matchPatternWithWildcardsAndSetsCasePreserving(normalizedInput, input, pattern, kb)
		if matched {
			fmt.Printf("DEBUG: Pattern '%s' matched with wildcards %v\n", pattern, wildcards)
			return category, wildcards, nil
		}
	}

	return nil, nil, fmt.Errorf("no matching pattern found")
}

func main() {
	// Create a test knowledge base
	kb := NewAIMLKnowledgeBase()
	kb.Categories = []Category{
		{
			Pattern:  "HELLO",
			Template: "Hello! How can I help you?",
		},
		{
			Pattern:  "MY NAME IS *",
			Template: "Nice to meet you, <star/>!",
		},
		{
			Pattern:  "I AM * YEARS OLD",
			Template: "You're <star/> years old!",
		},
	}

	// Index patterns
	for i := range kb.Categories {
		category := &kb.Categories[i]
		pattern := category.Pattern
		kb.Patterns[pattern] = category
	}

	// Test wildcard match
	fmt.Println("=== Testing wildcard match ===")
	category, wildcards, err := kb.MatchPattern("MY NAME IS JOHN")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Matched pattern: %s\n", category.Pattern)
	fmt.Printf("Wildcards: %v\n", wildcards)

	if wildcards["star1"] != "JOHN" {
		fmt.Printf("ERROR: Expected wildcard 'JOHN', got '%s'\n", wildcards["star1"])
	} else {
		fmt.Println("SUCCESS: Wildcard extraction working correctly")
	}
}
