package evidence_test

import (
	"encoding/json"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/evidence"
)

const (
	testTree = "0123456789abcdef0123456789abcdef01234567"

	verdictText = `{"findings":[],"passes":[],"reviewer":"test","tree":"` + testTree + `"}` + "\n"
	verifyText  = `{"at":"2026-01-01T00:00:00Z","by":"test","notes":[],"schema":1,"tree":"` + testTree + `","verdict":"verified"}` + "\n"
	proofText   = "# tree " + testTree + "\n# command echo ok\n# output\nok\n"
)

// fakeFiles is an in-memory Files. It records every directory the app asks to
// create so the test can prove the bundle writer stays inside the output's parent.
type fakeFiles struct {
	files   map[string][]byte
	dirs    map[string][]string
	written map[string][]byte
	made    []string
}

func newFakeFiles() *fakeFiles {
	return &fakeFiles{
		files:   map[string][]byte{},
		dirs:    map[string][]string{},
		written: map[string][]byte{},
	}
}

func (f *fakeFiles) Read(path string) ([]byte, error) {
	data, ok := f.files[filepath.Clean(path)]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return append([]byte(nil), data...), nil
}

func (f *fakeFiles) Exists(path string) (bool, error) {
	_, ok := f.files[filepath.Clean(path)]
	return ok, nil
}

func (f *fakeFiles) List(dir string) ([]string, error) {
	return f.dirs[filepath.Clean(dir)], nil
}

func (f *fakeFiles) MkdirAll(dir string) error {
	f.made = append(f.made, filepath.Clean(dir))
	return nil
}

func (f *fakeFiles) Write(path string, data []byte) error {
	key := filepath.Clean(path)
	f.files[key] = append([]byte(nil), data...)
	f.written[key] = append([]byte(nil), data...)
	return nil
}

type fakeGit struct {
	common string
	err    error
}

func (g fakeGit) CommonDir(string) (string, error) { return g.common, g.err }

func fixtures() (dir, root, out string, files *fakeFiles) {
	dir = filepath.Join(string(filepath.Separator), "repo")
	root = filepath.Join(dir, ".git", "odachi")
	out = filepath.Join(dir, "exports", "close.json")
	files = newFakeFiles()
	return dir, root, out, files
}

// TestExportsAllThreePartsAsCanonicalBytes is the whole contract: the bundle is
// keyed by the tree, carries the raw parts and their hashes, and has no clock.
// The bytes below are pinned, not recomputed with the implementation's encoder.
func TestExportsAllThreePartsAsCanonicalBytes(t *testing.T) {
	dir, root, out, files := fixtures()
	files.files[filepath.Join(root, "reviews", testTree+".json")] = []byte(verdictText)
	files.files[filepath.Join(root, "verifications", testTree+".json")] = []byte(verifyText)
	files.files[filepath.Join(dir, "proof", "0001-close.txt")] = []byte(proofText)
	files.dirs[filepath.Join(dir, "proof")] = []string{"0001-close.txt", "notes.md"}

	got := evidence.Run(evidence.Options{Dir: dir, Store: root, Tree: testTree, Out: out}, files, fakeGit{common: root})
	if got.Code != evidence.CodeOK {
		t.Fatalf("export: exit %d: %s", got.Code, got.Message)
	}

	const want = `{
  "schema": 1,
  "tree": "0123456789abcdef0123456789abcdef01234567",
  "verdict": "{\"findings\":[],\"passes\":[],\"reviewer\":\"test\",\"tree\":\"0123456789abcdef0123456789abcdef01234567\"}\n",
  "verification": "{\"at\":\"2026-01-01T00:00:00Z\",\"by\":\"test\",\"notes\":[],\"schema\":1,\"tree\":\"0123456789abcdef0123456789abcdef01234567\",\"verdict\":\"verified\"}\n",
  "proof": "# tree 0123456789abcdef0123456789abcdef01234567\n# command echo ok\n# output\nok\n",
  "sha256": {
    "verdict": "eb98c8f8c638a2a94bde159065c8bbd68ec0c122b014d7c5488838d00b83e71a",
    "verification": "cc2eec198efd4ca34a34fd8e9cd415269a56cbd8a82195b55b3496b15645a869",
    "proof": "3fa3ab272919d01a8d1c7451ebdeecbedcbd54c57474f1235dd0777b243c514b"
  }
}
`
	if string(files.written[out]) != want {
		t.Errorf("bundle bytes differ.\ngot:\n%s\nwant:\n%s", files.written[out], want)
	}
	if len(files.made) != 1 || files.made[0] != filepath.Dir(out) {
		t.Errorf("directories made = %v, want only the output's parent %s", files.made, filepath.Dir(out))
	}
}

