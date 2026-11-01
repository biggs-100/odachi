package check_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/check"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
)

// refutationFileFor names the file of one witness. Round 1 keeps the name a
// refutation always had; the next ones are numbered.
func refutationFileFor(t *testing.T, tree string, round int) string {
	t.Helper()
	rel, err := domainreview.RefutationPath(tree, round)
	if err != nil {
		t.Fatalf("RefutationPath(%d) = %v", round, err)
	}
	return filepath.Join(store, filepath.FromSlash(rel))
}

// arguedFinding is the same defect with no line quoted: the class that needs two
// witnesses before it falls.
func arguedFinding() domainreview.Finding {
	finding := blockingFinding()
	finding.Evidence = domainreview.EvidenceArgued
	return finding
}

// TestAnArguedBlockerNeedsTwoWitnesses is the rule this feature exists for. One
// refuter is one opinion about a claim nobody can point at; two citations are the
// quorum. The gate says how many it has, so the human knows whether to ask for
// another witness or to fix the code.
func TestAnArguedBlockerNeedsTwoWitnesses(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[reviewPath(currentTree)] = []byte(verdictWithLenses(t, currentTree, domainreview.Lenses(domainreview.Full), []domainreview.Finding{arguedFinding()}))
	files.files[refutationFileFor(t, currentTree, 1)] = []byte(refutedFile(t, currentTree, "el diff de la línea 12 muestra que la firma ya se validaba"))

	git := &fakeGit{tree: currentTree, changed: []string{"internal/auth/session.go"}}

	one := run(t, files, git, check.Options{RequireReview: true})
	if one.Code != 2 {
		t.Fatalf("one witness lifted an argued block: code = %d (%s)", one.Code, one.Message)
	}
	if !strings.Contains(strings.Join(one.Lines, "\n"), "1 of the 2 cited refutations") {
		t.Errorf("the block does not say how many witnesses it has:\n%s", strings.Join(one.Lines, "\n"))
	}

	files.files[refutationFileFor(t, currentTree, 2)] = []byte(refutedFile(t, currentTree, "el call site ya pasa un token firmado"))

	two := run(t, files, git, check.Options{RequireReview: true})
	if two.Code != 0 {
		t.Fatalf("two witnesses did not lift an argued block: code = %d (%s)", two.Code, two.Message)
	}
}

// witnessFile is one refuter's answer, sealed the way the store does it.
func witnessFile(t *testing.T, tree string, findings []domainreview.Finding, judgments ...domainreview.Judgment) string {
	t.Helper()
	sealed, err := domainreview.SealRefutation(domainreview.Refutation{Refuter: "claude", Refutations: judgments}, findings, tree, reviewInstant)
	if err != nil {
		t.Fatalf("sealing the witness: %v", err)
	}
	return string(domainreview.RenderRefutation(sealed))
}

// TestTwoFindingsOnOneLineNameTheRightOne is the row the review of this feature
// found: two findings on the same file and line, one lifted and one still
// blocking. The report must count the votes of the row that blocks -naming the
// first row by its place would say "2 of 2" about a finding nobody is looking at.
func TestTwoFindingsOnOneLineNameTheRightOne(t *testing.T) {
	first := arguedFinding()
	first.Why = "the session is accepted without checking the signature"
	second := arguedFinding()
	second.Why = "the same line also logs the token"
	findings := []domainreview.Finding{first, second}

	files := newFakeFiles(readyFiles())
	files.files[reviewPath(currentTree)] = []byte(verdictWithLenses(t, currentTree, domainreview.Lenses(domainreview.Full), findings))
	files.files[refutationFileFor(t, currentTree, 1)] = []byte(witnessFile(t, currentTree, findings,
		domainreview.Judgment{ID: "F1", Verdict: domainreview.Refuted, Why: "line 12 shows it was already there"},
		domainreview.Judgment{ID: "F2", Verdict: domainreview.Refuted, Why: "the log line is new but names no token"},
	))
	files.files[refutationFileFor(t, currentTree, 2)] = []byte(witnessFile(t, currentTree, findings,
		domainreview.Judgment{ID: "F1", Verdict: domainreview.Refuted, Why: "the call site passes a signed token"},
	))

	git := &fakeGit{tree: currentTree, changed: []string{"internal/auth/session.go"}}
	result := run(t, files, git, check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2: F2 has one witness (%s)", result.Code, result.Message)
	}
	lines := strings.Join(result.Lines, "\n")
	if !strings.Contains(lines, "F2 has 1 of the 2 cited refutations") {
		t.Errorf("the report does not count the row that blocks:\n%s", lines)
	}
	if strings.Contains(lines, "F1 has 2 of the 2") {
		t.Errorf("the report counts the row that was lifted:\n%s", lines)
	}
}

// TestACitedBlockerFallsWithOneWitness: the reviewer that quoted the line earns
// the short road. Making it pay for a second witness would punish the one that did
// the work, and the counter-citation is checkable at a glance.
func TestACitedBlockerFallsWithOneWitness(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[reviewPath(currentTree)] = []byte(verdictWithLenses(t, currentTree, domainreview.Lenses(domainreview.Full), []domainreview.Finding{blockingFinding()}))
	files.files[refutationFileFor(t, currentTree, 1)] = []byte(refutedFile(t, currentTree, "la línea 12 está dentro de un helper que el diff también agrega"))

	git := &fakeGit{tree: currentTree, changed: []string{"internal/auth/session.go"}}
	result := run(t, files, git, check.Options{RequireReview: true})

	if result.Code != 0 {
		t.Fatalf("a cited block survived one citation: code = %d (%s)", result.Code, result.Message)
	}
}
