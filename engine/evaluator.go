package engine

import (
	"encoding/xml"
	"errors"
	"math/rand"
	"time"
	"fmt"
	"strings"
	"os"
)

type Session struct {
	Vars      map[string]string
	Wildcards map[string][]string // e.g., "pattern", "that", "topic"
}

func NewSession() *Session {
	return &Session{
		Vars:      make(map[string]string),
		Wildcards: make(map[string][]string),
	}
}

// Evaluator evaluates AIML templates with session context
// Supports: <template>, <set>, <get>, <srai>, <think>, <condition>, <random>, <star>
type Evaluator struct {
	Session   *Session
	SraiFunc  func(string) (string, error) // For <srai> recursion
	Config    *Config
	SessionID string
	Bot       *Bot // Reference to parent bot for sets/maps
}

func NewEvaluator(session *Session, sraiFunc func(string) (string, error)) *Evaluator {
	return &Evaluator{
		Session:  session,
		SraiFunc: sraiFunc,
	}
}

func NewEvaluatorWithConfig(session *Session, sraiFunc func(string) (string, error), config *Config, sessionID string, bot *Bot) *Evaluator {
	return &Evaluator{
		Session:   session,
		SraiFunc:  sraiFunc,
		Config:    config,
		SessionID: sessionID,
		Bot:       bot,
	}
}

// Template XML node
// Used for generic parsing
//
type node struct {
	XMLName xml.Name
	Attr    []xml.Attr    `xml:",any,attr"`
	Nodes   []node        `xml:",any"`
	Text    string        `xml:",chardata"`
}

func (e *Evaluator) debugf(format string, args ...interface{}) {
	if e.Config != nil && e.Config.Debug {
		prefix := ""
		if e.SessionID != "" {
			prefix = fmt.Sprintf("[session=%s] ", e.SessionID)
		}
		fmt.Fprintf(os.Stderr, prefix+format+"\n", args...)
	}
}

func (e *Evaluator) EvaluateTemplate(template string) (string, error) {
	n := node{}
	err := xml.Unmarshal([]byte("<template>"+template+"</template>"), &n)
	if err != nil {
		return "", err
	}
	if e.Config != nil && e.Config.Debug {
		e.debugf("Evaluating template: %q", template)
	}

	// Find leading and trailing text by locating the first and last tag in the template
	leading, trailing := "", ""
	firstTag := strings.Index(template, "<")
	lastTag := strings.LastIndex(template, ">")
	if firstTag == -1 || lastTag == -1 {
		// No tags, treat entire template as leading text
		leading = template
	} else {
		if firstTag > 0 {
			leading = template[:firstTag]
		}
		if lastTag >= 0 && lastTag+1 < len(template) {
			trailing = template[lastTag+1:]
		}
	}

	var sb strings.Builder
	if leading != "" {
		sb.WriteString(leading)
	}
	children, err := e.evalNodeChildren(n.Nodes)
	if err != nil {
		return "", err
	}
	sb.WriteString(children)
	if trailing != "" {
		sb.WriteString(trailing)
	}
	return sb.String(), nil
}

func (e *Evaluator) evalNodeChildren(nodes []node) (string, error) {
	var sb strings.Builder
	for _, n := range nodes {
		if n.XMLName.Local == "srai" {
			out, err := e.evalNode(n)
			if err != nil {
				return "", err
			}
			sb.WriteString(out)
			continue
		}
		if n.XMLName.Local == "set" {
			_, err := e.evalNode(n)
			if err != nil {
				return "", err
			}
			continue
		}
		if n.XMLName.Local == "" && n.Text != "" {
			sb.WriteString(n.Text)
			continue
		}
		if n.XMLName.Local != "" && n.XMLName.Local != "set" && n.XMLName.Local != "srai" {
			out, err := e.evalNode(n)
			if err != nil {
				return "", err
			}
			sb.WriteString(out)
		}
	}
	return sb.String(), nil
}

