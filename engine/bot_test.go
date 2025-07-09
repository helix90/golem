package engine

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"golem/parser"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper to load categories into the bot as if parsed from AIML
func loadCategories(bot *Bot, cats []parser.Category) {
	for _, cat := range cats {
		bot.matchTree.Insert(cat)
	}
}

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
	loadCategories(bot, []parser.Category{cat})

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
	loadCategories(bot, []parser.Category{cat})

	// Print parsed node tree for inspection
	type node struct {
		XMLName xml.Name
		Attr    []xml.Attr `xml:",any,attr"`
		Nodes   []node     `xml:",any"`
		Text    string     `xml:",chardata"`
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

func TestBotLoadSet_JSONAndLegacy(t *testing.T) {
	bot := NewBot(false)

	// JSON array
	jsonSet := `["FOO", "BAR", "BAZ"]`
	f, err := os.CreateTemp("", "testsetjson-*.set")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	f.WriteString(jsonSet)
	f.Close()
	if err := bot.LoadSet(f.Name()); err != nil {
		t.Fatalf("LoadSet JSON failed: %v", err)
	}
	setName := strings.ToUpper(strings.TrimSuffix(filepath.Base(f.Name()), filepath.Ext(f.Name())))
	for _, v := range []string{"FOO", "BAR", "BAZ"} {
		if _, ok := bot.Sets[setName][v]; !ok {
			t.Errorf("Expected set to contain %q", v)
		}
	}

	// Legacy line-by-line
	legacySet := "ONE\nTWO\nTHREE\n"
	f2, err := os.CreateTemp("", "testsetlegacy-*.set")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(f2.Name())
	f2.WriteString(legacySet)
	f2.Close()
	if err := bot.LoadSet(f2.Name()); err != nil {
		t.Fatalf("LoadSet legacy failed: %v", err)
	}
	setName2 := strings.ToUpper(strings.TrimSuffix(filepath.Base(f2.Name()), filepath.Ext(f2.Name())))
	for _, v := range []string{"ONE", "TWO", "THREE"} {
		if _, ok := bot.Sets[setName2][v]; !ok {
			t.Errorf("Expected legacy set to contain %q", v)
		}
	}
}

func TestBotLoadMap_JSONAndLegacy(t *testing.T) {
	bot := NewBot(false)

	// JSON object
	jsonMap := `{"A": "1", "B": "2", "C": "3"}`
	f, err := os.CreateTemp("", "testmapjson-*.map")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	f.WriteString(jsonMap)
	f.Close()
	if err := bot.LoadMap(f.Name()); err != nil {
		t.Fatalf("LoadMap JSON failed: %v", err)
	}
	mapName := strings.ToUpper(strings.TrimSuffix(filepath.Base(f.Name()), filepath.Ext(f.Name())))
	for k, v := range map[string]string{"A": "1", "B": "2", "C": "3"} {
		if got, ok := bot.Maps[mapName][k]; !ok || got != v {
			t.Errorf("Expected map[%q]=%q, got %q", k, v, got)
		}
	}

	// Legacy line-by-line
	legacyMap := "X 10\nY 20\nZ 30\n"
	f2, err := os.CreateTemp("", "testmaplegacy-*.map")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(f2.Name())
	f2.WriteString(legacyMap)
	f2.Close()
	if err := bot.LoadMap(f2.Name()); err != nil {
		t.Fatalf("LoadMap legacy failed: %v", err)
	}
	mapName2 := strings.ToUpper(strings.TrimSuffix(filepath.Base(f2.Name()), filepath.Ext(f2.Name())))
	for k, v := range map[string]string{"X": "10", "Y": "20", "Z": "30"} {
		if got, ok := bot.Maps[mapName2][k]; !ok || got != v {
			t.Errorf("Expected legacy map[%q]=%q, got %q", k, v, got)
		}
	}
}

func TestBot_SetTagSupport(t *testing.T) {
	bot := NewBot(false)

	// Load a test set for ANIMAL
	setFile, err := os.CreateTemp("", "ANIMAL.set")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(setFile.Name())
	setFile.WriteString(`["DOG", "CAT", "BIRD", "FISH"]`)
	setFile.Close()
	if err := bot.LoadSet(setFile.Name()); err != nil {
		t.Fatalf("LoadSet failed: %v", err)
	}

	// Test patterns with <set> tags
	testCases := []struct {
		pattern  string
		input    string
		expected string
		desc     string
	}{
		{
			pattern:  "<set>ANIMAL</set>",
			input:    "DOG",
			expected: "A dog is a great pet!",
			desc:     "basic set match",
		},
		{
			pattern:  "<set>ANIMAL</set>",
			input:    "CAT",
			expected: "A cat is a great pet!",
			desc:     "another set member",
		},
		{
			pattern:  "I LOVE <set>ANIMAL</set>",
			input:    "I LOVE BIRD",
			expected: "Birds are wonderful!",
			desc:     "set in middle of pattern",
		},
		{
			pattern:  "MY <set>ANIMAL</set> IS CUTE",
			input:    "MY FISH IS CUTE",
			expected: "Fish are indeed cute!",
			desc:     "set with surrounding words",
		},
		{
			pattern:  "TELL ME ABOUT <set>ANIMAL</set>",
			input:    "TELL ME ABOUT DOG",
			expected: "Dogs are loyal companions.",
			desc:     "set with wildcard capture",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			cat := parser.Category{
				Pattern:  tc.pattern,
				That:     "",
				Topic:    "",
				Template: tc.expected,
			}
			// Debug: print normalized pattern tokens at insertion
			patternTokens := splitAIML(cat.Pattern)
			fmt.Fprintf(os.Stderr, "[TEST DEBUG] Insert pattern: %q -> tokens: %q\n", cat.Pattern, patternTokens)
			loadCategories(bot, []parser.Category{cat})

			// Debug: print normalized input tokens at match time
			inputTokens := splitAIML(tc.input)
			fmt.Fprintf(os.Stderr, "[TEST DEBUG] Match input: %q -> tokens: %q\n", tc.input, inputTokens)

			response, err := bot.Respond(tc.input, "test-session")
			if err != nil {
				t.Fatalf("Bot.Respond() returned an error: %v", err)
			}
			if response != tc.expected {
				t.Errorf("Expected response to be '%s', got: '%s'", tc.expected, response)
			}
		})
	}
}

