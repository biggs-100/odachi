package verify

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	domaingrant "github.com/biggs-100/odachi/internal/domain/grant"
)

// verifierBindingDomain separates the canonical verifier encoding from any other
// use of the same digest.
const verifierBindingDomain = "odachi/verify/v1"

// VerifierRole is the only role a managed verifier grant is bound to. It is a
// constant of the encoding, not a caller field: every verifier grant says
// verifier, and no request can say otherwise.
const VerifierRole = "verifier"

// VerifierBindingFields is the canonical verifier input a managed grant is bound
// to: the frozen candidate tree and base, the runtime and model, and the
// identity of the verifier that answers.
type VerifierBindingFields struct {
	Tree     string
	Base     string
	Runtime  string
	Model    string
	Verifier string
}

// VerifierBinding is the canonical SHA-256 digest of one verifier process. Every
// field is length-prefixed and the role is written as a constant, so the same
// verifier always binds the same digest and a different tree, base, runtime,
// model or verifier binds a different one.
func VerifierBinding(fields VerifierBindingFields) (domaingrant.Binding, error) {
	switch {
	case strings.TrimSpace(fields.Tree) == "":
		return domaingrant.Binding{}, fmt.Errorf("%w: the candidate tree is missing", domaingrant.ErrInvalidBinding)
	case strings.TrimSpace(fields.Base) == "":
		return domaingrant.Binding{}, fmt.Errorf("%w: the candidate base is missing", domaingrant.ErrInvalidBinding)
	case strings.TrimSpace(fields.Runtime) == "":
		return domaingrant.Binding{}, fmt.Errorf("%w: the runtime is missing", domaingrant.ErrInvalidBinding)
	case strings.TrimSpace(fields.Model) == "":
		return domaingrant.Binding{}, fmt.Errorf("%w: the model is missing", domaingrant.ErrInvalidBinding)
	case strings.TrimSpace(fields.Verifier) == "":
		return domaingrant.Binding{}, fmt.Errorf("%w: the verifier identity is missing", domaingrant.ErrInvalidBinding)
	}

	hasher := sha256.New()
	writeVerifierField(hasher, verifierBindingDomain)
	writeVerifierField(hasher, fields.Tree)
	writeVerifierField(hasher, fields.Base)
	writeVerifierField(hasher, VerifierRole)
	writeVerifierField(hasher, fields.Runtime)
	writeVerifierField(hasher, fields.Model)
	writeVerifierField(hasher, fields.Verifier)

	var binding domaingrant.Binding
	copy(binding[:], hasher.Sum(nil))
	if err := binding.Validate(); err != nil {
		return domaingrant.Binding{}, err
	}
	return binding, nil
}

// writeVerifierField writes one length-prefixed field, so a value containing the
// separator cannot collide with a different split of the same bytes.
func writeVerifierField(writer io.Writer, value string) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = writer.Write(length[:])
	_, _ = writer.Write([]byte(value))
}

// VerifierSession is the fresh runtime fact a session creator observed: the
// identity OpenCode minted and the model it reported for that session.
type VerifierSession struct {
	ID    domaingrant.SessionID
	Model string
}

// VerifierSessions creates one fresh runtime session per verifier process. It
// takes no session, continue, fork or resume value, so a caller cannot reopen an
// existing one.
type VerifierSessions interface {
	CreateFreshSession(ctx context.Context) (VerifierSession, error)
}

// VerifierGrants is the one-shot authority boundary the verifier consumes. The
// production value is the grant service; a test supplies a fake.
type VerifierGrants interface {
	Issue(ctx context.Context, binding domaingrant.Binding) (domaingrant.Grant, error)
	Claim(ctx context.Context, id domaingrant.ID) (domaingrant.Grant, error)
	BindSession(ctx context.Context, id domaingrant.ID, session domaingrant.SessionID) (domaingrant.Grant, error)
	Complete(ctx context.Context, id domaingrant.ID) (domaingrant.Grant, error)
	Fail(ctx context.Context, id domaingrant.ID) (domaingrant.Grant, error)
}

