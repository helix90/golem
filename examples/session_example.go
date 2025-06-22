package main

import (
	"fmt"
	"log"
	"golem/engine"
)

func main() {
	bot := engine.NewBot(false)
	err := bot.LoadAIML("testdata/simple.aiml")
	if err != nil {
		log.Fatalf("Failed to load AIML: %v", err)
	}

	sessionID := "user123"

	// First interaction
	input := "Hello"
	response, err := bot.Respond(input, sessionID)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("User: %s\nBot: %s\n", input, response)

	// Set a variable in the session
	bot.SessionManager().SetVar(sessionID, "mood", "happy")

	// Use a pattern that references the variable
	input = "How are you?"
	response, err = bot.Respond(input, sessionID)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("User: %s\nBot: %s\n", input, response)
} 