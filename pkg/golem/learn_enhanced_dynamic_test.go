package golem

import (
	"fmt"
	"strings"
	"testing"
)

// TestLearnTagDynamicEvaluation tests the enhanced learn tag with dynamic evaluation
func TestLearnTagDynamicEvaluation(t *testing.T) {
	g := New(false)
	kb := NewAIMLKnowledgeBase()
	g.SetKnowledgeBase(kb)

	// Create a session for testing
	session := g.CreateSession("test_session")

	// Set up variables for dynamic learning
	g.ProcessTemplateWithContext(`<set name="pattern1">HELLO *</set>`, map[string]string{}, session)
	g.ProcessTemplateWithContext(`<set name="response1">Hi there! How can I help you?</set>`, map[string]string{}, session)

	// Test dynamic learn tag with eval tags
	template := `<learn>
		<category>
			<pattern><eval><get name="pattern1"/></eval></pattern>
			<template><eval><get name="response1"/></eval></template>
		</category>
	</learn>`

	result := g.ProcessTemplateWithContext(template, map[string]string{}, session)

	// The learn tag should be removed after processing
	if strings.Contains(result, "<learn>") || strings.Contains(result, "</learn>") {
		t.Errorf("Learn tag not removed from template: %s", result)
	}

	// Test that the learned category works
	testInput := "HELLO WORLD"
	response, err := g.ProcessInput(testInput, session)
	if err != nil {
		t.Errorf("Error processing input: %v", err)
	}

	expectedResponse := "Hi there! How can I help you?"
	if response != expectedResponse {
		t.Errorf("Expected response '%s', got '%s'", expectedResponse, response)
	}
}

// TestLearnTagMultipleDynamicCategories tests learning multiple categories with dynamic evaluation
func TestLearnTagMultipleDynamicCategories(t *testing.T) {
	g := New(false)
	kb := NewAIMLKnowledgeBase()
	g.SetKnowledgeBase(kb)

	session := g.CreateSession("test_session")

	// Set up multiple variables for dynamic learning
	g.ProcessTemplateWithContext(`<set name="pattern1">WHAT IS *</set>`, map[string]string{}, session)
	g.ProcessTemplateWithContext(`<set name="response1">I don't know what <star/> is.</set>`, map[string]string{}, session)
	g.ProcessTemplateWithContext(`<set name="pattern2">TELL ME ABOUT *</set>`, map[string]string{}, session)
	g.ProcessTemplateWithContext(`<set name="response2">Let me tell you about <star/>.</set>`, map[string]string{}, session)

	// Test multiple dynamic categories
	template := `<learn>
		<category>
			<pattern><eval><get name="pattern1"/></eval></pattern>
			<template><eval><get name="response1"/></eval></template>
		</category>
		<category>
			<pattern><eval><get name="pattern2"/></eval></pattern>
			<template><eval><get name="response2"/></eval></template>
		</category>
	</learn>`

	result := g.ProcessTemplateWithContext(template, map[string]string{}, session)

	// The learn tag should be removed after processing
	if strings.Contains(result, "<learn>") || strings.Contains(result, "</learn>") {
		t.Errorf("Learn tag not removed from template: %s", result)
	}

	// Test both learned categories
	testCases := []struct {
		input    string
		expected string
	}{
		{"WHAT IS AI", "I don't know what AI is."},
		{"TELL ME ABOUT MACHINE LEARNING", "Let me tell you about MACHINE LEARNING."},
	}

	for _, tc := range testCases {
		response, err := g.ProcessInput(tc.input, session)
		if err != nil {
			t.Errorf("Error processing input '%s': %v", tc.input, err)
			continue
		}

		if response != tc.expected {
			t.Errorf("For input '%s', expected response '%s', got '%s'", tc.input, tc.expected, response)
		}
	}
}

// TestLearnTagWithComplexEval tests learn tag with complex eval expressions
func TestLearnTagWithComplexEval(t *testing.T) {
	g := New(false)
	kb := NewAIMLKnowledgeBase()
	g.SetKnowledgeBase(kb)

	session := g.CreateSession("test_session")

	// Set up variables for complex evaluation
	g.ProcessTemplateWithContext(`<set name="base_pattern">GREET *</set>`, map[string]string{}, session)
	g.ProcessTemplateWithContext(`<set name="base_response">Hello <star/>! Nice to meet you.</set>`, map[string]string{}, session)

	// Test complex eval with multiple tags
	template := `<learn>
		<category>
			<pattern><eval><uppercase><get name="base_pattern"/></uppercase></eval></pattern>
			<template><eval><formal><get name="base_response"/></formal></eval></template>
		</category>
	</learn>`

	result := g.ProcessTemplateWithContext(template, map[string]string{}, session)

	// The learn tag should be removed after processing
	if strings.Contains(result, "<learn>") || strings.Contains(result, "</learn>") {
		t.Errorf("Learn tag not removed from template: %s", result)
	}

	// Test the learned category
	testInput := "GREET ALICE"
	response, err := g.ProcessInput(testInput, session)
	if err != nil {
		t.Errorf("Error processing input: %v", err)
	}

	// Note: The <formal> tag is applied to the template structure during learning,
	// but the wildcard value is inserted as-is when the pattern is matched.
	// So "ALICE" from the input remains uppercase.
	expectedResponse := "Hello ALICE! Nice To Meet You."
	if response != expectedResponse {
		t.Errorf("Expected response '%s', got '%s'", expectedResponse, response)
	}
}