func (e *Evaluator) evalNode(n node) (string, error) {
	switch n.XMLName.Local {
	case "template":
		return e.evalNodeChildren(n.Nodes)
	case "set":
		name := ""
		for _, a := range n.Attr {
			if a.Name.Local == "name" {
				name = a.Value
			}
		}
		val := ""
		if len(n.Nodes) > 0 {
			var err error
			val, err = e.evalNodeChildren(n.Nodes)
			if err != nil {
				return "", err
			}
		} else if n.Text != "" {
			val = n.Text
		}
		if name != "" {
			// If this is a loaded set, add the value to the set
			setName := strings.ToUpper(name)
			valUC := strings.ToUpper(val)
			if e.Bot != nil && e.Bot.Sets != nil && e.Bot.Sets[setName] != nil {
				e.Bot.Sets[setName][valUC] = struct{}{}
				if e.Config != nil && e.Config.Debug {
					e.debugf("<set> (set) %q += %q", setName, valUC)
				}
			} else {
				e.Session.Vars[name] = val
				if e.Config != nil && e.Config.Debug {
					e.debugf("<set> (var) %q = %q", name, val)
				}
			}
		}
		return "", nil
	case "get":
		name := ""
		for _, a := range n.Attr {
			if a.Name.Local == "name" {
				name = a.Value
			}
		}
		val := e.Session.Vars[name]
		if e.Config != nil && e.Config.Debug {
			e.debugf("<get> %q = %q", name, val)
		}
		if val == "" {
			return "", nil
		}
		return val, nil
	case "srai":
		if e.SraiFunc == nil {
			return "", errors.New("SraiFunc not set")
		}
		input := n.Text
		if children, err := e.evalNodeChildren(n.Nodes); err == nil && children != "" {
			input = children
		}
		if e.Config != nil && e.Config.Debug {
			e.debugf("<srai> input: %q", input)
		}
		return e.SraiFunc(strings.TrimSpace(input))
	case "think":
		_, err := e.evalNodeChildren(n.Nodes)
		if err == nil && n.Text != "" {
			_ = n.Text
		}
		if e.Config != nil && e.Config.Debug {
			e.debugf("<think> (side effect only)")
		}
		return "", err
	case "condition":
		if e.Config != nil && e.Config.Debug {
			e.debugf("<condition> evaluation")
		}
		return e.evalCondition(n)
	case "random":
		if e.Config != nil && e.Config.Debug {
			e.debugf("<random> evaluation")
		}
		return e.evalRandom(n)
	case "star":
		if e.Config != nil && e.Config.Debug {
			e.debugf("<star> evaluation")
		}
		return e.evalStar(n)
	case "li":
		return e.evalNodeChildren(n.Nodes)
	case "map":
		name := ""
		for _, a := range n.Attr {
			if a.Name.Local == "name" {
				name = a.Value
			}
		}
		key := ""
		if len(n.Nodes) > 0 {
			var err error
			key, err = e.evalNodeChildren(n.Nodes)
			if err != nil {
				return "", err
			}
		} else if n.Text != "" {
			key = n.Text
		}
		mapName := strings.ToUpper(name)
		keyUC := strings.ToUpper(key)
		if mapName != "" && keyUC != "" && e.Bot != nil && e.Bot.Maps != nil && e.Bot.Maps[mapName] != nil {
			val := e.Bot.Maps[mapName][keyUC]
			if e.Config != nil && e.Config.Debug {
				e.debugf("<map> %q[%q] = %q", mapName, keyUC, val)
			}
			return val, nil
		}
		return "", nil
	default:
		var sb strings.Builder
		sb.WriteString(n.Text)
		for _, child := range n.Nodes {
			if child.Text != "" {
				sb.WriteString(child.Text)
			}
			if child.XMLName.Local != "" || len(child.Nodes) > 0 {
				out, err := e.evalNode(child)
				if err != nil {
					return "", err
				}
				sb.WriteString(out)
			}
		}
		return sb.String(), nil
	}
}

