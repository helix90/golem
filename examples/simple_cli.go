package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"golem/engine"
)

func main() {
	bot := engine.NewBot(false)
	err := bot.LoadAIML("testdata/simple.aiml")
	if err != nil {
		log.Fatalf("Failed to load AIML: %v", err)
	}

	fmt.Println("AIML Bot: Type 'quit' to exit.")
	scanner := bufio.NewScanner(os.Stdin)
	sessionID := "cli-session"
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		if input == "quit" {
			break
		}
		response, err := bot.Respond(input, sessionID)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		fmt.Println(response)
	}
} 