// TestLearnTagWithWildcards tests learn tag with wildcard evaluation
func TestLearnTagWithWildcards(t *testing.T) {
	g := New(false)
	kb := NewAIMLKnowledgeBase()
	g.SetKnowledgeBase(kb)

	session := g.CreateSession("test_session")

	// Set up wildcard variables
	g.ProcessTemplateWithContext(`<set name="wildcard_pattern">* LIKES *</set>`, map[string]string{}, session)
	g.ProcessTemplateWithContext(`<set name="wildcard_response"><star index="1"/> likes <star index="2"/>.</set>`, map[string]string{}, session)

	// Test wildcard evaluation in learn tag
	template := `<learn>
		<category>
			<pattern><eval><get name="wildcard_pattern"/></eval></pattern>
			<template><eval><get name="wildcard_response"/></eval></template>
		</category>
	</learn>`

	result := g.ProcessTemplateWithContext(template, map[string]string{}, session)

	// The learn tag should be removed after processing
	if strings.Contains(result, "<learn>") || strings.Contains(result, "</learn>") {
		t.Errorf("Learn tag not removed from template: %s", result)
	}

	// Test the learned category with wildcards
	testInput := "ALICE LIKES CHOCOLATE"
	response, err := g.ProcessInput(testInput, session)
	if err != nil {
		t.Errorf("Error processing input: %v", err)
	}

	expectedResponse := "ALICE likes CHOCOLATE."
	if response != expectedResponse {
		t.Errorf("Expected response '%s', got '%s'", expectedResponse, response)
	}
}

// TestLearnTagWithConditionalEval tests learn tag with conditional evaluation
func TestLearnTagWithConditionalEval(t *testing.T) {
	g := New(false)
	kb := NewAIMLKnowledgeBase()
	g.SetKnowledgeBase(kb)

	session := g.CreateSession("test_session")

	// Set up conditional variables
	g.ProcessTemplateWithContext(`<set name="condition">true</set>`, map[string]string{}, session)
	g.ProcessTemplateWithContext(`<set name="true_pattern">YES *</set>`, map[string]string{}, session)
	g.ProcessTemplateWithContext(`<set name="true_response">Yes, <star/> is correct!</set>`, map[string]string{}, session)

	// Test conditional evaluation in learn tag
	template := `<learn>
		<category>
			<pattern><eval><condition name="condition" value="true"><get name="true_pattern"/></condition></eval></pattern>
			<template><eval><condition name="condition" value="true"><get name="true_response"/></condition></eval></template>
		</category>
	</learn>`

	result := g.ProcessTemplateWithContext(template, map[string]string{}, session)

	// The learn tag should be removed after processing
	if strings.Contains(result, "<learn>") || strings.Contains(result, "</learn>") {
		t.Errorf("Learn tag not removed from template: %s", result)
	}

	// Test the learned category
	testInput := "YES THAT IS RIGHT"
	response, err := g.ProcessInput(testInput, session)
	if err != nil {
		t.Errorf("Error processing input: %v", err)
	}

	expectedResponse := "Yes, THAT IS RIGHT is correct!"
	if response != expectedResponse {
		t.Errorf("Expected response '%s', got '%s'", expectedResponse, response)
	}
}

// TestLearnTagErrorHandling tests error handling in dynamic learn tags
func TestLearnTagErrorHandling(t *testing.T) {
	g := New(false)
	kb := NewAIMLKnowledgeBase()
	g.SetKnowledgeBase(kb)

	session := g.CreateSession("test_session")

	// Test with invalid eval content
	template := `<learn>
		<category>
			<pattern><eval><get name="nonexistent"/></eval></pattern>
			<template><eval><get name="nonexistent"/></eval></template>
		</category>
	</learn>`

	result := g.ProcessTemplateWithContext(template, map[string]string{}, session)

	// The learn tag should be removed even on error
	if strings.Contains(result, "<learn>") || strings.Contains(result, "</learn>") {
		t.Errorf("Learn tag not removed from template: %s", result)
	}

	// Test with malformed eval content
	template2 := `<learn>
		<category>
			<pattern><eval>INVALID SYNTAX</eval></pattern>
			<template><eval>INVALID SYNTAX</eval></template>
		</category>
	</learn>`

	result2 := g.ProcessTemplateWithContext(template2, map[string]string{}, session)

	// The learn tag should be removed even on error
	if strings.Contains(result2, "<learn>") || strings.Contains(result2, "</learn>") {
		t.Errorf("Learn tag not removed from template: %s", result2)
	}
}

