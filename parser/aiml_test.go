package parser

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestParseFile(t *testing.T) {
	parser := NewParser(true)
	
	categories, err := parser.ParseFile("../testdata/simple.aiml")
	if err != nil {
		t.Fatalf("Failed to parse AIML file: %v", err)
	}

	// Should have 5 valid categories (2 malformed ones should be skipped)
	expectedCount := 5
	if len(categories) != expectedCount {
		t.Errorf("Expected %d categories, got %d", expectedCount, len(categories))
	}

	// Test specific categories
	expectedCategories := []Category{
		{
			Pattern:  "HELLO",
			That:     "",
			Topic:    "",
			Template: "Hello! How are you today?",
		},
		{
			Pattern:  "WHAT IS YOUR NAME",
			That:     "",
			Topic:    "",
			Template: "My name is Golem, nice to meet you!",
		},
		{
			Pattern:  "HOW ARE YOU",
			That:     "HELLO",
			Topic:    "",
			Template: "I'm doing well, thank you for asking!",
		},
		{
			Pattern:  "TELL ME A JOKE",
			That:     "",
			Topic:    "HUMOR",
			Template: "Why don't scientists trust atoms? Because they make up everything!",
		},
		{
			Pattern:  "GOODBYE",
			That:     "",
			Topic:    "",
			Template: "Goodbye! Have a great day!",
		},
	}

	for i, expected := range expectedCategories {
		if i >= len(categories) {
			t.Errorf("Missing category %d", i)
			continue
		}
		
		actual := categories[i]
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Category %d mismatch:\nExpected: %+v\nGot:      %+v", 
				i, expected, actual)
		}
	}

	field, _ := reflect.TypeOf(Category{}).FieldByName("Template")
	fmt.Fprintf(os.Stderr, "Template struct tag at runtime: %q\n", field.Tag)
}

func TestParseReader(t *testing.T) {
	parser := NewParser(false)
	
	aimlContent := `<?xml version="1.0" encoding="UTF-8"?>
<aiml version="1.0.1">
    <category>
        <pattern>TEST PATTERN</pattern>
        <template>Test template response</template>
    </category>
</aiml>`
	
	reader := strings.NewReader(aimlContent)
	categories, err := parser.ParseReader(reader)
	if err != nil {
		t.Fatalf("Failed to parse AIML content: %v", err)
	}

	expectedCount := 1
	if len(categories) != expectedCount {
		t.Errorf("Expected %d categories, got %d", expectedCount, len(categories))
	}

	expected := Category{
		Pattern:  "TEST PATTERN",
		That:     "",
		Topic:    "",
		Template: "Test template response",
	}

	if !reflect.DeepEqual(categories[0], expected) {
		t.Errorf("Category mismatch:\nExpected: %+v\nGot:      %+v", 
			expected, categories[0])
	}

	field, _ := reflect.TypeOf(Category{}).FieldByName("Template")
	fmt.Fprintf(os.Stderr, "Template struct tag at runtime: %q\n", field.Tag)
}

func TestParseFileWithMalformedEntries(t *testing.T) {
	parser := NewParser(true)
	
	// Create a temporary AIML content with malformed entries
	malformedContent := `<?xml version="1.0" encoding="UTF-8"?>
<aiml version="1.0.1">
    <category>
        <pattern>VALID PATTERN</pattern>
        <template>Valid template</template>
    </category>
    <category>
        <pattern></pattern>
        <template>Empty pattern</template>
    </category>
    <category>
        <pattern>EMPTY TEMPLATE</pattern>
        <template></template>
    </category>
    <category>
        <pattern>ANOTHER VALID</pattern>
        <template>Another valid template</template>
    </category>
</aiml>`
	
	reader := strings.NewReader(malformedContent)
	categories, err := parser.ParseReader(reader)
	if err != nil {
		t.Fatalf("Failed to parse AIML content: %v", err)
	}

	// Should only have 2 valid categories
	expectedCount := 2
	if len(categories) != expectedCount {
		t.Errorf("Expected %d categories, got %d", expectedCount, len(categories))
	}

	// Check that only valid categories are included
	validPatterns := []string{"VALID PATTERN", "ANOTHER VALID"}
	for i, category := range categories {
		if category.Pattern != validPatterns[i] {
			t.Errorf("Expected pattern %s, got %s", validPatterns[i], category.Pattern)
		}
	}

	field, _ := reflect.TypeOf(Category{}).FieldByName("Template")
	fmt.Fprintf(os.Stderr, "Template struct tag at runtime: %q\n", field.Tag)
}

func TestParseFileNonExistent(t *testing.T) {
	parser := NewParser(false)
	
	_, err := parser.ParseFile("nonexistent.aiml")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
	
	if !strings.Contains(err.Error(), "failed to open file") {
		t.Errorf("Expected error to contain 'failed to open file', got: %v", err)
	}
}

func TestNewParser(t *testing.T) {
	// Test creating parser with debug disabled
	parser := NewParser(false)
	if parser.debug {
		t.Error("Expected debug to be false")
	}
	
	// Test creating parser with debug enabled
	parserDebug := NewParser(true)
	if !parserDebug.debug {
		t.Error("Expected debug to be true")
	}
}

func TestCategoryValidation(t *testing.T) {
	parser := NewParser(true)
	
	testCases := []struct {
		name     string
		category Category
		expected bool
	}{
		{
			name: "valid category",
			category: Category{
				Pattern:  "TEST",
				Template: "Response",
			},
			expected: true,
		},
		{
			name: "empty pattern",
			category: Category{
				Pattern:  "",
				Template: "Response",
			},
			expected: false,
		},
		{
			name: "whitespace pattern",
			category: Category{
				Pattern:  "   ",
				Template: "Response",
			},
			expected: false,
		},
		{
			name: "empty template",
			category: Category{
				Pattern:  "TEST",
				Template: "",
			},
			expected: false,
		},
		{
			name: "whitespace template",
			category: Category{
				Pattern:  "TEST",
				Template: "   ",
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parser.validateCategory(tc.category, 0)
			if result != tc.expected {
				t.Errorf("Expected validation result %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestParseTemplateWithInnerXML(t *testing.T) {
	parser := NewParser(false)

	aimlContent := `<?xml version="1.0" encoding="UTF-8"?>
<aiml version="1.0.1">
    <category>
        <pattern>WHAT THE *</pattern>
        <template><srai>WHAT IS THE <star/></srai></template>
    </category>
</aiml>`

	reader := strings.NewReader(aimlContent)
	categories, err := parser.ParseReader(reader)
	if err != nil {
		t.Fatalf("Failed to parse AIML content: %v", err)
	}

	if len(categories) != 1 {
		t.Fatalf("Expected 1 category, got %d", len(categories))
	}

	expectedInnerXML := `<srai>WHAT IS THE <star/></srai>`
	if categories[0].Template != expectedInnerXML {
		t.Errorf("Expected template inner XML %q, got %q", expectedInnerXML, categories[0].Template)
	}

	field, _ := reflect.TypeOf(Category{}).FieldByName("Template")
	fmt.Fprintf(os.Stderr, "Template struct tag at runtime: %q\n", field.Tag)
} 