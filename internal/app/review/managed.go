package review

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	domaingrant "github.com/biggs-100/odachi/internal/domain/grant"
)

// reviewBindingDomain separates the canonical review encoding from any other
// use of the same digest.
const reviewBindingDomain = "odachi/review/v1"

// ReviewRole is the only role a managed review grant is bound to. It is a
// constant of the encoding, not a caller field: every review grant says
// reviewer, and no plan can say otherwise.
const ReviewRole = "reviewer"

// ReviewBindingFields is the canonical review input a managed grant is bound
// to: the frozen candidate tree and base, the runtime and model, the lens set,
// and the draw and part the reviewer process answers.
type ReviewBindingFields struct {
	Tree    string
	Base    string
	Runtime string
	Model   string
	Lenses  []string
	Draw    int
	Draws   int
	Part    int
	Parts   int
}

// ReviewBinding is the canonical SHA-256 digest of one reviewer process. Every
// field is length-prefixed, the lens set is sorted, and the role is written as
// a constant, so the same process always binds the same digest and a different
// tree, base, runtime, model, lens, draw or part binds a different one.
func ReviewBinding(fields ReviewBindingFields) (domaingrant.Binding, error) {
	switch {
	case strings.TrimSpace(fields.Tree) == "":
		return domaingrant.Binding{}, fmt.Errorf("%w: the candidate tree is missing", domaingrant.ErrInvalidBinding)
	case strings.TrimSpace(fields.Base) == "":
		return domaingrant.Binding{}, fmt.Errorf("%w: the candidate base is missing", domaingrant.ErrInvalidBinding)
	case strings.TrimSpace(fields.Runtime) == "":
		return domaingrant.Binding{}, fmt.Errorf("%w: the runtime is missing", domaingrant.ErrInvalidBinding)
	case strings.TrimSpace(fields.Model) == "":
		return domaingrant.Binding{}, fmt.Errorf("%w: the model is missing", domaingrant.ErrInvalidBinding)
	case len(fields.Lenses) == 0:
		return domaingrant.Binding{}, fmt.Errorf("%w: the review names no lens", domaingrant.ErrInvalidBinding)
	case fields.Draw < 1:
		return domaingrant.Binding{}, fmt.Errorf("%w: the draw is missing", domaingrant.ErrInvalidBinding)
	case fields.Draws < fields.Draw:
		return domaingrant.Binding{}, fmt.Errorf("%w: the draw count is smaller than the draw", domaingrant.ErrInvalidBinding)
	case fields.Part < 1:
		return domaingrant.Binding{}, fmt.Errorf("%w: the part is missing", domaingrant.ErrInvalidBinding)
	case fields.Parts < fields.Part:
		return domaingrant.Binding{}, fmt.Errorf("%w: the part count is smaller than the part", domaingrant.ErrInvalidBinding)
	}
	for _, lens := range fields.Lenses {
		if strings.TrimSpace(lens) == "" {
			return domaingrant.Binding{}, fmt.Errorf("%w: a lens name is empty", domaingrant.ErrInvalidBinding)
		}
	}

	hasher := sha256.New()
	writeReviewField(hasher, reviewBindingDomain)
	writeReviewField(hasher, fields.Tree)
	writeReviewField(hasher, fields.Base)
	writeReviewField(hasher, ReviewRole)
	writeReviewField(hasher, fields.Runtime)
	writeReviewField(hasher, fields.Model)
	writeReviewField(hasher, strconv.Itoa(fields.Draw))
	writeReviewField(hasher, strconv.Itoa(fields.Draws))
	writeReviewField(hasher, strconv.Itoa(fields.Part))
	writeReviewField(hasher, strconv.Itoa(fields.Parts))
	writeReviewLenses(hasher, fields.Lenses)

	var binding domaingrant.Binding
	copy(binding[:], hasher.Sum(nil))
	if err := binding.Validate(); err != nil {
		return domaingrant.Binding{}, err
	}
	return binding, nil
}

// writeReviewField writes one length-prefixed field, so a value containing the
// separator cannot collide with a different split of the same bytes.
func writeReviewField(writer io.Writer, value string) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = writer.Write(length[:])
	_, _ = writer.Write([]byte(value))
}

