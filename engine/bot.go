package engine

import (
	"fmt"
	"os"
	"strings"
	"golem/parser"
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
}

// NewBot creates a new Bot instance
func NewBot(debug bool) *Bot {
	cfg := &Config{Debug: debug}
	return &Bot{
		matchTree: NewMatchTree(),
		parser:    parser.NewParser(debug),
		sessions:  NewSessionManager(),
		config:    cfg,
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
	}, b.config, sessionID)
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