func (e *Evaluator) evalCondition(n node) (string, error) {
	// Support:
	// <condition name="var" value="val">...</condition>
	// <condition><li value="...">...</li></condition>
	// <condition><li>...</li></condition>
	// <condition set="setname" value="val">...</condition>
	// <condition map="mapname" key="key">...</condition>
	if len(n.Attr) > 0 {
		varName, varVal := "", ""
		setName, setVal := "", ""
		mapName, mapKey := "", ""
		for _, a := range n.Attr {
			switch strings.ToLower(a.Name.Local) {
			case "name":
				varName = a.Value
			case "value":
				varVal = a.Value
			case "set":
				setName = a.Value
			case "map":
				mapName = a.Value
			case "key":
				mapKey = a.Value
			}
		}
		if varName != "" && varVal != "" {
			if e.Config != nil && e.Config.Debug {
				e.debugf("<condition> checking var %q == %q (actual: %q)", varName, varVal, e.Session.Vars[varName])
			}
			if e.Session.Vars[varName] == varVal {
				if e.Config != nil && e.Config.Debug {
					e.debugf("<condition> var match: %q == %q", varName, varVal)
				}
				if len(n.Nodes) > 0 {
					return e.evalNodeChildren(n.Nodes)
				} else if n.Text != "" {
					return n.Text, nil
				}
				return "", nil
			}
			return "", nil
		}
		if setName != "" && setVal != "" && e.Bot != nil && e.Bot.Sets != nil && e.Bot.Sets[strings.ToUpper(setName)] != nil {
			setUC := strings.ToUpper(setName)
			valUC := strings.ToUpper(setVal)
			if e.Config != nil && e.Config.Debug {
				e.debugf("<condition> checking set %q contains %q; set contents: %+v", setUC, valUC, e.Bot.Sets[setUC])
			}
			if _, ok := e.Bot.Sets[setUC][valUC]; ok {
				if e.Config != nil && e.Config.Debug {
					e.debugf("<condition> set match: %q contains %q", setUC, valUC)
				}
				if len(n.Nodes) > 0 {
					return e.evalNodeChildren(n.Nodes)
				} else if n.Text != "" {
					return n.Text, nil
				}
				return "", nil
			}
			return "", nil
		}
		if mapName != "" && mapKey != "" && e.Bot != nil && e.Bot.Maps != nil && e.Bot.Maps[strings.ToUpper(mapName)] != nil {
			if e.Config != nil && e.Config.Debug {
				e.debugf("<condition> checking map %q[%q] (actual: %q)", strings.ToUpper(mapName), strings.ToUpper(mapKey), e.Bot.Maps[strings.ToUpper(mapName)][strings.ToUpper(mapKey)])
			}
			if val, ok := e.Bot.Maps[strings.ToUpper(mapName)][strings.ToUpper(mapKey)]; ok && val != "" {
				if e.Config != nil && e.Config.Debug {
					e.debugf("<condition> map match: %q[%q] = %q", strings.ToUpper(mapName), strings.ToUpper(mapKey), val)
				}
				if len(n.Nodes) > 0 {
					return e.evalNodeChildren(n.Nodes)
				} else if n.Text != "" {
					return n.Text, nil
				}
				return "", nil
			}
			return "", nil
		}
	}
	// Handle <condition><li ...>...</li></condition>
	for _, li := range n.Nodes {
		if li.XMLName.Local != "li" {
			continue
		}
		varName, varVal := "", ""
		setName, setVal := "", ""
		mapName, mapKey := "", ""
		for _, a := range li.Attr {
			switch strings.ToLower(a.Name.Local) {
			case "name":
				varName = a.Value
			case "value":
				varVal = a.Value
			case "set":
				setName = a.Value
			case "map":
				mapName = a.Value
			case "key":
				mapKey = a.Value
			}
		}
		if varName != "" && varVal != "" {
			if e.Session.Vars[varName] == varVal {
				if len(li.Nodes) > 0 {
					return e.evalNodeChildren(li.Nodes)
				} else if li.Text != "" {
					return li.Text, nil
				}
			}
			continue
		}
		if setName != "" && setVal != "" && e.Bot != nil && e.Bot.Sets != nil && e.Bot.Sets[strings.ToUpper(setName)] != nil {
			setUC := strings.ToUpper(setName)
			valUC := strings.ToUpper(setVal)
			if e.Config != nil && e.Config.Debug {
				e.debugf("<condition> checking set %q contains %q; set contents: %+v", setUC, valUC, e.Bot.Sets[setUC])
			}
			if _, ok := e.Bot.Sets[setUC][valUC]; ok {
				if len(li.Nodes) > 0 {
					return e.evalNodeChildren(li.Nodes)
				} else if li.Text != "" {
					return li.Text, nil
				}
			}
			continue
		}
		if mapName != "" && mapKey != "" && e.Bot != nil && e.Bot.Maps != nil && e.Bot.Maps[strings.ToUpper(mapName)] != nil {
			if val, ok := e.Bot.Maps[strings.ToUpper(mapName)][strings.ToUpper(mapKey)]; ok && val != "" {
				if len(li.Nodes) > 0 {
					return e.evalNodeChildren(li.Nodes)
				} else if li.Text != "" {
					return li.Text, nil
				}
			}
			continue
		}
		// Default <li> (no attributes): match if no other <li> matched
		if len(li.Attr) == 0 {
			if len(li.Nodes) > 0 {
				return e.evalNodeChildren(li.Nodes)
			} else if li.Text != "" {
				return li.Text, nil
			}
		}
	}
	return "", nil
}

func (e *Evaluator) evalRandom(n node) (string, error) {
	var options []node
	for _, li := range n.Nodes {
		if li.XMLName.Local == "li" {
			options = append(options, li)
		}
	}
	if len(options) == 0 {
		return "", nil
	}
	rand.Seed(time.Now().UnixNano())
	chosen := options[rand.Intn(len(options))]
	if chosen.Text != "" {
		return chosen.Text, nil
	}
	return e.evalNodeChildren(chosen.Nodes)
}

func (e *Evaluator) evalStar(n node) (string, error) {
	index := 1
	for _, a := range n.Attr {
		if a.Name.Local == "index" {
			fmt.Sscanf(a.Value, "%d", &index)
		}
	}
	words := e.Session.Wildcards["pattern"]
	if len(words) == 0 {
		return "", nil
	}
	if index <= 0 || index > len(words) {
		return "", nil
	}
	if len(n.Attr) == 0 {
		return strings.Join(words, " "), nil
	}
	return words[index-1], nil
}

// TODO: Implement handlers for each tag type
// func (e *Evaluator) handleSet(...)
// func (e *Evaluator) handleGet(...)
// func (e *Evaluator) handleSrai(...)
// func (e *Evaluator) handleThink(...)
// func (e *Evaluator) handleCondition(...)
// func (e *Evaluator) handleRandom(...)
// func (e *Evaluator) handleStar(...) 