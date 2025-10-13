package golem

import (
	"math/rand"
	"regexp"
	"strings"
	"unicode"
)

// TextProcessing provides text manipulation utilities for AIML templates
type TextProcessing struct {
	golem *Golem
}

// NewTextProcessing creates a new text processing instance
func NewTextProcessing(golem *Golem) *TextProcessing {
	return &TextProcessing{golem: golem}
}

// SubstitutePronouns performs pronoun substitution for person tags
func (tp *TextProcessing) SubstitutePronouns(text string) string {
	// Common pronoun substitutions
	substitutions := map[string]string{
		"i am":     "you are",
		"i'm":      "you're",
		"i was":    "you were",
		"i have":   "you have",
		"i've":     "you've",
		"i will":   "you will",
		"i'll":     "you'll",
		"i would":  "you would",
		"i'd":      "you'd",
		"i can":    "you can",
		"i could":  "you could",
		"i should": "you should",
		"i must":   "you must",
		"i need":   "you need",
		"i want":   "you want",
		"i like":   "you like",
		"i love":   "you love",
		"i hate":   "you hate",
		"i think":  "you think",
		"i know":   "you know",
		"i feel":   "you feel",
		"i see":    "you see",
		"i hear":   "you hear",
		"i smell":  "you smell",
		"i taste":  "you taste",
		"i touch":  "you touch",
		"my":       "your",
		"mine":     "yours",
		"myself":   "yourself",
		"me":       "you",
	}

	result := text
	for old, new := range substitutions {
		// Case-insensitive replacement
		re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(old) + `\b`)
		result = re.ReplaceAllString(result, new)
	}

	return result
}

// SubstitutePronouns2 performs first-to-third person pronoun substitution
func (tp *TextProcessing) SubstitutePronouns2(text string) string {
	substitutions := map[string]string{
		"i am":     "he is",
		"i'm":      "he's",
		"i was":    "he was",
		"i have":   "he has",
		"i've":     "he's",
		"i will":   "he will",
		"i'll":     "he'll",
		"i would":  "he would",
		"i'd":      "he'd",
		"i can":    "he can",
		"i could":  "he could",
		"i should": "he should",
		"i must":   "he must",
		"i need":   "he needs",
		"i want":   "he wants",
		"i like":   "he likes",
		"i love":   "he loves",
		"i hate":   "he hates",
		"i think":  "he thinks",
		"i know":   "he knows",
		"i feel":   "he feels",
		"i see":    "he sees",
		"i hear":   "he hears",
		"i smell":  "he smells",
		"i taste":  "he tastes",
		"i touch":  "he touches",
		"my":       "his",
		"mine":     "his",
		"myself":   "himself",
		"me":       "him",
	}

	result := text
	for old, new := range substitutions {
		re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(old) + `\b`)
		result = re.ReplaceAllString(result, new)
	}

	return result
}

// SubstituteGenderPronouns performs gender pronoun substitution
func (tp *TextProcessing) SubstituteGenderPronouns(text string) string {
	substitutions := map[string]string{
		"he":        "she",
		"him":       "her",
		"his":       "her",
		"himself":   "herself",
		"boy":       "girl",
		"man":       "woman",
		"guy":       "gal",
		"gentleman": "lady",
		"father":    "mother",
		"dad":       "mom",
		"daddy":     "mommy",
		"son":       "daughter",
		"brother":   "sister",
		"uncle":     "aunt",
		"nephew":    "niece",
		"husband":   "wife",
		"boyfriend": "girlfriend",
		"actor":     "actress",
		"waiter":    "waitress",
		"host":      "hostess",
		"prince":    "princess",
		"king":      "queen",
		"duke":      "duchess",
		"earl":      "countess",
		"baron":     "baroness",
		"lord":      "lady",
		"sir":       "madam",
		"mr":        "ms",
		"mister":    "miss",
	}

	result := text
	for old, new := range substitutions {
		re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(old) + `\b`)
		result = re.ReplaceAllString(result, new)
	}

	return result
}

// UppercaseTextPreservingTags converts text to uppercase while preserving XML tags
func (tp *TextProcessing) UppercaseTextPreservingTags(input string) string {
	var result strings.Builder
	inTag := false

	for _, char := range input {
		if char == '<' {
			inTag = true
		} else if char == '>' {
			inTag = false
		}

		if inTag {
			result.WriteRune(char)
		} else {
			result.WriteRune(unicode.ToUpper(char))
		}
	}

	return result.String()
}

// FormatFormalText formats text in a formal style
func (tp *TextProcessing) FormatFormalText(input string) string {
	// Basic formal formatting - capitalize first letter of each sentence
	sentences := strings.Split(input, ".")
	var result strings.Builder

	for i, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}

		if i > 0 {
			result.WriteString(". ")
		}

		// Capitalize first letter
		if len(sentence) > 0 {
			first := unicode.ToUpper(rune(sentence[0]))
			rest := sentence[1:]
			result.WriteRune(first)
			result.WriteString(rest)
		}
	}

	return result.String()
}

