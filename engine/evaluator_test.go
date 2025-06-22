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

func TestEvaluator_SetMapCondition(t *testing.T) {
	sess := &Session{Vars: make(map[string]string), Wildcards: make(map[string][]string)}
	cfg := &Config{Debug: true}

	// Test <set> with set
	bot := NewBot(false)
	bot.Sets["COLORS"] = map[string]struct{}{ "RED": {}, "BLUE": {} }
	eval := NewEvaluatorWithConfig(sess, nil, cfg, "", bot)
	_, err := eval.EvaluateTemplate(`<set name="COLORS">GREEN</set>`)
	if err != nil {
		t.Fatalf("set tag failed: %v", err)
	}
	if _, ok := bot.Sets["COLORS"]["GREEN"]; !ok {
		t.Errorf("Expected 'GREEN' to be added to set 'COLORS'")
	}

	// Test <map>
	bot = NewBot(false)
	bot.Maps["COUNTRIES"] = map[string]string{ "FR": "FRANCE", "DE": "GERMANY" }
	eval = NewEvaluatorWithConfig(sess, nil, cfg, "", bot)
	out, err := eval.EvaluateTemplate(`<map name="COUNTRIES">FR</map>`)
	if err != nil {
		t.Fatalf("map tag failed: %v", err)
	}
	if out != "FRANCE" {
		t.Errorf("Expected map lookup to return 'FRANCE', got %q", out)
	}

	// Test <condition> with set
	bot = NewBot(false)
	bot.Sets["COLORS"] = map[string]struct{}{ "RED": {}, "BLUE": {} }
	eval = NewEvaluatorWithConfig(sess, nil, cfg, "", bot)
	out, err = eval.EvaluateTemplate(`<condition set="COLORS" value="BLUE">Blue is in the set.</condition>`)
	if err != nil {
		t.Fatalf("condition set failed: %v", err)
	}
	if out != "Blue is in the set." {
		t.Errorf("Expected set condition to match, got %q", out)
	}

	// Test <condition> with map
	bot = NewBot(false)
	bot.Maps["COUNTRIES"] = map[string]string{ "FR": "FRANCE", "DE": "GERMANY" }
	eval = NewEvaluatorWithConfig(sess, nil, cfg, "", bot)
	out, err = eval.EvaluateTemplate(`<condition map="COUNTRIES" key="DE">Hallo!</condition>`)
	if err != nil {
		t.Fatalf("condition map failed: %v", err)
	}
	if out != "Hallo!" {
		t.Errorf("Expected map condition to match, got %q", out)
	}

	// Test <condition> with <li> children
	bot = NewBot(false)
	bot.Sets["COLORS"] = map[string]struct{}{ "RED": {}, "BLUE": {} }
	bot.Maps["COUNTRIES"] = map[string]string{ "FR": "FRANCE", "DE": "GERMANY" }
	eval = NewEvaluatorWithConfig(sess, nil, cfg, "", bot)
	tmpl := `<condition>
	<li set="COLORS" value="RED">Red found</li>
	<li map="COUNTRIES" key="FR">Bonjour</li>
	<li>Default</li>
</condition>`
	out, err = eval.EvaluateTemplate(tmpl)
	if err != nil {
		t.Fatalf("condition li failed: %v", err)
	}
	if out != "Red found" {
		t.Errorf("Expected first matching li to be 'Red found', got %q", out)
	}
} 