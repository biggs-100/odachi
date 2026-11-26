package verify_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/app/verify"
	domaingrant "github.com/biggs-100/odachi/internal/domain/grant"
)

// testGrant is a deterministic Go-issued identity: the receipt carries a typed
// grant, and its JSON spelling is the canonical base64url form.
func testGrant() domaingrant.ID {
	var id domaingrant.ID
	for i := range id {
		id[i] = byte(i + 1)
	}
	return id
}

// validReceipt is one complete tool-sealed runtime receipt: every fact the tool
// observed for one verifier run over one frozen candidate.
func validReceipt() verify.SealedRuntimeReceipt {
	return verify.SealedRuntimeReceipt{
		Grant:        testGrant(),
		Session:      domaingrant.SessionID("ses_verify_1"),
		Runtime:      "opencode",
		Model:        "opencode/model",
		SessionModel: "opencode/observed-model",
		Role:         verify.VerifierRole,
		Tree:         tree,
		Base:         verifierBase,
		OutputTree:   tree,
		Snapshot:     snapshotBefore,
		StartedAt:    startedAt.Format(time.RFC3339),
		FinishedAt:   finishedAt.Format(time.RFC3339),
		Outcome:      verify.OutcomeCompleted,
	}
}

// receiptJSON renders a receipt through the canonical encoder, then mutates the
// decoded object, so one test can drop or forge exactly one fact.
func receiptJSON(t *testing.T, mutate func(map[string]any)) string {
	t.Helper()
	encoded, err := verify.RenderRuntimeReceipt(validReceipt())
	if err != nil {
		t.Fatalf("RenderRuntimeReceipt: %v", err)
	}
	var fields map[string]any
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatalf("the canonical receipt is not JSON: %v", err)
	}
	if mutate != nil {
		mutate(fields)
	}
	mutated, err := json.Marshal(fields)
	if err != nil {
		t.Fatalf("marshal the mutated receipt: %v", err)
	}
	return string(mutated)
}

// TestSealedRuntimeReceiptAcceptsEveryToolFact: one complete receipt validates,
// its canonical bytes name every fact, and the bytes parse back to the same
// receipt.
func TestSealedRuntimeReceiptAcceptsEveryToolFact(t *testing.T) {
	receipt := validReceipt()
	if err := receipt.Validate(); err != nil {
		t.Fatalf("a complete receipt was refused: %v", err)
	}

	encoded, err := verify.RenderRuntimeReceipt(receipt)
	if err != nil {
		t.Fatalf("RenderRuntimeReceipt: %v", err)
	}
	again, err := verify.RenderRuntimeReceipt(receipt)
	if err != nil {
		t.Fatalf("RenderRuntimeReceipt: %v", err)
	}
	if string(encoded) != string(again) {
		t.Fatal("two renders of one receipt are different bytes")
	}
	for _, key := range []string{
		`"grant"`, `"session"`, `"runtime"`, `"model"`, `"session_model"`, `"role"`,
		`"tree"`, `"base"`, `"output_tree"`, `"snapshot"`, `"started_at"`, `"finished_at"`, `"outcome"`,
	} {
		if !strings.Contains(string(encoded), key) {
			t.Errorf("the canonical receipt does not carry %s:\n%s", key, encoded)
		}
	}
	parsed, err := verify.ParseRuntimeReceipt(encoded)
	if err != nil {
		t.Fatalf("ParseRuntimeReceipt: %v", err)
	}
	if parsed != receipt {
		t.Errorf("round trip = %+v, want %+v", parsed, receipt)
	}
}

// TestSealedRuntimeReceiptRefusesEveryMissingFact: a receipt is evidence, so a
// fact the tool did not observe cannot be left out: one missing field is one
// receipt that attests nothing.
func TestSealedRuntimeReceiptRefusesEveryMissingFact(t *testing.T) {
	cases := map[string]struct {
		key  string
		want string
	}{
		"grant":         {"grant", "no grant"},
		"session":       {"session", "no session"},
		"runtime":       {"runtime", "no runtime"},
		"model":         {"model", "no planned model"},
		"session model": {"session_model", "no observed session model"},
		"role":          {"role", "role"},
		"tree":          {"tree", "no candidate tree"},
		"base":          {"base", "no candidate base"},
		"output tree":   {"output_tree", "no output tree"},
		"snapshot":      {"snapshot", "no snapshot"},
		"started at":    {"started_at", "no start"},
		"finished at":   {"finished_at", "no finish"},
		"outcome":       {"outcome", "outcome"},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			raw := receiptJSON(t, func(fields map[string]any) { delete(fields, testCase.key) })
			_, err := verify.ParseRuntimeReceipt([]byte(raw))
			if err == nil {
				t.Fatalf("a receipt without %s parsed", testCase.key)
			}
			if !strings.Contains(strings.ToLower(err.Error()), testCase.want) {
				t.Errorf("error %q does not name %q", err, testCase.want)
			}
		})
	}
}