func TestBot_SetTagWithWildcards(t *testing.T) {
	bot := NewBot(false)

	// Load a test set for COLORS
	setFile, err := os.CreateTemp("", "COLORS.set")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(setFile.Name())
	setFile.WriteString(`["RED", "BLUE", "GREEN"]`)
	setFile.Close()
	if err := bot.LoadSet(setFile.Name()); err != nil {
		t.Fatalf("LoadSet failed: %v", err)
	}

	// Test pattern with set and wildcards
	cat := parser.Category{
		Pattern:  "I LIKE <set>COLORS</set> AND *",
		That:     "",
		Topic:    "",
		Template: "You like <star/> and <star index=\"2\"/>!",
	}
	loadCategories(bot, []parser.Category{cat})

	response, err := bot.Respond("I LIKE RED AND PIZZA", "test-session")
	if err != nil {
		t.Fatalf("Bot.Respond() returned an error: %v", err)
	}
	expected := "You like RED and PIZZA!"
	if response != expected {
		t.Errorf("Expected response to be '%s', got: '%s'", expected, response)
	}
}

func TestBot_SetTagNoMatch(t *testing.T) {
	bot := NewBot(false)

	// Load a test set for ANIMALS
	setFile, err := os.CreateTemp("", "ANIMALS.set")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(setFile.Name())
	setFile.WriteString(`["DOG", "CAT"]`)
	setFile.Close()
	if err := bot.LoadSet(setFile.Name()); err != nil {
		t.Fatalf("LoadSet failed: %v", err)
	}

	// Test pattern with set
	cat := parser.Category{
		Pattern:  "<set>ANIMALS</set>",
		That:     "",
		Topic:    "",
		Template: "That's an animal!",
	}
	loadCategories(bot, []parser.Category{cat})

	// Try input that's not in the set
	response, err := bot.Respond("ELEPHANT", "test-session")
	if err != nil {
		t.Fatalf("Bot.Respond() returned an error: %v", err)
	}
	// Should get default response since no match
	expectedContains := "don't have any knowledge loaded"
	if !strings.Contains(response, expectedContains) {
		t.Errorf("Expected response to contain '%s', got: '%s'", expectedContains, response)
	}
}

