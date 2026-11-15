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

// Malformed templates must never panic the AST parser.
func TestASTParserUnterminatedAttributeNoPanic(t *testing.T) {
	inputs := []string{
		`<get name="`,
		`<get name="x`,
		`<get name='`,
		`<get name=\"`,
		`<get name=\"x`,
		`<star index=`,
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Parser panicked on %q: %v", input, r)
				}
			}()
			if _, err := NewASTParser(input).Parse(); err != nil {
				t.Logf("Parse returned error: %v", err)
			}
		})
	}
}

func FuzzASTParser(f *testing.F) {
	seeds := []string{
		`<get name="x"/>`,
		`<uppercase><star/> and <get name="y"/></uppercase>`,
		`<condition name="a"><li value="1">one</li><li>other</li></condition>`,
		`<!-- c --><![CDATA[<raw>]]>text`,
		`<get name=\"x\"/>`,
		`<get name="`,
	}
	for _, seed := range seeds {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		root, _ := NewASTParser(input).Parse()
		_ = root.String()
	})
}

// Pattern validation is case-insensitive and accepts apostrophes, common
// punctuation, and non-ASCII letters. An invalid category is skipped without
// discarding the rest of the file.
func TestDirectoryLoadValidation(t *testing.T) {
	dir := t.TempDir()
	content := `<aiml version="2.0">
<category><pattern>hello there</pattern><template>lowercase</template></category>
<category><pattern>LET'S TALK ABOUT *</pattern><template>apostrophe</template></category>
<category><pattern>CAFÉ</pattern><template>accented</template></category>
<category><pattern>MAIL ME @ HOME</pattern><template>invalid</template></category>
</aiml>`
	if err := os.WriteFile(filepath.Join(dir, "mixed.aiml"), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	g := New(false)
	kb, err := g.LoadAIMLFromDirectory(dir)
	if err != nil {
		t.Fatalf("Failed to load directory: %v", err)
	}
	g.SetKnowledgeBase(kb)

	if got := len(kb.Categories); got != 3 {
		t.Fatalf("Expected 3 valid categories, got %d: %+v", got, kb.Categories)
	}

	session := g.CreateSession("test")
	for input, want := range map[string]string{
		"hello there":           "lowercase",
		"let's talk about cats": "apostrophe",
		"café":                  "accented",
	} {
		response, err := g.ProcessInput(input, session)
		if err != nil {
			t.Errorf("ProcessInput(%q) failed: %v", input, err)
			continue
		}
		if response != want {
			t.Errorf("ProcessInput(%q): expected '%s', got '%s'", input, want, response)
		}
	}
}

// LoadAIMLFromString validates categories like the file loaders do.
func TestLoadAIMLFromStringValidates(t *testing.T) {
	g := New(false)
	err := g.LoadAIMLFromString(`<aiml version="2.0">
<category><pattern></pattern><template>empty pattern</template></category>
</aiml>`)
	if err == nil {
		t.Fatal("Expected error when the only category has an empty pattern")
	}

	err = g.LoadAIMLFromString(`<aiml version="2.0">
<category><pattern>MAIL ME @ HOME</pattern><template>invalid</template></category>
<category><pattern>HELLO</pattern><template>valid</template></category>
</aiml>`)
	if err != nil {
		t.Fatalf("Failed to load AIML: %v", err)
	}
	categories := g.GetKnowledgeBase().Categories
	if len(categories) != 1 || categories[0].Pattern != "HELLO" {
		t.Errorf("Expected only the HELLO category, got %+v", categories)
	}
}
