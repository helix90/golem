package engine

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

// Session holds per-user variables and wildcards
type Session struct {
	Vars      map[string]string
	Wildcards map[string][]string
}

// Evaluator evaluates AIML templates with session context
// Supports: <template>, <set>, <get>, <srai>, <think>, <condition>, <random>, <star>
type Evaluator struct {
	Session   *Session
	SraiFunc  func(string) (string, error)
	Config    *Config
	SessionID string
	Bot       *Bot
}

// NewEvaluator creates a new Evaluator
func NewEvaluator(session *Session, sraiFunc func(string) (string, error)) *Evaluator {
	return &Evaluator{
		Session:  session,
		SraiFunc: sraiFunc,
	}
}

// NewEvaluatorWithConfig creates a new Evaluator with config, sessionID, and bot
func NewEvaluatorWithConfig(session *Session, sraiFunc func(string) (string, error), config *Config, sessionID string, bot *Bot) *Evaluator {
	return &Evaluator{
		Session:   session,
		SraiFunc:  sraiFunc,
		Config:    config,
		SessionID: sessionID,
		Bot:       bot,
	}
}

// node is used for generic XML parsing
type node struct {
	XMLName xml.Name
	Attr    []xml.Attr `xml:",any,attr"`
	Nodes   []node     `xml:",any"`
	Text    string     `xml:",chardata"`
}

