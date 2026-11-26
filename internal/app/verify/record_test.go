package verify_test

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/app/verify"
	verifydomain "github.com/biggs-100/odachi/internal/domain/verify"
)

// managedAnswer is the verifier's schema-1 answer as the tool parses it.
func managedAnswer() verifydomain.Record {
	return verifydomain.Record{
		Tree:    tree,
		By:      "opencode/model",
		Verdict: verifydomain.Verified,
		Notes:   []string{"go test ./internal/app/verify"},
	}
}

// legacyRecordText is one schema-1 verification as the unmanaged verb stores it.
const legacyRecordText = `{"at":"2026-11-16T12:00:00Z","by":"otro/modelo","notes":[],"schema":1,"tree":"` + tree + `","verdict":"verified"}` + "\n"

// TestSealManagedRecordStampsSchemaTwoAndTheToolReceipt: a managed answer becomes
// a schema-2 record that carries the validated receipt beside the answer, sealed
// with the tool's hour.
func TestSealManagedRecordStampsSchemaTwoAndTheToolReceipt(t *testing.T) {
	receipt := validReceipt()
	now := time.Date(2026, 11, 16, 12, 30, 0, 0, time.UTC)

	record, err := verify.SealManagedRecord(managedAnswer(), receipt, now)
	if err != nil {
		t.Fatalf("SealManagedRecord: %v", err)
	}
	if record.Schema != verify.RecordSchema || verify.RecordSchema != 2 {
		t.Fatalf("schema = %d (const %d), want 2", record.Schema, verify.RecordSchema)
	}
	if record.Runtime == nil {
		t.Fatal("the schema-2 record carries no runtime receipt")
	}
	if *record.Runtime != receipt {
		t.Errorf("runtime = %+v, want the tool receipt %+v", *record.Runtime, receipt)
	}
	if record.At != now.Format(time.RFC3339) {
		t.Errorf("at = %q, want the tool hour %q", record.At, now.Format(time.RFC3339))
	}
	if record.Tree != tree || record.By != receipt.SessionModel || record.Verdict != verifydomain.Verified {
		t.Errorf("the answer did not survive the stamp: %+v", record)
	}

	rendered, err := verify.RenderVerificationRecord(record)
	if err != nil {
		t.Fatalf("RenderVerificationRecord: %v", err)
	}
	if !strings.Contains(string(rendered), `"schema": 2`) || !strings.Contains(string(rendered), `"runtime"`) {
		t.Errorf("the stored record is not schema 2 with a receipt:\n%s", rendered)
	}
	parsed, err := verify.ParseVerificationRecord(rendered)
	if err != nil {
		t.Fatalf("ParseVerificationRecord: %v", err)
	}
	if parsed.Runtime == nil || *parsed.Runtime != receipt {
		t.Errorf("the receipt did not round trip: %+v", parsed.Runtime)
	}
}

// TestSealManagedRecordRefusesAnUnsealedReceipt: a receipt the tool cannot
// validate is refused before it is stamped, so no invalid evidence is stored.
func TestSealManagedRecordRefusesAnUnsealedReceipt(t *testing.T) {
	receipt := validReceipt()
	receipt.Snapshot = "forged-snapshot"

	record, err := verify.SealManagedRecord(managedAnswer(), receipt, startedAt)
	if err == nil {
		t.Fatalf("an unsealed receipt was stamped: %+v", record)
	}
	if !strings.Contains(err.Error(), "snapshot") {
		t.Errorf("error %q does not name the unsealed snapshot", err)
	}
}

// TestSealManagedRecordRefusesAnAnswerForAnotherTree: the tool read one tree, and
// an answer that names another covers bytes nobody read.
func TestSealManagedRecordRefusesAnAnswerForAnotherTree(t *testing.T) {
	answer := managedAnswer()
	answer.Tree = "fedcba9876543210fedcba9876543210fedcba98"

	if _, err := verify.SealManagedRecord(answer, validReceipt(), startedAt); err == nil {
		t.Fatal("an answer for another tree was stamped")
	}
}

