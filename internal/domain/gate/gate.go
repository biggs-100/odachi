// Package gate answers one question: can this work be called done? It is pure
// and ordered, so the first unanswered thing is always the one the human sees.
package gate

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// Code is the exit code of `odachi check`.
type Code int

const (
	// Ready means proof exists for the exact tree.
	Ready Code = 0
	// NeedsWork means the agent can act without asking the human anything.
	NeedsWork Code = 1
	// NeedsDecision means only the human can unblock this.
	NeedsDecision Code = 2
)

// Reasons are stable identifiers; the Message is what a human reads.
const (
	ReasonReady           = "ready"
	ReasonNoBrief         = "no_brief"
	ReasonBriefIncomplete = "brief_incomplete"
	ReasonOutOfScope      = "out_of_scope"
	ReasonFenceLoosened   = "fence_loosened"
	ReasonOpenQuestion    = "open_question"
	// ReasonUnplanned is a requirement the feature declares and no step names: the
	// plan and the vision disagree, and only the agent can close the gap.
	ReasonUnplanned    = "requirement_unplanned"
	ReasonProofMissing = "proof_missing"
	ReasonProofStale   = "proof_stale"
	ReasonProofFailed  = "proof_failed"
	// ReasonQuoteMismatch is a requirement whose words are not the brief's: the tool
	// wrote the sentence twice and only the human edits either, so a drift is an
	// agent rewriting what it was asked for (0314).
	ReasonQuoteMismatch = "requirement_not_the_human_words"
	// ReasonUndoMissed is a suite that stayed green with the change undone: the
	// tests do not notice the code being reverted, so a green proof certifies
	// nothing about this behavior (0052).
	ReasonUndoMissed = "undo_missed"
	// ReasonUndoMissing is a close whose row was never taken for this tree. The
	// close owes it the way it owes the proof, and a stored proof without it is
	// the skip that would make the row a coincidence of one invocation.
	ReasonUndoMissing = "undo_missing"
	// ReasonMutationFailed is a guard proven by breaking it: the run took a string
	// out of the source of truth and the suite either stayed green or failed for a
	// cause that did not name what the run expected. A guard that does not fire is
	// a comment with a test's name on it (0297).
	ReasonMutationFailed = "mutation_failed"
	// ReasonVerifySlice is a close that owes a verifier and carries more authored
	// lines than one verifier can finish: the record would judge a tree too large
	// to read, so the gate names the number it measured before anybody pays for the
	// verdict, and the rest belongs to a feature of its own (0331).
	ReasonVerifySlice = "verification_slice"
	// ReasonVerifyMissing is a close whose commits declared a risk and whose tree
	// carries no verification an independent verifier signed: who writes does not
	// verify itself, and the record is what makes it checkable (0315).
	ReasonVerifyMissing = "verification_missing"
	// ReasonVerifyRefuted is the record that exists and says the requirement does
	// not hold. It is not the same sentence as a missing record, and calling it one
	// would send the human to produce what is already there (0315).
	ReasonVerifyRefuted = "verification_refuted"
)

// Commands that resolve a block. Every blocked result names one.
const (
	nextStart = "odachi start"
	// nextAsk carries the shape of the answer. A bare `odachi ask` prints usage:
	// naming it would be a dead end for a human too, and the bench measures exactly
	// that.
	nextAsk        = "odachi ask \"<your decision>\""
	nextCheckProof = "odachi check --proof <file>"
	nextVerify     = "odachi verify --exec \"<command>\""
)

// SliceLines is how many lines a candidate may author when the tree that closes
// the feature owes it a verifier: 400, the number the method already writes for
// one task. Past it the verdict would read a tree nobody can finish in one
// sitting, and the rest of the work belongs to a feature of its own (0331).
const SliceLines = 400

// Reasons is every reason the gate can answer with, in the order the constants
// declare them. It exists so the rulebook can point at one: a rule enforced by
// "the gate refuses" names the refusal, and a name that is not in this list is a
// promise the tool does not keep (0314).
func Reasons() []string {
	return []string{
		ReasonReady,
		ReasonNoBrief,
		ReasonBriefIncomplete,
		ReasonOutOfScope,
		ReasonFenceLoosened,
		ReasonOpenQuestion,
		ReasonUnplanned,
		ReasonQuoteMismatch,
		ReasonProofMissing,
		ReasonProofStale,
		ReasonProofFailed,
		ReasonUndoMissed,
		ReasonUndoMissing,
		ReasonMutationFailed,
		ReasonVerifySlice,
		ReasonVerifyMissing,
		ReasonVerifyRefuted,
	}
}

