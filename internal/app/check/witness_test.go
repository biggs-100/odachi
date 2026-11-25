package check_test

import (
	"errors"
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

// TestAWitnessThatCannotBeReadIsNotAnEmptyOne: an absent answer is the normal
// state; a file that is there and cannot be read is not. Reading it as "nobody
// answered" would let a permissions problem hide a witness the gate should count,
// and the human would see a number that is wrong without a word about why.
func TestAWitnessThatCannotBeReadIsNotAnEmptyOne(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[reviewPath(currentTree)] = []byte(verdictWithLenses(t, currentTree, domainreview.Lenses(domainreview.Full), []domainreview.Finding{arguedFinding()}))
	files.files[refutationFileFor(t, currentTree, 1)] = []byte(refutedFile(t, currentTree, "internal/auth/session.go:12 el diff de la línea 12 muestra que la firma ya se validaba"))
	files.denyRead = func(path string) error {
		if strings.Contains(filepath.ToSlash(path), ".refuted") {
			return errors.New("permission denied")
		}
		return nil
	}

	git := citationGit()
	result := run(t, files, git, check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2 (%s)", result.Code, result.Message)
	}
	joined := result.Message + "\n" + strings.Join(result.Lines, "\n")
	if !strings.Contains(joined, "refuted.json") || !strings.Contains(joined, "permission denied") {
		t.Errorf("the refusal does not name the file and the error:\n%s", joined)
	}
}

// TestARefutationStoredUnderThisTreeButNamingAnotherIsIgnored: the filename is
// not authority when the document itself names a different candidate.
func TestARefutationStoredUnderThisTreeButNamingAnotherIsIgnored(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[reviewPath(currentTree)] = []byte(verdictWithLenses(t, currentTree, domainreview.Lenses(domainreview.Full), []domainreview.Finding{blockingFinding()}))
	files.files[refutationPath(currentTree)] = []byte(refutedFile(t, otherTree, "internal/auth/session.go:12 the signature was already verified"))

	git := citationGit()
	result := run(t, files, git, check.Options{RequireReview: true})

	if result.Code != 2 {
		t.Fatalf("code = %d, want 2: a foreign-tree refutation must not lift the block (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, otherTree) || !strings.Contains(result.Message, currentTree) {
		t.Errorf("the refusal does not name both trees: %s", result.Message)
	}
}

// TestAnArguedBlockerNeedsTwoWitnesses is the rule this feature exists for. One
// refuter is one opinion about a claim nobody can point at; two citations are the
// quorum. The gate says how many it has, so the human knows whether to ask for
// another witness or to fix the code.
func TestAnArguedBlockerNeedsTwoWitnesses(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[reviewPath(currentTree)] = []byte(verdictWithLenses(t, currentTree, domainreview.Lenses(domainreview.Full), []domainreview.Finding{arguedFinding()}))
	files.files[refutationFileFor(t, currentTree, 1)] = []byte(refutedFile(t, currentTree, "internal/auth/session.go:12 el diff de la línea 12 muestra que la firma ya se validaba"))

	git := citationGit()

	one := run(t, files, git, check.Options{RequireReview: true})
	if one.Code != 2 {
		t.Fatalf("one witness lifted an argued block: code = %d (%s)", one.Code, one.Message)
	}
	if !strings.Contains(strings.Join(one.Lines, "\n"), "1 of the 2 cited refutations") {
		t.Errorf("the block does not say how many witnesses it has:\n%s", strings.Join(one.Lines, "\n"))
	}

	files.files[refutationFileFor(t, currentTree, 2)] = []byte(refutedFileBy(t, "mimo", currentTree, "internal/auth/session.go:12 el call site ya pasa un token firmado"))

	two := run(t, files, git, check.Options{RequireReview: true})
	if two.Code != 0 {
		t.Fatalf("two witnesses did not lift an argued block: code = %d (%s)", two.Code, two.Message)
	}
}

// witnessFile is one refuter's answer, sealed the way the store does it.
func witnessFile(t *testing.T, refuter, tree string, findings []domainreview.Finding, judgments ...domainreview.Judgment) string {
	t.Helper()
	sealed, err := domainreview.SealRefutation(domainreview.Refutation{Refuter: refuter, Refutations: judgments}, findings, domainreview.CitationEvidence{Tree: tree, Diff: citationDiff}, reviewInstant)
	if err != nil {
		t.Fatalf("sealing the witness: %v", err)
	}
	return string(domainreview.RenderRefutation(sealed))
}

// TestCitationContractAtTheGate exercises the same domain decision through the
// close path: only a checkable citation can lift the blocker.
func TestCitationContractAtTheGate(t *testing.T) {
	const spacedPath = "internal/auth/session store.go"
	spacedDiff := "diff --git a/internal/auth/session store.go b/internal/auth/session store.go\n--- a/internal/auth/session store.go\n+++ b/internal/auth/session store.go\n@@ -12 +12 @@\n-viejo\n+verified signature\n"
	for _, tc := range []struct {
		name    string
		path    string
		diff    string
		why     string
		refTree string
		want    int
	}{
		{name: "valid line", diff: citationDiff, why: "internal/auth/session.go:12 verifies the signature", want: 0},
		{name: "valid quote", diff: citationDiff, why: `internal/auth/session.go "verified signature"`, want: 0},
		{name: "absent quote", diff: citationDiff, why: `internal/auth/session.go "invented signature"`, want: 2},
		{name: "wrong tree", diff: citationDiff, why: "internal/auth/session.go:12 verifies the signature", refTree: otherTree, want: 2},
		{name: "missing evidence", why: "internal/auth/session.go:12 verifies the signature", want: 2},
		{name: "empty diff", diff: " \n", why: "internal/auth/session.go:12 verifies the signature", want: 2},
		{name: "spaced path", path: spacedPath, diff: spacedDiff, why: spacedPath + ":12 verifies the signature", want: 0},
		{name: "repeated spaced path mention", path: spacedPath, diff: spacedDiff, why: spacedPath + " is the file; " + spacedPath + ":12 verifies the signature", want: 0},
		{name: "free prose", diff: citationDiff, why: "the signature was already verified", want: 2},
		{name: "malformed locator", diff: citationDiff, why: "internal/auth/session.go:twelve", want: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			finding := blockingFinding()
			if tc.path != "" {
				finding.Path = tc.path
			}
			refTree := tc.refTree
			if refTree == "" {
				refTree = currentTree
			}
			files := newFakeFiles(readyFiles())
			files.files[reviewPath(currentTree)] = []byte(verdictWithLenses(t, currentTree, domainreview.Lenses(domainreview.Full), []domainreview.Finding{finding}))
			files.files[refutationPath(currentTree)] = domainreview.RenderRefutation(domainreview.Refutation{
				Tree:        refTree,
				Refuter:     "claude",
				Refutations: []domainreview.Judgment{{ID: "F1", Verdict: domainreview.Refuted, Why: tc.why}},
			})
			git := citationGit()
			git.changed = []string{finding.Path}
			git.diff = tc.diff

			result := run(t, files, git, check.Options{RequireReview: true})
			if result.Code != tc.want {
				t.Fatalf("code = %d, want %d (%s)", result.Code, tc.want, result.Message)
			}
		})
	}
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
	files.files[refutationFileFor(t, currentTree, 1)] = []byte(witnessFile(t, "claude", currentTree, findings,
		domainreview.Judgment{ID: "F1", Verdict: domainreview.Refuted, Why: "internal/auth/session.go:12 line 12 shows it was already there"},
		domainreview.Judgment{ID: "F2", Verdict: domainreview.Refuted, Why: "internal/auth/session.go:12 the log line is new but names no token"},
	))
	files.files[refutationFileFor(t, currentTree, 2)] = []byte(witnessFile(t, "mimo", currentTree, findings,
		domainreview.Judgment{ID: "F1", Verdict: domainreview.Refuted, Why: "internal/auth/session.go:12 the call site passes a signed token"},
	))

	git := citationGit()
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
	files.files[refutationFileFor(t, currentTree, 1)] = []byte(refutedFile(t, currentTree, "internal/auth/session.go:12 la línea 12 está dentro de un helper que el diff también agrega"))

	git := citationGit()
	result := run(t, files, git, check.Options{RequireReview: true})

	if result.Code != 0 {
		t.Fatalf("a cited block survived one citation: code = %d (%s)", result.Code, result.Message)
	}
}
