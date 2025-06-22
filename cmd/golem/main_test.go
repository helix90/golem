package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"io/ioutil"
)

func TestCLI_EndToEnd(t *testing.T) {
	// Create a temp AIML file
	aiml := `<?xml version="1.0" encoding="UTF-8"?>
<aiml version="1.0.1">
  <category>
    <pattern>HELLO</pattern>
    <template>Hello, human!</template>
  </category>
  <category>
    <pattern>WHAT IS YOUR NAME</pattern>
    <template>My name is Golem.</template>
  </category>
</aiml>`
	tmp, err := ioutil.TempFile("", "test_*.aiml")
	if err != nil {
		t.Fatalf("Failed to create temp AIML: %v", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write([]byte(aiml)); err != nil {
		t.Fatalf("Failed to write AIML: %v", err)
	}
	tmp.Close()

	// Run the CLI as a subprocess
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcessCLI", "--", "-load", tmp.Name())
	cmd.Env = append(os.Environ(), "GOLEM_CLI_HELPER=1")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to get stdin: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to get stdout: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start CLI: %v", err)
	}

	// Send input and read output
	inputs := []string{"hello", "what is your name", "exit"}
	outputs := []string{"Hello, human!", "My name is Golem."}
	go func() {
		for _, in := range inputs {
			stdin.Write([]byte(in + "\n"))
		}
		stdin.Close()
	}()
	buf, _ := ioutil.ReadAll(stdout)
	outStr := string(buf)

	for _, expect := range outputs {
		if !strings.Contains(outStr, expect) {
			t.Errorf("Expected output to contain %q, got: %s", expect, outStr)
		}
	}
}

// TestHelperProcessCLI is not a real test. It's used as a subprocess entry point.
func TestHelperProcessCLI(t *testing.T) {
	if os.Getenv("GOLEM_CLI_HELPER") != "1" {
		return
	}
	// Remove the test flags and run main
	os.Args = os.Args[len(os.Args)-3:]
	main()
	os.Exit(0)
} 