// writeReviewLenses writes a count followed by the sorted, length-prefixed
// names. The lenses are a set: the same set binds the same digest whatever
// order the plan lists it in.
func writeReviewLenses(writer io.Writer, lenses []string) {
	canonical := slices.Clone(lenses)
	slices.Sort(canonical)
	var count [8]byte
	binary.BigEndian.PutUint64(count[:], uint64(len(canonical)))
	_, _ = writer.Write(count[:])
	for _, lens := range canonical {
		writeReviewField(writer, lens)
	}
}

// ReviewSession is the fresh runtime fact a session creator observed: the
// identity OpenCode minted and the model it reported for that session.
type ReviewSession struct {
	ID    domaingrant.SessionID
	Model string
}

// ReviewSessions creates one fresh runtime session per reviewer process. It
// takes no session, continue, fork or resume value, so a caller cannot reopen an
// existing one.
type ReviewSessions interface {
	CreateFreshSession(ctx context.Context) (ReviewSession, error)
}

// ReviewGrants is the one-shot authority boundary each reviewer process
// consumes. The production value is the grant service; a test supplies a fake.
type ReviewGrants interface {
	Issue(ctx context.Context, binding domaingrant.Binding) (domaingrant.Grant, error)
	Claim(ctx context.Context, id domaingrant.ID) (domaingrant.Grant, error)
	BindSession(ctx context.Context, id domaingrant.ID, session domaingrant.SessionID) (domaingrant.Grant, error)
	Complete(ctx context.Context, id domaingrant.ID) (domaingrant.Grant, error)
	Fail(ctx context.Context, id domaingrant.ID) (domaingrant.Grant, error)
}

// ReviewExecutionRequest is what one reviewer process receives after its grant
// is claimed and its fresh session is bound. It carries only that fresh session:
// there is no resume, continue, fork or caller-supplied session field. Model is
// the model the grant binds; SessionModel is the model the runtime observed for
// the fresh session.
type ReviewExecutionRequest struct {
	Prompt       string
	Dir          string
	Tree         string
	Base         string
	Lenses       []string
	Model        string
	SessionModel string
	Session      domaingrant.SessionID
	Draw         int
	Draws        int
	Part         int
	Parts        int
}

// ReviewExecution is one reviewer's answer.
type ReviewExecution struct {
	Output   string
	ExitCode int
}

// ReviewExecutor runs the reviewer once inside the fresh session Go bound. It
// is the only edge that reaches a model; tests inject a fake and spawn nobody.
type ReviewExecutor interface {
	Execute(ctx context.Context, request ReviewExecutionRequest) (ReviewExecution, error)
}

// ReviewFreezer exports one candidate tree into a directory the reviewers may
// read. The managed review calls it exactly once, before any grant is issued.
type ReviewFreezer interface {
	ArchiveRun(root, tree string) (run, candidate string, err error)
	Discard(dir string) error
}

// ManagedUnit is one reviewer process in a managed review: a lens set, the
// model the grant binds, the draw and part it answers, and the prompt it is
// asked. Two units of the same review must differ in lens set, draw or part.
type ManagedUnit struct {
	Lenses []string
	Model  string
	Draw   int
	Draws  int
	Part   int
	Parts  int
	Prompt string
}

// ManagedReview is one managed review of one frozen candidate: the tree and
// base every unit reads, the runtime that creates sessions, and the units.
type ManagedReview struct {
	Dir     string
	Tree    string
	Base    string
	Runtime string
	Units   []ManagedUnit
}

// ManagedAttempt is the evidence one reviewer process leaves: its grant, the
// canonical binding, its fresh session, the bound and observed models, and the
// lens/draw/part it answered.
type ManagedAttempt struct {
	Grant        domaingrant.ID
	Binding      domaingrant.Binding
	Session      domaingrant.SessionID
	Model        string
	SessionModel string
	Lenses       []string
	Draw         int
	Draws        int
	Part         int
	Parts        int
}

// ManagedReviewResult is what a managed or legacy caller reads. Managed and
// Fresh are true only when at least one reviewer process issued, claimed and
// bound its own authority; the legacy path and every pre-bind refusal keep them
// false.
type ManagedReviewResult struct {
	Code     int
	Message  string
	Frozen   string
	Run      string
	Tree     string
	Base     string
	Output   string
	Managed  bool
	Fresh    bool
	Attempts []ManagedAttempt
}