// TestSealManagedRecordStampsTheToolTreeWhenTheAnswerNamesNone: the answer may
// leave the tree out; the receipt's tree is the tool's fact and fills it.
func TestSealManagedRecordStampsTheToolTreeWhenTheAnswerNamesNone(t *testing.T) {
	answer := managedAnswer()
	answer.Tree = ""

	record, err := verify.SealManagedRecord(answer, validReceipt(), startedAt)
	if err != nil {
		t.Fatalf("SealManagedRecord: %v", err)
	}
	if record.Tree != validReceipt().Tree {
		t.Errorf("tree = %q, want the tool tree %q", record.Tree, validReceipt().Tree)
	}
}

// TestSealManagedRecordDerivesTheAuthorFromTheToolReceipt: the answer's "by" is a
// claim the model wrote about itself. A managed record is authored by the model
// the tool observed, so a forged answer author never reaches the seal.
func TestSealManagedRecordDerivesTheAuthorFromTheToolReceipt(t *testing.T) {
	answer := managedAnswer()
	answer.By = "anthropic/claude-forged"
	receipt := validReceipt()

	record, err := verify.SealManagedRecord(answer, receipt, startedAt)
	if err != nil {
		t.Fatalf("SealManagedRecord: %v", err)
	}
	if record.By != receipt.SessionModel {
		t.Errorf("by = %q, want the tool-observed %q", record.By, receipt.SessionModel)
	}
	if record.By == answer.By {
		t.Errorf("the forged answer author %q was copied into the record", answer.By)
	}

	rendered, err := verify.RenderVerificationRecord(record)
	if err != nil {
		t.Fatalf("RenderVerificationRecord: %v", err)
	}
	parsed, err := verify.ParseVerificationRecord(rendered)
	if err != nil {
		t.Fatalf("ParseVerificationRecord: %v", err)
	}
	if parsed.By != receipt.SessionModel {
		t.Errorf("the stored author = %q, want %q", parsed.By, receipt.SessionModel)
	}
}

// TestSealManagedRecordRefusesAnAnswerThatIsNotAVerification: the seal validates
// the record it built, so an answer with a verdict outside the closed vocabulary
// is refused instead of stored.
func TestSealManagedRecordRefusesAnAnswerThatIsNotAVerification(t *testing.T) {
	answer := managedAnswer()
	answer.Verdict = "maybe"

	record, err := verify.SealManagedRecord(answer, validReceipt(), startedAt)
	if err == nil {
		t.Fatalf("an answer with an unknown verdict was sealed: %+v", record)
	}
	if !strings.Contains(err.Error(), "only answers") {
		t.Errorf("error %q does not name the closed verdict vocabulary", err)
	}
}

// TestParseVerificationRecordKeepsLegacySchemaOne: a record written before the
// receipt existed still reads, renders and validates exactly as it did.
func TestParseVerificationRecordKeepsLegacySchemaOne(t *testing.T) {
	record, err := verify.ParseVerificationRecord([]byte(legacyRecordText))
	if err != nil {
		t.Fatalf("ParseVerificationRecord: %v", err)
	}
	if record.Schema != 1 || record.Runtime != nil {
		t.Fatalf("legacy record = schema %d runtime %v, want 1 and none", record.Schema, record.Runtime)
	}
	rendered, err := verify.RenderVerificationRecord(record)
	if err != nil {
		t.Fatalf("RenderVerificationRecord: %v", err)
	}
	legacy, err := verifydomain.Parse([]byte(legacyRecordText))
	if err != nil {
		t.Fatalf("the legacy record is not what the domain reads: %v", err)
	}
	if string(rendered) != string(verifydomain.Render(legacy)) {
		t.Errorf("the legacy record was rewritten.\ngot:\n%s\nwant:\n%s", rendered, verifydomain.Render(legacy))
	}
}

