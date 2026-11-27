package verify_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	domaingrant "github.com/biggs-100/odachi/internal/domain/grant"
	"github.com/biggs-100/odachi/internal/domain/verify"
)

// receiptBase is the candidate base a managed verifier binds beside its tree.
const receiptBase = "b4c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9"

// receiptSnapshot is the stable content identity the frozen candidate answers.
const receiptSnapshot = "a1a2a3a4a5a6a7a8a9a0b1b2b3b4b5b6b7b8b9b0b1b2b3b4b5b6b7b8b9b0c1c2"

// receiptGrant is a deterministic Go-issued identity for a fixture receipt.
func receiptGrant(t *testing.T) string {
	t.Helper()
	var id domaingrant.ID
	for i := range id {
		id[i] = byte(i + 1)
	}
	return id.String()
}

// receiptObject is one complete tool-sealed receipt for the given tree.
func receiptObject(t *testing.T, tree, by string) map[string]any {
	t.Helper()
	model := "model"
	if provider, _, found := strings.Cut(by, "/"); found && provider != "" {
		model = provider + "/planned"
	}
	return map[string]any{
		"grant":         receiptGrant(t),
		"session":       "ses_domain_1",
		"runtime":       "opencode",
		"model":         model,
		"session_model": by,
		"role":          "verifier",
		"tree":          tree,
		"base":          receiptBase,
		"output_tree":   tree,
		"snapshot":      receiptSnapshot,
		"started_at":    "2026-11-16T12:00:00Z",
		"finished_at":   "2026-11-16T12:00:05Z",
		"outcome":       "completed",
	}
}

// schemaTwoRecord builds one schema-2 verification for the given tree.
func schemaTwoRecord(t *testing.T, tree, by string, mutate func(map[string]any)) []byte {
	t.Helper()
	record := map[string]any{
		"schema":  2,
		"tree":    tree,
		"by":      by,
		"at":      "2026-11-16T12:00:00Z",
		"verdict": "verified",
		"notes":   []string{},
		"runtime": receiptObject(t, tree, by),
	}
	if mutate != nil {
		mutate(record)
	}
	data, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("marshal the schema-2 record: %v", err)
	}
	return data
}

// schemaTwoRuntime mutates the runtime object inside a schema-2 record.
func schemaTwoRuntime(t *testing.T, data []byte, mutate func(map[string]any)) []byte {
	t.Helper()
	var record map[string]any
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatalf("read the schema-2 record: %v", err)
	}
	runtime, ok := record["runtime"].(map[string]any)
	if !ok {
		t.Fatalf("the schema-2 record carries no runtime object: %s", data)
	}
	mutate(runtime)
	record["runtime"] = runtime
	mutated, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("marshal the mutated schema-2 record: %v", err)
	}
	return mutated
}

// TestASchemaTwoRecordParsesAndRendersItsReceipt: the domain owns the managed
// record too. A schema-2 file with a complete tool receipt parses, validates and
// keeps the receipt through one canonical render.
func TestASchemaTwoRecordParsesAndRendersItsReceipt(t *testing.T) {
	data := schemaTwoRecord(t, tree, "opencode/observed", nil)

	record, err := verify.Parse(data)
	if err != nil {
		t.Fatalf("Parse() of a complete schema-2 record = %v", err)
	}
	if err := record.Validate(); err != nil {
		t.Fatalf("Validate() of a complete schema-2 record = %v", err)
	}
	rendered := verify.Render(record)
	if !strings.Contains(string(rendered), `"runtime"`) || !strings.Contains(string(rendered), `"schema": 2`) {
		t.Errorf("the rendered schema-2 record lost its receipt or version:\n%s", rendered)
	}
	again, err := verify.Parse(rendered)
	if err != nil {
		t.Fatalf("Parse() of the rendered schema-2 record = %v", err)
	}
	if second := verify.Render(again); string(second) != string(rendered) {
		t.Errorf("two renders of one schema-2 record differ.\ngot:\n%s\nwant:\n%s", second, rendered)
	}
}

// TestASchemaTwoRecordWithoutAReceiptIsRefused: schema 2 exists to carry the
// receipt, so a file that names it and drops the receipt is unsealed evidence.
func TestASchemaTwoRecordWithoutAReceiptIsRefused(t *testing.T) {
	data := schemaTwoRecord(t, tree, "opencode/observed", func(record map[string]any) {
		delete(record, "runtime")
	})

	record, err := verify.Parse(data)
	if err != nil {
		t.Fatalf("Parse() of a schema-2 record without a receipt = %v", err)
	}
	if err := record.Validate(); err == nil {
		t.Fatal("Validate() accepted a schema-2 record without a receipt")
	} else if !strings.Contains(strings.ToLower(err.Error()), "runtime receipt") {
		t.Errorf("the refusal does not name the missing receipt: %v", err)
	}
}

