package engine

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"golem/parser"
)

func TestBotRespondWithoutCategories(t *testing.T) {
	// Create a new bot without loading any AIML
	bot := NewBot(false)

	// Test various inputs
	testCases := []string{
		"Hello",
		"What is your name?",
		"How are you?",
		"Tell me a joke",
		"",
		"   ",
	}

	for _, input := range testCases {
		t.Run("input_"+strings.ReplaceAll(input, " ", "_"), func(t *testing.T) {
			response, err := bot.Respond(input, "test-session")
			
			// Should not return an error
			if err != nil {
				t.Errorf("Bot.Respond() returned an error: %v", err)
			}

			// Should return a default message about no knowledge loaded
			expectedContains := "don't have any knowledge loaded"
			if !strings.Contains(response, expectedContains) {
				t.Errorf("Expected response to contain '%s', got: '%s'", expectedContains, response)
			}

			// Response should not be empty
			if strings.TrimSpace(response) == "" {
				t.Error("Bot.Respond() returned empty response")
			}
		})
	}
}

func TestBotRespondWithDebug(t *testing.T) {
	// Create a bot with debug enabled
	bot := NewBot(true)

	// Insert a category so the bot can respond with 'Hi bar!'
	tmpl := "Hi <set name=\"foo\">bar</set><get name=\"foo\"/>!"
	cat := parser.Category{
		Pattern:  "TEST INPUT",
		That:     "",
		Topic:    "",
		Template: tmpl,
	}
	bot.matchTree.Insert(cat)

	response, err := bot.Respond("test input", "debug-session")
	
	if err != nil {
		t.Errorf("Bot.Respond() returned an error: %v", err)
	}

	if response != "Hi bar!" {
		t.Errorf("Expected response to be 'Hi bar!', got: %s", response)
	}
}

func TestBotLoadAIMLNonExistentFile(t *testing.T) {
	bot := NewBot(false)
	
	err := bot.LoadAIML("/path/to/nonexistent/file.aiml")
	
	if err == nil {
		t.Error("Expected LoadAIML to return an error for non-existent file")
	}
	
	if !strings.Contains(err.Error(), "no such file or directory") {
		t.Errorf("Expected error to contain 'no such file or directory', got: %v", err)
	}
}

func TestNewBot(t *testing.T) {
	// Test creating bot with debug disabled
	bot := NewBot(false)
	if bot == nil {
		t.Error("Expected bot to be non-nil after construction")
	}

	// Test creating bot with debug enabled
	botDebug := NewBot(true)
	if botDebug == nil {
		t.Error("Expected botDebug to be non-nil after construction")
	}
}

func TestBotRespondWithDebugTrace(t *testing.T) {
	bot := NewBot(true)
	// Create a simple AIML category in memory
	// The template includes text before, between, and after tags.
	// The evaluator should preserve the correct text/tag order: 'Hi bar!'
	tmpl := "Hi <set name=\"foo\">bar</set><get name=\"foo\"/>!"
	cat := parser.Category{
		Pattern:  "HELLO",
		That:     "",
		Topic:    "",
		Template: tmpl,
	}
	bot.matchTree.Insert(cat)

	// Print parsed node tree for inspection
	type node struct {
		XMLName xml.Name
		Attr    []xml.Attr    `xml:",any,attr"`
		Nodes   []node        `xml:",any"`
		Text    string        `xml:",chardata"`
	}
	n := node{}
	err := xml.Unmarshal([]byte("<template>"+tmpl+"</template>"), &n)
	if err != nil {
		t.Fatalf("Failed to parse template XML: %v", err)
	}
	fmt.Fprintf(os.Stderr, "Parsed node tree: %+v\n", n)

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Respond
	response, err := bot.Respond("hello", "debug-session")
	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	io.Copy(&buf, r)
	trace := buf.String()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(trace, "Matched category") {
		t.Errorf("Expected trace to contain 'Matched category', got: %s", trace)
	}
	if !strings.Contains(trace, "Wildcard captures") {
		t.Errorf("Expected trace to contain 'Wildcard captures', got: %s", trace)
	}
	if !strings.Contains(trace, "<set>") || !strings.Contains(trace, "<get>") {
		t.Errorf("Expected trace to contain evaluation steps, got: %s", trace)
	}
	if !strings.Contains(trace, "session=debug-session") {
		t.Errorf("Expected trace to contain session context, got: %s", trace)
	}
	// The evaluator now preserves the correct text/tag order, so the output matches AIML expectations.
	if response != "Hi bar!" {
		t.Errorf("Expected response to be 'Hi bar!', got: %s", response)
	}
} 