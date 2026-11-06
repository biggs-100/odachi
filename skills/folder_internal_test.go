package skills

import (
	"slices"
	"testing"
	"testing/fstest"
)

// TestFolderCarriesTheAssets: a skill folder also carries what it copies -a
// template, a schema- whatever its extension, sorted with the rest.
func TestFolderCarriesTheAssets(t *testing.T) {
	fsys := fstest.MapFS{
		"router/SKILL.md":           {Data: []byte("main")},
		"router/assets/doc.md":      {Data: []byte("template")},
		"router/assets/schema.json": {Data: []byte("{}")},
		"router/assets/.hidden":     {Data: []byte("dot")},
	}

	files, err := folder(fsys, "router")
	if err != nil {
		t.Fatalf("folder(): %v", err)
	}

	want := []File{
		{Path: "SKILL.md", Content: []byte("main")},
		{Path: "assets/.hidden", Content: []byte("dot")},
		{Path: "assets/doc.md", Content: []byte("template")},
		{Path: "assets/schema.json", Content: []byte("{}")},
	}
	if !slices.EqualFunc(files, want, func(a, b File) bool {
		return a.Path == b.Path && string(a.Content) == string(b.Content)
	}) {
		t.Errorf("folder() = %v, want %v", files, want)
	}
}

// TestFolderCarriesTheReferences: one skill folder is its SKILL.md plus the
// markdown under references/, sorted by path. A file the check would refuse
// never ships.
func TestFolderCarriesTheReferences(t *testing.T) {
	fsys := fstest.MapFS{
		"router/SKILL.md":             {Data: []byte("main")},
		"router/references/b.md":      {Data: []byte("b")},
		"router/references/a.md":      {Data: []byte("a")},
		"router/references/notes.txt": {Data: []byte("not a reference")},
		"other/SKILL.md":              {Data: []byte("other")},
	}

	files, err := folder(fsys, "router")
	if err != nil {
		t.Fatalf("folder(): %v", err)
	}

	want := []File{
		{Path: "SKILL.md", Content: []byte("main")},
		{Path: "references/a.md", Content: []byte("a")},
		{Path: "references/b.md", Content: []byte("b")},
	}
	if !slices.EqualFunc(files, want, func(a, b File) bool {
		return a.Path == b.Path && string(a.Content) == string(b.Content)
	}) {
		t.Errorf("folder() = %v, want %v", files, want)
	}
}
