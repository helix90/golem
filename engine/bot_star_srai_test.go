package engine

import (
	"fmt"
	"golem/parser"
	"os"
	"reflect"
	"testing"
)

func TestEvaluator_Star(t *testing.T) {
	sess := &Session{Vars: make(map[string]string), Wildcards: map[string][]string{"pattern": {"foo", "bar", "baz"}}}
	eval := NewEvaluator(sess, nil)

	out, err := eval.EvaluateTemplate(`<star/>`)
	if err != nil {
		t.Fatalf("star failed: %v", err)
	}
	if out != "foo bar baz" {
		t.Errorf("Expected 'foo bar baz', got %q", out)
	}

	out, err = eval.EvaluateTemplate(`<star index="2"/>`)
	if err != nil {
		t.Fatalf("star index failed: %v", err)
	}
	if out != "bar" {
		t.Errorf("Expected 'bar', got %q", out)
	}
}

func TestEvaluator_Srai(t *testing.T) {
	sess := &Session{Vars: make(map[string]string), Wildcards: make(map[string][]string)}
	mockSrai := func(input string) (string, error) {
		if input == "WHAT IS YOUR NAME" {
			return "My name is Golem.", nil
		}
		return "", nil
	}
	eval := NewEvaluator(sess, mockSrai)

	out, err := eval.EvaluateTemplate(`<srai>WHAT IS YOUR NAME</srai>`)
	if err != nil {
		t.Fatalf("srai failed: %v", err)
	}
	if out != "My name is Golem." {
		t.Errorf("Expected 'My name is Golem.', got %q", out)
	}
}

func TestEvaluator_SetGetVarAttribute(t *testing.T) {
	sess := &Session{Vars: make(map[string]string), Wildcards: make(map[string][]string)}
	eval := NewEvaluator(sess, nil)

	_, err := eval.EvaluateTemplate(`<set var="animal">cat</set>`)
	if err != nil {
		t.Fatalf("set var failed: %v", err)
	}
	out, err := eval.EvaluateTemplate(`<get var="animal"/>`)
	if err != nil {
		t.Fatalf("get var failed: %v", err)
	}
	if out != "cat" {
		t.Errorf("Expected 'cat', got %q", out)
	}
}

func TestEvaluator_That_And_Thatstar(t *testing.T) {
	// Setup a Bot with a SessionManager and a session with That set
	fakeBot := &Bot{sessions: NewSessionManager()}
	userSession := fakeBot.sessions.GetOrCreateSession("test-session")
	userSession.That = "The last response!"

	sess := &Session{Vars: make(map[string]string), Wildcards: map[string][]string{"that": {"foo", "bar", "baz"}}}
	eval := NewEvaluatorWithConfig(sess, nil, &Config{Debug: false}, "test-session", fakeBot)

	// Test <that/>
	out, err := eval.EvaluateTemplate(`<that/>`)
	if err != nil {
		t.Fatalf("that failed: %v", err)
	}
	if out != "The last response!" {
		t.Errorf("Expected 'The last response!', got %q", out)
	}

	// Test <thatstar/>
	out, err = eval.EvaluateTemplate(`<thatstar/>`)
	if err != nil {
		t.Fatalf("thatstar failed: %v", err)
	}
	if out != "foo bar baz" {
		t.Errorf("Expected 'foo bar baz', got %q", out)
	}

	// Test <thatstar index="2"/>
	out, err = eval.EvaluateTemplate(`<thatstar index="2"/>`)
	if err != nil {
		t.Fatalf("thatstar index failed: %v", err)
	}
	if out != "bar" {
		t.Errorf("Expected 'bar', got %q", out)
	}
}

func init() {
	field, _ := reflect.TypeOf(parser.Category{}).FieldByName("Template")
	fmt.Fprintf(os.Stderr, "Template struct tag at runtime: %q\n", field.Tag)
}
