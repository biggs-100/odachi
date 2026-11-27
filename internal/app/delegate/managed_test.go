package delegate_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/delegate"
	domaingrant "github.com/biggs-100/odachi/internal/domain/grant"
)

// eventLog is the shared, ordered record of what the fake boundaries observed.
// The managed attempt is sequential, so a plain slice is enough under -race.
type eventLog struct {
	events []string
}

func (log *eventLog) add(event string) { log.events = append(log.events, event) }

func (log *eventLog) snapshot() []string { return slices.Clone(log.events) }

// fakeGrants is the one-shot authority boundary. It keeps the claim tombstone
// and can replay the previous issued grant to prove a second attempt cannot
// consume it.
type fakeGrants struct {
	log           *eventLog
	claims        map[string]bool
	issued        domaingrant.Grant
	claimed       domaingrant.Grant
	bound         domaingrant.Grant
	issueErr      error
	claimErr      error
	bindErr       error
	completeErr   error
	failErr       error
	replay        bool
	bindings      []domaingrant.Binding
	completeCalls int
	failCalls     int
}

func newFakeGrants(log *eventLog) *fakeGrants {
	return &fakeGrants{log: log, claims: make(map[string]bool)}
}

func (fake *fakeGrants) Issue(_ context.Context, binding domaingrant.Binding) (domaingrant.Grant, error) {
	fake.log.add("issue")
	fake.bindings = append(fake.bindings, binding)
	if fake.issueErr != nil {
		return domaingrant.Grant{}, fake.issueErr
	}
	if fake.replay {
		return fake.issued, nil
	}
	issued, err := domaingrant.NewGrant(binding)
	if err != nil {
		return domaingrant.Grant{}, err
	}
	fake.issued = issued
	return issued, nil
}

func (fake *fakeGrants) Claim(_ context.Context, id domaingrant.ID) (domaingrant.Grant, error) {
	fake.log.add("claim")
	if fake.claimErr != nil {
		return domaingrant.Grant{}, fake.claimErr
	}
	if fake.claims[id.String()] {
		return domaingrant.Grant{}, domaingrant.ErrAlreadyClaimed
	}
	if fake.issued.ID != id {
		return domaingrant.Grant{}, errors.New("the claim names a foreign grant")
	}
	claimed, err := fake.issued.Claim()
	if err != nil {
		return domaingrant.Grant{}, err
	}
	fake.claims[id.String()] = true
	fake.claimed = claimed
	return claimed, nil
}

func (fake *fakeGrants) BindSession(_ context.Context, _ domaingrant.ID, session domaingrant.SessionID) (domaingrant.Grant, error) {
	fake.log.add("bind")
	if fake.bindErr != nil {
		return domaingrant.Grant{}, fake.bindErr
	}
	bound, err := fake.claimed.BindSession(session)
	if err != nil {
		return domaingrant.Grant{}, err
	}
	fake.bound = bound
	return bound, nil
}

func (fake *fakeGrants) Complete(_ context.Context, _ domaingrant.ID) (domaingrant.Grant, error) {
	fake.log.add("complete")
	fake.completeCalls++
	if fake.completeErr != nil {
		return domaingrant.Grant{}, fake.completeErr
	}
	return fake.bound.Complete()
}

func (fake *fakeGrants) Fail(_ context.Context, _ domaingrant.ID) (domaingrant.Grant, error) {
	fake.log.add("fail")
	fake.failCalls++
	if fake.failErr != nil {
		return domaingrant.Grant{}, fake.failErr
	}
	current := fake.bound
	if current.State == domaingrant.StateIssued {
		current = fake.claimed
	}
	if current.State == domaingrant.StateIssued {
		return domaingrant.Grant{}, domaingrant.ErrInvalidTransition
	}
	return current.Fail()
}

// fakeSession is one canned fresh-session answer.
type fakeSession struct {
	session delegate.Session
}

// fakeSessions is the fresh runtime boundary. It takes no session value, and it
// refuses a repeated identity the way the T6 adapter does.
type fakeSessions struct {
	log       *eventLog
	calls     int
	seen      map[domaingrant.SessionID]bool
	responses []fakeSession
	err       error
	created   []delegate.Session
}

func newFakeSessions(log *eventLog) *fakeSessions {
	return &fakeSessions{log: log, seen: make(map[domaingrant.SessionID]bool)}
}