// ManagedReviewDeps are the four boundaries a managed review needs: the freezer,
// the grant service, the fresh-session runtime and the reviewer executor.
type ManagedReviewDeps struct {
	Freezer  ReviewFreezer
	Grants   ReviewGrants
	Sessions ReviewSessions
	Executor ReviewExecutor
}

// validate refuses an unwired review before it freezes anything.
func (deps ManagedReviewDeps) validate() error {
	switch {
	case deps.Freezer == nil:
		return errors.New("managed review needs a freezer")
	case deps.Grants == nil:
		return errors.New("managed review needs a grant service")
	case deps.Sessions == nil:
		return errors.New("managed review needs a fresh-session runtime")
	case deps.Executor == nil:
		return errors.New("managed review needs a reviewer executor")
	default:
		return nil
	}
}

// validate refuses a plan that cannot bind one reviewer process before the
// candidate is frozen.
func (request ManagedReview) validate() error {
	switch {
	case strings.TrimSpace(request.Tree) == "":
		return errors.New("the managed review names no candidate tree")
	case strings.TrimSpace(request.Base) == "":
		return errors.New("the managed review names no candidate base")
	case strings.TrimSpace(request.Runtime) == "":
		return errors.New("the managed review names no runtime")
	case len(request.Units) == 0:
		return errors.New("the managed review has no reviewer unit")
	}
	for i, unit := range request.Units {
		if len(unit.Lenses) == 0 {
			return fmt.Errorf("reviewer unit %d names no lens", i+1)
		}
		if strings.TrimSpace(unit.Model) == "" {
			return fmt.Errorf("reviewer unit %d names no model", i+1)
		}
		if strings.TrimSpace(unit.Prompt) == "" {
			return fmt.Errorf("reviewer unit %d has no prompt", i+1)
		}
		if unit.Draw < 1 || unit.Draws < unit.Draw {
			return fmt.Errorf("reviewer unit %d names an impossible draw", i+1)
		}
		if unit.Part < 1 || unit.Parts < unit.Part {
			return fmt.Errorf("reviewer unit %d names an impossible part", i+1)
		}
	}
	return nil
}