// IsReason reports whether the name is one the gate answers with.
func IsReason(name string) bool {
	return slices.Contains(Reasons(), name)
}

// State is everything Evaluate needs, gathered by the caller. The gate itself
// reads no files and runs no commands.
type State struct {
	HasBrief     bool
	BriefMissing []string

	TreeHash      string
	ProofTreeHash string
	ProofOutput   string

	// ProofFailed says the command that produced the proof exited non-zero. The
	// gate cannot see an exit code, so the caller that ran the command reports
	// it, together with the last line it printed and the command itself: the
	// continuation has to name the exact command, not a generic one.
	ProofFailed   bool
	ProofLastLine string
	ProofCommand  string

	// ProofTimedOut is the other way a proof command fails: it never finished.
	// The sentence says how long the tool waited, which is what the human has to
	// raise or what they have to fix.
	ProofTimedOut       bool
	ProofTimeoutSeconds int

	// UndoMissed says the suite stayed green with the change undone.
	// UndoCommand is what ran, so the continuation names it exactly instead of a
	// generic command.
	//
	// An undo that could not be judged -the reverted copy did not build- is not
	// here on purpose: it is neither a pass nor a miss, and blocking on it would
	// stop every change that adds a symbol its tests call. The caller reports it
	// as a note, by name, and the row's silence is what it must never be.
	UndoMissed  bool
	UndoCommand string
	// MutationFailed says the guard did not fire: the suite stayed green with the
	// string taken out, or it failed for a cause that did not name what the run
	// expected. MutationNote carries which of the two, and the three fields after it
	// let the continuation reproduce the exact question (0297).
	MutationFailed   bool
	MutationNote     string
	MutationCommand  string
	MutationFile     string
	MutationRemoved  string
	MutationExpected string
	// UndoMissing says the close has no row for this tree. The close owes it the
	// way it owes the proof, and a stored proof without it is the skip that turns
	// the row into a coincidence of one invocation (0052).
	UndoMissing bool

	OpenQuestions int
	// Unplanned names the requirements the feature declares and no step names. It
	// is gathered when the feature closes: at the close, a requirement nobody
	// planned is a requirement nobody built, and the plan is the artifact that
	// says so. Mid-feature the list is empty on purpose -a requirement added by a
	// change waits for its step, and the steps are planned before the work.
	Unplanned []string
	// QuoteMismatch names the requirements whose words are not the brief's, each
	// with both sentences. It is gathered at the close: mid-feature the requirement
	// may still be waiting for a `ask --change`, and the quote is the human's
	// decision, not a shape of the plan.
	QuoteMismatch []string
	// VerificationMissing names the risks whose commits asked for a verifier and
	// whose record this tree does not have, or the file it cannot read. It is
	// gathered at the close, like the quote: mid-feature the record would name a
	// tree that still moves (0315).
	VerificationMissing []string
	// VerificationLines is what the candidate costs, when the close owes a
	// verifier: the authored lines of the feature's diff, measured before the
	// record is looked for. Past SliceLines the verifier would judge a tree nobody
	// can finish, and the gate names the number it read -but only while the record
	// is missing, because a tree already signed has no verdict left to spend (0331).
	VerificationLines int
	// VerificationRefuted holds the notes of a record that answered refuted: the
	// verifier ran, it read these bytes, and the requirement does not hold.
	VerificationRefuted []string
	Violations          []string
	// ProofDeferred says the feature is still open, so the proof is not owed yet:
	// it is owed by the tree that closes the feature, which is the one a pull
	// request and CI judge. Absence and staleness stop blocking; a proof this run
	// took and watched fail still blocks, because a failure somebody just saw is
	// not a formality of the closing.
	ProofDeferred bool
	// ProofPaperwork says the proof is stale only because documents moved since it,
	// and the feature had already closed: a release note cannot fail a suite, and
	// the tree that judges the change -the pull request, with the suite run from
	// the base ref- is where it is proved again. Staleness stops blocking; a
	// missing proof does not become paperwork, and neither does the tree that
	// closes the feature.
	ProofPaperwork bool
	// FenceLoosened names the rules, paths or dependencies the base fence had
	// and the candidate no longer has. Only the caller with --fence-from can
	// answer it: without a base ref, the worktree fence is the policy.
	FenceLoosened []string
}