func TestBot_SetTagInThatSection(t *testing.T) {
	bot := NewBot(true)

	// Load a test set for AFFIRMATIVE
	setFile, err := os.CreateTemp("", "AFFIRMATIVE.set")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(setFile.Name())
	setFile.WriteString(`["YES", "SURE", "OKAY"]`)
	setFile.Close()
	if err := bot.LoadSet(setFile.Name()); err != nil {
		t.Fatalf("LoadSet failed: %v", err)
	}

	// Add a category to respond to 'YES' (simulate a real bot response)
	catAffirm := parser.Category{
		Pattern:  "YES",
		That:     "*",
		Topic:    "",
		Template: "Affirmative received!",
	}
	loadCategories(bot, []parser.Category{catAffirm})

	// Test pattern with set in that section
	cat := parser.Category{
		Pattern:  "DO YOU WANT TO PLAY",
		That:     "<set>AFFIRMATIVE</set>",
		Topic:    "",
		Template: "Great! Let's play!",
	}
	loadCategories(bot, []parser.Category{cat})

	// First, get the bot's response to 'YES' (sets 'that' to 'YES' pattern)
	resp1, err := bot.Respond("YES", "test-session")
	if err != nil {
		t.Fatalf("Bot.Respond() returned an error: %v", err)
	}
	if resp1 != "Affirmative received!" {
		t.Errorf("Expected response to be 'Affirmative received!', got: '%s'", resp1)
	}

	// Now test the pattern with that context (should match <set>AFFIRMATIVE</set> in 'that')
	response, err := bot.Respond("DO YOU WANT TO PLAY", "test-session")
	if err != nil {
		t.Fatalf("Bot.Respond() returned an error: %v", err)
	}
	expected := "Great! Let's play!"
	if response != expected {
		t.Errorf("Expected response to be '%s', got: '%s'", expected, response)
	}
}

func TestBotThinkTag(t *testing.T) {
	bot := NewBot(false)

	tmpl := `<think><set name="foo">bar</set></think>Hello <get name="foo"/>!`
	cat := parser.Category{
		Pattern:  "TEST THINK",
		That:     "",
		Topic:    "",
		Template: tmpl,
	}
	loadCategories(bot, []parser.Category{cat})

	response, err := bot.Respond("test think", "think-session")
	if err != nil {
		t.Fatalf("Bot.Respond() returned an error: %v", err)
	}
	if response != "Hello bar!" {
		t.Errorf("Expected response to be 'Hello bar!', got: %s", response)
	}
}