// VerifierExecutionRequest is what the verifier receives after its grant is
// claimed and its fresh session is bound. It carries only that fresh session:
// there is no resume, continue, fork or caller-supplied session field. Model is
// the model the grant binds; SessionModel is the model the runtime observed for
// the fresh session.
type VerifierExecutionRequest struct {
	Prompt       string
	Dir          string
	Tree         string
	Base         string
	Model        string
	SessionModel string
	Session      domaingrant.SessionID
	Verifier     string
}

// VerifierExecution is the verifier's answer.
type VerifierExecution struct {
	Output   string
	ExitCode int
}

// VerifierExecutor runs the verifier once inside the fresh session Go bound. It
// is the only edge that reaches a model; tests inject a fake and spawn nobody.
type VerifierExecutor interface {
	Execute(ctx context.Context, request VerifierExecutionRequest) (VerifierExecution, error)
}

// VerifierFreezer exports one candidate tree into a directory the verifier may
// read. It is the same freeze `review` uses, so both verbs read one candidate.
type VerifierFreezer interface {
	ArchiveRun(root, tree string) (run, candidate string, err error)
	Discard(dir string) error
}

// Snapshotter answers the content identity of a frozen directory. The managed
// verifier reads it before and after execution; a different answer means the
// bytes moved under the verifier and nothing may be sealed.
type Snapshotter interface {
	Digest(dir string) (string, error)
}

// Clock is the tool's clock. The verifier seals the instants it observed, never
// an hour the answer claims.
type Clock interface {
	Now() time.Time
}

// ManagedVerify is one managed verification of one frozen candidate: the tree
// and base every run binds, the runtime that creates sessions, the verifier that
// answers, and the prompt it reads.
type ManagedVerify struct {
	Dir      string
	Tree     string
	Base     string
	Runtime  string
	Model    string
	Verifier string
	Prompt   string
}

// Outcome is the terminal fact a sealed runtime receipt carries. A run that did
// not complete seals nothing, so the only sealed outcome is completed.
type Outcome string

// OutcomeCompleted says the verifier ran to a zero exit over a stable snapshot
// under one bound grant and session.
const OutcomeCompleted Outcome = "completed"

// SealedRuntimeReceipt is the tool-derived runtime fact of one completed
// verifier run. Every field is observed or issued by Go; none of them is read
// from the verifier's answer. It is not exported or persisted yet: validation
// and export belong to the next slice.
type SealedRuntimeReceipt struct {
	Grant        domaingrant.ID
	Session      domaingrant.SessionID
	Runtime      string
	Model        string
	SessionModel string
	Role         string
	Tree         string
	Base         string
	Snapshot     string
	StartedAt    string
	FinishedAt   string
	Outcome      Outcome
}

// ManagedVerifyResult is what a managed verifier caller reads. Managed, Fresh
// and Sealed are set only when one run issued, claimed and bound its own
// authority, executed over an unchanged snapshot and completed its grant; every
// refusal leaves all three clear.
type ManagedVerifyResult struct {
	Code     int
	Message  string
	Tree     string
	Base     string
	Output   string
	ExitCode int
	Managed  bool
	Fresh    bool
	Sealed   *SealedRuntimeReceipt
}

// ManagedVerifyDeps are the six boundaries a managed verifier needs: the
// freezer, the grant service, the fresh-session runtime, the verifier executor,
// the snapshot reader and the clock.
type ManagedVerifyDeps struct {
	Freezer   VerifierFreezer
	Grants    VerifierGrants
	Sessions  VerifierSessions
	Executor  VerifierExecutor
	Snapshots Snapshotter
	Clock     Clock
}

// validate refuses an unwired verifier before it freezes anything.
func (deps ManagedVerifyDeps) validate() error {
	switch {
	case deps.Freezer == nil:
		return errors.New("managed verify needs a freezer")
	case deps.Grants == nil:
		return errors.New("managed verify needs a grant service")
	case deps.Sessions == nil:
		return errors.New("managed verify needs a fresh-session runtime")
	case deps.Executor == nil:
		return errors.New("managed verify needs a verifier executor")
	case deps.Snapshots == nil:
		return errors.New("managed verify needs a snapshot reader")
	case deps.Clock == nil:
		return errors.New("managed verify needs a clock")
	default:
		return nil
	}
}

