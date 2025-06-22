package engine

import (
	"testing"
)

func TestEvaluator_Template(t *testing.T) {
	sess := NewSession()
	eval := NewEvaluator(sess, nil)
	out, err := eval.EvaluateTemplate("Hello there!")
	if err != nil || out != "Hello there!" {
		t.Errorf("Expected 'Hello there!', got '%s', err: %v", out, err)
	}
}

func TestEvaluator_SetAndGet(t *testing.T) {
	sess := NewSession()
	eval := NewEvaluator(sess, nil)
	_, _ = eval.EvaluateTemplate(`<set name="foo">bar</set>`)
	out, _ := eval.EvaluateTemplate(`<get name="foo"/>`)
	if out != "bar" {
		t.Errorf("Expected 'bar', got '%s'", out)
	}
}

func TestEvaluator_Srai(t *testing.T) {
	sess := NewSession()
	mockSrai := func(input string) (string, error) {
		if input == "WHAT IS YOUR NAME" {
			return "My name is Golem.", nil
		}
		return "", nil
	}
	eval := NewEvaluator(sess, mockSrai)
	out, _ := eval.EvaluateTemplate(`<srai>WHAT IS YOUR NAME</srai>`)
	if out != "My name is Golem." {
		t.Errorf("Expected 'My name is Golem.', got '%s'", out)
	}
}

func TestEvaluator_Think(t *testing.T) {
	sess := NewSession()
	eval := NewEvaluator(sess, nil)
	out, _ := eval.EvaluateTemplate(`<think><set name="foo">secret</set></think>Visible`)
	if out != "Visible" {
		t.Errorf("Expected 'Visible', got '%s'", out)
	}
	val, _ := eval.EvaluateTemplate(`<get name="foo"/>`)
	if val != "secret" {
		t.Errorf("Expected 'secret', got '%s'", val)
	}
}

func TestEvaluator_Condition(t *testing.T) {
	sess := NewSession()
	sess.Vars["weather"] = "sunny"
	eval := NewEvaluator(sess, nil)
	tmpl := `<condition name="weather"><li value="sunny">It's sunny!</li><li>Unknown</li></condition>`
	out, _ := eval.EvaluateTemplate(tmpl)
	if out != "It's sunny!" {
		t.Errorf("Expected 'It's sunny!', got '%s'", out)
	}
}

func TestEvaluator_Random(t *testing.T) {
	sess := NewSession()
	eval := NewEvaluator(sess, nil)
	tmpl := `<random><li>One</li><li>Two</li><li>Three</li></random>`
	out, _ := eval.EvaluateTemplate(tmpl)
	if out != "One" && out != "Two" && out != "Three" {
		t.Errorf("Expected one of the options, got '%s'", out)
	}
}

func TestEvaluator_Star(t *testing.T) {
	sess := NewSession()
	sess.Wildcards["pattern"] = []string{"foo", "bar"}
	eval := NewEvaluator(sess, nil)
	out, _ := eval.EvaluateTemplate(`<star/>`)
	if out != "foo bar" {
		t.Errorf("Expected 'foo bar', got '%s'", out)
	}
	out, _ = eval.EvaluateTemplate(`<star index="2"/>`)
	if out != "bar" {
		t.Errorf("Expected 'bar', got '%s'", out)
	}
} 