package engine

import (
	"fmt"
	"os"
	"strings"
	"golem/parser"
	"bufio"
	"path/filepath"
)

// Config represents the configuration for the bot
type Config struct {
	Debug bool
}

// Bot represents an AIML chatbot
type Bot struct {
	matchTree *MatchTree
	parser    *parser.Parser
	sessions  *SessionManager
	config    *Config
	Sets      map[string]map[string]struct{} // set name -> set of values
	Maps      map[string]map[string]string   // map name -> key->value
}

// NewBot creates a new Bot instance
func NewBot(debug bool) *Bot {
	cfg := &Config{Debug: debug}
	return &Bot{
		matchTree: NewMatchTree(),
		parser:    parser.NewParser(debug),
		sessions:  NewSessionManager(),
		config:    cfg,
		Sets:      make(map[string]map[string]struct{}),
		Maps:      make(map[string]map[string]string),
	}
}

// LoadAIML loads AIML categories from the specified path
func (b *Bot) LoadAIML(path string) error {
	if b.config != nil && b.config.Debug {
		fmt.Fprintf(os.Stderr, "Loading AIML from: %s\n", path)
	}
	cats, err := b.parser.ParseFile(path)
	if err != nil {
		return err
	}
	for _, cat := range cats {
		b.matchTree.Insert(cat)
	}
	if b.config != nil && b.config.Debug {
		fmt.Fprintf(os.Stderr, "Loaded %d categories from %s\n", len(cats), path)
	}
	return nil
}

// Respond generates a response for the given input and session
func (b *Bot) Respond(input string, sessionID string) (string, error) {
	if b.config != nil && b.config.Debug {
		fmt.Fprintf(os.Stderr, "[session=%s] Input: %s\n", sessionID, input)
	}
	inputNorm := strings.TrimSpace(strings.ToUpper(input))
	sess := b.sessions.GetOrCreateSession(sessionID)
	that := sess.That
	topic := sess.Topic
	res, found := b.matchTree.MatchWithMeta(inputNorm, that, topic)
	if !found {
		return "I don't have any knowledge loaded yet. Please load some AIML files first.", nil
	}
	if b.config != nil && b.config.Debug {
		fmt.Fprintf(os.Stderr, "[session=%s] Matched category: pattern=%q, that=%q, topic=%q\n", sessionID, res.MatchedPattern, res.MatchedThat, res.MatchedTopic)
		fmt.Fprintf(os.Stderr, "[session=%s] Wildcard captures: %+v\n", sessionID, res.WildcardCaptures)
	}
	// Prepare evaluator with session context
	userSession := &Session{
		Vars:      sess.Vars,
		Wildcards: res.WildcardCaptures,
	}
	eval := NewEvaluatorWithConfig(userSession, func(sraiInput string) (string, error) {
		// SRAI recursion: match again with new input, same session
		return b.Respond(sraiInput, sessionID)
	}, b.config, sessionID, b)
	output, err := eval.EvaluateTemplate(res.Template)
	if err != nil {
		return "[Error evaluating template]", err
	}
	// Update session context
	b.sessions.UpdateThat(sessionID, output)
	if res.MatchedTopic != "" {
		b.sessions.UpdateTopic(sessionID, res.MatchedTopic)
	}
	return output, nil
}

// LoadSet loads a .set file (one value per line, set name is file base name)
func (b *Bot) LoadSet(path string) error {
	name := strings.ToUpper(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)))
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	if b.Sets[name] == nil {
		b.Sets[name] = make(map[string]struct{})
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		val := strings.ToUpper(strings.TrimSpace(scanner.Text()))
		if val != "" {
			b.Sets[name][val] = struct{}{}
		}
	}
	return scanner.Err()
}

// LoadMap loads a .map file (key value per line, map name is file base name)
func (b *Bot) LoadMap(path string) error {
	name := strings.ToUpper(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)))
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	if b.Maps[name] == nil {
		b.Maps[name] = make(map[string]string)
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			key := strings.ToUpper(parts[0])
			val := strings.ToUpper(strings.Join(parts[1:], " "))
			b.Maps[name][key] = val
		}
	}
	return scanner.Err()
} 