// EvaluateTemplate is a stub for now
func (e *Evaluator) EvaluateTemplate(template string) (string, error) {
	decoder := xml.NewDecoder(strings.NewReader("<template>" + template + "</template>"))
	var sb strings.Builder
	var setName string
	for {
		t, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		switch tok := t.(type) {
		case xml.CharData:
			sb.WriteString(string(tok))
		case xml.StartElement:
			switch tok.Name.Local {
			case "set":
				for _, attr := range tok.Attr {
					if attr.Name.Local == "name" || attr.Name.Local == "var" {
						setName = attr.Value
						break
					}
				}
				// Read the value inside <set>
				var val string
				for {
					innerT, _ := decoder.Token()
					if innerT == nil {
						break
					}
					switch innerTok := innerT.(type) {
					case xml.CharData:
						val += string(innerTok)
					case xml.EndElement:
						if innerTok.Name.Local == "set" {
							if setName != "" {
								e.Session.Vars[setName] = val
								if e.Config != nil && e.Config.Debug {
									e.debugf("<set> (var) %q = %q", setName, val)
								}
								sb.WriteString(val)
							}
							setName = ""
							break
						}
					}
					if setName == "" {
						break
					}
				}
			case "get":
				var getName string
				for _, attr := range tok.Attr {
					if attr.Name.Local == "name" || attr.Name.Local == "var" || attr.Name.Local == "car" {
						getName = attr.Value
						break
					}
				}
				if e.Config != nil && e.Config.Debug {
					e.debugf("<get> %q = %q", getName, e.Session.Vars[getName])
				}
				sb.WriteString(e.Session.Vars[getName])
				// Consume the self-closing <get/> or <get></get>
				for {
					innerT, _ := decoder.Token()
					if innerT == nil {
						break
					}
					if endTok, ok := innerT.(xml.EndElement); ok && endTok.Name.Local == "get" {
						break
					}
				}
			case "star":
				index := 1
				for _, attr := range tok.Attr {
					if attr.Name.Local == "index" {
						fmt.Sscanf(attr.Value, "%d", &index)
					}
				}
				words := e.Session.Wildcards["pattern"]
				if len(words) == 0 {
					// nothing captured
				} else if len(tok.Attr) == 0 {
					sb.WriteString(words[0])
				} else if index > 0 && index <= len(words) {
					sb.WriteString(words[index-1])
				}
				// Consume the self-closing <star/> or <star></star>
				for {
					innerT, _ := decoder.Token()
					if innerT == nil {
						break
					}
					if endTok, ok := innerT.(xml.EndElement); ok && endTok.Name.Local == "star" {
						break
					}
				}
			case "that":
				// <that/> returns the last bot response (from session context)
				if e.Bot != nil && e.SessionID != "" {
					that := e.Bot.sessions.GetThat(e.SessionID)
					sb.WriteString(that)
				}
				// Consume the self-closing <that/> or <that></that>
				for {
					innerT, _ := decoder.Token()
					if innerT == nil {
						break
					}
					if endTok, ok := innerT.(xml.EndElement); ok && endTok.Name.Local == "that" {
						break
					}
				}
			case "thatstar":
				index := 1
				for _, attr := range tok.Attr {
					if attr.Name.Local == "index" {
						fmt.Sscanf(attr.Value, "%d", &index)
					}
				}
				words := e.Session.Wildcards["that"]
				if len(words) == 0 {
					// nothing captured
				} else if len(tok.Attr) == 0 {
					sb.WriteString(words[0])
				} else if index > 0 && index <= len(words) {
					sb.WriteString(words[index-1])
				}
				// Consume the self-closing <thatstar/> or <thatstar></thatstar>
				for {
					innerT, _ := decoder.Token()
					if innerT == nil {
						break
					}
					if endTok, ok := innerT.(xml.EndElement); ok && endTok.Name.Local == "thatstar" {
						break
					}
				}
			case "srai":
				// Read the content inside <srai>
				var sraiContent strings.Builder
				for {
					innerT, _ := decoder.Token()
					if innerT == nil {
						break
					}
					switch innerTok := innerT.(type) {
					case xml.CharData:
						sraiContent.WriteString(string(innerTok))
					case xml.EndElement:
						if innerTok.Name.Local == "srai" {
							break
						}
					}
					if t, ok := innerT.(xml.EndElement); ok && t.Name.Local == "srai" {
						break
					}
				}
				if e.SraiFunc != nil {
					result, err := e.SraiFunc(strings.TrimSpace(sraiContent.String()))
					if err == nil {
						sb.WriteString(result)
					}
				}
			case "think":
				// Evaluate the contents of <think> for side effects, but do not output
				var thinkContent strings.Builder
				depth := 1
				for depth > 0 {
					innerT, err := decoder.Token()
					if err != nil {
						return "", err
					}
					switch innerTok := innerT.(type) {
					case xml.StartElement:
						if innerTok.Name.Local == "think" {
							depth++
						}
						// Write out the start tag
						thinkContent.WriteString("<" + innerTok.Name.Local)
						for _, attr := range innerTok.Attr {
							thinkContent.WriteString(fmt.Sprintf(" %s=\"%s\"", attr.Name.Local, attr.Value))
						}
						thinkContent.WriteString(">")
					case xml.EndElement:
						if innerTok.Name.Local == "think" {
							depth--
							if depth == 0 {
								break
							}
						}
						thinkContent.WriteString("</" + innerTok.Name.Local + ">")
					case xml.CharData:
						thinkContent.WriteString(string(innerTok))
					}
					if depth == 0 {
						break
					}
				}
				// Recursively evaluate the contents for side effects only
				_, err := e.EvaluateTemplate(thinkContent.String())
				if err != nil {
					return "", err
				}
				// Do not append anything to sb (no output)
			case "uniq":
				var subj, pred, obj string
				for {
					innerT, err := decoder.Token()
					if err == io.EOF {
						break
					}
					if err != nil {
						return "", err
					}
					switch innerTok := innerT.(type) {
					case xml.StartElement:
						switch innerTok.Name.Local {
						case "subj":
							subj, _ = e.evalInnerXML(decoder, &innerTok)
						case "pred":
							pred, _ = e.evalInnerXML(decoder, &innerTok)
						case "obj":
							obj, _ = e.evalInnerXML(decoder, &innerTok)
						}
					case xml.EndElement:
						if innerTok.Name.Local == "uniq" {
							goto uniq_done
						}
					}
				}
			uniq_done:
				if e.Bot == nil {
					break
				}
				if len(obj) > 0 && obj[0] == '?' {
					// Query: look up (subj, pred)
					if e.Bot.KnowledgeBase[subj] != nil {
						val := e.Bot.KnowledgeBase[subj][pred]
						if val != "" {
							// Set the variable (without '?')
							if e.Session != nil {
								varName := obj[1:]
								e.Session.Vars[varName] = val
							}
							sb.WriteString(val)
						}
					}
				} else {
					// Assertion: store (subj, pred, obj)
					if e.Bot.KnowledgeBase[subj] == nil {
						e.Bot.KnowledgeBase[subj] = make(map[string]string)
					}
					e.Bot.KnowledgeBase[subj][pred] = obj
					sb.WriteString(obj)
				}
				// End <uniq>
			}
		}
	}
	return strings.TrimSpace(sb.String()), nil
}

// Add this method to Evaluator
func (e *Evaluator) debugf(format string, args ...interface{}) {
	if e.Config != nil && e.Config.Debug {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

// Helper to evaluate the inner XML/text of a tag
func (e *Evaluator) evalInnerXML(decoder *xml.Decoder, start *xml.StartElement) (string, error) {
	var sb strings.Builder
	depth := 1
	for depth > 0 {
		t, err := decoder.Token()
		if err != nil {
			return "", err
		}
		switch tok := t.(type) {
		case xml.StartElement:
			depth++
			// Recursively evaluate nested tags
			val, err := e.evalInnerXML(decoder, &tok)
			if err != nil {
				return "", err
			}
			sb.WriteString(val)
		case xml.EndElement:
			if tok.Name.Local == start.Name.Local {
				depth--
			}
		case xml.CharData:
			sb.WriteString(string(tok))
		}
	}
	return strings.TrimSpace(sb.String()), nil
}