// validate refuses a verifier that cannot bind one process before the candidate
// is frozen.
func (request ManagedVerify) validate() error {
	switch {
	case strings.TrimSpace(request.Tree) == "":
		return errors.New("the managed verify names no candidate tree")
	case strings.TrimSpace(request.Base) == "":
		return errors.New("the managed verify names no candidate base")
	case strings.TrimSpace(request.Runtime) == "":
		return errors.New("the managed verify names no runtime")
	case strings.TrimSpace(request.Model) == "":
		return errors.New("the managed verify names no model")
	case strings.TrimSpace(request.Verifier) == "":
		return errors.New("the managed verify names no verifier")
	case strings.TrimSpace(request.Prompt) == "":
		return errors.New("the managed verify has no prompt")
	default:
		return nil
	}
}

// RunManagedVerify freezes one candidate tree and runs one verifier over it. The
// grant is issued and claimed before the fresh session is created, the session
// is bound before execution, and the snapshot is read again after execution. A
// refusal before the bind starts no verifier; a snapshot that moved, a verifier
// that failed or a session the runtime could not mint seals no evidence and
// terminates the grant.
//
// The executor's output is kept as the answer it is. It is never read for the
// grant, session, runtime, model, role, tree, base, snapshot, hours or outcome:
// those are the tool's facts, and a model's claim about them is ignored.
func RunManagedVerify(ctx context.Context, request ManagedVerify, deps ManagedVerifyDeps) ManagedVerifyResult {
	result := ManagedVerifyResult{Tree: request.Tree, Base: request.Base}
	if err := deps.validate(); err != nil {
		return blockedManagedVerify(result, err.Error())
	}
	if err := request.validate(); err != nil {
		return blockedManagedVerify(result, err.Error())
	}
	if request.Dir == "" {
		request.Dir = "."
	}

	binding, err := VerifierBinding(VerifierBindingFields{
		Tree:     request.Tree,
		Base:     request.Base,
		Runtime:  request.Runtime,
		Model:    request.Model,
		Verifier: request.Verifier,
	})
	if err != nil {
		return blockedManagedVerify(result, "could not bind the verifier input: "+oneLine(err.Error()))
	}

	// One freeze for the one process: the verifier reads a copy, never the live
	// worktree. The freeze comes after every reason not to start a verifier, so a
	// request nobody can pay for creates no directory.
	run, frozen, err := deps.Freezer.ArchiveRun(request.Dir, request.Tree)
	if err != nil {
		return blockedManagedVerify(result, "could not freeze the candidate for the verifier: "+oneLine(err.Error())+"; no verifier was started")
	}
	if strings.TrimSpace(run) == "" || strings.TrimSpace(frozen) == "" {
		return blockedManagedVerify(result, "the freeze named no candidate directory; no verifier was started")
	}
	defer func() { _ = deps.Freezer.Discard(run) }()

	before, err := deps.Snapshots.Digest(frozen)
	if err != nil {
		return blockedManagedVerify(result, "could not read the frozen candidate: "+oneLine(err.Error())+"; no verifier was started")
	}
	if strings.TrimSpace(before) == "" {
		return blockedManagedVerify(result, "the frozen candidate has no snapshot identity; no verifier was started")
	}

	issued, err := deps.Grants.Issue(ctx, binding)
	if err != nil {
		return blockedManagedVerify(result, "could not issue the verifier grant: "+oneLine(err.Error()))
	}
	if issued.State != domaingrant.StateIssued || issued.Binding != binding {
		return blockedManagedVerify(result, "the issued grant is not bound to the verifier input")
	}

	claimed, err := deps.Grants.Claim(ctx, issued.ID)
	if err != nil {
		// A consumed, missing, corrupt or foreign grant is a refusal: no session
		// is created for this run and nobody runs.
		return blockedManagedVerify(result, "could not claim the verifier grant: "+oneLine(err.Error()))
	}
	if claimed.ID != issued.ID || claimed.State != domaingrant.StateClaimed || claimed.Binding != binding {
		return blockedManagedVerify(result, "the claimed grant does not match the verifier input")
	}

	session, err := deps.Sessions.CreateFreshSession(ctx)
	if err != nil {
		failVerifierGrant(ctx, deps.Grants, issued.ID)
		return blockedManagedVerify(result, "could not create a fresh verifier session: "+oneLine(err.Error()))
	}
	if err := validateVerifierSession(session); err != nil {
		failVerifierGrant(ctx, deps.Grants, issued.ID)
		return blockedManagedVerify(result, "the runtime reported an unusable session for the verifier: "+oneLine(err.Error()))
	}
	// BindSession validates the session grammar before storage; doing it here makes
	// a malformed identity a refusal before any session marker exists.
	if _, err := claimed.BindSession(session.ID); err != nil {
		failVerifierGrant(ctx, deps.Grants, issued.ID)
		return blockedManagedVerify(result, "the runtime session is not bindable for the verifier: "+oneLine(err.Error()))
	}

	bound, err := deps.Grants.BindSession(ctx, issued.ID, session.ID)
	if err != nil {
		failVerifierGrant(ctx, deps.Grants, issued.ID)
		return blockedManagedVerify(result, "could not bind the fresh session for the verifier: "+oneLine(err.Error()))
	}
	if bound.ID != issued.ID || bound.State != domaingrant.StateBound || bound.Session != session.ID {
		failVerifierGrant(ctx, deps.Grants, issued.ID)
		return blockedManagedVerify(result, "the grant did not bind the fresh verifier session")
	}

	started := deps.Clock.Now()
	execution, err := deps.Executor.Execute(ctx, VerifierExecutionRequest{
		Prompt:       request.Prompt,
		Dir:          frozen,
		Tree:         request.Tree,
		Base:         request.Base,
		Model:        request.Model,
		SessionModel: session.Model,
		Session:      bound.Session,
		Verifier:     request.Verifier,
	})
	if err != nil {
		failVerifierGrant(ctx, deps.Grants, issued.ID)
		result.Output = execution.Output
		result.ExitCode = execution.ExitCode
		return blockedManagedVerify(result, "the verifier failed: "+oneLine(err.Error()))
	}
	if execution.ExitCode != 0 {
		failVerifierGrant(ctx, deps.Grants, issued.ID)
		result.Output = execution.Output
		result.ExitCode = execution.ExitCode
		return blockedManagedVerify(result, fmt.Sprintf("the verifier exited with code %d", execution.ExitCode))
	}

	after, err := deps.Snapshots.Digest(frozen)
	if err != nil {
		failVerifierGrant(ctx, deps.Grants, issued.ID)
		result.Output = execution.Output
		return blockedManagedVerify(result, "could not re-read the frozen candidate: "+oneLine(err.Error())+"; nothing was sealed")
	}
	if strings.TrimSpace(after) == "" || after != before {
		failVerifierGrant(ctx, deps.Grants, issued.ID)
		result.Output = execution.Output
		return blockedManagedVerify(result, "the frozen candidate changed under the verifier: nothing was sealed")
	}
	finished := deps.Clock.Now()

	completed, err := deps.Grants.Complete(ctx, issued.ID)
	if err != nil {
		failVerifierGrant(ctx, deps.Grants, issued.ID)
		return blockedManagedVerify(result, "the verifier finished but the grant did not complete: "+oneLine(err.Error()))
	}
	if completed.ID != issued.ID || completed.State != domaingrant.StateCompleted {
		failVerifierGrant(ctx, deps.Grants, issued.ID)
		return blockedManagedVerify(result, "the grant did not complete the verifier")
	}

	sealed := &SealedRuntimeReceipt{
		Grant:        issued.ID,
		Session:      bound.Session,
		Runtime:      request.Runtime,
		Model:        request.Model,
		SessionModel: session.Model,
		Role:         VerifierRole,
		Tree:         request.Tree,
		Base:         request.Base,
		Snapshot:     before,
		StartedAt:    started.UTC().Format(time.RFC3339),
		FinishedAt:   finished.UTC().Format(time.RFC3339),
		Outcome:      OutcomeCompleted,
	}
	result.Code = CodeOK
	result.Message = fmt.Sprintf("verified %s over snapshot %s", request.Tree, before)
	result.Output = execution.Output
	result.ExitCode = execution.ExitCode
	result.Managed = true
	result.Fresh = true
	result.Sealed = sealed
	return result
}

