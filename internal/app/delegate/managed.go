package delegate

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	domaingrant "github.com/biggs-100/odachi/internal/domain/grant"
)

// bindingDomain separates the canonical delegation encoding from any other use
// of the same digest.
const bindingDomain = "odachi/delegate/v1"

// BindingFields is the canonical delegation input a managed grant is bound to:
// the task path and digest, the step, the requirement it cites, the agent, and
// the surfaces and commands the worker receives.
type BindingFields struct {
	TaskPath    string
	TaskDigest  [sha256.Size]byte
	Step        string
	Requirement string
	Agent       string
	Surfaces    []string
	Commands    []string
}

// Binding is the canonical SHA-256 binding of one delegation input. Every field
// is length-prefixed and the two collections are sorted, so the same input
// always produces the same digest and a different field produces a different
// one. It refuses an input that names no task, digest or step.
func Binding(fields BindingFields) (domaingrant.Binding, error) {
	if strings.TrimSpace(fields.TaskPath) == "" {
		return domaingrant.Binding{}, fmt.Errorf("%w: the task path is missing", domaingrant.ErrInvalidBinding)
	}
	if fields.TaskDigest == ([sha256.Size]byte{}) {
		return domaingrant.Binding{}, fmt.Errorf("%w: the task digest is missing", domaingrant.ErrInvalidBinding)
	}
	if strings.TrimSpace(fields.Step) == "" {
		return domaingrant.Binding{}, fmt.Errorf("%w: the step is missing", domaingrant.ErrInvalidBinding)
	}

	hasher := sha256.New()
	writeField(hasher, bindingDomain)
	writeField(hasher, fields.TaskPath)
	writeField(hasher, hex.EncodeToString(fields.TaskDigest[:]))
	writeField(hasher, fields.Step)
	writeField(hasher, fields.Requirement)
	writeField(hasher, fields.Agent)
	writeList(hasher, fields.Surfaces)
	writeList(hasher, fields.Commands)

	var binding domaingrant.Binding
	copy(binding[:], hasher.Sum(nil))
	if err := binding.Validate(); err != nil {
		return domaingrant.Binding{}, err
	}
	return binding, nil
}

// writeField writes one length-prefixed field, so a value containing the
// separator cannot collide with a different split of the same bytes.
func writeField(writer io.Writer, value string) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = writer.Write(length[:])
	_, _ = writer.Write([]byte(value))
}

// writeList writes a count followed by the sorted, length-prefixed elements.
// Surfaces and commands are sets: the same set binds the same digest whatever
// order the parent typed it in.
func writeList(writer io.Writer, values []string) {
	canonical := slices.Clone(values)
	slices.Sort(canonical)
	var count [8]byte
	binary.BigEndian.PutUint64(count[:], uint64(len(canonical)))
	_, _ = writer.Write(count[:])
	for _, value := range canonical {
		writeField(writer, value)
	}
}

// Grants is the one-shot authority boundary the managed attempt consumes. The
// production value is the grant service; a test supplies a fake.
type Grants interface {
	Issue(ctx context.Context, binding domaingrant.Binding) (domaingrant.Grant, error)
	Claim(ctx context.Context, id domaingrant.ID) (domaingrant.Grant, error)
	BindSession(ctx context.Context, id domaingrant.ID, session domaingrant.SessionID) (domaingrant.Grant, error)
	Complete(ctx context.Context, id domaingrant.ID) (domaingrant.Grant, error)
	Fail(ctx context.Context, id domaingrant.ID) (domaingrant.Grant, error)
}

// Session is the fresh runtime fact a session creator observed: the identity
// OpenCode minted and the model it reported for that session.
type Session struct {
	ID    domaingrant.SessionID
	Model string
}

// Sessions creates one fresh runtime session per attempt. It takes no session,
// continue, fork or resume value, so a caller cannot reopen an existing one.
type Sessions interface {
	CreateFreshSession(ctx context.Context) (Session, error)
}

// ExecutionRequest is what one managed worker receives after the grant is
// claimed and the fresh session is bound. It carries only that fresh session:
// there is no resume, continue, fork or caller-supplied session field.
type ExecutionRequest struct {
	Prompt  string
	Dir     string
	TaskDoc string
	Step    string
	Session domaingrant.SessionID
	Model   string
}

// Execution is the worker's answer.
type Execution struct {
	Output   string
	ExitCode int
}

// Executor runs the worker once inside the fresh session Go bound. It is the
// only edge that reaches the model; tests inject a fake and spawn nobody.
type Executor interface {
	Execute(ctx context.Context, request ExecutionRequest) (Execution, error)
}

// ManagedAttempt is the evidence one managed worker attempt leaves: the grant,
// the canonical binding, the fresh session and the observed model.
type ManagedAttempt struct {
	Grant   domaingrant.ID
	Binding domaingrant.Binding
	Session domaingrant.SessionID
	Model   string
	Managed bool
	Fresh   bool
}

