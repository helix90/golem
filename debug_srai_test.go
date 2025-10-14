package main

import (
	"fmt"

	"github.com/helix90/golem/pkg/golem"
)

func main() {
	g := golem.New(false)

	// Set up knowledge base
	kb := golem.NewAIMLKnowledgeBase()
	kb.Categories = []golem.Category{
		{Pattern: "HELLO", Template: "Hello! How can I help you today?"},
	}
	kb.Patterns = make(map[string]*golem.Category)
	for i := range kb.Categories {
		kb.Patterns[kb.Categories[i].Pattern] = &kb.Categories[i]
	}
	g.SetKnowledgeBase(kb)

	// Create session
	session := g.CreateSession("test_session")
	session.RequestHistory = []string{"HELLO"}

	// Test template
	template := "You said: <input/>, let me respond: <srai><input/></srai>"

	fmt.Printf("Template: %s\n", template)
	fmt.Printf("RequestHistory: %v\n", session.RequestHistory)

	result := g.ProcessTemplateWithContext(template, map[string]string{}, session)

	fmt.Printf("Result: %s\n", result)
}
