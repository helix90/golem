package testdata

import (
	"golem/engine"
	"golem/parser"
	"os"
	"path/filepath"
	"testing"
)

func TestConformance_SimpleAIML(t *testing.T) {
	bot, err := engine.NewBot(engine.Config{})
	if err != nil {
		t.Fatalf("Failed to create bot: %v", err)
	}
	path := filepath.Join("testdata", "simple.aiml")
	if err := bot.LoadAIML(path); err != nil {
		t.Fatalf("Failed to load AIML: %v", err)
	}
	tests := []struct {
		input  string
		that   string
		topic  string
		expect string
	}{
		{"HELLO", "", "", "Hello! How are you today?"},
		{"WHAT IS YOUR NAME", "", "", "My name is Golem, nice to meet you!"},
		{"HOW ARE YOU", "HELLO", "", "I'm doing well, thank you for asking!"},
		{"TELL ME A JOKE", "", "HUMOR", "Why don't scientists trust atoms? Because they make up everything!"},
		{"GOODBYE", "", "", "Goodbye! Have a great day!"},
		// Malformed categories should not match
		{"", "", "", ""},
		{"EMPTY TEMPLATE", "", "", ""},
	}
	for _, tc := range tests {
		cat, found := bot.MatchTree.MatchWithMeta(tc.input, tc.that, tc.topic)
		if tc.expect == "" {
			if found && cat != nil && cat.Template != "" {
				t.Errorf("Expected no match for input %q, that %q, topic %q, but got: %q", tc.input, tc.that, tc.topic, cat.Template)
			}
			continue
		}
		if !found || cat == nil {
			t.Errorf("Expected match for input %q, that %q, topic %q, but got none", tc.input, tc.that, tc.topic)
			continue
		}
		if cat.Template != tc.expect {
			t.Errorf("For input %q, that %q, topic %q: expected %q, got %q", tc.input, tc.that, tc.topic, tc.expect, cat.Template)
		}
	}
} 