// blockedManagedVerify is the refusal a managed verify returns. It clears every
// freshness and sealed claim, so a refusal never carries evidence it did not
// earn.
func blockedManagedVerify(result ManagedVerifyResult, message string) ManagedVerifyResult {
	result.Code = CodeBlocked
	result.Message = message
	result.Managed = false
	result.Fresh = false
	result.Sealed = nil
	return result
}

// validateVerifierSession refuses an identity the verifier cannot bind: a
// missing identity or a missing observed model. The grammar is validated by the
// domain through BindSession; duplicate and foreign identities are refused by
// the runtime port before it returns one.
func validateVerifierSession(session VerifierSession) error {
	if session.ID == "" {
		return fmt.Errorf("%w: the fresh verifier session identity is missing", domaingrant.ErrInvalidSession)
	}
	if strings.TrimSpace(session.Model) == "" {
		return errors.New("the observed verifier model is missing")
	}
	return nil
}

// failVerifierGrant terminates a claimed or bound grant after the run cannot
// continue. It uses a detached context: a canceled run must not leave a claimed
// grant behind. A grant that is already terminal refuses the write, and that
// refusal is deliberately not surfaced over the run's own result.
func failVerifierGrant(ctx context.Context, grants VerifierGrants, id domaingrant.ID) {
	_, _ = grants.Fail(context.WithoutCancel(ctx), id)
}

