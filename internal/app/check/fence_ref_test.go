package check_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/check"
	"github.com/biggs-100/odachi/internal/domain/gate"
)

// TestTheFenceCanComeFromATrustedRef: a CI gate reads the base branch's fence,
// so the change that asks to pass cannot loosen the rule it must obey.
func TestTheFenceCanComeFromATrustedRef(t *testing.T) {
	files := newFakeFiles(readyFiles())
	git := &fakeGit{
		tree:    currentTree,
		changed: []string{"internal/payments/stripe.go"},
		filesAt: map[string]string{
			"origin/main:.odachi/fence.toml": "[[rule]]\nnot = \"pagos\"\npaths = [\"**/payments/**\"]\n",
		},
	}

	worktree := run(t, files, git, check.Options{})
	if worktree.Code != 0 {
		t.Fatalf("without --fence-from: code = %d (%s)", worktree.Code, worktree.Message)
	}

	trusted := run(t, files, git, check.Options{FenceFrom: "origin/main"})
	if trusted.Code == 0 {
		t.Fatal("the trusted fence did not block the change")
	}
	if !strings.Contains(trusted.Message, "pagos") {
		t.Errorf("message = %q, want the rule named", trusted.Message)
	}
}

// TestALoosenedFenceNeedsANewBrief: the fence may only lose a rule, a path or a
// dependency when the brief changes too, because start writes them together.
// Emptying the fence with the same brief is the probe the review found green.
func TestALoosenedFenceNeedsANewBrief(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, ".odachi", "fence.toml")] = []byte("")
	git := &fakeGit{
		tree: currentTree,
		filesAt: map[string]string{
			"origin/main:brief.md":           completeBrief,
			"origin/main:.odachi/fence.toml": "[[rule]]\nnot = \"monorepo scopes\"\npaths = [\"internal/tenants/*\"]\n",
		},
	}

	result := run(t, files, git, check.Options{FenceFrom: "origin/main"})

	if result.Code != int(gate.NeedsDecision) || result.Reason != gate.ReasonFenceLoosened {
		t.Fatalf("Code/Reason = %d/%q, want %d/%q (%s)", result.Code, result.Reason, gate.NeedsDecision, gate.ReasonFenceLoosened, result.Message)
	}
	if !strings.Contains(result.Message, "monorepo scopes") || !strings.Contains(result.Message, "the rule is gone") {
		t.Errorf("message = %q, want the rule and what was lost", result.Message)
	}
}

// TestANewBriefMayBringANewFence: a new feature rewrites both files, and the
// fence of the previous feature is not supposed to survive it. The non-goals are
// what the fence is made of, so they are what has to change.
func TestANewBriefMayBringANewFence(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, ".odachi", "fence.toml")] = []byte("")
	git := &fakeGit{
		tree: currentTree,
		filesAt: map[string]string{
			"origin/main:brief.md":           "# the previous feature\n\n## Not included (first version)\n- something else\n",
			"origin/main:.odachi/fence.toml": "[[rule]]\nnot = \"monorepo scopes\"\npaths = [\"internal/tenants/*\"]\n",
		},
	}

	result := run(t, files, git, check.Options{FenceFrom: "origin/main"})

	if result.Code != 0 {
		t.Fatalf("a new brief with a new fence was blocked: %s", result.Message)
	}
}

// TestATrailingSpaceIsNotANewBrief: the probe the review found. A whitespace-only
// edit to the header is not a new vision, so it cannot unlock the fence.
func TestATrailingSpaceIsNotANewBrief(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, "brief.md")] = []byte(strings.Replace(completeBrief, "## NO incluye (primera versión)", "## NO incluye (primera versión)   ", 1))
	files.files[filepath.Join(dir, ".odachi", "fence.toml")] = []byte("")
	git := &fakeGit{
		tree: currentTree,
		filesAt: map[string]string{
			"origin/main:brief.md":           completeBrief,
			"origin/main:.odachi/fence.toml": "[[rule]]\nnot = \"monorepo scopes\"\npaths = [\"internal/tenants/*\"]\n",
		},
	}

	result := run(t, files, git, check.Options{FenceFrom: "origin/main"})

	if result.Code == 0 {
		t.Fatal("a trailing space in the brief unlocked the fence")
	}
	if result.Reason != gate.ReasonFenceLoosened {
		t.Errorf("Reason = %q, want %q (%s)", result.Reason, gate.ReasonFenceLoosened, result.Message)
	}
}

// TestAChangedOutcomeIsNotANewPolicy: the fence is made of the non-goals. A fix
// that rewords what must exist does not get to drop a rule.
func TestAChangedOutcomeIsNotANewPolicy(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, "brief.md")] = []byte(strings.Replace(completeBrief, "- reservar turnos", "- reservar turnos y cancelarlos", 1))
	files.files[filepath.Join(dir, ".odachi", "fence.toml")] = []byte("")
	git := &fakeGit{
		tree: currentTree,
		filesAt: map[string]string{
			"origin/main:brief.md":           completeBrief,
			"origin/main:.odachi/fence.toml": "[[rule]]\nnot = \"monorepo scopes\"\npaths = [\"internal/tenants/*\"]\n",
		},
	}

	result := run(t, files, git, check.Options{FenceFrom: "origin/main"})

	if result.Code == 0 {
		t.Fatal("a changed outcome unlocked the fence")
	}
	if result.Reason != gate.ReasonFenceLoosened {
		t.Errorf("Reason = %q, want %q (%s)", result.Reason, gate.ReasonFenceLoosened, result.Message)
	}
}