// TestSealedRuntimeReceiptRefusesUnsafeIdentities: grant, session and runtime
// become storage names and log lines, so each one is parsed into its type before
// it is evidence.
func TestSealedRuntimeReceiptRefusesUnsafeIdentities(t *testing.T) {
	cases := map[string]struct {
		key   string
		value string
		want  string
	}{
		"grant is not an issued identity": {"grant", "forged-grant", "not a Go-issued identity"},
		"grant is empty":                  {"grant", "", "no grant"},
		"session with a space":            {"session", "ses verify", "not a safe identity"},
		"session with a bidi override":    {"session", "ses\u202everify", "not a safe identity"},
		"session with a format control":   {"session", "ses\u200bverify", "not a safe identity"},
		"session too long":                {"session", strings.Repeat("a", 257), "not a safe identity"},
		"runtime with a space":            {"runtime", "open code", "not a canonical runtime name"},
		"runtime with a slash":            {"runtime", "opencode/managed", "not a canonical runtime name"},
		"planned model with shell code":   {"model", "opencode/model;rm -rf ~", "not a model identity"},
		"observed model with a space":     {"session_model", "opencode/model name", "not a model identity"},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			raw := receiptJSON(t, func(fields map[string]any) { fields[testCase.key] = testCase.value })
			_, err := verify.ParseRuntimeReceipt([]byte(raw))
			if err == nil {
				t.Fatalf("a receipt with an unsafe %s parsed", testCase.key)
			}
			if !strings.Contains(err.Error(), testCase.want) {
				t.Errorf("error %q does not name %q", err, testCase.want)
			}
		})
	}
}

// TestSealedRuntimeReceiptRefusesNonCanonicalHashes: the trees and the snapshot
// key a store path and name the bytes read, so only the exact hex the toolchain
// writes is a fact.
func TestSealedRuntimeReceiptRefusesNonCanonicalHashes(t *testing.T) {
	cases := map[string]struct {
		key   string
		value string
		want  string
	}{
		"upper case tree":     {"tree", strings.ToUpper(tree), "not a canonical tree hash"},
		"short tree":          {"tree", tree[:39], "not a canonical tree hash"},
		"non hex base":        {"base", strings.Repeat("z", 40), "not a canonical tree hash"},
		"snapshot is a tree":  {"snapshot", tree, "not a SHA-256 content identity"},
		"snapshot upper case": {"snapshot", strings.ToUpper(snapshotBefore), "not a SHA-256 content identity"},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			raw := receiptJSON(t, func(fields map[string]any) {
				fields[testCase.key] = testCase.value
				if testCase.key == "tree" {
					fields["output_tree"] = testCase.value
				}
			})
			_, err := verify.ParseRuntimeReceipt([]byte(raw))
			if err == nil {
				t.Fatalf("a non-canonical %s parsed", testCase.key)
			}
			if !strings.Contains(err.Error(), testCase.want) {
				t.Errorf("error %q does not name %q", err, testCase.want)
			}
		})
	}
}

// TestSealedRuntimeReceiptRefusesABackwardsOrLooseClock: the two instants are the
// tool's, in one canonical spelling, and a run cannot end before it starts.
func TestSealedRuntimeReceiptRefusesABackwardsOrLooseClock(t *testing.T) {
	cases := map[string]struct {
		mutate func(map[string]any)
		want   string
	}{
		"finished before started": {
			mutate: func(fields map[string]any) {
				fields["started_at"] = finishedAt.Format(time.RFC3339)
				fields["finished_at"] = startedAt.Add(-time.Minute).Format(time.RFC3339)
			},
			want: "before it started",
		},
		"started at is prose": {
			mutate: func(fields map[string]any) { fields["started_at"] = "this afternoon" },
			want:   "not canonical RFC3339",
		},
		"finished at carries an offset": {
			mutate: func(fields map[string]any) { fields["finished_at"] = "2026-11-16T14:00:05+02:00" },
			want:   "not canonical RFC3339",
		},
		"finished at carries a fraction": {
			mutate: func(fields map[string]any) { fields["finished_at"] = "2026-11-16T12:00:05.5Z" },
			want:   "not canonical RFC3339",
		},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			raw := receiptJSON(t, testCase.mutate)
			_, err := verify.ParseRuntimeReceipt([]byte(raw))
			if err == nil {
				t.Fatal("a receipt with a loose clock parsed")
			}
			if !strings.Contains(err.Error(), testCase.want) {
				t.Errorf("error %q does not name %q", err, testCase.want)
			}
		})
	}
}

