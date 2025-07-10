package engine

import (
	"fmt"
	"golem/parser"
	"os"
	"regexp"
	"strings"
)

// MatchNode represents a node in the match tree
// Each node can have children for words, '*' and '_', and may store a category
// The tree is 3-level: pattern -> that -> topic
// (AIML spec: topic and that are optional, default to "*" if not present)
type MatchNode struct {
	children map[string]*MatchNode
	star     *MatchNode // '*' wildcard
	under    *MatchNode // '_' wildcard
	category *parser.Category
}

// MatchTree is the root of the match tree
// It stores the root node for all patterns
// Insert and Match operate on pattern, that, topic
// All keys are uppercased and trimmed
type MatchTree struct {
	root  *MatchNode
	Debug bool
}

func (t *MatchTree) debugf(format string, args ...interface{}) {
	if t != nil && t.Debug {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

// Helper to select the most specific match: exact > set > '_' > '*'
type matchType int

const (
	starMatch matchType = iota // lowest specificity
	underMatch
	setMatch
	exactMatch // highest specificity
)

type matchResultWithType struct {
	res   *MatchResult
	mtype matchType
}

func selectBestMatch(matches []matchResultWithType) (*MatchResult, bool) {
	best := -1
	var bestRes *MatchResult
	maxPatternCapture := -1
	for _, m := range matches {
		if m.res != nil {
			mtype := int(m.mtype)
			patternCapLen := 0
			if m.res.WildcardCaptures != nil {
				patternCapLen = len(m.res.WildcardCaptures["pattern"])
			}
			if mtype > best || (mtype == best && patternCapLen > maxPatternCapture) {
				best = mtype
				bestRes = m.res
				maxPatternCapture = patternCapLen
			}
		}
	}
	return bestRes, bestRes != nil
}

func NewMatchTree() *MatchTree {
	return &MatchTree{root: newMatchNode(), Debug: false}
}

func newMatchNode() *MatchNode {
	return &MatchNode{children: make(map[string]*MatchNode)}
}

// Insert adds a category to the tree, indexed by pattern, that, topic
func (t *MatchTree) Insert(cat parser.Category) {
	// Normalize pattern, that, topic to uppercase and trimmed
	cat.Pattern = strings.ToUpper(strings.TrimSpace(cat.Pattern))
	cat.That = strings.ToUpper(strings.TrimSpace(cat.That))
	cat.Topic = strings.ToUpper(strings.TrimSpace(cat.Topic))

	if cat.That == "" {
		cat.That = "*"
	}
	if cat.Topic == "" {
		cat.Topic = "*"
	}

	patternWords := splitAIML(cat.Pattern)
	thatWords := splitAIML(cat.That)
	topicWords := splitAIML(cat.Topic)

	t.debugf("[DEBUG] Insert: Pattern=%q That=%q Topic=%q", cat.Pattern, cat.That, cat.Topic)
	t.debugf("[DEBUG] Insert tokens: patternWords=%q thatWords=%q topicWords=%q", patternWords, thatWords, topicWords)
	// New: print normalized 'that' and its tokens at insertion time
	t.debugf("[DEBUG] Insert normalized that: %q", cat.That)
	t.debugf("[DEBUG] Insert that tokens: %q", thatWords)

	if len(thatWords) == 0 {
		thatWords = []string{"*"}
	}
	if len(topicWords) == 0 {
		topicWords = []string{"*"}
	}

	// Insert into tree: pattern -> that -> topic
	pnode := t.root
	for _, w := range patternWords {
		pnode = pnode.childFor(w)
	}
	tnode := pnode
	for _, w := range thatWords {
		tnode = tnode.childFor(w)
	}
	topicNode := tnode
	for _, w := range topicWords {
		topicNode = topicNode.childFor(w)
	}
	// Instead of newCat := cat; topicNode.category = &newCat
	topicNode.category = new(parser.Category)
	*topicNode.category = cat
}

// childFor returns the child node for a word, creating it if necessary
func (n *MatchNode) childFor(word string) *MatchNode {
	word = strings.ToUpper(strings.TrimSpace(word))
	switch word {
	case "*":
		if n.star == nil {
			n.star = newMatchNode()
		}
		return n.star
	case "_":
		if n.under == nil {
			n.under = newMatchNode()
		}
		return n.under
	default:
		if n.children[word] == nil {
			n.children[word] = newMatchNode()
		}
		return n.children[word]
	}
}

// splitAIML splits a pattern/that/topic into words, uppercased, and converts <set>NAME</set> (any case) to __SET_NAME__ tokens
func splitAIML(s string) []string {
	s = strings.TrimSpace(s)
	if strings.Contains(strings.ToLower(s), "<set>") {
		// Replace <set>NAME</set> (any case) with __SET_NAME__
		re := regexp.MustCompile(`(?i)<set>([^<]+)</set>`) // case-insensitive
		s = re.ReplaceAllStringFunc(s, func(m string) string {
			matches := re.FindStringSubmatch(m)
			if len(matches) == 2 {
				return " __SET_" + strings.ToUpper(strings.TrimSpace(matches[1])) + "__ "
			}
			return m
		})
	}
	words := strings.Fields(strings.ToUpper(s))
	return words
}

// MatchResult contains metadata about a successful match
// WildcardCaptures: map from wildcard position (pattern/that/topic) to the captured words
// MatchedPattern/That/Topic: the actual pattern/that/topic that matched (with wildcards expanded)
type MatchResult struct {
	Category         *parser.Category
	MatchedPattern   string
	MatchedThat      string
	MatchedTopic     string
	Template         string
	WildcardCaptures map[string][]string // keys: pattern, that, topic
}

const matchRecursionLimit = 20

// MatchWithMeta now takes both pattern and input tokens separately for proper set matching
func (t *MatchTree) MatchWithMeta(input, that, topic string, sets map[string]map[string]struct{}) (*MatchResult, bool) {
	inputWords := splitAIML(input)
	thatWords := splitAIML(that)
	topicWords := splitAIML(topic)
	if len(thatWords) == 0 {
		thatWords = []string{"*"}
	}
	if len(topicWords) == 0 {
		topicWords = []string{"*"}
	}

	t.debugf("[DEBUG] Match: Input=%q That=%q Topic=%q", strings.Join(inputWords, " "), strings.Join(thatWords, " "), strings.Join(topicWords, " "))
	t.debugf("[DEBUG] Match tokens: inputWords=%q thatWords=%q topicWords=%q", inputWords, thatWords, topicWords)
	// New: print normalized 'that' and its tokens at match time
	t.debugf("[DEBUG] Match normalized that: %q", that)
	t.debugf("[DEBUG] Match that tokens: %q", thatWords)

	// Start recursive match from the root node with all tokens
	res, found := matchNodeMetaPattern(t, t.root, inputWords, inputWords, thatWords, topicWords, 0, map[string][]string{"pattern": {}, "that": {}, "topic": {}}, sets)
	if found {
		return res, true
	}
	return nil, false
}

func matchNodeMetaPattern(t *MatchTree, node *MatchNode, pattern, input, that, topic []string, depth int, captures map[string][]string, sets map[string]map[string]struct{}) (*MatchResult, bool) {
	if depth > matchRecursionLimit {
		return nil, false
	}
	if len(pattern) > 0 {
		w := pattern[0]
		matches := []matchResultWithType{}
		// 1. Exact match
		if len(input) > 0 {
			if child, ok := node.children[w]; ok && w == input[0] {
				if res, found := matchNodeMetaPattern(t, child, pattern[1:], input[1:], that, topic, depth+1, captures, sets); found {
					matches = append(matches, matchResultWithType{res, exactMatch})
				}
			}
		}
		// 2. Set match (try all set children)
		if len(input) > 0 {
			for childKey, childNode := range node.children {
				if strings.HasPrefix(childKey, "__SET_") && strings.HasSuffix(childKey, "__") {
					setName := strings.TrimSuffix(strings.TrimPrefix(childKey, "__SET_"), "__")
					if set, exists := sets[setName]; exists {
						_, isMember := set[input[0]]
						if isMember {
							capCopy := copyCaptures(captures)
							capCopy["pattern"] = append(capCopy["pattern"], input[0])
							if res, found := matchNodeMetaPattern(t, childNode, pattern[1:], input[1:], that, topic, depth+1, capCopy, sets); found {
								matches = append(matches, matchResultWithType{res, setMatch})
							}
						}
					}
				}
			}
		}
		// 3. '_' wildcard (single word)
		if node.under != nil && len(input) > 0 {
			capCopy := copyCaptures(captures)
			capCopy["pattern"] = append(capCopy["pattern"], input[0])
			if res, found := matchNodeMetaPattern(t, node.under, pattern[1:], input[1:], that, topic, depth+1, capCopy, sets); found {
				matches = append(matches, matchResultWithType{res, underMatch})
			}
		}
		// 4. '*' wildcard (zero or more words, all splits)
		if node.star != nil {
			for i := 0; i <= len(input); i++ {
				capCopy := copyCaptures(captures)
				capCopy["pattern"] = append(capCopy["pattern"], input[:i]...)
				if res, found := matchNodeMetaPattern(t, node.star, pattern[1:], input[i:], that, topic, depth+1, capCopy, sets); found {
					matches = append(matches, matchResultWithType{res, starMatch})
				}
			}
		}
		return selectBestMatch(matches)
	}
	// Pattern exhausted, now match that
	return matchNodeMetaThatSection(t, node, that, topic, depth, captures, sets)
}

func matchNodeMetaThatSection(t *MatchTree, node *MatchNode, that, topic []string, depth int, captures map[string][]string, sets map[string]map[string]struct{}) (*MatchResult, bool) {
	capCopy := copyCaptures(captures)
	capCopy["that"] = []string{}
	return matchNodeMetaThatSectionRecurse(t, nil, node, that, topic, depth, capCopy, sets)
}

func matchNodeMetaThatSectionRecurse(t *MatchTree, parent *MatchNode, node *MatchNode, that, topic []string, depth int, captures map[string][]string, sets map[string]map[string]struct{}) (*MatchResult, bool) {
	if depth > matchRecursionLimit {
		return nil, false
	}
	if len(that) > 0 {
		w := that[0]
		matches := []matchResultWithType{}
		// DEBUG: Print which that node is being tried
		t.debugf("[DEBUG] That recurse: depth=%d, thatWord=%q, that=%q", depth, w, that)
		// 1. Exact match
		if child, ok := node.children[w]; ok {
			t.debugf("[DEBUG] That recurse: trying exact match for %q at depth %d", w, depth)
			if res, found := matchNodeMetaThatSectionRecurse(t, node, child, that[1:], topic, depth+1, captures, sets); found {
				matches = append(matches, matchResultWithType{res, exactMatch})
			}
		}
		// 2. Set match (try all set children)
		for childKey, childNode := range node.children {
			if strings.HasPrefix(childKey, "__SET_") && strings.HasSuffix(childKey, "__") {
				setName := strings.TrimSuffix(strings.TrimPrefix(childKey, "__SET_"), "__")
				if set, exists := sets[setName]; exists {
					_, isMember := set[that[0]]
					t.debugf("[DEBUG] That recurse: trying set match for %q in set %q at depth %d", w, setName, depth)
					if isMember {
						capCopy := copyCaptures(captures)
						capCopy["that"] = append(capCopy["that"], that[0])
						if res, found := matchNodeMetaThatSectionRecurse(t, node, childNode, that[1:], topic, depth+1, capCopy, sets); found {
							matches = append(matches, matchResultWithType{res, setMatch})
						}
					}
				}
			}
		}
		// 3. '_' wildcard (single word)
		if node.under != nil {
			t.debugf("[DEBUG] That recurse: trying _ wildcard at depth %d", depth)
			capCopy := copyCaptures(captures)
			capCopy["that"] = append(capCopy["that"], that[0])
			if res, found := matchNodeMetaThatSectionRecurse(t, node, node.under, that[1:], topic, depth+1, capCopy, sets); found {
				matches = append(matches, matchResultWithType{res, underMatch})
			}
		}
		// 4. '*' wildcard (zero or more words, all splits)
		var starThatMatchFound bool
		if node.star != nil {
			for i := 0; i <= len(that); i++ {
				t.debugf("[DEBUG] That recurse: trying * wildcard for i=%d at depth %d", i, depth)
				capCopy := copyCaptures(captures)
				capCopy["that"] = append(capCopy["that"], that[:i]...)
				if res, found := matchNodeMetaThatSectionRecurse(t, node, node.star, that[i:], topic, depth+1, capCopy, sets); found {
					matches = append(matches, matchResultWithType{res, starMatch})
					starThatMatchFound = true
				}
			}
		}
		// FINAL FALLBACK: always try '*' node as a fallback if no matches found, without consuming that tokens
		if !starThatMatchFound && node.star != nil {
			t.debugf("[DEBUG] That recurse: final fallback to * node at depth %d", depth)
			capCopy := copyCaptures(captures)
			if res, found := matchNodeMetaThatSectionRecurse(t, node, node.star, that, topic, depth+1, capCopy, sets); found {
				matches = append(matches, matchResultWithType{res, starMatch})
			}
		}
		// NEW: If no matches found, and parent has a star child, and node is not that child, try fallback from parent.star
		if len(matches) == 0 && parent != nil && parent.star != nil && parent.star != node {
			t.debugf("[DEBUG] That recurse: parent fallback to * node at depth %d", depth)
			capCopy := copyCaptures(captures)
			if res, found := matchNodeMetaThatSectionRecurse(t, parent, parent.star, that, topic, depth+1, capCopy, sets); found {
				matches = append(matches, matchResultWithType{res, starMatch})
			}
		}
		return selectBestMatch(matches)
	}
	// That exhausted, now match topic
	// Fallback: try star child if present
	if node.star != nil {
		capCopy := copyCaptures(captures)
		if res, found := matchNodeMetaThatSectionRecurse(t, node, node.star, that, topic, depth+1, capCopy, sets); found {
			return res, true
		}
	}
	return matchNodeMetaTopicSection(t, node, topic, depth, captures, sets)
}

func matchNodeMetaTopicSection(t *MatchTree, node *MatchNode, topic []string, depth int, captures map[string][]string, sets map[string]map[string]struct{}) (*MatchResult, bool) {
	capCopy := copyCaptures(captures)
	capCopy["topic"] = []string{}
	return matchNodeMetaTopicSectionRecurse(t, nil, node, topic, depth, capCopy, sets)
}

func matchNodeMetaTopicSectionRecurse(t *MatchTree, parent *MatchNode, node *MatchNode, topic []string, depth int, captures map[string][]string, sets map[string]map[string]struct{}) (*MatchResult, bool) {
	if depth > matchRecursionLimit {
		return nil, false
	}
	if len(topic) > 0 {
		w := topic[0]
		matches := []matchResultWithType{}
		// DEBUG: Print which topic node is being tried
		t.debugf("[DEBUG] Topic recurse: depth=%d, topicWord=%q", depth, w)
		// 1. Exact match
		if child, ok := node.children[w]; ok {
			t.debugf("[DEBUG] Topic recurse: trying exact match for %q at depth %d", w, depth)
			if res, found := matchNodeMetaTopicSectionRecurse(t, node, child, topic[1:], depth+1, captures, sets); found {
				matches = append(matches, matchResultWithType{res, exactMatch})
			}
		}
		// 2. Set match (try all set children)
		for childKey, childNode := range node.children {
			if strings.HasPrefix(childKey, "__SET_") && strings.HasSuffix(childKey, "__") {
				setName := strings.TrimSuffix(strings.TrimPrefix(childKey, "__SET_"), "__")
				if set, exists := sets[setName]; exists {
					_, isMember := set[topic[0]]
					if isMember {
						t.debugf("[DEBUG] Topic recurse: trying set match for %q in set %q at depth %d", w, setName, depth)
						if res, found := matchNodeMetaTopicSectionRecurse(t, node, childNode, topic[1:], depth+1, captures, sets); found {
							matches = append(matches, matchResultWithType{res, setMatch})
						}
					}
				}
			}
		}
		// 3. '_' wildcard (single word)
		if node.under != nil {
			t.debugf("[DEBUG] Topic recurse: trying _ wildcard at depth %d", depth)
			capCopy := copyCaptures(captures)
			capCopy["topic"] = append(capCopy["topic"], topic[0])
			if res, found := matchNodeMetaTopicSectionRecurse(t, node, node.under, topic[1:], depth+1, capCopy, sets); found {
				matches = append(matches, matchResultWithType{res, underMatch})
			}
		}
		// 4. '*' wildcard (zero or more words, all splits)
		var starTopicMatchFound bool
		if node.star != nil {
			for i := 0; i <= len(topic); i++ {
				t.debugf("[DEBUG] Topic recurse: trying * wildcard for i=%d at depth %d", i, depth)
				capCopy := copyCaptures(captures)
				capCopy["topic"] = append(capCopy["topic"], topic[:i]...)
				if res, found := matchNodeMetaTopicSectionRecurse(t, node, node.star, topic[i:], depth+1, capCopy, sets); found {
					matches = append(matches, matchResultWithType{res, starMatch})
					starTopicMatchFound = true
				}
			}
		}
		// FINAL FALLBACK: always try '*' node as a fallback if no matches found, without consuming topic tokens
		if !starTopicMatchFound && node.star != nil {
			t.debugf("[DEBUG] Topic recurse: final fallback to * node at depth %d", depth)
			capCopy := copyCaptures(captures)
			if res, found := matchNodeMetaTopicSectionRecurse(t, node, node.star, topic, depth+1, capCopy, sets); found {
				matches = append(matches, matchResultWithType{res, starMatch})
			}
		}
		// NEW: If no matches found, and parent has a star child, and node is not that child, try fallback from parent.star
		if len(matches) == 0 && parent != nil && parent.star != nil && parent.star != node {
			t.debugf("[DEBUG] Topic recurse: parent fallback to * node at depth %d", depth)
			capCopy := copyCaptures(captures)
			if res, found := matchNodeMetaTopicSectionRecurse(t, parent, parent.star, topic, depth+1, capCopy, sets); found {
				matches = append(matches, matchResultWithType{res, starMatch})
			}
		}
		return selectBestMatch(matches)
	}
	// All exhausted, return category if present
	if node.category != nil {
		t.debugf("[DEBUG] Matched: Pattern=%q That=%q Topic=%q Template=%q", node.category.Pattern, node.category.That, node.category.Topic, string(node.category.Template))
		return &MatchResult{
			Category:         node.category,
			MatchedPattern:   node.category.Pattern,
			MatchedThat:      node.category.That,
			MatchedTopic:     node.category.Topic,
			Template:         string(node.category.Template),
			WildcardCaptures: captures,
		}, true
	}
	// Fallback: try star child if present
	if node.star != nil {
		capCopy := copyCaptures(captures)
		if res, found := matchNodeMetaTopicSectionRecurse(t, node, node.star, topic, depth+1, capCopy, sets); found {
			return res, true
		}
	}
	return nil, false
}

func copyCaptures(src map[string][]string) map[string][]string {
	out := make(map[string][]string, len(src))
	for k, v := range src {
		out[k] = append([]string{}, v...)
	}
	return out
}

// Match is the legacy interface, returns only the template
func (t *MatchTree) Match(input, that, topic string) (*parser.Category, bool) {
	res, found := t.MatchWithMeta(input, that, topic, nil)
	if !found {
		return nil, false
	}
	return res.Category, true
}