// TestParseVerificationRecordRefusesSchemaTwoWithoutAReceipt: schema 2 exists to
// carry the receipt, so a schema-2 file without one is unsealed evidence.
func TestParseVerificationRecordRefusesSchemaTwoWithoutAReceipt(t *testing.T) {
	cases := map[string]string{
		"no runtime key": `{"schema":2,"tree":"` + tree + `","by":"otro/modelo","at":"2026-11-16T12:00:00Z","verdict":"verified","notes":[]}`,
		"null runtime":   `{"schema":2,"tree":"` + tree + `","by":"otro/modelo","at":"2026-11-16T12:00:00Z","verdict":"verified","notes":[],"runtime":null}`,
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := verify.ParseVerificationRecord([]byte(raw))
			if err == nil {
				t.Fatal("a schema-2 record without a receipt parsed")
			}
			if !strings.Contains(strings.ToLower(err.Error()), "runtime receipt") {
				t.Errorf("error %q does not name the missing receipt", err)
			}
		})
	}
}

// TestParseVerificationRecordRefusesAForgedReceipt: a record whose receipt names
// an unsafe identity, a moved output or an unknown fact is refused whole.
func TestParseVerificationRecordRefusesAForgedReceipt(t *testing.T) {
	valid, err := verify.SealManagedRecord(managedAnswer(), validReceipt(), startedAt)
	if err != nil {
		t.Fatalf("SealManagedRecord: %v", err)
	}
	rendered, err := verify.RenderVerificationRecord(valid)
	if err != nil {
		t.Fatalf("RenderVerificationRecord: %v", err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(rendered, &fields); err != nil {
		t.Fatalf("the seal is not JSON: %v", err)
	}
	var receipt map[string]any
	if err := json.Unmarshal(fields["runtime"], &receipt); err != nil {
		t.Fatalf("the receipt is not JSON: %v", err)
	}

	cases := map[string]func(map[string]any){
		"unsafe grant":     func(r map[string]any) { r["grant"] = "forged-grant" },
		"unknown outcome":  func(r map[string]any) { r["outcome"] = "forged-outcome" },
		"moved output":     func(r map[string]any) { r["output_tree"] = "fedcba9876543210fedcba9876543210fedcba98" },
		"missing snapshot": func(r map[string]any) { delete(r, "snapshot") },
		"extra field":      func(r map[string]any) { r["sealed"] = true },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			copyOfReceipt := make(map[string]any, len(receipt))
			for key, value := range receipt {
				copyOfReceipt[key] = value
			}
			mutate(copyOfReceipt)
			forgedReceipt, err := json.Marshal(copyOfReceipt)
			if err != nil {
				t.Fatalf("marshal the forged receipt: %v", err)
			}
			fields["runtime"] = forgedReceipt
			forged, err := json.Marshal(fields)
			if err != nil {
				t.Fatalf("marshal the forged record: %v", err)
			}
			if _, err := verify.ParseVerificationRecord(forged); err == nil {
				t.Fatalf("a record with a %s parsed", name)
			}
		})
	}
}