// Result is the verdict plus what to do next.
type Result struct {
	Code    Code
	Reason  string
	Message string
	Next    string
}

// Evaluate applies the documented order: no brief -> brief incomplete -> out of
// scope -> fence loosened -> open question -> proof missing -> proof stale ->
// proof failed -> undo missed -> ready. The order decides the Code too: a block
// only the human can clear is NeedsDecision, a block the agent can clear is
// NeedsWork.
func Evaluate(s State) Result {
	switch {
	case !s.HasBrief:
		return Result{NeedsDecision, ReasonNoBrief, "there is no brief: we do not know yet what must exist", nextStart}
	case len(s.BriefMissing) > 0:
		return Result{NeedsDecision, ReasonBriefIncomplete, "the brief is incomplete: missing " + strings.Join(s.BriefMissing, ", "), nextStart}
	case len(s.Violations) > 0:
		return Result{NeedsDecision, ReasonOutOfScope, "out of scope: " + strings.Join(s.Violations, "; "), nextAsk}
	case len(s.FenceLoosened) > 0:
		return Result{NeedsDecision, ReasonFenceLoosened,
			"the fence was loosened without new non-goals in the brief: " + strings.Join(s.FenceLoosened, "; ") +
				"; the fence may only change with the brief's non-goals. Restore it, or write the non-goals first",
			nextAsk}
	case s.OpenQuestions > 0:
		return Result{NeedsDecision, ReasonOpenQuestion, openQuestions(s.OpenQuestions), nextAsk}
	// The plan and the vision disagree: the feature declares a requirement and no
	// step names it, so a close would ship a promise nobody worked on. It is
	// NeedsWork and not a decision: the agent plans the step the human already
	// asked for.
	case len(s.Unplanned) > 0:
		verb := "names it"
		if len(s.Unplanned) > 1 {
			verb = "names them"
		}
		return Result{NeedsWork, ReasonUnplanned,
			"the feature declares " + strings.Join(s.Unplanned, ", ") + " and no step " + verb +
				": a requirement nobody planned is a requirement nobody built",
			nextPlan(s.Unplanned[0])}
	// The plan is written in the agent's words: the tool wrote the sentence twice,
	// only the human edits either, and a drift means the requirement was rewritten
	// instead of quoted. It is NeedsWork because the fix is a copy, and the
	// continuation names the other door: a human who changed their mind records it
	// with `ask --change`, and the quote stops being the original's (0314).
	case len(s.QuoteMismatch) > 0:
		return Result{NeedsWork, ReasonQuoteMismatch,
			"the plan does not quote the human: " + strings.Join(s.QuoteMismatch, "; ") +
				"; copy the brief's words into the requirement, or record the change with odachi ask --change",
			`odachi ask --change "<the human's words>" --req R1`}
	// A close that owes a verifier measures its candidate first: a verifier spent
	// on a tree nobody can finish is a verdict that cannot land, and the rest of
	// the work is a feature of its own, not the tail of this one (0331). The
	// question only exists while the record is missing: once the verifier signed
	// this tree there is nothing left to spend, and refusing it would name a door
	// that cannot shrink a diff somebody already read. It comes before the plain
	// missing case on purpose -an over-limit close with no record earns the better
	// diagnosis, the number and the slice, not the absence (0315).
	case s.VerificationLines > SliceLines && len(s.VerificationMissing) > 0:
		return Result{NeedsWork, ReasonVerifySlice,
			fmt.Sprintf("the candidate is %d authored lines and the limit for a close that owes a verifier is %d: the verifier would be handed a tree it cannot finish; the rest of the work belongs to a feature of its own",
				s.VerificationLines, SliceLines),
			nextStart}
	// A close that declared a risk owes the record an independent verifier leaves,
	// and the rule is the one the catalog already states: who writes does not verify
	// itself. It is NeedsWork because the agent can run the verifier; the record,
	// not a sentence in a task document, is what the gate reads (0315).
	case len(s.VerificationMissing) > 0:
		return Result{NeedsWork, ReasonVerifyMissing,
			"the commits declared a risk and this tree has no verification it can read: " + strings.Join(s.VerificationMissing, "; "),
			nextVerify}
	// The record exists and says the requirement does not hold. The tree does not
	// close on the words of the person who wrote it, and the notes say what showed
	// it (0315).
	case len(s.VerificationRefuted) > 0:
		return Result{NeedsWork, ReasonVerifyRefuted,
			"the verifier refuted this tree: " + strings.Join(s.VerificationRefuted, "; "),
			nextVerify}
	// A proof exists as soon as it names the tree it came from. An empty body is
	// still a proof: `go vet`, `tsc --noEmit` and `gofmt -l` all succeed while
	// printing nothing, and telling a human "the proof is missing" right after
	// they ran one is a dead end that names a command which cannot resolve it.
	case !s.ProofDeferred && s.ProofTreeHash == "":
		return Result{NeedsWork, ReasonProofMissing, "the proof for this change is missing", nextCheckProof}
	case !s.ProofDeferred && !s.ProofPaperwork && s.ProofTreeHash != s.TreeHash:
		return Result{NeedsWork, ReasonProofStale, "the proof is from another state of the code", nextCheckProof}
	case s.ProofTimedOut:
		return Result{NeedsWork, ReasonProofFailed, "the proof did not finish in " + strconv.Itoa(s.ProofTimeoutSeconds) + "s", nextRunWithTimeout(s.ProofCommand, s.ProofTimeoutSeconds)}
	case s.ProofFailed:
		return Result{NeedsWork, ReasonProofFailed, "the proof failed: " + proofLastLine(s.ProofLastLine), nextRun(s.ProofCommand)}
	// An undo that could not be judged is not a miss and not a pass: it is the
	// caller's note, and it never blocks. Blocking on it would stop every change
	// that adds a symbol its tests call, which is most of them (0052).
	case s.UndoMissing:
		return Result{NeedsWork, ReasonUndoMissing,
			"the row that reverts the change was not taken for this tree, and the close owes it",
			nextRun(s.ProofCommand)}
	case s.UndoMissed:
		return Result{NeedsWork, ReasonUndoMissed,
			"the tests do not notice the change being undone: the suite stayed green without it",
			nextRun(s.UndoCommand)}
	case s.MutationFailed:
		return Result{NeedsWork, ReasonMutationFailed,
			"the mutation row failed: " + s.MutationNote,
			fmt.Sprintf("odachi check --run %q --mutate %q --remove %q --expect %q",
				s.MutationCommand, s.MutationFile, s.MutationRemoved, s.MutationExpected)}
	default:
		if s.ProofDeferred && s.ProofTreeHash != s.TreeHash {
			return Result{Ready, ReasonReady, "ready for a step: the proof is owed when this feature closes, and the tree that closes it is the one CI judges", ""}
		}
		if s.ProofPaperwork && s.ProofTreeHash != s.TreeHash {
			return Result{Ready, ReasonReady, "ready: the proof belongs to the code it ran on, and only documents moved since", ""}
		}
		return Result{Ready, ReasonReady, "ready: there is proof for this tree", ""}
	}
}