// TestSealedRuntimeReceiptRefusesUnknownRoleOrOutcome: both vocabularies are
// closed. A role no actor issues, or an outcome no tool seals, is not evidence.
func TestSealedRuntimeReceiptRefusesUnknownRoleOrOutcome(t *testing.T) {
	cases := map[string]struct {
		key   string
		value string
		want  string
	}{
		"forged role":    {"role", "forged-role", "known roles"},
		"reviewer role":  {"role", "reviewer", "known roles"},
		"forged outcome": {"outcome", "forged-outcome", "only sealed outcome"},
		"failed outcome": {"outcome", "failed", "only sealed outcome"},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			raw := receiptJSON(t, func(fields map[string]any) { fields[testCase.key] = testCase.value })
			_, err := verify.ParseRuntimeReceipt([]byte(raw))
			if err == nil {
				t.Fatalf("an unknown %s parsed", testCase.key)
			}
			if !strings.Contains(err.Error(), testCase.want) {
				t.Errorf("error %q does not name %q", err, testCase.want)
			}
		})
	}
}

// TestSealedRuntimeReceiptKeepsTheModelsConsistent: one run names one provider
// family. The observed session model may be another model of the same provider -
// a fallback or a variant - but not a model of another provider.
func TestSealedRuntimeReceiptKeepsTheModelsConsistent(t *testing.T) {
	cases := map[string]struct {
		planned  string
		observed string
		want     string
	}{
		"same provider": {
			planned:  "opencode/model-a",
			observed: "opencode/model-b",
		},
		"same provider and a variant": {
			planned:  "opencode/model#high",
			observed: "opencode/model",
		},
		"observed without a provider cannot be compared": {
			planned:  "opencode/model",
			observed: "model-fallback",
		},
		"another provider": {
			planned:  "opencode/model",
			observed: "anthropic/claude",
			want:     "another provider",
		},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			raw := receiptJSON(t, func(fields map[string]any) {
				fields["model"] = testCase.planned
				fields["session_model"] = testCase.observed
			})
			_, err := verify.ParseRuntimeReceipt([]byte(raw))
			if testCase.want == "" {
				if err != nil {
					t.Fatalf("a consistent receipt was refused: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("an inconsistent receipt parsed")
			}
			if !strings.Contains(err.Error(), testCase.want) {
				t.Errorf("error %q does not name %q", err, testCase.want)
			}
		})
	}
}

// TestSealedRuntimeReceiptRefusesAVerifierOutputThatMoved: a verifier reads one
// frozen tree, so the tree it output is the tree it read. A receipt that says
// otherwise describes a run the freeze refused.
func TestSealedRuntimeReceiptRefusesAVerifierOutputThatMoved(t *testing.T) {
	raw := receiptJSON(t, func(fields map[string]any) {
		fields["output_tree"] = "fedcba9876543210fedcba9876543210fedcba98"
	})
	_, err := verify.ParseRuntimeReceipt([]byte(raw))
	if err == nil {
		t.Fatal("a verifier receipt with a different output tree parsed")
	}
	if !strings.Contains(err.Error(), "cannot move it") {
		t.Errorf("error %q does not say the verifier froze one tree", err)
	}
}

// TestParseRuntimeReceiptRefusesExtraOrPartialJSON: the receipt is one strict
// object, so a field the tool never wrote cannot ride along as evidence.
func TestParseRuntimeReceiptRefusesExtraOrPartialJSON(t *testing.T) {
	cases := map[string]string{
		"an unknown field": receiptJSON(t, func(fields map[string]any) { fields["sealed"] = true }),
		"two objects":      receiptJSON(t, nil) + receiptJSON(t, nil),
		"prose":            "the verifier looked at everything and it seems fine",
		"empty":            "   ",
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := verify.ParseRuntimeReceipt([]byte(raw)); err == nil {
				t.Fatalf("%s parsed as a runtime receipt", name)
			}
		})
	}
}
