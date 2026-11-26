package evidence_test

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/evidence"
	"github.com/biggs-100/odachi/internal/app/verify"
	domaingrant "github.com/biggs-100/odachi/internal/domain/grant"
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

// receiptGrant is a deterministic Go-issued identity, spelled the way the
// canonical receipt writes it.
func receiptGrant() string {
	var id domaingrant.ID
	for i := range id {
		id[i] = byte(i + 1)
	}
	return id.String()
}

// snapshotDigest is the SHA-256 content identity a frozen candidate answered.
const snapshotDigest = "a1a2a3a4a5a6a7a8a9a0b1b2b3b4b5b6b7b8b9b0b1b2b3b4b5b6b7b8b9b0c1c2"

// schemaTwoVerification is one schema-2 verification file: the schema-1 answer
// with the tool's runtime receipt stamped in. mutate can forge exactly one fact.
func schemaTwoVerification(t *testing.T, mutate func(map[string]any)) string {
	t.Helper()
	receipt := map[string]any{
		"grant":         receiptGrant(),
		"session":       "ses_export_1",
		"runtime":       "opencode",
		"model":         "opencode/model",
		"session_model": "opencode/observed-model",
		"role":          verify.VerifierRole,
		"tree":          testTree,
		"base":          "b4c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9",
		"output_tree":   testTree,
		"snapshot":      snapshotDigest,
		"started_at":    "2026-11-16T12:00:00Z",
		"finished_at":   "2026-11-16T12:00:05Z",
		"outcome":       string(verify.OutcomeCompleted),
	}
	if mutate != nil {
		mutate(receipt)
	}
	encoded, err := json.Marshal(receipt)
	if err != nil {
		t.Fatalf("marshal the runtime receipt: %v", err)
	}
	return `{"schema":2,"tree":"` + testTree + `","by":"opencode/observed-model","at":"2026-11-16T12:00:00Z","verdict":"verified","notes":[],"runtime":` + string(encoded) + `}` + "\n"
}

// TestExportsAValidatedRuntimeReceipt: a managed verification exports as schema 2
// with the validated receipt as its own canonical object and hash.
func TestExportsAValidatedRuntimeReceipt(t *testing.T) {
	dir, root, out, files := fixtures()
	record := schemaTwoVerification(t, nil)
	files.files[filepath.Join(root, "verifications", testTree+".json")] = []byte(record)

	got := evidence.Run(evidence.Options{Dir: dir, Store: root, Tree: testTree, Out: out}, files, fakeGit{common: root})
	if got.Code != evidence.CodeOK {
		t.Fatalf("export: exit %d: %s", got.Code, got.Message)
	}

	var bundle struct {
		Schema       int               `json:"schema"`
		Verification *string           `json:"verification"`
		Runtime      *json.RawMessage  `json:"runtime"`
		SHA256       map[string]string `json:"sha256"`
	}
	if err := json.Unmarshal(files.written[out], &bundle); err != nil {
		t.Fatalf("the bundle is not JSON: %v", err)
	}
	if bundle.Schema != evidence.SchemaRuntime {
		t.Errorf("schema = %d, want %d for a bundle with a receipt", bundle.Schema, evidence.SchemaRuntime)
	}
	if bundle.Verification == nil || *bundle.Verification != record {
		t.Errorf("the verification bytes did not survive the export")
	}
	if bundle.Runtime == nil {
		t.Fatal("the bundle carries no runtime receipt")
	}
	receipt, err := verify.ParseRuntimeReceipt(*bundle.Runtime)
	if err != nil {
		t.Fatalf("the exported receipt is not a sealed receipt: %v", err)
	}
	if receipt.Grant.String() != receiptGrant() || receipt.Session != "ses_export_1" || receipt.Snapshot != snapshotDigest {
		t.Errorf("the exported receipt = %+v, want the tool facts", receipt)
	}
	canonical, err := verify.RenderRuntimeReceipt(receipt)
	if err != nil {
		t.Fatalf("RenderRuntimeReceipt: %v", err)
	}
	if compact(t, *bundle.Runtime) != compact(t, canonical) {
		t.Errorf("the exported receipt is not the canonical receipt.\ngot:\n%s\nwant:\n%s", *bundle.Runtime, canonical)
	}
	// Pinned, not recomputed: the hash is over the canonical receipt, and a
	// change to its key set or order must show up as a different bundle.
	if want := "2441b2f527b7bb2c9cd8959939fcfb1a73e809185f047d993daf35031f3c8bab"; bundle.SHA256["runtime"] != want {
		t.Errorf("sha256[runtime] = %q, want the canonical receipt hash %s", bundle.SHA256["runtime"], want)
	}
}

// compact is the JSON with the bundle's indentation removed, so two spellings of
// the same object compare as one.
func compact(t *testing.T, data []byte) string {
	t.Helper()
	var out bytes.Buffer
	if err := json.Compact(&out, data); err != nil {
		t.Fatalf("compact JSON: %v", err)
	}
	return out.String()
}