// TestAReorderedNonGoalsIsTheSamePolicy: the list is a set. Reordering it is not
// a new vision, so it cannot unlock the fence.
func TestAReorderedNonGoalsIsTheSamePolicy(t *testing.T) {
	briefWith := func(nongoals ...string) string {
		return "# turnos\n\n## What must exist\n- reservar turnos\n\n## Done when\n- reservo y aparece\n\n## Not included (first version)\n- " +
			strings.Join(nongoals, "\n- ") + "\n"
	}

	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, "brief.md")] = []byte(briefWith("firmas", "pagos"))
	files.files[filepath.Join(dir, ".odachi", "fence.toml")] = []byte("")
	git := &fakeGit{
		tree: currentTree,
		filesAt: map[string]string{
			"origin/main:brief.md":           briefWith("pagos", "firmas"),
			"origin/main:.odachi/fence.toml": "[[rule]]\nnot = \"monorepo scopes\"\npaths = [\"internal/tenants/*\"]\n",
		},
	}

	result := run(t, files, git, check.Options{FenceFrom: "origin/main"})

	if result.Code == 0 {
		t.Fatal("reordering the non-goals unlocked the fence")
	}
}

// TestAMalformedCandidateFenceBlocksEvenWithANewBrief: a new brief does not
// excuse a fence the tool cannot read, because after the merge it is the policy
// and every later run would read it as the base.
func TestAMalformedCandidateFenceBlocksEvenWithANewBrief(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, ".odachi", "fence.toml")] = []byte("not toml [[[")
	git := &fakeGit{
		tree: currentTree,
		filesAt: map[string]string{
			"origin/main:brief.md":           "# the previous feature\n\n## Not included (first version)\n- something else\n",
			"origin/main:.odachi/fence.toml": "[[rule]]\nnot = \"monorepo scopes\"\npaths = [\"internal/tenants/*\"]\n",
		},
	}

	result := run(t, files, git, check.Options{FenceFrom: "origin/main"})

	if result.Code == 0 {
		t.Fatal("a new brief let a malformed fence through")
	}
	if !strings.Contains(result.Message, "could not read") {
		t.Errorf("message = %q, want it to name the unreadable file", result.Message)
	}
}

// TestAStricterFenceNeedsNoNewBrief: adding a rule or a path never weakens the
// policy, so the same brief is enough.
func TestAStricterFenceNeedsNoNewBrief(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, ".odachi", "fence.toml")] = []byte(
		"[[rule]]\nnot = \"monorepo scopes\"\npaths = [\"internal/tenants/*\", \"extra/*\"]\n" +
			"[[rule]]\nnot = \"a new rule\"\npaths = [\"x/*\"]\n")
	git := &fakeGit{
		tree: currentTree,
		filesAt: map[string]string{
			"origin/main:brief.md":           completeBrief,
			"origin/main:.odachi/fence.toml": "[[rule]]\nnot = \"monorepo scopes\"\npaths = [\"internal/tenants/*\"]\n",
		},
	}

	result := run(t, files, git, check.Options{FenceFrom: "origin/main"})

	if result.Code != 0 {
		t.Fatalf("a stricter fence was blocked: %s", result.Message)
	}
}

// TestAnUnreadableCandidateFenceBlocks: with --fence-from the candidate's copy is
// not the policy, but it is the policy after the merge. A file the tool cannot
// read must stop the change, not travel silently.
func TestAnUnreadableCandidateFenceBlocks(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, ".odachi", "fence.toml")] = []byte("not toml [[[")
	git := &fakeGit{
		tree: currentTree,
		filesAt: map[string]string{
			"origin/main:brief.md":           completeBrief,
			"origin/main:.odachi/fence.toml": "[[rule]]\nnot = \"monorepo scopes\"\npaths = [\"internal/tenants/*\"]\n",
		},
	}

	result := run(t, files, git, check.Options{FenceFrom: "origin/main"})

	if result.Code == 0 {
		t.Fatal("an unreadable candidate fence was accepted")
	}
	if !strings.Contains(result.Message, "could not read") {
		t.Errorf("message = %q, want it to name the unreadable file", result.Message)
	}
}

// TestAnUnreadableFenceRefIsABlock: a typo in the ref must not read as "there
// is no fence". Turning the gate off by mistake is the failure this refuses.
func TestAnUnreadableFenceRefIsABlock(t *testing.T) {
	files := newFakeFiles(readyFiles())
	git := &fakeGit{tree: currentTree, fileAtErr: errors.New("fatal: not a valid object name")}

	result := run(t, files, git, check.Options{FenceFrom: "origin/mian"})

	if result.Code == 0 {
		t.Fatal("a bad ref read as a project without a fence")
	}
	if !strings.Contains(result.Message, "--fence-from") {
		t.Errorf("message = %q, want the flag named", result.Message)
	}
}