// LegacyVerifierExecutor is the unmanaged `--exec` boundary: one command that
// receives a prompt and runs no grant machinery.
type LegacyVerifierExecutor interface {
	Run(ctx context.Context, command, prompt string) (VerifierExecution, error)
}

// RunUnmanagedVerify is the legacy `--exec` path: it runs the command without
// freezing a candidate, issuing a grant or creating a session, so the result is
// explicitly unmanaged and carries no freshness or sealed-runtime claim.
func RunUnmanagedVerify(ctx context.Context, command, prompt string, executor LegacyVerifierExecutor) ManagedVerifyResult {
	if strings.TrimSpace(command) == "" {
		return blockedManagedVerify(ManagedVerifyResult{}, "the unmanaged verify names no verifier command")
	}
	if executor == nil {
		return blockedManagedVerify(ManagedVerifyResult{}, "the unmanaged verify needs a verifier executor")
	}
	if strings.TrimSpace(prompt) == "" {
		return blockedManagedVerify(ManagedVerifyResult{}, "the unmanaged verify has no prompt")
	}

	execution, err := executor.Run(ctx, command, prompt)
	if err != nil {
		result := blockedManagedVerify(ManagedVerifyResult{}, "the unmanaged verifier failed: "+oneLine(err.Error()))
		result.Output = execution.Output
		result.ExitCode = execution.ExitCode
		return result
	}
	if execution.ExitCode != 0 {
		result := blockedManagedVerify(ManagedVerifyResult{}, fmt.Sprintf("the unmanaged verifier exited with code %d", execution.ExitCode))
		result.Output = execution.Output
		result.ExitCode = execution.ExitCode
		return result
	}
	return ManagedVerifyResult{
		Code:     CodeOK,
		Message:  fmt.Sprintf("the unmanaged verifier ran %q without a grant or a session", command),
		Output:   execution.Output,
		ExitCode: execution.ExitCode,
	}
}
