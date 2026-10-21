package skills_test

import (
	"os"
	"slices"
	"testing"

	"github.com/biggs-100/odachi/skills"
)

// bundled is the five starter skills the binary carries.
var bundled = []string{"acceptance", "data", "ship", "slices", "tests"}

// TestAllCarriesTheFiveStarterSkills: the binary knows them by name, sorted, so
// `pack` writes the same set on every machine and in the same order.
func TestAllCarriesTheFiveStarterSkills(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}

	names := make([]string, 0, len(all))
	for _, s := range all {
		names = append(names, s.Name)
		if len(s.Content) == 0 {
			t.Errorf("%s came back empty", s.Name)
		}
	}
	if !slices.Equal(names, bundled) {
		t.Fatalf("All() = %v, want %v", names, bundled)
	}
	if got := skills.Names(); !slices.Equal(got, bundled) {
		t.Errorf("Names() = %v, want %v", got, bundled)
	}
}

// TestEmbeddedBytesAreTheFilesInTheRepo: the embed is the source, not a copy of
// it. If a skill is edited on disk and the bytes here do not match, the binary
// is shipping a version nobody reviewed.
func TestEmbeddedBytesAreTheFilesInTheRepo(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}

	for _, s := range all {
		onDisk, err := os.ReadFile(s.Name + "/SKILL.md")
		if err != nil {
			t.Fatalf("read %s/SKILL.md: %v", s.Name, err)
		}
		if string(s.Content) != string(onDisk) {
			t.Errorf("%s: the embedded bytes differ from %s/SKILL.md", s.Name, s.Name)
		}
	}
}