// ManagedResult is what a managed or legacy caller reads. Managed and Fresh are
// true only when this attempt issued, claimed and bound its own authority; the
// legacy path and every refusal keep them false.
type ManagedResult struct {
	Code     int
	Message  string
	Prompt   string
	Task     string
	Output   string
	ExitCode int
	Managed  bool
	Fresh    bool
	Attempt  *ManagedAttempt
}

// ManagedDeps are the three boundaries a managed attempt needs: the grant
// service, the fresh-session runtime and the worker executor.
type ManagedDeps struct {
	Grants   Grants
	Sessions Sessions
	Executor Executor
}

// validate refuses an unwired attempt before it issues any authority.
func (deps ManagedDeps) validate() error {
	switch {
	case deps.Grants == nil:
		return errors.New("managed delegation needs a grant service")
	case deps.Sessions == nil:
		return errors.New("managed delegation needs a fresh-session runtime")
	case deps.Executor == nil:
		return errors.New("managed delegation needs an executor")
	default:
		return nil
	}
}

// RunManaged is the managed delegation boundary. It composes the prompt, binds
// the canonical delegation input into one grant, consumes that grant exactly
// once, creates one fresh session, binds the observed identity and model to the
// grant, and only then runs the worker. A refusal before the bind runs nobody.
// Every call is a new attempt: a correction issues its own grant and creates its
// own session, and a replayed or resumed authority is refused.
func RunManaged(ctx context.Context, opts Options, files Files, deps ManagedDeps) ManagedResult {
	delegation, result := prepare(opts, files)
	if result.Code != CodeOK {
		return ManagedResult{Code: result.Code, Message: result.Message, Task: result.Task}
	}
	if err := deps.validate(); err != nil {
		return blockedManaged(err.Error())
	}

	binding, err := Binding(BindingFields{
		TaskPath:    delegation.taskPath,
		TaskDigest:  delegation.taskDigest,
		Step:        delegation.step.ID,
		Requirement: delegation.step.Requirement,
		Agent:       delegation.step.Agent,
		Surfaces:    opts.Surfaces,
		Commands:    opts.Commands,
	})
	if err != nil {
		return blockedManaged("could not bind the delegation input: " + oneLine(err.Error()))
	}

	issued, err := deps.Grants.Issue(ctx, binding)
	if err != nil {
		return blockedManaged("could not issue the delegation grant: " + oneLine(err.Error()))
	}
	if issued.State != domaingrant.StateIssued || issued.Binding != binding {
		return blockedManaged("the issued grant is not bound to the delegation input")
	}

	claimed, err := deps.Grants.Claim(ctx, issued.ID)
	if err != nil {
		// A consumed, missing, corrupt or foreign grant is a refusal: no
		// session is created and nobody runs.
		return blockedManaged("could not claim the delegation grant: " + oneLine(err.Error()))
	}
	if claimed.ID != issued.ID || claimed.State != domaingrant.StateClaimed || claimed.Binding != binding {
		return blockedManaged("the claimed grant does not match the issued grant")
	}

	session, err := deps.Sessions.CreateFreshSession(ctx)
	if err != nil {
		failGrant(ctx, deps.Grants, issued.ID)
		return blockedManaged("could not create a fresh worker session: " + oneLine(err.Error()))
	}
	if err := validateFreshSession(session); err != nil {
		failGrant(ctx, deps.Grants, issued.ID)
		return blockedManaged("the runtime reported an unusable session: " + oneLine(err.Error()))
	}
	// BindSession validates the session grammar before storage; doing it here
	// makes a malformed identity a refusal before any marker is published.
	if _, err := claimed.BindSession(session.ID); err != nil {
		failGrant(ctx, deps.Grants, issued.ID)
		return blockedManaged("the runtime session is not bindable: " + oneLine(err.Error()))
	}

	bound, err := deps.Grants.BindSession(ctx, issued.ID, session.ID)
	if err != nil {
		failGrant(ctx, deps.Grants, issued.ID)
		return blockedManaged("could not bind the fresh session to the delegation grant: " + oneLine(err.Error()))
	}
	if bound.ID != issued.ID || bound.Session != session.ID || bound.State != domaingrant.StateBound {
		failGrant(ctx, deps.Grants, issued.ID)
		return blockedManaged("the grant did not bind the fresh session")
	}

	// The grant is issued, claimed and session-bound from here on: the fresh
	// attempt is real evidence even if the worker later fails.
	attempt := &ManagedAttempt{
		Grant:   issued.ID,
		Binding: binding,
		Session: bound.Session,
		Model:   session.Model,
		Managed: true,
		Fresh:   true,
	}

	execution, err := deps.Executor.Execute(ctx, ExecutionRequest{
		Prompt:  delegation.prompt,
		Dir:     opts.Dir,
		TaskDoc: delegation.taskPath,
		Step:    delegation.step.ID,
		Session: bound.Session,
		Model:   session.Model,
	})
	if err != nil {
		failGrant(ctx, deps.Grants, issued.ID)
		return blockedAttempt(delegation.prompt, opts.TaskDoc, execution, attempt,
			"the managed worker failed: "+oneLine(err.Error()))
	}
	// The real shell runner reports a non-zero exit with a nil error. A worker
	// that did not succeed is never a completion.
	if execution.ExitCode != 0 {
		failGrant(ctx, deps.Grants, issued.ID)
		return blockedAttempt(delegation.prompt, opts.TaskDoc, execution, attempt,
			fmt.Sprintf("the managed worker exited with code %d", execution.ExitCode))
	}

	if _, err := deps.Grants.Complete(ctx, issued.ID); err != nil {
		failGrant(ctx, deps.Grants, issued.ID)
		return blockedManaged("the worker finished but the grant did not complete: " + oneLine(err.Error()))
	}

	return ManagedResult{
		Code:     CodeOK,
		Prompt:   delegation.prompt,
		Task:     opts.TaskDoc,
		Output:   execution.Output,
		ExitCode: execution.ExitCode,
		Managed:  true,
		Fresh:    true,
		Attempt:  attempt,
	}
}

