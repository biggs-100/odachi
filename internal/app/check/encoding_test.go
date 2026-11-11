package check_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/check"
	"github.com/biggs-100/odachi/internal/domain/gate"
)

// TestAnArtifactThatIsNotUTF8Blocks is 0274: a record saved in another encoding is a
// record nobody can trust, and until this the gate judged it as if it were text.
func TestAnArtifactThatIsNotUTF8Blocks(t *testing.T) {
	cases := []struct {
		name  string
		path  string
		bytes []byte
	}{
		{name: "the brief", path: "brief.md", bytes: []byte("# turnos\n\xff\n")},
		{name: "the fence", path: "odd/fence.toml", bytes: []byte("[[rule]]\nnot = \"\x93pagos\"\n")},
		{name: "a task document", path: "odd/tasks/0007-turnos.md", bytes: []byte("# turnos\n\xfe\n")},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			files := newFakeFiles(readyFiles())
			files.files[filepath.Join(dir, filepath.FromSlash(c.path))] = c.bytes

			got := run(t, files, &fakeGit{tree: currentTree}, check.Options{})

			if got.Code == int(gate.Ready) {
				t.Fatalf("%s with bytes that are not UTF-8 was judged ready", c.path)
			}
			if !strings.Contains(got.Message, c.path) {
				t.Errorf("the refusal does not name %s: %q", c.path, got.Message)
			}
			if !strings.Contains(got.Message, "UTF-8") {
				t.Errorf("the refusal does not say what to save it as: %q", got.Message)
			}
		})
	}
}

// TestAFileSavedWithABOMBlocks: the BOM is valid UTF-8, and it still moves the first
// line out from under the parser, so the gate names it instead of reading around it.
func TestAFileSavedWithABOMBlocks(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, "brief.md")] = []byte("\xEF\xBB\xBF# turnos\n\n## What must exist\n- reservar turnos\n")

	got := run(t, files, &fakeGit{tree: currentTree}, check.Options{})

	if got.Code == int(gate.Ready) {
		t.Fatal("a brief saved with a BOM was judged ready")
	}
	if !strings.Contains(got.Message, "BOM") {
		t.Errorf("the refusal does not name the BOM: %q", got.Message)
	}
	if !strings.Contains(got.Message, "brief.md") {
		t.Errorf("the refusal does not name the file: %q", got.Message)
	}
}

// TestABriefThatCarriesADoubleEncodingBlocks is 0274's second half: the bytes decode,
// so no decoder can see the damage, and the gate names the sequence instead of judging
// a brief that no longer says what was written.
func TestABriefThatCarriesADoubleEncodingBlocks(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, "brief.md")] = []byte("# turnos\n\n## What must exist\n- reservar turnos \u00e2\u20ac\u00a6 sin pausa\n\n## Done when\n- reservo un turno\n")

	got := run(t, files, &fakeGit{tree: currentTree}, check.Options{})

	if got.Code == int(gate.Ready) {
		t.Fatal("a brief with a double-encoded sentence was judged ready")
	}
	if !strings.Contains(got.Message, "brief.md") {
		t.Errorf("the refusal does not name the file: %q", got.Message)
	}
	if !strings.Contains(got.Message, "encoding") {
		t.Errorf("the refusal does not say what happened: %q", got.Message)
	}
}

// TestATaskDocumentMayNameTheSequences: the task documents are prose, and a document
// that explains this defect has to be able to write the marker without blocking the
// gate that watches for it.
func TestATaskDocumentMayNameTheSequences(t *testing.T) {
	doc := "# turnos\n\nBase: abc1234\n\n## Requirements (verbatim, from the user)\nR1 \u00b7 \"reservar turnos \u00e2\u20ac\u00a6 sin pausa\"\n\n## Steps\n- [ ] T1 \u00b7 reservar turnos   R1 \u00b7 route:inline\n"
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, "odd/tasks/0007-turnos.md")] = []byte(doc)

	got := run(t, files, &fakeGit{tree: currentTree}, check.Options{})

	if strings.Contains(got.Message, "encoding") {
		t.Errorf("a task document that names the sequence was refused for it: %q", got.Message)
	}
}
