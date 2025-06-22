package engine

import (
	"strings"
	"golem/parser"
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
//
type MatchTree struct {
	root *MatchNode
}

func NewMatchTree() *MatchTree {
	return &MatchTree{root: newMatchNode()}
}

func newMatchNode() *MatchNode {
	return &MatchNode{children: make(map[string]*MatchNode)}
}

// Insert adds a category to the tree, indexed by pattern, that, topic
func (t *MatchTree) Insert(cat parser.Category) {
	patternWords := splitAIML(cat.Pattern)
	thatWords := splitAIML(cat.That)
	topicWords := splitAIML(cat.Topic)

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
	topicNode.category = &cat
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

// splitAIML splits a pattern/that/topic into words, uppercased, ignoring extra whitespace
func splitAIML(s string) []string {
	words := strings.Fields(strings.ToUpper(strings.TrimSpace(s)))
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

// Match finds the best matching category for the given pattern, that, and topic
// Returns the match result and whether a match was found
func (t *MatchTree) MatchWithMeta(input, that, topic string) (*MatchResult, bool) {
	inputWords := splitAIML(input)
	thatWords := splitAIML(that)
	topicWords := splitAIML(topic)
	if len(thatWords) == 0 {
		thatWords = []string{"*"}
	}
	if len(topicWords) == 0 {
		topicWords = []string{"*"}
	}

	// Try with provided that/topic, then fallback to '*'
	if res, found := matchNodeMeta(t.root, inputWords, thatWords, topicWords, 0, map[string][]string{"pattern": {}, "that": {}, "topic": {}}); found {
		return res, true
	}
	if res, found := matchNodeMeta(t.root, inputWords, []string{"*"}, topicWords, 0, map[string][]string{"pattern": {}, "that": {}, "topic": {}}); found {
		return res, true
	}
	if res, found := matchNodeMeta(t.root, inputWords, thatWords, []string{"*"}, 0, map[string][]string{"pattern": {}, "that": {}, "topic": {}}); found {
		return res, true
	}
	if res, found := matchNodeMeta(t.root, inputWords, []string{"*"}, []string{"*"}, 0, map[string][]string{"pattern": {}, "that": {}, "topic": {}}); found {
		return res, true
	}
	return nil, false
}

// matchNodeMeta recursively matches pattern, that, topic with support for wildcards and returns match metadata
func matchNodeMeta(node *MatchNode, pattern, that, topic []string, depth int, captures map[string][]string) (*MatchResult, bool) {
	if depth > matchRecursionLimit {
		return nil, false
	}
	if len(pattern) > 0 {
		w := pattern[0]
		// Try exact word
		if child, ok := node.children[w]; ok {
			if res, found := matchNodeMeta(child, pattern[1:], that, topic, depth+1, captures); found {
				return res, true
			}
		}
		// Try '_' wildcard
		if node.under != nil {
			capCopy := copyCaptures(captures)
			capCopy["pattern"] = append([]string{}, captures["pattern"]...)
			capCopy["pattern"] = append(capCopy["pattern"], w)
			if res, found := matchNodeMeta(node.under, pattern[1:], that, topic, depth+1, capCopy); found {
				return res, true
			}
		}
		// Try '*' wildcard (matches zero or more words, all possible splits)
		if node.star != nil {
			for i := 0; i <= len(pattern); i++ {
				capCopy := copyCaptures(captures)
				capCopy["pattern"] = append([]string{}, captures["pattern"]...)
				capCopy["pattern"] = append(capCopy["pattern"], pattern[:i]...)
				if res, found := matchNodeMeta(node.star, pattern[i:], that, topic, depth+1, capCopy); found {
					return res, true
				}
			}
		}
		return nil, false
	}
	// Pattern exhausted, now match that
	if len(pattern) == 0 && len(that) > 0 {
		// Only reset 'that' captures at the transition point
		return matchNodeMetaThatSection(node, that, topic, depth, captures)
	}
	// That exhausted, now match topic
	if len(pattern) == 0 && len(that) == 0 && len(topic) > 0 {
		// Only reset 'topic' captures at the transition point
		return matchNodeMetaTopicSection(node, topic, depth, captures)
	}
	// All exhausted, return category if present
	if node.category != nil {
		return &MatchResult{
			Category:         node.category,
			MatchedPattern:   node.category.Pattern,
			MatchedThat:      node.category.That,
			MatchedTopic:     node.category.Topic,
			Template:         node.category.Template,
			WildcardCaptures: captures,
		}, true
	}
	return nil, false
}

// Helper for matching the 'that' section, with correct wildcard capture
func matchNodeMetaThatSection(node *MatchNode, that, topic []string, depth int, captures map[string][]string) (*MatchResult, bool) {
	if len(that) == 0 {
		return matchNodeMeta(node, []string{}, that, topic, depth, captures)
	}
	capturesForThat := copyCaptures(captures)
	capturesForThat["that"] = []string{} // Reset only at the start
	// Now, recurse through the that section
	return matchNodeMetaThatSectionRecurse(node, that, topic, depth, capturesForThat)
}

func matchNodeMetaThatSectionRecurse(node *MatchNode, that, topic []string, depth int, captures map[string][]string) (*MatchResult, bool) {
	if len(that) == 0 {
		return matchNodeMeta(node, []string{}, []string{}, topic, depth, captures)
	}
	w := that[0]
	if child, ok := node.children[w]; ok {
		if res, found := matchNodeMetaThatSectionRecurse(child, that[1:], topic, depth+1, captures); found {
			return res, true
		}
	}
	if node.under != nil {
		capCopy := copyCaptures(captures)
		capCopy["that"] = append([]string{}, captures["that"]...)
		capCopy["that"] = append(capCopy["that"], w)
		if res, found := matchNodeMetaThatSectionRecurse(node.under, that[1:], topic, depth+1, capCopy); found {
			return res, true
		}
	}
	if node.star != nil {
		for i := 0; i <= len(that); i++ {
			capCopy := copyCaptures(captures)
			capCopy["that"] = append([]string{}, captures["that"]...)
			capCopy["that"] = append(capCopy["that"], that[:i]...)
			if res, found := matchNodeMetaThatSectionRecurse(node.star, that[i:], topic, depth+1, capCopy); found {
				return res, true
			}
		}
	}
	// Fallback: try '*' for this level
	if child, ok := node.children["*"]; ok {
		if res, found := matchNodeMetaThatSectionRecurse(child, that[1:], topic, depth+1, captures); found {
			return res, true
		}
	}
	return nil, false
}

// Helper for matching the 'topic' section, with correct wildcard capture
func matchNodeMetaTopicSection(node *MatchNode, topic []string, depth int, captures map[string][]string) (*MatchResult, bool) {
	if len(topic) == 0 {
		return matchNodeMeta(node, []string{}, []string{}, []string{}, depth, captures)
	}
	capturesForTopic := copyCaptures(captures)
	capturesForTopic["topic"] = []string{} // Reset only at the start
	return matchNodeMetaTopicSectionRecurse(node, topic, depth, capturesForTopic)
}

func matchNodeMetaTopicSectionRecurse(node *MatchNode, topic []string, depth int, captures map[string][]string) (*MatchResult, bool) {
	if len(topic) == 0 {
		return matchNodeMeta(node, []string{}, []string{}, []string{}, depth, captures)
	}
	w := topic[0]
	if child, ok := node.children[w]; ok {
		if res, found := matchNodeMetaTopicSectionRecurse(child, topic[1:], depth+1, captures); found {
			return res, true
		}
	}
	if node.under != nil {
		capCopy := copyCaptures(captures)
		capCopy["topic"] = append([]string{}, captures["topic"]...)
		capCopy["topic"] = append(capCopy["topic"], w)
		if res, found := matchNodeMetaTopicSectionRecurse(node.under, topic[1:], depth+1, capCopy); found {
			return res, true
		}
	}
	if node.star != nil {
		for i := 0; i <= len(topic); i++ {
			capCopy := copyCaptures(captures)
			capCopy["topic"] = append([]string{}, captures["topic"]...)
			capCopy["topic"] = append(capCopy["topic"], topic[:i]...)
			if res, found := matchNodeMetaTopicSectionRecurse(node.star, topic[i:], depth+1, capCopy); found {
				return res, true
			}
		}
	}
	// Fallback: try '*' for this level
	if child, ok := node.children["*"]; ok {
		if res, found := matchNodeMetaTopicSectionRecurse(child, topic[1:], depth+1, captures); found {
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
	res, found := t.MatchWithMeta(input, that, topic)
	if !found {
		return nil, false
	}
	return res.Category, true
} 