func (fake *fakeSessions) CreateFreshSession(_ context.Context) (delegate.Session, error) {
	fake.log.add("create-session")
	fake.calls++
	if fake.err != nil {
		return delegate.Session{}, fake.err
	}
	var session delegate.Session
	if fake.calls-1 < len(fake.responses) {
		session = fake.responses[fake.calls-1].session
	} else {
		session = delegate.Session{
			ID:    domaingrant.SessionID("ses_fresh_" + strconv.Itoa(fake.calls)),
			Model: "anthropic/claude-sonnet-4",
		}
	}
	if session.ID != "" {
		if fake.seen[session.ID] {
			return delegate.Session{}, errors.New("duplicate session identity")
		}
		fake.seen[session.ID] = true
	}
	fake.created = append(fake.created, session)
	return session, nil
}

// fakeExecutor records the one managed execution it receives.
type fakeExecutor struct {
	log      *eventLog
	calls    int
	requests []delegate.ExecutionRequest
	err      error
	output   string
	exitCode int
}

func (fake *fakeExecutor) Execute(_ context.Context, request delegate.ExecutionRequest) (delegate.Execution, error) {
	fake.log.add("execute")
	fake.calls++
	fake.requests = append(fake.requests, request)
	if fake.err != nil {
		return delegate.Execution{Output: fake.output, ExitCode: fake.exitCode}, fake.err
	}
	return delegate.Execution{Output: fake.output, ExitCode: fake.exitCode}, nil
}

// fakeLegacy is the unmanaged `--exec` boundary: a command and a prompt, with
// no grant and no session.
type fakeLegacy struct {
	calls    int
	command  string
	prompt   string
	output   string
	exitCode int
	err      error
}

func (fake *fakeLegacy) Run(_ context.Context, command, prompt string) (delegate.Execution, error) {
	fake.calls++
	fake.command = command
	fake.prompt = prompt
	if fake.err != nil {
		return delegate.Execution{Output: fake.output, ExitCode: fake.exitCode}, fake.err
	}
	return delegate.Execution{Output: fake.output, ExitCode: fake.exitCode}, nil
}

// managedDeps wires the three fake boundaries over one event log.
func managedDeps(log *eventLog) (delegate.ManagedDeps, *fakeGrants, *fakeSessions, *fakeExecutor) {
	grants := newFakeGrants(log)
	sessions := newFakeSessions(log)
	executor := &fakeExecutor{log: log, output: "worker-output"}
	return delegate.ManagedDeps{Grants: grants, Sessions: sessions, Executor: executor}, grants, sessions, executor
}

// TestManagedAttemptIssuesClaimsBindsThenExecutesAndCompletes: the grant is
// consumed and the fresh session is bound before the worker runs; only then is
// the grant completed with the observed session, model and output named.
func TestManagedAttemptIssuesClaimsBindsThenExecutesAndCompletes(t *testing.T) {
	log := &eventLog{}
	opts, files := setup(t)
	deps, grants, sessions, executor := managedDeps(log)

	result := delegate.RunManaged(t.Context(), opts, files, deps)
	if result.Code != delegate.CodeOK {
		t.Fatalf("RunManaged: %s", result.Message)
	}
	if !result.Managed || !result.Fresh {
		t.Fatalf("managed/fresh = %v/%v, want true/true", result.Managed, result.Fresh)
	}
	if result.Attempt == nil {
		t.Fatal("a managed attempt left no evidence")
	}
	if result.Output != "worker-output" {
		t.Fatalf("output = %q, want worker-output", result.Output)
	}

	want := []string{"issue", "claim", "create-session", "bind", "execute", "complete"}
	if got := log.snapshot(); !slices.Equal(got, want) {
		t.Fatalf("boundary order = %v, want %v", got, want)
	}

	if len(sessions.created) != 1 || len(executor.requests) != 1 {
		t.Fatalf("sessions = %d, executions = %d, want 1 and 1", len(sessions.created), len(executor.requests))
	}
	session := sessions.created[0]
	request := executor.requests[0]
	if request.Session != session.ID {
		t.Fatalf("executor session = %q, want the fresh bound %q", request.Session, session.ID)
	}
	if request.Model != session.Model {
		t.Fatalf("executor model = %q, want the observed %q", request.Model, session.Model)
	}
	if result.Attempt.Grant != grants.issued.ID {
		t.Fatalf("attempt grant = %s, want %s", result.Attempt.Grant, grants.issued.ID)
	}
	if result.Attempt.Session != session.ID {
		t.Fatalf("attempt session = %q, want %q", result.Attempt.Session, session.ID)
	}
	if result.Attempt.Model != session.Model {
		t.Fatalf("attempt model = %q, want %q", result.Attempt.Model, session.Model)
	}
}