// TestASchemaOneRecordCarryingAReceiptIsRefused: schema 1 is the unmanaged answer
// and claims no runtime. A model-written receipt beside it cannot become one.
func TestASchemaOneRecordCarryingAReceiptIsRefused(t *testing.T) {
	data := schemaTwoRecord(t, tree, "otro/modelo", nil)
	data = []byte(strings.Replace(string(data), `"schema":2`, `"schema":1`, 1))

	record, err := verify.Parse(data)
	if err != nil {
		t.Fatalf("Parse() of a schema-1 record carrying a receipt = %v", err)
	}
	if err := record.Validate(); err == nil {
		t.Fatal("Validate() accepted a schema-1 record carrying a runtime receipt")
	}
}

// TestASchemaTwoRecordForAnotherTreeIsRefused: the answer and its receipt must
// speak for the same bytes.
func TestASchemaTwoRecordForAnotherTreeIsRefused(t *testing.T) {
	other := "fedcba9876543210fedcba9876543210fedcba98"
	data := schemaTwoRecord(t, tree, "opencode/observed", func(record map[string]any) {
		record["tree"] = other
	})

	record, err := verify.Parse(data)
	if err != nil {
		t.Fatalf("Parse() = %v", err)
	}
	if err := record.Validate(); err == nil {
		t.Fatal("Validate() accepted a record whose receipt speaks for another tree")
	}
}

// TestASchemaTwoRecordWithAForgedAuthorIsRefused: a managed record is authored by
// the model the tool observed, never by the model the answer named.
func TestASchemaTwoRecordWithAForgedAuthorIsRefused(t *testing.T) {
	data := schemaTwoRecord(t, tree, "otro/modelo", nil)
	data, err := mutateRecordField(data, "by", "anthropic/claude-forged")
	if err != nil {
		t.Fatalf("forge the record author: %v", err)
	}

	record, err := verify.Parse(data)
	if err != nil {
		t.Fatalf("Parse() = %v", err)
	}
	if err := record.Validate(); err == nil {
		t.Fatal("Validate() accepted a record authored by a model its receipt did not observe")
	} else if !strings.Contains(err.Error(), "author") {
		t.Errorf("the refusal does not name the forged author: %v", err)
	}
}

// TestASchemaTwoRecordWithAnInvalidReceiptIsRefused: an unsafe identity or a
// snapshot that is not a SHA-256 never becomes evidence.
func TestASchemaTwoRecordWithAnInvalidReceiptIsRefused(t *testing.T) {
	cases := map[string]func(map[string]any){
		"unsafe session": func(runtime map[string]any) { runtime["session"] = "ses domain" },
		"short snapshot": func(runtime map[string]any) { runtime["snapshot"] = "abc" },
		"moved output": func(runtime map[string]any) {
			runtime["output_tree"] = "fedcba9876543210fedcba9876543210fedcba98"
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			data := schemaTwoRuntime(t, schemaTwoRecord(t, tree, "opencode/observed", nil), mutate)
			record, err := verify.Parse(data)
			if err == nil {
				err = record.Validate()
			}
			if err == nil {
				t.Fatalf("a schema-2 record with an invalid receipt (%s) was accepted", name)
			}
		})
	}
}

// TestTheSchemaOneBytesAreUnchanged: the managed work must not rewrite the legacy
// file. The schema-1 render is the same indented JSON it always was.
func TestTheSchemaOneBytesAreUnchanged(t *testing.T) {
	record := verify.Record{
		Tree:    tree,
		By:      "otro/modelo",
		Verdict: verify.Verified,
		Notes:   []string{},
	}.Seal(time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC))

	want := fmt.Sprintf(`{
  "schema": 1,
  "tree": %q,
  "by": "otro/modelo",
  "at": "2026-11-16T12:00:00Z",
  "verdict": "verified",
  "notes": []
}
`, tree)
	if got := string(verify.Render(record)); got != want {
		t.Errorf("the schema-1 bytes changed.\ngot:\n%s\nwant:\n%s", got, want)
	}
}

// TestASchemaThreeRecordIsRefused: a file from a newer binary is not guessed.
func TestASchemaThreeRecordIsRefused(t *testing.T) {
	data := []byte(`{"schema":3,"tree":"` + tree + `","by":"otro/modelo","at":"2026-11-16T12:00:00Z","verdict":"verified","notes":[]}`)

	if _, err := verify.Parse(data); err == nil {
		t.Fatal("Parse() accepted a schema-3 record with a binary that writes 1 and 2")
	} else if !strings.Contains(err.Error(), "upgrade odachi") {
		t.Errorf("the refusal does not send the human to upgrade: %v", err)
	}
}

// mutateRecordField replaces one top-level field of a JSON record.
func mutateRecordField(data []byte, key string, value string) ([]byte, error) {
	var record map[string]any
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, err
	}
	record[key] = value
	return json.Marshal(record)
}