// nextPlan names the command that plans the requirement nobody planned: the
// shape plan already documents, with the id the message named. A bare
// `odachi plan` would print usage, which the bench calls a dead end.
func nextPlan(requirement string) string {
	return `odachi plan "<text>" --req ` + requirement
}

// proofLastLine says what the failing command last printed. A command that
// failed in silence is still a failure, and the message never ends in a colon
// with nothing after it.
func proofLastLine(line string) string {
	line = strings.Join(strings.Fields(line), " ")
	if line == "" {
		return "no output"
	}
	return line
}

// nextRun names the command that produced the failing proof, quoted so a
// human can paste it back exactly. Without a known command the continuation
// keeps the placeholder the other blocked results use.
func nextRun(command string) string {
	command = strings.TrimSpace(command)
	if command == "" {
		return `odachi check --run "<command>"`
	}
	return `odachi check --run "` + strings.ReplaceAll(command, `"`, `\"`) + `"`
}

// nextRunWithTimeout is nextRun for the one case where repeating the command
// would repeat the failure: the command did not finish in time. The suggestion
// doubles the wait, so a human who pastes it gets a real chance to finish
// instead of the same timeout again.
func nextRunWithTimeout(command string, seconds int) string {
	next := nextRun(command)
	if seconds <= 0 {
		return next + " --timeout 300"
	}
	return next + " --timeout " + strconv.Itoa(seconds*2)
}

// openQuestions keeps the sentence readable for one and for many.
func openQuestions(n int) string {
	if n == 1 {
		return "1 unanswered question"
	}
	return fmt.Sprintf("%d unanswered questions", n)
}