// TestManagedAttemptBindsTheCanonicalDelegationInput: the grant binding is the
// exact digest of the task path, task digest, step, requirement, agent,
// surfaces and commands the worker receives.
func TestManagedAttemptBindsTheCanonicalDelegationInput(t *testing.T) {
	log := &eventLog{}
	opts, files := setup(t)
	deps, grants, _, _ := managedDeps(log)

	if result := delegate.RunManaged(t.Context(), opts, files, deps); result.Code != delegate.CodeOK {
		t.Fatalf("RunManaged: %s", result.Message)
	}

	taskPath, err := filepath.Abs(opts.TaskDoc)
	if err != nil {
		t.Fatalf("Abs: %v", err)
	}
	want, err := delegate.Binding(delegate.BindingFields{
		TaskPath:    taskPath,
		TaskDigest:  sha256.Sum256(files[opts.TaskDoc]),
		Step:        "T1",
		Requirement: "R1",
		Agent:       "odachi-worker",
		Surfaces:    opts.Surfaces,
		Commands:    opts.Commands,
	})
	if err != nil {
		t.Fatalf("Binding: %v", err)
	}
	if len(grants.bindings) != 1 {
		t.Fatalf("issued bindings = %d, want 1", len(grants.bindings))
	}
	if grants.bindings[0] != want {
		t.Fatalf("grant binding = %s, want %s", grants.bindings[0], want)
	}
}

// TestBindingChangesWithEveryCanonicalField: no canonical field is silently
// dropped, and the two collections are canonical sets: reordering them keeps
// the same digest.
func TestBindingChangesWithEveryCanonicalField(t *testing.T) {
	base := delegate.BindingFields{
		TaskPath:    `C:\work\task.md`,
		TaskDigest:  sha256.Sum256([]byte("doc")),
		Step:        "T1",
		Requirement: "R1",
		Agent:       "odachi-worker",
		Surfaces:    []string{"internal/a.go", "internal/b.go"},
		Commands:    []string{"go test ./..."},
	}
	baseline, err := delegate.Binding(base)
	if err != nil {
		t.Fatalf("Binding: %v", err)
	}

	change := func(mutate func(*delegate.BindingFields)) delegate.BindingFields {
		fields := base
		fields.Surfaces = slices.Clone(base.Surfaces)
		fields.Commands = slices.Clone(base.Commands)
		mutate(&fields)
		return fields
	}
	variants := map[string]delegate.BindingFields{
		"task path":   change(func(f *delegate.BindingFields) { f.TaskPath = `C:\work\other.md` }),
		"task digest": change(func(f *delegate.BindingFields) { f.TaskDigest = sha256.Sum256([]byte("other")) }),
		"step":        change(func(f *delegate.BindingFields) { f.Step = "T2" }),
		"requirement": change(func(f *delegate.BindingFields) { f.Requirement = "R2" }),
		"agent":       change(func(f *delegate.BindingFields) { f.Agent = "other-worker" }),
		"surfaces":    change(func(f *delegate.BindingFields) { f.Surfaces = []string{"internal/a.go"} }),
		"commands":    change(func(f *delegate.BindingFields) { f.Commands = []string{"go vet ./..."} }),
	}
	for name, fields := range variants {
		t.Run(name, func(t *testing.T) {
			got, err := delegate.Binding(fields)
			if err != nil {
				t.Fatalf("Binding: %v", err)
			}
			if got == baseline {
				t.Fatalf("changing %s did not change the binding", name)
			}
		})
	}

	reordered := change(func(f *delegate.BindingFields) {
		f.Surfaces = []string{"internal/b.go", "internal/a.go"}
	})
	got, err := delegate.Binding(reordered)
	if err != nil {
		t.Fatalf("Binding: %v", err)
	}
	if got != baseline {
		t.Fatalf("reordering a canonical set changed the binding")
	}
}

