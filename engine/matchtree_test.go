package engine

import (
	"golem/parser"
	"testing"
)

func makeCat(pattern, that, topic, template string) parser.Category {
	return parser.Category{
		Pattern:  pattern,
		That:     that,
		Topic:    topic,
		Template: template,
	}
}

// func TestMatchTree_InsertAndMatch(t *testing.T) {
// 	tree := NewMatchTree()
//
// 	// Insert categories
// 	tree.Insert(makeCat("HELLO", "", "", "Hi there!"))
// 	tree.Insert(makeCat("HOW ARE YOU", "HELLO", "", "I'm good, thanks!"))
// 	tree.Insert(makeCat("WHAT IS YOUR NAME", "", "", "My name is Golem."))
// 	tree.Insert(makeCat("TELL * JOKE", "", "HUMOR", "Here's a joke!"))
// 	tree.Insert(makeCat("BYE", "", "", "Goodbye!"))
// 	tree.Insert(makeCat("_ WEATHER", "", "", "Weather wildcard!"))
// 	tree.Insert(makeCat("*", "", "", "Catch-all!"))
//
// 	tests := []struct {
// 		input  string
// 		that   string
// 		topic  string
// 		expect string
// 		desc   string
// 	}{
// 		{"hello", "", "", "Hi there!", "Exact match"},
// 		{"how are you", "hello", "", "I'm good, thanks!", "Match with 'that' context"},
// 		{"what is your name", "", "", "My name is Golem.", "Exact match 2"},
// 		{"tell me a joke", "", "humor", "Here's a joke!", "Wildcard in pattern, topic match"},
// 		{"bye", "", "", "Goodbye!", "Simple match"},
// 		{"today weather", "", "", "Weather wildcard!", "_ wildcard at start"},
// 		{"something else", "", "", "Catch-all!", "Catch-all wildcard"},
// 		{"tell me a joke", "", "", "Catch-all!", "No topic, so catch-all"},
// 		{"how are you", "bye", "", "Catch-all!", "No that match, so catch-all"},
// 	}
//
// 	for _, test := range tests {
// 		t.Run(test.desc, func(t *testing.T) {
// 			cat, found := tree.Match(test.input, test.that, test.topic)
// 			if !found {
// 				t.Fatalf("No match found for input: %q, that: %q, topic: %q", test.input, test.that, test.topic)
// 			}
// 			sess := &Session{Vars: make(map[string]string), Wildcards: make(map[string][]string)}
// 			eval := NewEvaluator(sess, nil)
// 			output, err := eval.EvaluateTemplate(cat.Template)
// 			if err != nil {
// 				t.Fatalf("Evaluation error: %v", err)
// 			}
// 			if output != test.expect {
// 				t.Errorf("Expected output %q, got %q", test.expect, output)
// 			}
// 		})
// 	}
// }

func TestMatchTree_WildcardPriority(t *testing.T) {
	tree := NewMatchTree()
	tree.Insert(makeCat("HELLO *", "", "", "Wildcard after hello"))
	tree.Insert(makeCat("HELLO THERE", "", "", "Hello there!"))

	cat, found := tree.Match("hello there", "", "")
	if !found {
		t.Fatal("Expected a match for 'hello there'")
	}
	sess := &Session{Vars: make(map[string]string), Wildcards: make(map[string][]string)}
	eval := NewEvaluator(sess, nil)
	output, err := eval.EvaluateTemplate(cat.Template)
	if err != nil {
		t.Fatalf("Evaluation error: %v", err)
	}
	if output != "Hello there!" {
		t.Errorf("Expected 'Hello there!', got %q", output)
	}

	cat, found = tree.Match("hello world", "", "")
	if !found {
		t.Fatal("Expected a match for 'hello world'")
	}
	sess = &Session{Vars: make(map[string]string), Wildcards: make(map[string][]string)}
	eval = NewEvaluator(sess, nil)
	output, err = eval.EvaluateTemplate(cat.Template)
	if err != nil {
		t.Fatalf("Evaluation error: %v", err)
	}
	if output != "Wildcard after hello" {
		t.Errorf("Expected 'Wildcard after hello', got %q", output)
	}
}