// RunManagedReview freezes one candidate tree for the whole review and gives
// every unit and draw its own grant and fresh session over that frozen tree. A
// refusal before the session is bound starts no reviewer and claims no
// freshness; a reviewer that fails after the bind leaves its fresh attempt as
// blocked evidence and never completes its grant.
func RunManagedReview(ctx context.Context, request ManagedReview, deps ManagedReviewDeps) ManagedReviewResult {
	result := ManagedReviewResult{Tree: request.Tree, Base: request.Base}
	if err := deps.validate(); err != nil {
		return blockedReview(result, err.Error())
	}
	if err := request.validate(); err != nil {
		return blockedReview(result, err.Error())
	}
	if request.Dir == "" {
		request.Dir = "."
	}

	// One freeze for the whole review: every lens, part and draw reads the same
	// bytes. The freeze comes after every reason not to start a reviewer, so a
	// plan nobody can pay for creates no directory.
	run, frozen, err := deps.Freezer.ArchiveRun(request.Dir, request.Tree)
	if err != nil {
		return blockedReview(result, "could not freeze the candidate for the reviewer: "+oneLine(err.Error())+"; no reviewer was started")
	}
	if strings.TrimSpace(run) == "" || strings.TrimSpace(frozen) == "" {
		return blockedReview(result, "the freeze named no candidate directory; no reviewer was started")
	}
	defer func() { _ = deps.Freezer.Discard(run) }()
	result.Run = run
	result.Frozen = frozen

	// used is every fresh identity this review already bound. The runtime must
	// mint a new one per process, and a repeated identity is refused before the
	// second reviewer runs.
	used := make(map[domaingrant.SessionID]struct{}, len(request.Units))
	for _, unit := range request.Units {
		binding, err := ReviewBinding(ReviewBindingFields{
			Tree:    request.Tree,
			Base:    request.Base,
			Runtime: request.Runtime,
			Model:   unit.Model,
			Lenses:  unit.Lenses,
			Draw:    unit.Draw,
			Draws:   unit.Draws,
			Part:    unit.Part,
			Parts:   unit.Parts,
		})
		if err != nil {
			return blockedReview(result, "could not bind "+unitLabel(unit)+": "+oneLine(err.Error()))
		}

		issued, err := deps.Grants.Issue(ctx, binding)
		if err != nil {
			return blockedReview(result, "could not issue the grant for "+unitLabel(unit)+": "+oneLine(err.Error()))
		}
		if issued.State != domaingrant.StateIssued || issued.Binding != binding {
			return blockedReview(result, "the issued grant is not bound to "+unitLabel(unit))
		}

		claimed, err := deps.Grants.Claim(ctx, issued.ID)
		if err != nil {
			// A consumed, missing, corrupt or foreign grant is a refusal: no
			// session is created for this unit and nobody runs.
			return blockedReview(result, "could not claim the grant for "+unitLabel(unit)+": "+oneLine(err.Error()))
		}
		if claimed.ID != issued.ID || claimed.State != domaingrant.StateClaimed || claimed.Binding != binding {
			return blockedReview(result, "the claimed grant does not match "+unitLabel(unit))
		}

		session, err := deps.Sessions.CreateFreshSession(ctx)
		if err != nil {
			failReviewGrant(ctx, deps.Grants, issued.ID)
			return blockedReview(result, "could not create a fresh reviewer session for "+unitLabel(unit)+": "+oneLine(err.Error()))
		}
		if err := validateReviewSession(session); err != nil {
			failReviewGrant(ctx, deps.Grants, issued.ID)
			return blockedReview(result, "the runtime reported an unusable session for "+unitLabel(unit)+": "+oneLine(err.Error()))
		}
		if _, duplicate := used[session.ID]; duplicate {
			failReviewGrant(ctx, deps.Grants, issued.ID)
			return blockedReview(result, "the runtime reported a duplicate session for "+unitLabel(unit)+": "+string(session.ID))
		}
		// BindSession validates the session grammar before storage; doing it here
		// makes a malformed identity a refusal before any session marker exists.
		if _, err := claimed.BindSession(session.ID); err != nil {
			failReviewGrant(ctx, deps.Grants, issued.ID)
			return blockedReview(result, "the runtime session is not bindable for "+unitLabel(unit)+": "+oneLine(err.Error()))
		}

		bound, err := deps.Grants.BindSession(ctx, issued.ID, session.ID)
		if err != nil {
			failReviewGrant(ctx, deps.Grants, issued.ID)
			return blockedReview(result, "could not bind the fresh session for "+unitLabel(unit)+": "+oneLine(err.Error()))
		}
		if bound.ID != issued.ID || bound.State != domaingrant.StateBound || bound.Session != session.ID {
			failReviewGrant(ctx, deps.Grants, issued.ID)
			return blockedReview(result, "the grant did not bind the fresh session for "+unitLabel(unit))
		}
		used[session.ID] = struct{}{}

		// The grant is issued, claimed and session-bound from here on: the fresh
		// attempt is real evidence even if the reviewer later fails.
		result.Attempts = append(result.Attempts, ManagedAttempt{
			Grant:        issued.ID,
			Binding:      binding,
			Session:      bound.Session,
			Model:        unit.Model,
			SessionModel: session.Model,
			Lenses:       slices.Clone(unit.Lenses),
			Draw:         unit.Draw,
			Draws:        unit.Draws,
			Part:         unit.Part,
			Parts:        unit.Parts,
		})

		execution, err := deps.Executor.Execute(ctx, ReviewExecutionRequest{
			Prompt:       unit.Prompt,
			Dir:          frozen,
			Tree:         request.Tree,
			Base:         request.Base,
			Lenses:       slices.Clone(unit.Lenses),
			Model:        unit.Model,
			SessionModel: session.Model,
			Session:      bound.Session,
			Draw:         unit.Draw,
			Draws:        unit.Draws,
			Part:         unit.Part,
			Parts:        unit.Parts,
		})
		if err != nil {
			failReviewGrant(ctx, deps.Grants, issued.ID)
			result.Output = execution.Output
			return blockedReview(result, "the reviewer failed for "+unitLabel(unit)+": "+oneLine(err.Error()))
		}
		if execution.ExitCode != 0 {
			failReviewGrant(ctx, deps.Grants, issued.ID)
			result.Output = execution.Output
			return blockedReview(result, fmt.Sprintf("the reviewer exited with code %d for %s", execution.ExitCode, unitLabel(unit)))
		}

		completed, err := deps.Grants.Complete(ctx, issued.ID)
		if err != nil {
			failReviewGrant(ctx, deps.Grants, issued.ID)
			result.Output = execution.Output
			return blockedReview(result, "the reviewer finished but the grant did not complete for "+unitLabel(unit)+": "+oneLine(err.Error()))
		}
		if completed.ID != issued.ID || completed.State != domaingrant.StateCompleted {
			failReviewGrant(ctx, deps.Grants, issued.ID)
			result.Output = execution.Output
			return blockedReview(result, "the grant did not complete "+unitLabel(unit))
		}
		result.Output = execution.Output
	}

	result.Code = CodeOK
	result.Message = fmt.Sprintf("reviewed %d units over candidate %s", len(request.Units), request.Tree)
	result.Managed = len(result.Attempts) > 0
	result.Fresh = result.Managed
	return result
}