// TestEveryManagedAttemptUsesADistinctGrantAndFreshSession: a correction is
// another attempt, so it issues its own grant and creates its own session.
func TestEveryManagedAttemptUsesADistinctGrantAndFreshSession(t *testing.T) {
	log := &eventLog{}
	opts, files := setup(t)
	deps, grants, sessions, executor := managedDeps(log)

	first := delegate.RunManaged(t.Context(), opts, files, deps)
	second := delegate.RunManaged(t.Context(), opts, files, deps)
	if first.Code != delegate.CodeOK || second.Code != delegate.CodeOK {
		t.Fatalf("attempts = %d/%d: %s | %s", first.Code, second.Code, first.Message, second.Message)
	}
	if first.Attempt == nil || second.Attempt == nil {
		t.Fatal("an attempt left no evidence")
	}
	if first.Attempt.Grant == second.Attempt.Grant {
		t.Fatalf("both attempts used grant %s", first.Attempt.Grant)
	}
	if first.Attempt.Session == second.Attempt.Session {
		t.Fatalf("both attempts used session %q", first.Attempt.Session)
	}
	if sessions.calls != 2 {
		t.Fatalf("session creations = %d, want 2", sessions.calls)
	}
	if executor.calls != 2 {
		t.Fatalf("executions = %d, want 2", executor.calls)
	}
	if len(grants.bindings) != 2 || grants.bindings[0] != grants.bindings[1] {
		t.Fatalf("the two attempts did not bind the same delegation input")
	}
}

// TestManagedAttemptRefusesAReplayedGrantBeforeCreatingASession: a grant that
// was already consumed cannot start a second worker, and the refusal claims no
// freshness.
func TestManagedAttemptRefusesAReplayedGrantBeforeCreatingASession(t *testing.T) {
	log := &eventLog{}
	opts, files := setup(t)
	deps, grants, sessions, executor := managedDeps(log)

	if result := delegate.RunManaged(t.Context(), opts, files, deps); result.Code != delegate.CodeOK {
		t.Fatalf("first RunManaged: %s", result.Message)
	}
	grants.replay = true
	result := delegate.RunManaged(t.Context(), opts, files, deps)
	if result.Code != delegate.CodeBlocked {
		t.Fatalf("replay was not refused: %+v", result)
	}
	if result.Managed || result.Fresh || result.Attempt != nil {
		t.Fatal("a refused replay claimed freshness")
	}
	if sessions.calls != 1 {
		t.Fatalf("session creations = %d, want 1", sessions.calls)
	}
	if executor.calls != 1 {
		t.Fatalf("executions = %d, want 1", executor.calls)
	}
	if !strings.Contains(result.Message, "claim") {
		t.Fatalf("message %q does not name the refused claim", result.Message)
	}
}

// TestManagedAttemptFailsClosedWhenTheFreshSessionIsUnusable: a missing,
// malformed, duplicate, foreign, empty-id or model-less identity stops before
// execution and terminates the claimed grant.
func TestManagedAttemptFailsClosedWhenTheFreshSessionIsUnusable(t *testing.T) {
	cases := map[string]func(*fakeSessions){
		"missing":     func(s *fakeSessions) { s.err = errors.New("session identity is missing") },
		"malformed":   func(s *fakeSessions) { s.err = errors.New("session identity is malformed") },
		"duplicate":   func(s *fakeSessions) { s.err = errors.New("duplicate session identity") },
		"foreign":     func(s *fakeSessions) { s.err = errors.New("session identity is not fresh") },
		"empty id":    func(s *fakeSessions) { s.responses = []fakeSession{{session: delegate.Session{Model: "m"}}} },
		"empty model": func(s *fakeSessions) { s.responses = []fakeSession{{session: delegate.Session{ID: "ses_x"}}} },
	}
	for name, configure := range cases {
		t.Run(name, func(t *testing.T) {
			log := &eventLog{}
			opts, files := setup(t)
			deps, grants, sessions, executor := managedDeps(log)
			configure(sessions)

			result := delegate.RunManaged(t.Context(), opts, files, deps)
			if result.Code != delegate.CodeBlocked {
				t.Fatalf("an unusable session produced %+v", result)
			}
			if result.Managed || result.Fresh || result.Attempt != nil {
				t.Fatal("an unusable session claimed freshness")
			}
			if executor.calls != 0 {
				t.Fatalf("the worker ran %d times, want 0", executor.calls)
			}
			if grants.failCalls != 1 {
				t.Fatalf("grant failures = %d, want 1", grants.failCalls)
			}
		})
	}
}