// ExplodeText separates characters with spaces
func (tp *TextProcessing) ExplodeText(input string) string {
	var result strings.Builder
	for i, char := range input {
		if i > 0 {
			result.WriteRune(' ')
		}
		result.WriteRune(char)
	}
	return result.String()
}

// CapitalizeText capitalizes the first letter of each word
func (tp *TextProcessing) CapitalizeText(input string) string {
	words := strings.Fields(input)
	var result strings.Builder

	for i, word := range words {
		if i > 0 {
			result.WriteRune(' ')
		}

		if len(word) > 0 {
			first := unicode.ToUpper(rune(word[0]))
			rest := strings.ToLower(word[1:])
			result.WriteRune(first)
			result.WriteString(rest)
		}
	}

	return result.String()
}

// ReverseText reverses the order of characters
func (tp *TextProcessing) ReverseText(input string) string {
	runes := []rune(input)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// CreateAcronym creates an acronym from the first letters of words
func (tp *TextProcessing) CreateAcronym(input string) string {
	words := strings.Fields(input)
	var result strings.Builder

	for _, word := range words {
		if len(word) > 0 {
			result.WriteRune(unicode.ToUpper(rune(word[0])))
		}
	}

	return result.String()
}

// TrimText removes leading and trailing whitespace
func (tp *TextProcessing) TrimText(input string) string {
	return strings.TrimSpace(input)
}

// ExtractSubstring extracts a substring between start and end strings
func (tp *TextProcessing) ExtractSubstring(input, startStr, endStr string) string {
	startIdx := strings.Index(input, startStr)
	if startIdx == -1 {
		return ""
	}

	startIdx += len(startStr)
	endIdx := strings.Index(input[startIdx:], endStr)
	if endIdx == -1 {
		return input[startIdx:]
	}

	return input[startIdx : startIdx+endIdx]
}

// ReplaceText replaces all occurrences of search with replace
func (tp *TextProcessing) ReplaceText(input, search, replace string) string {
	return strings.ReplaceAll(input, search, replace)
}

// PluralizeText attempts to pluralize words
func (tp *TextProcessing) PluralizeText(input string) string {
	words := strings.Fields(input)
	var result strings.Builder

	for i, word := range words {
		if i > 0 {
			result.WriteRune(' ')
		}
		result.WriteString(tp.PluralizeWord(word))
	}

	return result.String()
}

// PluralizeWord pluralizes a single word
func (tp *TextProcessing) PluralizeWord(word string) string {
	word = strings.ToLower(word)

	// Handle irregular plurals
	irregulars := map[string]string{
		"child":    "children",
		"man":      "men",
		"woman":    "women",
		"person":   "people",
		"foot":     "feet",
		"tooth":    "teeth",
		"mouse":    "mice",
		"goose":    "geese",
		"ox":       "oxen",
		"sheep":    "sheep",
		"deer":     "deer",
		"fish":     "fish",
		"moose":    "moose",
		"series":   "series",
		"species":  "species",
		"aircraft": "aircraft",
	}

	if plural, exists := irregulars[word]; exists {
		return plural
	}

	// Handle words ending in -y
	if strings.HasSuffix(word, "y") && len(word) > 1 {
		lastTwo := word[len(word)-2:]
		if !strings.ContainsAny(lastTwo, "aeiou") {
			return word[:len(word)-1] + "ies"
		}
	}

	// Handle words ending in -s, -sh, -ch, -x, -z
	if strings.HasSuffix(word, "s") || strings.HasSuffix(word, "sh") ||
		strings.HasSuffix(word, "ch") || strings.HasSuffix(word, "x") ||
		strings.HasSuffix(word, "z") {
		return word + "es"
	}

	// Handle words ending in -f or -fe
	if strings.HasSuffix(word, "f") {
		return word[:len(word)-1] + "ves"
	}
	if strings.HasSuffix(word, "fe") {
		return word[:len(word)-2] + "ves"
	}

	// Default case - just add 's'
	return word + "s"
}

// ShuffleText shuffles the order of words
func (tp *TextProcessing) ShuffleText(input string) string {
	words := strings.Fields(input)

	// Fisher-Yates shuffle
	for i := len(words) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		words[i], words[j] = words[j], words[i]
	}

	return strings.Join(words, " ")
}

// CalculateLength calculates the length of content
func (tp *TextProcessing) CalculateLength(content, lengthType string) string {
	switch strings.ToLower(lengthType) {
	case "words":
		words := strings.Fields(content)
		return string(rune(len(words)))
	case "characters":
		return string(rune(len(content)))
	case "sentences":
		sentences := strings.Split(content, ".")
		count := 0
		for _, s := range sentences {
			if strings.TrimSpace(s) != "" {
				count++
			}
		}
		return string(rune(count))
	default:
		return string(rune(len(content)))
	}
}