// TestExportsOnlyAProof keeps the missing parts as null instead of inventing
// empty strings, and hashes only what exists.
func TestExportsOnlyAProof(t *testing.T) {
	dir, _, out, files := fixtures()
	files.files[filepath.Join(dir, "proof", "0001-close.txt")] = []byte(proofText)
	files.dirs[filepath.Join(dir, "proof")] = []string{"0001-close.txt"}

	got := evidence.Run(evidence.Options{Dir: dir, Tree: testTree, Out: out}, files, fakeGit{common: filepath.Join(dir, ".git", "odachi")})
	if got.Code != evidence.CodeOK {
		t.Fatalf("export: exit %d: %s", got.Code, got.Message)
	}

	var bundle struct {
		Verdict      *string        `json:"verdict"`
		Verification *string        `json:"verification"`
		Proof        *string        `json:"proof"`
		SHA256       map[string]any `json:"sha256"`
	}
	if err := json.Unmarshal(files.written[out], &bundle); err != nil {
		t.Fatalf("the bundle is not JSON: %v", err)
	}
	if bundle.Verdict != nil || bundle.Verification != nil {
		t.Errorf("missing parts = %v / %v, want null", bundle.Verdict, bundle.Verification)
	}
	if bundle.Proof == nil || *bundle.Proof != proofText {
		t.Errorf("proof = %v, want the raw proof text", bundle.Proof)
	}
	if len(bundle.SHA256) != 1 || bundle.SHA256["proof"] != "3fa3ab272919d01a8d1c7451ebdeecbedcbd54c57474f1235dd0777b243c514b" {
		t.Errorf("sha256 = %v, want only the proof's hash", bundle.SHA256)
	}
}

// TestRefusesATreeWithNoEvidence says so instead of writing an empty bundle.
func TestRefusesATreeWithNoEvidence(t *testing.T) {
	dir, _, out, files := fixtures()
	got := evidence.Run(evidence.Options{Dir: dir, Tree: testTree, Out: out}, files, fakeGit{common: filepath.Join(dir, ".git", "odachi")})
	if got.Code == evidence.CodeOK {
		t.Fatal("export with no evidence = 0, want a refusal")
	}
	if !strings.Contains(got.Message, testTree) {
		t.Errorf("message %q does not name the tree", got.Message)
	}
	if len(files.written) != 0 {
		t.Errorf("a refusal still wrote %v", files.written)
	}
}

// TestRefusesABadTreeHash pins the shape the tree argument must have.
func TestRefusesABadTreeHash(t *testing.T) {
	dir, _, out, files := fixtures()
	got := evidence.Run(evidence.Options{Dir: dir, Tree: "not-a-hash", Out: out}, files, fakeGit{common: filepath.Join(dir, ".git", "odachi")})
	if got.Code == evidence.CodeOK {
		t.Fatal("export with a bad tree = 0, want a refusal")
	}
	if !strings.Contains(got.Message, "40 or 64") {
		t.Errorf("message %q does not name the expected shape", got.Message)
	}
	if len(files.written) != 0 {
		t.Errorf("a refusal still wrote %v", files.written)
	}
}

// TestRefusesToOverwriteAnOutput names the file it protected.
func TestRefusesToOverwriteAnOutput(t *testing.T) {
	dir, _, out, files := fixtures()
	files.files[filepath.Join(dir, "proof", "0001-close.txt")] = []byte(proofText)
	files.dirs[filepath.Join(dir, "proof")] = []string{"0001-close.txt"}
	files.files[out] = []byte("keep me")

	got := evidence.Run(evidence.Options{Dir: dir, Tree: testTree, Out: out}, files, fakeGit{common: filepath.Join(dir, ".git", "odachi")})
	if got.Code == evidence.CodeOK {
		t.Fatal("export over an existing file = 0, want a refusal")
	}
	if !strings.Contains(got.Message, out) {
		t.Errorf("message %q does not name the output %s", got.Message, out)
	}
	if string(files.files[out]) != "keep me" {
		t.Errorf("the existing output was overwritten: %q", files.files[out])
	}
}