// TestManagedAttemptDistinguishesPreExecutionFromPostBindFailures: a refused bind
// is a pre-execution refusal and claims no freshness, while a worker failure after
// the session is bound is real evidence of a fresh attempt even though it is
// blocked.
func TestManagedAttemptDistinguishesPreExecutionFromPostBindFailures(t *testing.T) {
	cases := map[string]struct {
		configure    func(*fakeGrants, *fakeExecutor)
		freshAttempt bool
		executions   int
	}{
		"bind": {
			configure:  func(g *fakeGrants, _ *fakeExecutor) { g.bindErr = errors.New("bind refused") },
			executions: 0,
		},
		"execute": {
			configure:    func(_ *fakeGrants, e *fakeExecutor) { e.err = errors.New("worker crashed") },
			freshAttempt: true,
			executions:   1,
		},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			log := &eventLog{}
			opts, files := setup(t)
			deps, grants, sessions, executor := managedDeps(log)
			testCase.configure(grants, executor)

			result := delegate.RunManaged(t.Context(), opts, files, deps)
			if result.Code != delegate.CodeBlocked {
				t.Fatalf("a failed attempt produced %+v", result)
			}
			if result.Managed != testCase.freshAttempt || result.Fresh != testCase.freshAttempt {
				t.Fatalf("managed/fresh = %v/%v, want %v/%v", result.Managed, result.Fresh, testCase.freshAttempt, testCase.freshAttempt)
			}
			if testCase.freshAttempt {
				if result.Attempt == nil {
					t.Fatal("a bound fresh attempt left no evidence")
				}
				if result.Attempt.Grant != grants.issued.ID || result.Attempt.Session != sessions.created[0].ID {
					t.Fatalf("attempt = %+v, want grant and fresh session", result.Attempt)
				}
			} else if result.Attempt != nil {
				t.Fatal("a pre-execution refusal claimed freshness")
			}
			if grants.completeCalls != 0 {
				t.Fatalf("complete called %d times, want 0", grants.completeCalls)
			}
			if grants.failCalls != 1 {
				t.Fatalf("grant failures = %d, want 1", grants.failCalls)
			}
			if executor.calls != testCase.executions {
				t.Fatalf("the worker ran %d times, want %d", executor.calls, testCase.executions)
			}
		})
	}
}

// TestManagedNonZeroExitFailsTheGrantWithoutCompleting: the real shell runner
// reports a non-zero exit with a nil error, so the managed path must read the
// exit code, fail the grant, never complete it, and keep the bound fresh attempt,
// output and exit code as blocked evidence.
func TestManagedNonZeroExitFailsTheGrantWithoutCompleting(t *testing.T) {
	log := &eventLog{}
	opts, files := setup(t)
	deps, grants, sessions, executor := managedDeps(log)
	executor.output = "worker-exploded"
	executor.exitCode = 1

	result := delegate.RunManaged(t.Context(), opts, files, deps)
	if result.Code != delegate.CodeBlocked {
		t.Fatalf("a non-zero worker exit produced %+v", result)
	}
	if !result.Managed || !result.Fresh {
		t.Fatalf("managed/fresh = %v/%v, want true/true", result.Managed, result.Fresh)
	}
	if result.Attempt == nil {
		t.Fatal("a bound fresh attempt left no evidence")
	}
	if result.Output != "worker-exploded" {
		t.Fatalf("output = %q, want the worker output", result.Output)
	}
	if result.ExitCode != 1 {
		t.Fatalf("exit code = %d, want 1", result.ExitCode)
	}
	if !strings.Contains(result.Message, "1") {
		t.Fatalf("message %q does not name the exit code", result.Message)
	}
	if grants.completeCalls != 0 {
		t.Fatalf("complete called %d times, want 0", grants.completeCalls)
	}
	if grants.failCalls != 1 {
		t.Fatalf("grant failures = %d, want 1", grants.failCalls)
	}
	if len(sessions.created) != 1 {
		t.Fatalf("session creations = %d, want 1", len(sessions.created))
	}
	if result.Attempt.Grant != grants.issued.ID {
		t.Fatalf("attempt grant = %s, want %s", result.Attempt.Grant, grants.issued.ID)
	}
	if result.Attempt.Session != sessions.created[0].ID {
		t.Fatalf("attempt session = %q, want %q", result.Attempt.Session, sessions.created[0].ID)
	}
	if result.Attempt.Model != sessions.created[0].Model {
		t.Fatalf("attempt model = %q, want %q", result.Attempt.Model, sessions.created[0].Model)
	}
}

