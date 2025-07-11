package parser

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

// Category represents an AIML category entry
type Category struct {
	Pattern  string         `xml:"pattern"`
	That     string         `xml:"that"`
	Topic    string         `xml:"topic"`
	Template string         `xml:"template,innerxml"`
	Unknown  []xml.CharData `xml:",any"` // catch-all for unknown tags (fallback)
}

// AIML represents the root AIML element
type AIML struct {
	Categories []Category `xml:"category"`
}

// Parser represents an AIML parser
type Parser struct {
	debug bool
}

// NewParser creates a new AIML parser
func NewParser(debug bool) *Parser {
	return &Parser{
		debug: debug,
	}
}

// ParseFile parses an AIML file and returns the categories
func (p *Parser) ParseFile(filename string) ([]Category, error) {
	if p.debug {
		fmt.Fprintf(os.Stderr, "Parsing AIML file: %s\n", filename)
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return p.ParseReader(file)
}

// ParseReader parses AIML content from an io.Reader
func (p *Parser) ParseReader(reader io.Reader) ([]Category, error) {
	var aiml AIML

	decoder := xml.NewDecoder(reader)
	decoder.Strict = false // Allow malformed XML

	err := decoder.Decode(&aiml)
	if err != nil {
		return nil, fmt.Errorf("failed to decode XML: %w", err)
	}

	// Validate and clean up categories
	var validCategories []Category
	for i, category := range aiml.Categories {
		if p.validateCategory(category, i) {
			// Clean up whitespace
			category.Pattern = strings.TrimSpace(category.Pattern)
			category.That = strings.TrimSpace(category.That)
			category.Topic = strings.TrimSpace(category.Topic)
			tempStr := strings.TrimSpace(category.Template)
			category.Template = tempStr

			validCategories = append(validCategories, category)
		}
	}

	if p.debug {
		fmt.Fprintf(os.Stderr, "Parsed %d valid categories from %d total entries\n",
			len(validCategories), len(aiml.Categories))
	}

	return validCategories, nil
}

// validateCategory checks if a category is valid and logs warnings for malformed entries
func (p *Parser) validateCategory(category Category, index int) bool {
	// Check for required pattern
	if strings.TrimSpace(category.Pattern) == "" {
		if p.debug {
			fmt.Fprintf(os.Stderr, "Warning: Category %d has empty pattern, skipping\n", index)
		}
		return false
	}

	// Check for required template
	templateStr := strings.TrimSpace(category.Template)
	if templateStr == "" {
		if p.debug {
			fmt.Fprintf(os.Stderr, "Warning: Category %d has empty template, skipping\n", index)
		}
		return false
	}

	return true
}