func TestEvaluator_SetGetVarAndCarAttributes(t *testing.T) {
	bot := NewBot(false)

	// Test <set var="foo">bar</set> and <get var="foo"/>
	cat1 := parser.Category{
		Pattern:  "SET VAR FOO",
		That:     "",
		Topic:    "",
		Template: `<set var="foo">bar</set>`,
	}
	cat2 := parser.Category{
		Pattern:  "GET VAR FOO",
		That:     "",
		Topic:    "",
		Template: `<get var="foo"/>`,
	}

	// Test <set name="baz">qux</set> and <get name="baz"/>
	cat3 := parser.Category{
		Pattern:  "SET NAME BAZ",
		That:     "",
		Topic:    "",
		Template: `<set name="baz">qux</set>`,
	}
	cat4 := parser.Category{
		Pattern:  "GET NAME BAZ",
		That:     "",
		Topic:    "",
		Template: `<get name="baz"/>`,
	}

	// Test <get car="foo"/>
	cat5 := parser.Category{
		Pattern:  "GET CAR FOO",
		That:     "",
		Topic:    "",
		Template: `<get car="foo"/>`,
	}

	loadCategories(bot, []parser.Category{cat1, cat2, cat3, cat4, cat5})

	// Set var foo
	resp, err := bot.Respond("SET VAR FOO", "test-session")
	if err != nil {
		t.Fatalf("Bot.Respond() returned an error: %v", err)
	}
	if resp != "bar" {
		t.Errorf("Expected response to be 'bar', got: '%s'", resp)
	}

	// Get var foo
	resp, err = bot.Respond("GET VAR FOO", "test-session")
	if err != nil {
		t.Fatalf("Bot.Respond() returned an error: %v", err)
	}
	if resp != "bar" {
		t.Errorf("Expected response to be 'bar', got: '%s'", resp)
	}

	// Set name baz
	resp, err = bot.Respond("SET NAME BAZ", "test-session")
	if err != nil {
		t.Fatalf("Bot.Respond() returned an error: %v", err)
	}
	if resp != "qux" {
		t.Errorf("Expected response to be 'qux', got: '%s'", resp)
	}

	// Get name baz
	resp, err = bot.Respond("GET NAME BAZ", "test-session")
	if err != nil {
		t.Fatalf("Bot.Respond() returned an error: %v", err)
	}
	if resp != "qux" {
		t.Errorf("Expected response to be 'qux', got: '%s'", resp)
	}

	// Get car foo (should return 'bar')
	resp, err = bot.Respond("GET CAR FOO", "test-session")
	if err != nil {
		t.Fatalf("Bot.Respond() returned an error: %v", err)
	}
	if resp != "bar" {
		t.Errorf("Expected response to be 'bar', got: '%s'", resp)
	}
}

func TestEvaluator_UniqTag(t *testing.T) {
	bot := NewBot(false)

	// Assert a triple: CAT has sound MEOW
	cat1 := parser.Category{
		Pattern:  "ASSERT CAT SOUND",
		That:     "",
		Topic:    "",
		Template: `<uniq><subj>CAT</subj><pred>sound</pred><obj>MEOW</obj></uniq>`,
	}
	// Query the triple: CAT has sound ?sound
	cat2 := parser.Category{
		Pattern:  "QUERY CAT SOUND",
		That:     "",
		Topic:    "",
		Template: `<uniq><subj>CAT</subj><pred>sound</pred><obj>?sound</obj></uniq>`,
	}
	// Use <get var="sound"/> to retrieve the variable
	cat3 := parser.Category{
		Pattern:  "GET SOUND",
		That:     "",
		Topic:    "",
		Template: `<get var="sound"/>`,
	}

	loadCategories(bot, []parser.Category{cat1, cat2, cat3})

	// Assert triple
	resp, err := bot.Respond("ASSERT CAT SOUND", "test-session")
	if err != nil {
		t.Fatalf("Bot.Respond() returned an error: %v", err)
	}
	if resp != "MEOW" {
		t.Errorf("Expected response to be 'MEOW', got: '%s'", resp)
	}

	// Query triple (should set ?sound and return 'MEOW')
	resp, err = bot.Respond("QUERY CAT SOUND", "test-session")
	if err != nil {
		t.Fatalf("Bot.Respond() returned an error: %v", err)
	}
	if resp != "MEOW" {
		t.Errorf("Expected response to be 'MEOW', got: '%s'", resp)
	}

	// Get variable (should return 'MEOW')
	resp, err = bot.Respond("GET SOUND", "test-session")
	if err != nil {
		t.Fatalf("Bot.Respond() returned an error: %v", err)
	}
	if resp != "MEOW" {
		t.Errorf("Expected response to be 'MEOW', got: '%s'", resp)
	}
}