// TestLegacyNonZeroExitIsBlockedWithoutFreshness: the unmanaged `--exec` path
// must also read the exit code, return blocked with the output and code, and
// still claim no grant and no freshness.
func TestLegacyNonZeroExitIsBlockedWithoutFreshness(t *testing.T) {
	opts, files := setup(t)
	legacy := &fakeLegacy{output: "legacy-exploded", exitCode: 3}

	result := delegate.RunUnmanaged(t.Context(), opts, files, "agent --exec", legacy)
	if result.Code != delegate.CodeBlocked {
		t.Fatalf("a non-zero legacy exit produced %+v", result)
	}
	if result.Managed || result.Fresh || result.Attempt != nil {
		t.Fatal("the unmanaged path claimed freshness")
	}
	if result.Output != "legacy-exploded" {
		t.Fatalf("output = %q, want the legacy output", result.Output)
	}
	if result.ExitCode != 3 {
		t.Fatalf("exit code = %d, want 3", result.ExitCode)
	}
	if !strings.Contains(result.Message, "3") {
		t.Fatalf("message %q does not name the exit code", result.Message)
	}
	if legacy.calls != 1 {
		t.Fatalf("legacy runs = %d, want 1", legacy.calls)
	}
}

// TestManagedAttemptRefusesAnInlineStepBeforeIssuingAGrant: the managed path
// keeps the prompt builder's refusal of a step the parent runs.
func TestManagedAttemptRefusesAnInlineStepBeforeIssuingAGrant(t *testing.T) {
	log := &eventLog{}
	opts, files := setup(t)
	opts.Step = "T2"
	deps, grants, sessions, executor := managedDeps(log)

	result := delegate.RunManaged(t.Context(), opts, files, deps)
	if result.Code != delegate.CodeBlocked {
		t.Fatal("an inline step got a managed attempt")
	}
	if len(grants.bindings) != 0 || sessions.calls != 0 || executor.calls != 0 {
		t.Fatal("an inline step reached a managed boundary")
	}
}

// TestManagedAttemptRefusesMissingBoundaries: no grant, no fresh runtime and no
// executor means no attempt.
func TestManagedAttemptRefusesMissingBoundaries(t *testing.T) {
	opts, files := setup(t)
	result := delegate.RunManaged(t.Context(), opts, files, delegate.ManagedDeps{})
	if result.Code != delegate.CodeBlocked {
		t.Fatalf("an unwired attempt produced %+v", result)
	}
	if result.Fresh || result.Attempt != nil {
		t.Fatal("an unwired attempt claimed freshness")
	}
}

// TestFreshSessionBoundaryExposesNoResumeOperation: the managed runtime surface
// has one operation, it takes no session identity, and it offers no resume,
// continue or fork door.
func TestFreshSessionBoundaryExposesNoResumeOperation(t *testing.T) {
	typ := reflect.TypeFor[delegate.Sessions]()
	if typ.NumMethod() != 1 {
		t.Fatalf("Sessions has %d methods, want 1", typ.NumMethod())
	}
	method := typ.Method(0)
	if method.Name != "CreateFreshSession" {
		t.Fatalf("the only runtime operation is %q, want CreateFreshSession", method.Name)
	}
	if method.Type.NumIn() != 1 {
		t.Fatalf("CreateFreshSession takes %d inputs, want just the context", method.Type.NumIn())
	}
	for _, forbidden := range []string{"Resume", "Continue", "Fork"} {
		if _, ok := typ.MethodByName(forbidden); ok {
			t.Fatalf("the runtime surface exposes %s", forbidden)
		}
	}
}

// TestLegacyUnmanagedPathMakesNoFreshnessClaim: the `--exec` form composes and
// runs through the legacy boundary without a grant or a session, and says so.
func TestLegacyUnmanagedPathMakesNoFreshnessClaim(t *testing.T) {
	opts, files := setup(t)
	legacy := &fakeLegacy{output: "legacy-output"}

	result := delegate.RunUnmanaged(t.Context(), opts, files, "agent --exec", legacy)
	if result.Code != delegate.CodeOK {
		t.Fatalf("RunUnmanaged: %s", result.Message)
	}
	if result.Managed || result.Fresh || result.Attempt != nil {
		t.Fatal("the unmanaged path claimed freshness")
	}
	if legacy.calls != 1 {
		t.Fatalf("legacy runs = %d, want 1", legacy.calls)
	}
	if legacy.command != "agent --exec" {
		t.Fatalf("legacy command = %q", legacy.command)
	}
	if !strings.Contains(legacy.prompt, "## Skills to load before work") {
		t.Fatalf("the legacy prompt lost the skills section:\n%s", legacy.prompt)
	}
	if !strings.Contains(result.Prompt, "# T1 · implementar el parser") {
		t.Fatalf("the result did not keep the composed prompt:\n%s", result.Prompt)
	}
	if result.Output != "legacy-output" {
		t.Fatalf("output = %q, want legacy-output", result.Output)
	}
}