// CountOccurrences counts occurrences of search string in content
func (tp *TextProcessing) CountOccurrences(content, search string) string {
	count := strings.Count(content, search)
	return string(rune(count))
}

// SplitText splits text by delimiter
func (tp *TextProcessing) SplitText(content, delimiter, limitStr string) string {
	parts := strings.Split(content, delimiter)

	if limitStr != "" {
		// Handle limit if specified
		// For simplicity, just return the first part
		if len(parts) > 0 {
			return parts[0]
		}
	}

	return strings.Join(parts, delimiter)
}

// JoinText joins text with delimiter
func (tp *TextProcessing) JoinText(content, delimiter string) string {
	// For now, just return the content as-is
	// This could be enhanced to handle more complex joining
	return content
}

// IndentText indents text by specified level
func (tp *TextProcessing) IndentText(content string, level int, char string) string {
	if char == "" {
		char = " "
	}

	indent := strings.Repeat(char, level)
	lines := strings.Split(content, "\n")
	var result strings.Builder

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}
		if strings.TrimSpace(line) != "" {
			result.WriteString(indent)
		}
		result.WriteString(line)
	}

	return result.String()
}

// DedentText removes indentation from text
func (tp *TextProcessing) DedentText(content string, level int, char string) string {
	if char == "" {
		char = " "
	}

	dedent := strings.Repeat(char, level)
	lines := strings.Split(content, "\n")
	var result strings.Builder

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}
		if strings.HasPrefix(line, dedent) {
			result.WriteString(line[level:])
		} else {
			result.WriteString(line)
		}
	}

	return result.String()
}

// UniqueText removes duplicate words
func (tp *TextProcessing) UniqueText(content string, delimiter string) string {
	words := strings.Fields(content)
	seen := make(map[string]bool)
	var unique []string

	for _, word := range words {
		if !seen[word] {
			seen[word] = true
			unique = append(unique, word)
		}
	}

	return strings.Join(unique, " ")
}

// RandomInt generates a random integer between 0 and max-1
func (tp *TextProcessing) RandomInt(max int) int {
	if max <= 0 {
		return 0
	}
	return rand.Intn(max)
}

// NormalizeTextForOutput normalizes text for output
func (tp *TextProcessing) NormalizeTextForOutput(input string) string {
	// Basic normalization - trim whitespace and fix common issues
	result := strings.TrimSpace(input)

	// Fix multiple spaces
	re := regexp.MustCompile(`\s+`)
	result = re.ReplaceAllString(result, " ")

	// Fix common punctuation issues
	result = strings.ReplaceAll(result, " .", ".")
	result = strings.ReplaceAll(result, " ,", ",")
	result = strings.ReplaceAll(result, " !", "!")
	result = strings.ReplaceAll(result, " ?", "?")

	return result
}

// DenormalizeText denormalizes text
func (tp *TextProcessing) DenormalizeText(input string) string {
	// Basic denormalization - restore some natural spacing
	result := strings.ReplaceAll(input, ".", ". ")
	result = strings.ReplaceAll(result, ",", ", ")
	result = strings.ReplaceAll(result, "!", "! ")
	result = strings.ReplaceAll(result, "?", "? ")

	// Clean up extra spaces
	re := regexp.MustCompile(`\s+`)
	result = re.ReplaceAllString(result, " ")

	return strings.TrimSpace(result)
}

// SplitSentences splits text into sentences
func (tp *TextProcessing) SplitSentences(text string) []string {
	// Simple sentence splitting on periods
	sentences := strings.Split(text, ".")
	var result []string

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence != "" {
			result = append(result, sentence)
		}
	}

	return result
}

// CapitalizeSentences capitalizes the first letter of each sentence
func (tp *TextProcessing) CapitalizeSentences(text string) string {
	sentences := tp.SplitSentences(text)
	var result strings.Builder

	for i, sentence := range sentences {
		if i > 0 {
			result.WriteString(". ")
		}

		if len(sentence) > 0 {
			first := unicode.ToUpper(rune(sentence[0]))
			rest := sentence[1:]
			result.WriteRune(first)
			result.WriteString(rest)
		}
	}

	return result.String()
}

// CapitalizeWords capitalizes the first letter of each word
func (tp *TextProcessing) CapitalizeWords(text string) string {
	words := strings.Fields(text)
	var result strings.Builder

	for i, word := range words {
		if i > 0 {
			result.WriteRune(' ')
		}

		if len(word) > 0 {
			first := unicode.ToUpper(rune(word[0]))
			rest := strings.ToLower(word[1:])
			result.WriteRune(first)
			result.WriteString(rest)
		}
	}

	return result.String()
}

// IsWord checks if a string is a word (contains only letters)
func (tp *TextProcessing) IsWord(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return len(s) > 0
}
