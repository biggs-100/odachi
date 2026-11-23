package skills_test

import (
	"encoding/json"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/skills"
)

// TestDesignDeclaresItsOutput: the delivery skill says what it hands back, and
// the formats of the folder it writes live in one reference so the body keeps
// its sixty-line budget.
func TestDesignDeclaresItsOutput(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}

	var design skills.Skill
	for _, s := range all {
		if s.Name == "design" {
			design = s
		}
	}
	if design.Name == "" {
		t.Fatal("design is not in the payload")
	}

	want := []string{"SKILL.md", "references/files.md"}
	got := make([]string, 0, len(design.Files))
	for _, file := range design.Files {
		got = append(got, file.Path)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("design carries %v, want %v", got, want)
	}

	main := string(design.Main())
	for _, section := range []string{"## Output", "## References", "`references/files.md`"} {
		if !strings.Contains(main, section) {
			t.Errorf("the design skill does not say %q", section)
		}
	}
	files := string(design.Files[1].Content)
	for _, want := range []string{"design.json", "comments.json", "screenshot"} {
		if !strings.Contains(files, want) {
			t.Errorf("references/files.md does not say %q", want)
		}
	}
}

// TestDesignNamesOnlyReferencesThatExist: the names the design skill tells the
// reader to load are real references in the payload; a name with no file is a
// broken instruction that sends the reader nowhere.
func TestDesignNamesOnlyReferencesThatExist(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}

	var design skills.Skill
	names := make([]string, 0)
	for _, s := range all {
		if s.Name == "design" {
			design = s
		}
		for _, f := range s.Files {
			names = append(names, path.Base(f.Path))
		}
	}
	if design.Name == "" {
		t.Fatal("design is not in the payload")
	}

	const marker = "read it with `ui-craft`"
	var line string
	for candidate := range strings.SplitSeq(string(design.Main()), "\n") {
		if strings.Contains(candidate, marker) {
			line = candidate
		}
	}
	if line == "" {
		t.Fatalf("the design skill has no line containing %q", marker)
	}

	index := 0
	for part := range strings.SplitSeq(line, "`") {
		if index%2 == 1 {
			token := strings.TrimSpace(part)
			if token != "ui-craft" && !slices.Contains(names, token+".md") {
				t.Errorf("the design skill names %q, but no skill carries %q", token, token+".md")
			}
		}
		index++
	}
}

// TestDesignFilesExampleUsesBareScreenshots: the manifest example the skill ships
// uses bare screenshot basenames, the form the design domain accepts; a path in
// the example teaches a manifest the reader will refuse.
func TestDesignFilesExampleUsesBareScreenshots(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}

	var design skills.Skill
	for _, s := range all {
		if s.Name == "design" {
			design = s
		}
	}
	if design.Name == "" {
		t.Fatal("design is not in the payload")
	}

	var filesContent string
	for _, f := range design.Files {
		if f.Path == "references/files.md" {
			filesContent = string(f.Content)
		}
	}
	if filesContent == "" {
		t.Fatal("references/files.md is not in the design payload")
	}

	_, after, found := strings.Cut(filesContent, "```json")
	if !found {
		t.Fatal("references/files.md has no json example")
	}
	example, _, found := strings.Cut(after, "```")
	if !found {
		t.Fatal("the json example in references/files.md is not fenced")
	}

	var manifest struct {
		Artboards []struct {
			Screenshot string `json:"screenshot"`
		} `json:"artboards"`
	}
	if err := json.Unmarshal([]byte(example), &manifest); err != nil {
		t.Fatalf("the json example is not valid: %v", err)
	}
	if len(manifest.Artboards) == 0 {
		t.Fatal("the json example has no artboards")
	}
	for _, artboard := range manifest.Artboards {
		if want := filepath.Base(artboard.Screenshot); artboard.Screenshot != want {
			t.Errorf("the json example uses screenshot %q, want bare basename %q", artboard.Screenshot, want)
		}
	}
}