// blockedReview is the refusal a managed review returns, carrying the fresh
// attempts that already landed and claiming freshness only for them.
func blockedReview(result ManagedReviewResult, message string) ManagedReviewResult {
	result.Code = CodeBlocked
	result.Message = message
	result.Managed = len(result.Attempts) > 0
	result.Fresh = result.Managed
	return result
}

// unitLabel names one reviewer process in a message the human reads.
func unitLabel(unit ManagedUnit) string {
	return fmt.Sprintf("%s draw %d of %d part %d of %d", strings.Join(unit.Lenses, ","), unit.Draw, unit.Draws, unit.Part, unit.Parts)
}

// validateReviewSession refuses an identity the review cannot bind: a missing
// identity or a missing observed model. The grammar is validated by the domain
// through BindSession; duplicate and foreign identities are refused by the plan
// loop and the runtime port.
func validateReviewSession(session ReviewSession) error {
	if session.ID == "" {
		return fmt.Errorf("%w: the fresh reviewer session identity is missing", domaingrant.ErrInvalidSession)
	}
	if strings.TrimSpace(session.Model) == "" {
		return errors.New("the observed reviewer model is missing")
	}
	return nil
}

// failReviewGrant terminates a claimed or bound grant after the unit cannot
// continue. It uses a detached context: a canceled review must not leave a
// claimed grant behind. A grant that is already terminal refuses the write, and
// that refusal is deliberately not surfaced over the unit's own result.
func failReviewGrant(ctx context.Context, grants ReviewGrants, id domaingrant.ID) {
	_, _ = grants.Fail(context.WithoutCancel(ctx), id)
}

// LegacyReviewExecutor is the unmanaged `--exec` boundary: one command that
// receives a prompt and runs no grant machinery.
type LegacyReviewExecutor interface {
	Run(ctx context.Context, command, prompt string) (ReviewExecution, error)
}

// RunUnmanagedReview is the legacy `--exec` path: it runs every unit through the
// command without issuing a grant or creating a session, so the result is
// explicitly unmanaged and carries no freshness claim.
func RunUnmanagedReview(ctx context.Context, command string, units []ManagedUnit, executor LegacyReviewExecutor) ManagedReviewResult {
	result := ManagedReviewResult{}
	if strings.TrimSpace(command) == "" {
		return blockedReview(result, "the unmanaged review names no reviewer command")
	}
	if executor == nil {
		return blockedReview(result, "the unmanaged review needs a reviewer executor")
	}
	if len(units) == 0 {
		return blockedReview(result, "the unmanaged review has no reviewer unit")
	}
	for _, unit := range units {
		execution, err := executor.Run(ctx, command, unit.Prompt)
		if err != nil {
			result.Output = execution.Output
			return blockedReview(result, "the unmanaged reviewer failed for "+unitLabel(unit)+": "+oneLine(err.Error()))
		}
		if execution.ExitCode != 0 {
			result.Output = execution.Output
			return blockedReview(result, fmt.Sprintf("the unmanaged reviewer exited with code %d for %s", execution.ExitCode, unitLabel(unit)))
		}
		result.Output = execution.Output
	}
	result.Code = CodeOK
	result.Message = fmt.Sprintf("the unmanaged review ran %d units without a grant or a session", len(units))
	return result
}