func TestMatchTree_MatchWithMeta_MetadataAndWildcards(t *testing.T) {
	tree := NewMatchTree()
	tree.Insert(makeCat("HELLO *", "", "", "Hi *!"))
	tree.Insert(makeCat("HOW ARE _", "HELLO *", "", "I'm good after *"))
	tree.Insert(makeCat("WHAT IS YOUR NAME", "", "", "My name is Golem."))
	tree.Insert(makeCat("*", "", "", "Catch-all!"))

	tests := []struct {
		input           string
		that            string
		topic           string
		expectTemplate  string
		expectPattern   string
		expectThat      string
		expectWildcards map[string][]string
		desc            string
	}{
		{
			input:           "hello world",
			that:            "",
			topic:           "",
			expectTemplate:  "Hi *!",
			expectPattern:   "HELLO *",
			expectThat:      "",
			expectWildcards: map[string][]string{"pattern": {"WORLD"}},
			desc:            "Pattern wildcard capture",
		},
		{
			input:           "how are you",
			that:            "hello bob",
			topic:           "",
			expectTemplate:  "I'm good after *",
			expectPattern:   "HOW ARE _",
			expectThat:      "HELLO *",
			expectWildcards: map[string][]string{"pattern": {"YOU"}, "that": {"BOB"}},
			desc:            "Pattern _ and that * wildcard capture",
		},
		{
			input:           "what is your name",
			that:            "",
			topic:           "",
			expectTemplate:  "My name is Golem.",
			expectPattern:   "WHAT IS YOUR NAME",
			expectThat:      "",
			expectWildcards: map[string][]string{"pattern": {}},
			desc:            "Exact match, no wildcards",
		},
		{
			input:           "something else",
			that:            "",
			topic:           "",
			expectTemplate:  "Catch-all!",
			expectPattern:   "*",
			expectThat:      "",
			expectWildcards: map[string][]string{"pattern": {"SOMETHING", "ELSE"}},
			desc:            "Catch-all wildcard",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			res, found := tree.MatchWithMeta(test.input, test.that, test.topic, nil)
			if !found {
				t.Fatalf("No match found for input: %q, that: %q, topic: %q", test.input, test.that, test.topic)
			}
			sess := &Session{Vars: make(map[string]string), Wildcards: make(map[string][]string)}
			eval := NewEvaluator(sess, nil)
			output, err := eval.EvaluateTemplate(res.Template)
			if err != nil {
				t.Fatalf("Evaluation error: %v", err)
			}
			if output != test.expectTemplate {
				t.Errorf("Expected output %q, got %q", test.expectTemplate, output)
			}
			if res.MatchedPattern != test.expectPattern {
				t.Errorf("Expected matched pattern %q, got %q", test.expectPattern, res.MatchedPattern)
			}
			// If test.that is empty, expect MatchedThat to be "*"
			expectThat := test.expectThat
			if test.that == "" {
				expectThat = "*"
			}
			if res.MatchedThat != expectThat {
				t.Errorf("Expected matched that %q, got %q", expectThat, res.MatchedThat)
			}
			for k, v := range test.expectWildcards {
				if got, ok := res.WildcardCaptures[k]; !ok || !equalStringSlices(got, v) {
					t.Errorf("Expected wildcard captures[%s] = %v, got %v", k, v, got)
				}
			}
		})
	}
}

func TestMatchTree_ThatAndWildcardMatching(t *testing.T) {
	tree := NewMatchTree()

	// Insert categories with 'that' and wildcards
	tree.Insert(makeCat("HOW ARE YOU", "HELLO *", "", "I'm good after *"))
	tree.Insert(makeCat("HOW ARE _", "HELLO THERE", "", "I'm good after there"))
	tree.Insert(makeCat("HOW ARE *", "*", "", "Wildcard everywhere"))
	tree.Insert(makeCat("*", "*", "", "Catch-all!"))

	tests := []struct {
		input           string
		that            string
		topic           string
		expectTemplate  string
		expectPattern   string
		expectThat      string
		expectWildcards map[string][]string
		desc            string
	}{
		{
			input:           "how are you",
			that:            "hello bob",
			topic:           "",
			expectTemplate:  "I'm good after *",
			expectPattern:   "HOW ARE YOU",
			expectThat:      "HELLO *",
			expectWildcards: map[string][]string{"pattern": {}, "that": {"BOB"}},
			desc:            "Exact pattern, wildcard in that",
		},
		{
			input:           "how are they",
			that:            "hello there",
			topic:           "",
			expectTemplate:  "I'm good after there",
			expectPattern:   "HOW ARE _",
			expectThat:      "HELLO THERE",
			expectWildcards: map[string][]string{"pattern": {"THEY"}},
			desc:            "Pattern _ wildcard, exact that",
		},
		{
			input:           "how are you",
			that:            "something else",
			topic:           "",
			expectTemplate:  "Wildcard everywhere",
			expectPattern:   "HOW ARE *",
			expectThat:      "*",
			expectWildcards: map[string][]string{"pattern": {"YOU"}, "that": {"SOMETHING", "ELSE"}},
			desc:            "Pattern * and that * wildcards",
		},
		{
			input:           "unknown input",
			that:            "random",
			topic:           "",
			expectTemplate:  "Catch-all!",
			expectPattern:   "*",
			expectThat:      "*",
			expectWildcards: map[string][]string{"pattern": {"UNKNOWN", "INPUT"}, "that": {"RANDOM"}},
			desc:            "Catch-all fallback",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			res, found := tree.MatchWithMeta(test.input, test.that, test.topic, nil)
			if !found {
				t.Fatalf("No match found for input: %q, that: %q, topic: %q", test.input, test.that, test.topic)
			}
			sess := &Session{Vars: make(map[string]string), Wildcards: make(map[string][]string)}
			eval := NewEvaluator(sess, nil)
			output, err := eval.EvaluateTemplate(res.Template)
			if err != nil {
				t.Fatalf("Evaluation error: %v", err)
			}
			if output != test.expectTemplate {
				t.Errorf("Expected output %q, got %q", test.expectTemplate, output)
			}
			if res.MatchedPattern != test.expectPattern {
				t.Errorf("Expected matched pattern %q, got %q", test.expectPattern, res.MatchedPattern)
			}
			// If test.that is empty, expect MatchedThat to be "*"
			expectThat := test.expectThat
			if test.that == "" {
				expectThat = "*"
			}
			if res.MatchedThat != expectThat {
				t.Errorf("Expected matched that %q, got %q", expectThat, res.MatchedThat)
			}
			for k, v := range test.expectWildcards {
				if got, ok := res.WildcardCaptures[k]; !ok || !equalStringSlices(got, v) {
					t.Errorf("Expected wildcard captures[%s] = %v, got %v", k, v, got)
				}
			}
		})
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