// TestRefusesToExportAnInvalidOrUnsealedReceipt: an inconclusive receipt is not
// evidence, so the export refuses instead of writing a bundle that lies.
func TestRefusesToExportAnInvalidOrUnsealedReceipt(t *testing.T) {
	cases := map[string]string{
		"null receipt": `{"schema":2,"tree":"` + testTree + `","by":"test","at":"2026-11-16T12:00:00Z","verdict":"verified","notes":[],"runtime":null}`,
		"empty receipt": schemaTwoVerification(t, func(receipt map[string]any) {
			for key := range receipt {
				delete(receipt, key)
			}
		}),
		"forged grant": schemaTwoVerification(t, func(receipt map[string]any) {
			receipt["grant"] = "forged-grant"
		}),
		"forged snapshot": schemaTwoVerification(t, func(receipt map[string]any) {
			receipt["snapshot"] = "forged-snapshot"
		}),
		"moved output": schemaTwoVerification(t, func(receipt map[string]any) {
			receipt["output_tree"] = "fedcba9876543210fedcba9876543210fedcba98"
		}),
		"backwards clock": schemaTwoVerification(t, func(receipt map[string]any) {
			receipt["started_at"] = "2026-11-16T12:00:09Z"
			receipt["finished_at"] = "2026-11-16T12:00:05Z"
		}),
		"unknown outcome": schemaTwoVerification(t, func(receipt map[string]any) {
			receipt["outcome"] = "forged-outcome"
		}),
		"schema two without a receipt": `{"schema":2,"tree":"` + testTree + `","by":"test","at":"2026-11-16T12:00:00Z","verdict":"verified","notes":[]}`,
		"a newer schema":               `{"schema":3,"tree":"` + testTree + `","by":"test","at":"2026-11-16T12:00:00Z","verdict":"verified","notes":[]}`,
		"prose":                        "the verifier says it is fine",
	}
	for name, record := range cases {
		t.Run(name, func(t *testing.T) {
			dir, root, out, files := fixtures()
			files.files[filepath.Join(root, "verifications", testTree+".json")] = []byte(record)

			got := evidence.Run(evidence.Options{Dir: dir, Store: root, Tree: testTree, Out: out}, files, fakeGit{common: root})
			if got.Code == evidence.CodeOK {
				t.Fatalf("an unsealed receipt exported: %s", got.Message)
			}
			if len(files.written) != 0 {
				t.Errorf("a refusal still wrote %v", files.written)
			}
		})
	}
}

// TestRefusesToExportARecordWithAForgedAuthor: a managed verification is authored by
// the model the tool observed. A record whose "by" differs from the receipt's
// session_model is refused by the same validation the store reads, so the export
// writes nothing.
func TestRefusesToExportARecordWithAForgedAuthor(t *testing.T) {
	dir, root, out, files := fixtures()
	record := schemaTwoVerification(t, nil)
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(record), &fields); err != nil {
		t.Fatalf("the verification is not JSON: %v", err)
	}
	fields["by"] = json.RawMessage(`"anthropic/claude-forged"`)
	forged, err := json.Marshal(fields)
	if err != nil {
		t.Fatalf("marshal the forged verification: %v", err)
	}
	files.files[filepath.Join(root, "verifications", testTree+".json")] = forged

	got := evidence.Run(evidence.Options{Dir: dir, Store: root, Tree: testTree, Out: out}, files, fakeGit{common: root})
	if got.Code == evidence.CodeOK {
		t.Fatalf("a record with a forged author exported: %s", got.Message)
	}
	if !strings.Contains(strings.ToLower(got.Message), "author") {
		t.Errorf("message %q does not name the forged author", got.Message)
	}
	if len(files.written) != 0 {
		t.Errorf("a refusal still wrote %v", files.written)
	}
}

// TestRefusesToExportAVerificationOfAnotherTree: the file is named by the tree it
// proves; a record that speaks for other bytes cannot ride under this name.
func TestRefusesToExportAVerificationOfAnotherTree(t *testing.T) {
	dir, root, out, files := fixtures()
	other := "fedcba9876543210fedcba9876543210fedcba98"
	files.files[filepath.Join(root, "verifications", testTree+".json")] = []byte(
		`{"schema":1,"tree":"` + other + `","by":"test","at":"2026-01-01T00:00:00Z","verdict":"verified","notes":[]}`)

	got := evidence.Run(evidence.Options{Dir: dir, Store: root, Tree: testTree, Out: out}, files, fakeGit{common: root})
	if got.Code == evidence.CodeOK {
		t.Fatal("a verification of another tree exported")
	}
	if len(files.written) != 0 {
		t.Errorf("a refusal still wrote %v", files.written)
	}
}

// TestLegacySchemaOneExportsWithoutARuntimeReceipt: the unmanaged record keeps the
// bundle it always exported: schema 1, no runtime key, no runtime hash.
func TestLegacySchemaOneExportsWithoutARuntimeReceipt(t *testing.T) {
	dir, root, out, files := fixtures()
	files.files[filepath.Join(root, "verifications", testTree+".json")] = []byte(verifyText)

	got := evidence.Run(evidence.Options{Dir: dir, Store: root, Tree: testTree, Out: out}, files, fakeGit{common: root})
	if got.Code != evidence.CodeOK {
		t.Fatalf("export: exit %d: %s", got.Code, got.Message)
	}
	var bundle map[string]json.RawMessage
	if err := json.Unmarshal(files.written[out], &bundle); err != nil {
		t.Fatalf("the bundle is not JSON: %v", err)
	}
	if string(bundle["schema"]) != "1" {
		t.Errorf("schema = %s, want 1", bundle["schema"])
	}
	if _, present := bundle["runtime"]; present {
		t.Error("the legacy bundle carries a runtime key")
	}
	var hashes map[string]string
	if err := json.Unmarshal(bundle["sha256"], &hashes); err != nil {
		t.Fatalf("sha256 is not an object: %v", err)
	}
	if _, present := hashes["runtime"]; present {
		t.Error("the legacy bundle hashes a runtime receipt")
	}
}