// blockedAttempt is the evidence a managed worker leaves after a post-bind
// failure. The grant was issued, claimed and session-bound, so the fresh attempt
// is preserved even though the result is blocked and claims no success.
func blockedAttempt(prompt, task string, execution Execution, attempt *ManagedAttempt, message string) ManagedResult {
	return ManagedResult{
		Code:     CodeBlocked,
		Message:  message,
		Prompt:   prompt,
		Task:     task,
		Output:   execution.Output,
		ExitCode: execution.ExitCode,
		Managed:  true,
		Fresh:    true,
		Attempt:  attempt,
	}
}

// validateFreshSession refuses an identity the attempt cannot bind: a missing
// identity or a missing observed model. The grammar is validated by the domain
// through BindSession; foreign and duplicate identities are refused by the
// runtime port before it returns one.
func validateFreshSession(session Session) error {
	if session.ID == "" {
		return fmt.Errorf("%w: the fresh session identity is missing", domaingrant.ErrInvalidSession)
	}
	if strings.TrimSpace(session.Model) == "" {
		return errors.New("the observed model is missing")
	}
	return nil
}

// failGrant terminates a claimed or bound grant after the attempt cannot
// continue. It uses a detached context: a canceled attempt must not leave a
// claimed grant behind for a later worker. A grant that is already terminal
// refuses the write, and that refusal is deliberately not surfaced over the
// attempt's own result.
func failGrant(ctx context.Context, grants Grants, id domaingrant.ID) {
	_, _ = grants.Fail(context.WithoutCancel(ctx), id)
}

// blockedManaged is the refusal a managed or legacy attempt returns.
func blockedManaged(message string) ManagedResult {
	return ManagedResult{Code: CodeBlocked, Message: message}
}

// LegacyExecutor is the unmanaged `--exec` boundary: one command that receives
// the composed prompt and runs no grant machinery.
type LegacyExecutor interface {
	Run(ctx context.Context, command, prompt string) (Execution, error)
}

// RunUnmanaged is the legacy `--exec` path. It composes the same prompt and
// hands it to the command without issuing a grant or creating a session. Its
// result is explicitly unmanaged and carries no freshness claim.
func RunUnmanaged(ctx context.Context, opts Options, files Files, command string, executor LegacyExecutor) ManagedResult {
	delegation, result := prepare(opts, files)
	if result.Code != CodeOK {
		return ManagedResult{Code: result.Code, Message: result.Message, Task: result.Task}
	}
	if executor == nil {
		return blockedManaged("the unmanaged delegation needs an executor")
	}
	execution, err := executor.Run(ctx, command, delegation.prompt)
	if err != nil {
		return ManagedResult{
			Code:     CodeBlocked,
			Message:  "the unmanaged worker failed: " + oneLine(err.Error()),
			Prompt:   delegation.prompt,
			Task:     opts.TaskDoc,
			Output:   execution.Output,
			ExitCode: execution.ExitCode,
		}
	}
	if execution.ExitCode != 0 {
		return ManagedResult{
			Code:     CodeBlocked,
			Message:  fmt.Sprintf("the unmanaged worker exited with code %d", execution.ExitCode),
			Prompt:   delegation.prompt,
			Task:     opts.TaskDoc,
			Output:   execution.Output,
			ExitCode: execution.ExitCode,
		}
	}
	return ManagedResult{
		Code:     CodeOK,
		Prompt:   delegation.prompt,
		Task:     opts.TaskDoc,
		Output:   execution.Output,
		ExitCode: execution.ExitCode,
	}
}
