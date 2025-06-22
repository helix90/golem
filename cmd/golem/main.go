package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"golem/engine"
	"archive/zip"
	"io"
	"io/ioutil"
	"path/filepath"
)

func main() {
	var (
		loadPath = flag.String("load", "", "Path to AIML file, directory, or zip to load")
		debug    = flag.Bool("debug", false, "Enable debug output")
	)
	flag.Parse()

	// Initialize the bot
	bot := engine.NewBot(*debug)

	// Load AIML if specified
	if *loadPath != "" {
		if err := loadAIMLFiles(bot, *loadPath, *debug); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading AIML: %v\n", err)
			os.Exit(1)
		}
		if *debug {
			fmt.Fprintf(os.Stderr, "AIML loaded from: %s\n", *loadPath)
		}
	}

	// Generate a simple session ID
	sessionID := "default"

	fmt.Println("Golem AIML Bot")
	fmt.Println("Type 'quit' or 'exit' to exit")
	fmt.Println("Type your message and press Enter:")
	fmt.Println()

	// Read from stdin in a loop
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Check for exit commands
		if input == "quit" || input == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		// Get bot response
		response, err := bot.Respond(input, sessionID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}

		fmt.Println(response)
		fmt.Println()
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
}

// loadAIMLFiles loads AIML from a file, directory, or zip
func loadAIMLFiles(bot *engine.Bot, path string, debug bool) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return err
		}
		for _, f := range files {
			if !f.IsDir() {
				ext := filepath.Ext(f.Name())
				switch ext {
				case ".aiml":
					if err := bot.LoadAIML(filepath.Join(path, f.Name())); err != nil {
						return err
					}
				case ".set":
					if err := bot.LoadSet(filepath.Join(path, f.Name())); err != nil {
						return err
					}
					fmt.Fprintf(os.Stderr, "Loaded .set file: %s\n", f.Name())
				case ".map":
					if err := bot.LoadMap(filepath.Join(path, f.Name())); err != nil {
						return err
					}
					fmt.Fprintf(os.Stderr, "Loaded .map file: %s\n", f.Name())
				}
			}
		}
		return nil
	}
	if filepath.Ext(path) == ".zip" {
		return loadAIMLFromZip(bot, path, debug)
	}
	// Single file
	return bot.LoadAIML(path)
}

// loadAIMLFromZip extracts and loads all AIML files from a zip archive
func loadAIMLFromZip(bot *engine.Bot, zipPath string, debug bool) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		if filepath.Ext(f.Name) == ".aiml" {
			if debug {
				fmt.Fprintf(os.Stderr, "Extracting %s from zip\n", f.Name)
			}
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()
			tmp, err := ioutil.TempFile("", "golem_zip_*.aiml")
			if err != nil {
				return err
			}
			defer os.Remove(tmp.Name())
			if _, err := io.Copy(tmp, rc); err != nil {
				tmp.Close()
				return err
			}
			tmp.Close()
			if err := bot.LoadAIML(tmp.Name()); err != nil {
				return err
			}
		}
	}
	return nil
} 