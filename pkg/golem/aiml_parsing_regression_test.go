package golem

import (
	"os"
	"path/filepath"
	"testing"
)

const thatCollisionAIML = `<aiml version="2.0">
<category><pattern>DO YOU LIKE IT</pattern><template>Do you like cats?</template></category>
<category><pattern>YES</pattern><that>DO YOU LIKE CATS</that><template>cats yes</template></category>
<category><pattern>YES</pattern><template>generic yes</template></category>
</aiml>`

// Categories sharing a pattern but differing by <that> must all be indexed,
// regardless of which loader is used.
func TestLoadersKeepCategoriesSharingPattern(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "that.aiml")
	if err := os.WriteFile(file, []byte(thatCollisionAIML), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	loaders := map[string]func(g *Golem) error{
		"LoadAIMLFromString": func(g *Golem) error {
			return g.LoadAIMLFromString(thatCollisionAIML)
		},
		"LoadAIML": func(g *Golem) error {
			kb, err := g.LoadAIML(file)
			if err == nil {
				g.SetKnowledgeBase(kb)
			}
			return err
		},
		"LoadAIMLFromDirectory": func(g *Golem) error {
			kb, err := g.LoadAIMLFromDirectory(dir)
			if err == nil {
				g.SetKnowledgeBase(kb)
			}
			return err
		},
	}

	for name, load := range loaders {
		t.Run(name, func(t *testing.T) {
			g := New(false)
			if err := load(g); err != nil {
				t.Fatalf("Failed to load AIML: %v", err)
			}

			if got := len(g.GetKnowledgeBase().Patterns); got != 3 {
				t.Errorf("Expected 3 indexed patterns, got %d", got)
			}

			session := g.CreateSession("test")
			if _, err := g.ProcessInput("do you like it", session); err != nil {
				t.Fatalf("Failed to process input: %v", err)
			}
			response, err := g.ProcessInput("yes", session)
			if err != nil {
				t.Fatalf("Failed to process input: %v", err)
			}
			if response != "cats yes" {
				t.Errorf("Expected 'cats yes', got '%s'", response)
			}
		})
	}
}

// Categories inside comments must not be loaded, even when the comment spans
// multiple lines.
func TestCommentedOutCategoriesNotLoaded(t *testing.T) {
	g := New(false)
	err := g.LoadAIMLFromString(`<aiml version="2.0">
<!--
<category><pattern>HIDDEN</pattern><template>multi-line comment</template></category>
-->
<!-- <category><pattern>HIDDEN TOO</pattern><template>single-line comment</template></category> -->
<category><pattern>HELLO</pattern><template>real</template></category>
</aiml>`)
	if err != nil {
		t.Fatalf("Failed to load AIML: %v", err)
	}

	categories := g.GetKnowledgeBase().Categories
	if len(categories) != 1 {
		t.Fatalf("Expected 1 category, got %d: %+v", len(categories), categories)
	}
	if categories[0].Pattern != "HELLO" {
		t.Errorf("Expected pattern 'HELLO', got '%s'", categories[0].Pattern)
	}
}