// TestLearnfTagDynamicEvaluation tests the enhanced learnf tag with dynamic evaluation
func TestLearnfTagDynamicEvaluation(t *testing.T) {
	g := New(false)
	kb := NewAIMLKnowledgeBase()
	g.SetKnowledgeBase(kb)

	session := g.CreateSession("test_session")

	// Set up variables for dynamic persistent learning
	g.ProcessTemplateWithContext(`<set name="persistent_pattern">PERSISTENT *</set>`, map[string]string{}, session)
	g.ProcessTemplateWithContext(`<set name="persistent_response">This is a persistent response for <star/>.</set>`, map[string]string{}, session)

	// Test dynamic learnf tag
	template := `<learnf>
		<category>
			<pattern><eval><get name="persistent_pattern"/></eval></pattern>
			<template><eval><get name="persistent_response"/></eval></template>
		</category>
	</learnf>`

	result := g.ProcessTemplateWithContext(template, map[string]string{}, session)

	// The learnf tag should be removed after processing
	if strings.Contains(result, "<learnf>") || strings.Contains(result, "</learnf>") {
		t.Errorf("Learnf tag not removed from template: %s", result)
	}

	// Test that the learned category works
	testInput := "PERSISTENT TEST"
	response, err := g.ProcessInput(testInput, session)
	if err != nil {
		t.Errorf("Error processing input: %v", err)
	}

	expectedResponse := "This is a persistent response for TEST."
	if response != expectedResponse {
		t.Errorf("Expected response '%s', got '%s'", expectedResponse, response)
	}
}

// TestLearnTagIntegrationWithOtherTags tests integration with other tag types
func TestLearnTagIntegrationWithOtherTags(t *testing.T) {
	g := New(false)
	kb := NewAIMLKnowledgeBase()
	g.SetKnowledgeBase(kb)

	session := g.CreateSession("test_session")

	// Set up variables using other tag types
	g.ProcessTemplateWithContext(`<set name="time_pattern">WHAT TIME IS IT</set>`, map[string]string{}, session)
	g.ProcessTemplateWithContext(`<set name="time_response">The current time is <date/>.</set>`, map[string]string{}, session)

	// Test integration with other tags
	template := `<learn>
		<category>
			<pattern><eval><uppercase><get name="time_pattern"/></uppercase></eval></pattern>
			<template><eval><get name="time_response"/></eval></template>
		</category>
	</learn>`

	result := g.ProcessTemplateWithContext(template, map[string]string{}, session)

	// The learn tag should be removed after processing
	if strings.Contains(result, "<learn>") || strings.Contains(result, "</learn>") {
		t.Errorf("Learn tag not removed from template: %s", result)
	}

	// Test the learned category
	testInput := "what time is it"
	response, err := g.ProcessInput(testInput, session)
	if err != nil {
		t.Errorf("Error processing input: %v", err)
	}

	// The response should contain a date (we can't predict the exact format)
	if !strings.Contains(response, "The current time is") {
		t.Errorf("Expected response to contain 'The current time is', got '%s'", response)
	}
}

// TestLearnTagPerformance tests performance with multiple dynamic categories
func TestLearnTagPerformance(t *testing.T) {
	g := New(false)
	kb := NewAIMLKnowledgeBase()
	g.SetKnowledgeBase(kb)

	session := g.CreateSession("test_session")

	// Set up many variables for performance testing
	for i := 0; i < 10; i++ {
		pattern := fmt.Sprintf("PATTERN%d *", i)
		response := fmt.Sprintf("Response %d for <star/>.", i)

		g.ProcessTemplateWithContext(fmt.Sprintf(`<set name="pattern%d">%s</set>`, i, pattern), map[string]string{}, session)
		g.ProcessTemplateWithContext(fmt.Sprintf(`<set name="response%d">%s</set>`, i, response), map[string]string{}, session)
	}

	// Create a large learn template
	var learnContent strings.Builder
	learnContent.WriteString("<learn>\n")
	for i := 0; i < 10; i++ {
		learnContent.WriteString(fmt.Sprintf(`
		<category>
			<pattern><eval><get name="pattern%d"/></eval></pattern>
			<template><eval><get name="response%d"/></eval></template>
		</category>`, i, i))
	}
	learnContent.WriteString("\n</learn>")

	// Test performance
	result := g.ProcessTemplateWithContext(learnContent.String(), map[string]string{}, session)

	// The learn tag should be removed after processing
	if strings.Contains(result, "<learn>") || strings.Contains(result, "</learn>") {
		t.Errorf("Learn tag not removed from template")
	}

	// Test a few learned categories
	testCases := []struct {
		input    string
		expected string
	}{
		{"PATTERN0 TEST", "Response 0 for TEST."},
		{"PATTERN5 TEST", "Response 5 for TEST."},
		{"PATTERN9 TEST", "Response 9 for TEST."},
	}

	for _, tc := range testCases {
		response, err := g.ProcessInput(tc.input, session)
		if err != nil {
			t.Errorf("Error processing input '%s': %v", tc.input, err)
			continue
		}

		if response != tc.expected {
			t.Errorf("For input '%s', expected response '%s', got '%s'", tc.input, tc.expected, response)
		}
	}
}