// TestParseVerificationRecordRefusesAForgedAuthor: a schema-2 record is authored by
// the model the tool observed. A record whose "by" differs from the receipt's
// session_model is a claim the receipt does not back, so it is refused whole.
func TestParseVerificationRecordRefusesAForgedAuthor(t *testing.T) {
	valid, err := verify.SealManagedRecord(managedAnswer(), validReceipt(), startedAt)
	if err != nil {
		t.Fatalf("SealManagedRecord: %v", err)
	}
	rendered, err := verify.RenderVerificationRecord(valid)
	if err != nil {
		t.Fatalf("RenderVerificationRecord: %v", err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(rendered, &fields); err != nil {
		t.Fatalf("the seal is not JSON: %v", err)
	}
	fields["by"] = json.RawMessage(`"anthropic/claude-forged"`)
	forged, err := json.Marshal(fields)
	if err != nil {
		t.Fatalf("marshal the forged record: %v", err)
	}

	if _, err := verify.ParseVerificationRecord(forged); err == nil {
		t.Fatal("a record with a forged author parsed")
	} else if !strings.Contains(err.Error(), "author") {
		t.Errorf("error %q does not name the forged author", err)
	}
}

// TestParseVerificationRecordRefusesANewerSchema: a file from a newer binary is
// not guessed; the human upgrades odachi.
func TestParseVerificationRecordRefusesANewerSchema(t *testing.T) {
	raw := `{"schema":3,"tree":"` + tree + `","by":"otro/modelo","at":"2026-11-16T12:00:00Z","verdict":"verified","notes":[]}`
	_, err := verify.ParseVerificationRecord([]byte(raw))
	if err == nil {
		t.Fatal("a schema-3 record parsed with a binary that writes 2")
	}
	if !strings.Contains(err.Error(), "upgrade odachi") {
		t.Errorf("error %q does not send the human to upgrade", err)
	}
}

// TestAModelCannotPopulateTheReceiptFromItsAnswer: the answer is the model's
// words. A receipt-shaped key inside it does not become the tool's receipt.
func TestAModelCannotPopulateTheReceiptFromItsAnswer(t *testing.T) {
	forged := `{"schema":1,"tree":"` + tree + `","by":"otro/modelo","verdict":"verified","notes":[],` +
		`"runtime":{"grant":"forged-grant","session":"ses_forged","outcome":"completed"}}`
	if _, err := verify.ParseVerificationRecord([]byte(forged)); err == nil {
		t.Fatal("a schema-1 answer carrying its own receipt parsed as evidence")
	}
	if _, err := verifydomain.Parse([]byte(forged)); err == nil {
		t.Fatal("the domain accepted a model-written receipt")
	}
}

// TestLegacyUnmanagedVerifyWritesNoRuntimeReceipt: `verify --exec` stays
// unmanaged: its stored record is schema 1 and carries no runtime receipt.
func TestLegacyUnmanagedVerifyWritesNoRuntimeReceipt(t *testing.T) {
	files := newFakeFiles()
	shell := &fakeShell{stdout: legacyRecordText}

	result := verify.Run(options(), files, fakeGit{tree: tree}, shell)
	if result.Code != verify.CodeOK {
		t.Fatalf("run: exit %d (%s)", result.Code, result.Message)
	}
	stored := files.files[filepath.Join(store, "verifications", tree+".json")]
	if !strings.Contains(stored, `"schema": 1`) || strings.Contains(stored, `"runtime"`) {
		t.Errorf("the unmanaged record claims a runtime receipt:\n%s", stored)
	}
	record, err := verify.ParseVerificationRecord([]byte(stored))
	if err != nil {
		t.Fatalf("the unmanaged record no longer reads: %v", err)
	}
	if record.Schema != 1 || record.Runtime != nil {
		t.Errorf("the unmanaged record = schema %d runtime %v, want 1 and none", record.Schema, record.Runtime)
	}
}

// TestManagedVerifySealsAReceiptSchemaTwoAccepts: everything the managed verifier
// seals is a valid receipt, names the tree it read as its output, and becomes a
// schema-2 record without a correction.
func TestManagedVerifySealsAReceiptSchemaTwoAccepts(t *testing.T) {
	log := &managedEvents{}
	fixture := newManagedFixture(log)

	result := verify.RunManagedVerify(t.Context(), managedRequest(), fixture.deps)
	if result.Code != verify.CodeOK || result.Sealed == nil {
		t.Fatalf("managed verify = %+v, want a sealed receipt", result)
	}
	if err := result.Sealed.Validate(); err != nil {
		t.Fatalf("the tool sealed an invalid receipt: %v", err)
	}
	if result.Sealed.OutputTree != tree {
		t.Errorf("output tree = %q, want the frozen candidate %q", result.Sealed.OutputTree, tree)
	}
	record, err := verify.SealManagedRecord(managedAnswer(), *result.Sealed, startedAt)
	if err != nil {
		t.Fatalf("SealManagedRecord of the tool's own receipt: %v", err)
	}
	if record.Schema != 2 || record.Runtime == nil {
		t.Fatalf("the sealed record = %+v, want schema 2 with the receipt", record)
	}
}
