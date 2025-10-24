package golem

import (
	"testing"
	"time"
)

// TestTreeProcessorPersonTag tests the <person> tag with tree processor
func TestTreeProcessorPersonTag(t *testing.T) {
	g := New(false)
	g.EnableTreeProcessing() // Enable AST-based processing

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "Basic first person to second person",
			template: "<person>I am happy</person>",
			expected: "you are happy",
		},
		{
			name:     "Basic second person to first person",
			template: "<person>you are happy</person>",
			expected: "I am happy",
		},
		{
			name:     "Multiple pronouns",
			template: "<person>I gave you my book</person>",
			expected: "you gave I your book",
		},
		{
			name:     "Possessive pronouns",
			template: "<person>my car and your bike</person>",
			expected: "your car and my bike",
		},
		{
			name:     "Contractions",
			template: "<person>I'm going to give you my car</person>",
			expected: "you're going to give I your car",
		},
		{
			name:     "Complex sentence",
			template: "<person>I think you should take my advice</person>",
			expected: "you think I should take your advice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := g.CreateSession("test_person_" + tt.name)
			result := g.ProcessTemplateWithContext(tt.template, nil, session)

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestTreeProcessorPerson2Tag tests the <person2> tag with tree processor
func TestTreeProcessorPerson2Tag(t *testing.T) {
	g := New(false)
	g.EnableTreeProcessing() // Enable AST-based processing

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "Basic first person to third person",
			template: "<person2>I am happy</person2>",
			expected: "they are happy",
		},
		{
			name:     "Multiple pronouns",
			template: "<person2>I gave my book to myself</person2>",
			expected: "they gave their book to themselves",
		},
		{
			name:     "Plural first person",
			template: "<person2>We are going to our house</person2>",
			expected: "They are going to their house",
		},
		{
			name:     "Possessive pronouns",
			template: "<person2>This is mine and that is ours</person2>",
			expected: "This is theirs and that is theirs",
		},
		{
			name:     "Contractions",
			template: "<person2>I'm going to give you my car</person2>",
			expected: "they're going to give you their car",
		},
		{
			name:     "Complex sentence",
			template: "<person2>I think my idea is better</person2>",
			expected: "they think their idea is better",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := g.CreateSession("test_person2_" + tt.name)
			result := g.ProcessTemplateWithContext(tt.template, nil, session)

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestTreeProcessorGenderTag tests the <gender> tag with tree processor
func TestTreeProcessorGenderTag(t *testing.T) {
	g := New(false)
	g.EnableTreeProcessing() // Enable AST-based processing

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "Basic masculine to feminine",
			template: "<gender>he is a doctor</gender>",
			expected: "she is a doctor",
		},
		{
			name:     "Basic feminine to masculine",
			template: "<gender>she is a teacher</gender>",
			expected: "he is a teacher",
		},
		{
			name:     "Multiple pronouns",
			template: "<gender>he said she was here</gender>",
			expected: "she said he was here",
		},
		{
			name:     "Possessive pronouns",
			template: "<gender>his book and her pen</gender>",
			expected: "her book and his pen",
		},
		{
			name:     "Reflexive pronouns",
			template: "<gender>he helped himself and she helped herself</gender>",
			expected: "she helped herself and he helped himself",
		},
		{
			name:     "Contractions",
			template: "<gender>he's going and she's staying</gender>",
			expected: "she's going and he's staying",
		},
		{
			name:     "Complex sentence",
			template: "<gender>he said his friend told him she would help</gender>",
			expected: "she said her friend told her he would help",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := g.CreateSession("test_gender_" + tt.name)
			result := g.ProcessTemplateWithContext(tt.template, nil, session)

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestTreeProcessorTextTagsIntegration tests person, person2, and gender tags in AIML categories
func TestTreeProcessorTextTagsIntegration(t *testing.T) {
	g := New(false)
	g.EnableTreeProcessing()

	aimlContent := `<?xml version="1.0" encoding="UTF-8"?>
<aiml version="2.0">
	<category>
		<pattern>TELL ME WHAT I SAID</pattern>
		<template><person>you said hello</person></template>
	</category>
	
	<category>
		<pattern>CONVERT TO THIRD PERSON</pattern>
		<template><person2>I am happy</person2></template>
	</category>
	
	<category>
		<pattern>SWAP GENDER</pattern>
		<template><gender>he said she was here</gender></template>
	</category>
	
	<category>
		<pattern>COMBINED TEST</pattern>
		<template><person><gender>he told you</gender></person></template>
	</category>
</aiml>`

	err := g.LoadAIMLFromString(aimlContent)
	if err != nil {
		t.Fatalf("Failed to load AIML: %v", err)
	}

	session := &ChatSession{
		ID:              "test-text-tags",
		Variables:       make(map[string]string),
		History:         make([]string, 0),
		CreatedAt:       time.Now().Format(time.RFC3339),
		LastActivity:    time.Now().Format(time.RFC3339),
		Topic:           "",
		ThatHistory:     make([]string, 0),
		ResponseHistory: make([]string, 0),
		RequestHistory:  make([]string, 0),
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Person tag in category",
			input:    "tell me what i said",
			expected: "I said hello",
		},
		{
			name:     "Person2 tag in category",
			input:    "convert to third person",
			expected: "they are happy",
		},
		{
			name:     "Gender tag in category",
			input:    "swap gender",
			expected: "she said he was here",
		},
		{
			name:     "Nested person and gender tags",
			input:    "combined test",
			expected: "she told I",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := g.ProcessInput(tt.input, session)
			if err != nil {
				t.Fatalf("Failed to process input: %v", err)
			}

			if response != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, response)
			}
		})
	}
}

// TestTreeProcessorTextTagsWithWildcards tests text tags with wildcards
func TestTreeProcessorTextTagsWithWildcards(t *testing.T) {
	g := New(false)
	g.EnableTreeProcessing()

	aimlContent := `<?xml version="1.0" encoding="UTF-8"?>
<aiml version="2.0">
	<category>
		<pattern>PERSON ECHO *</pattern>
		<template><person><star/></person></template>
	</category>
	
	<category>
		<pattern>PERSON2 ECHO *</pattern>
		<template><person2><star/></person2></template>
	</category>
	
	<category>
		<pattern>GENDER ECHO *</pattern>
		<template><gender><star/></gender></template>
	</category>
</aiml>`

	err := g.LoadAIMLFromString(aimlContent)
	if err != nil {
		t.Fatalf("Failed to load AIML: %v", err)
	}

	session := &ChatSession{
		ID:              "test-wildcards",
		Variables:       make(map[string]string),
		History:         make([]string, 0),
		CreatedAt:       time.Now().Format(time.RFC3339),
		LastActivity:    time.Now().Format(time.RFC3339),
		Topic:           "",
		ThatHistory:     make([]string, 0),
		ResponseHistory: make([]string, 0),
		RequestHistory:  make([]string, 0),
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Person with wildcard",
			input:    "person echo I love you",
			expected: "you love I",
		},
		{
			name:     "Person2 with wildcard",
			input:    "person2 echo I am happy",
			expected: "they are happy",
		},
		{
			name:     "Gender with wildcard",
			input:    "gender echo he said hello",
			expected: "she said hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := g.ProcessInput(tt.input, session)
			if err != nil {
				t.Fatalf("Failed to process input: %v", err)
			}

			if response != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, response)
			}
		})
	}
}
