package verify_test

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/app/verify"
	domaingrant "github.com/biggs-100/odachi/internal/domain/grant"
)

// verifierBase is the base a managed verifier binds beside the candidate tree.
const verifierBase = "b4c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9"

// snapshotBefore is the stable digest a frozen candidate answers before and
// after a verifier that changes nothing.
const snapshotBefore = "a1a2a3a4a5a6a7a8a9a0b1b2b3b4b5b6b7b8b9b0b1b2b3b4b5b6b7b8b9b0"

var (
	startedAt  = time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	finishedAt = time.Date(2026, 11, 16, 12, 0, 5, 0, time.UTC)
)

// managedEvents is the ordered record of what the fake boundaries observed. A
// managed verify is sequential, so a plain slice is enough under -race.
type managedEvents struct {
	events []string
}

func (log *managedEvents) add(event string) { log.events = append(log.events, event) }

func (log *managedEvents) snapshot() []string { return slices.Clone(log.events) }

// fakeManagedFreezer is the one freeze a managed verifier is allowed. It names
// the run and the candidate inside it, and records the discard.
type fakeManagedFreezer struct {
	log       *managedEvents
	calls     int
	archived  []string
	run       string
	candidate string
	err       error
	discarded []string
}

func (fake *fakeManagedFreezer) ArchiveRun(_, tree string) (string, string, error) {
	fake.log.add("freeze")
	fake.calls++
	fake.archived = append(fake.archived, tree)
	if fake.err != nil {
		return "", "", fake.err
	}
	return fake.run, fake.candidate, nil
}

func (fake *fakeManagedFreezer) Discard(dir string) error {
	fake.log.add("discard")
	fake.discarded = append(fake.discarded, dir)
	return nil
}

// fakeSnapshotter answers a fixed digest per call, so a test can make the tree
// move between the before and after checks.
type fakeSnapshotter struct {
	log     *managedEvents
	calls   int
	digests []string
	errAt   int
	err     error
}

func (fake *fakeSnapshotter) Digest(_ string) (string, error) {
	fake.log.add("snapshot")
	fake.calls++
	if fake.errAt == fake.calls {
		return "", fake.err
	}
	if fake.calls-1 < len(fake.digests) {
		return fake.digests[fake.calls-1], nil
	}
	return snapshotBefore, nil
}

// fakeManagedGrants is the one-shot authority boundary. It keeps each consumed
// identity, so a replay is refused the way the durable service refuses it.
type fakeManagedGrants struct {
	log       *managedEvents
	issued    []domaingrant.Grant
	claimed   map[string]domaingrant.Grant
	bound     map[string]domaingrant.Grant
	bindings  []domaingrant.Binding
	issueErr  error
	claimErr  error
	bindErr   error
	complete  error
	failErr   error
	completes int
	failures  int
}

func newFakeManagedGrants(log *managedEvents) *fakeManagedGrants {
	return &fakeManagedGrants{
		log:     log,
		claimed: make(map[string]domaingrant.Grant),
		bound:   make(map[string]domaingrant.Grant),
	}
}

func (fake *fakeManagedGrants) Issue(_ context.Context, binding domaingrant.Binding) (domaingrant.Grant, error) {
	fake.log.add("issue")
	fake.bindings = append(fake.bindings, binding)
	if fake.issueErr != nil {
		return domaingrant.Grant{}, fake.issueErr
	}
	issued, err := domaingrant.NewGrant(binding)
	if err != nil {
		return domaingrant.Grant{}, err
	}
	fake.issued = append(fake.issued, issued)
	return issued, nil
}

func (fake *fakeManagedGrants) Claim(_ context.Context, id domaingrant.ID) (domaingrant.Grant, error) {
	fake.log.add("claim")
	if fake.claimErr != nil {
		return domaingrant.Grant{}, fake.claimErr
	}
	if _, ok := fake.claimed[id.String()]; ok {
		return domaingrant.Grant{}, domaingrant.ErrAlreadyClaimed
	}
	for i := range fake.issued {
		if fake.issued[i].ID != id {
			continue
		}
		claimed, err := fake.issued[i].Claim()
		if err != nil {
			return domaingrant.Grant{}, err
		}
		fake.claimed[id.String()] = claimed
		return claimed, nil
	}
	return domaingrant.Grant{}, errors.New("the claim names a foreign grant")
}

func (fake *fakeManagedGrants) BindSession(_ context.Context, id domaingrant.ID, session domaingrant.SessionID) (domaingrant.Grant, error) {
	fake.log.add("bind")
	if fake.bindErr != nil {
		return domaingrant.Grant{}, fake.bindErr
	}
	claimed, ok := fake.claimed[id.String()]
	if !ok {
		return domaingrant.Grant{}, domaingrant.ErrInvalidTransition
	}
	bound, err := claimed.BindSession(session)
	if err != nil {
		return domaingrant.Grant{}, err
	}
	fake.bound[id.String()] = bound
	return bound, nil
}

func (fake *fakeManagedGrants) Complete(_ context.Context, id domaingrant.ID) (domaingrant.Grant, error) {
	fake.log.add("complete")
	fake.completes++
	if fake.complete != nil {
		return domaingrant.Grant{}, fake.complete
	}
	bound, ok := fake.bound[id.String()]
	if !ok {
		return domaingrant.Grant{}, domaingrant.ErrInvalidTransition
	}
	return bound.Complete()
}

func (fake *fakeManagedGrants) Fail(_ context.Context, id domaingrant.ID) (domaingrant.Grant, error) {
	fake.log.add("fail")
	fake.failures++
	if fake.failErr != nil {
		return domaingrant.Grant{}, fake.failErr
	}
	current, ok := fake.bound[id.String()]
	if !ok {
		current, ok = fake.claimed[id.String()]
	}
	if !ok {
		return domaingrant.Grant{}, domaingrant.ErrInvalidTransition
	}
	return current.Fail()
}

// fakeManagedSessions is the fresh runtime boundary. It can refuse outright or
// hand back a canned identity, so an unusable session is exercised at the
// verifier boundary.
type fakeManagedSessions struct {
	log       *managedEvents
	calls     int
	responses []verify.VerifierSession
	err       error
	created   []verify.VerifierSession
}

func (fake *fakeManagedSessions) CreateFreshSession(_ context.Context) (verify.VerifierSession, error) {
	fake.log.add("create-session")
	fake.calls++
	if fake.err != nil {
		return verify.VerifierSession{}, fake.err
	}
	session := verify.VerifierSession{
		ID:    domaingrant.SessionID("ses_verify_" + strconv.Itoa(fake.calls)),
		Model: "opencode/observed-model",
	}
	if fake.calls-1 < len(fake.responses) {
		session = fake.responses[fake.calls-1]
	}
	fake.created = append(fake.created, session)
	return session, nil
}

// fakeManagedExecutor records the request it was handed. It can fail or exit
// non-zero, and its output can carry forged executor claims.
type fakeManagedExecutor struct {
	log      *managedEvents
	calls    int
	requests []verify.VerifierExecutionRequest
	output   string
	exitCode int
	err      error
}

func (fake *fakeManagedExecutor) Execute(_ context.Context, request verify.VerifierExecutionRequest) (verify.VerifierExecution, error) {
	fake.log.add("execute")
	fake.calls++
	fake.requests = append(fake.requests, request)
	return verify.VerifierExecution{Output: fake.output, ExitCode: fake.exitCode}, fake.err
}

// fakeClock answers the two instants a run seals: the start and the finish.
type fakeClock struct {
	times []time.Time
	calls int
}

func (fake *fakeClock) Now() time.Time {
	if fake.calls < len(fake.times) {
		now := fake.times[fake.calls]
		fake.calls++
		return now
	}
	fake.calls++
	return finishedAt
}

// fakeLegacyVerifier is the unmanaged `--exec` boundary: a command and a
// prompt, with no grant and no session.
type fakeLegacyVerifier struct {
	calls    int
	command  string
	prompt   string
	output   string
	exitCode int
	err      error
}

func (fake *fakeLegacyVerifier) Run(_ context.Context, command, prompt string) (verify.VerifierExecution, error) {
	fake.calls++
	fake.command = command
	fake.prompt = prompt
	return verify.VerifierExecution{Output: fake.output, ExitCode: fake.exitCode}, fake.err
}

// managedFixture wires the six fake boundaries over one event log.
type managedFixture struct {
	deps      verify.ManagedVerifyDeps
	freezer   *fakeManagedFreezer
	grants    *fakeManagedGrants
	sessions  *fakeManagedSessions
	executor  *fakeManagedExecutor
	snapshots *fakeSnapshotter
	clock     *fakeClock
}

func newManagedFixture(log *managedEvents) *managedFixture {
	freezer := &fakeManagedFreezer{
		log:       log,
		run:       "/temp/odachi-verify-run",
		candidate: "/temp/odachi-verify-run/candidate",
	}
	grants := newFakeManagedGrants(log)
	sessions := &fakeManagedSessions{log: log}
	executor := &fakeManagedExecutor{
		log:    log,
		output: `{"schema":1,"by":"otro/modelo","verdict":"verified","notes":[]}`,
	}
	snapshots := &fakeSnapshotter{log: log, digests: []string{snapshotBefore, snapshotBefore}}
	clock := &fakeClock{times: []time.Time{startedAt, finishedAt}}
	return &managedFixture{
		deps: verify.ManagedVerifyDeps{
			Freezer:   freezer,
			Grants:    grants,
			Sessions:  sessions,
			Executor:  executor,
			Snapshots: snapshots,
			Clock:     clock,
		},
		freezer:   freezer,
		grants:    grants,
		sessions:  sessions,
		executor:  executor,
		snapshots: snapshots,
		clock:     clock,
	}
}

// managedRequest is one managed verification of the candidate tree against its
// base on one runtime, with one verifier identity.
func managedRequest() verify.ManagedVerify {
	return verify.ManagedVerify{
		Dir:      "/proyecto",
		Tree:     tree,
		Base:     verifierBase,
		Runtime:  "opencode",
		Model:    "opencode/model",
		Verifier: "odachi-verify",
		Prompt:   "# verify the tree",
	}
}

// TestManagedVerifyFreezesOneCandidateAndRunsOnlyThere: the verifier freezes the
// very tree the review would freeze and runs inside that frozen directory, never
// the live worktree.
func TestManagedVerifyFreezesOneCandidateAndRunsOnlyThere(t *testing.T) {
	log := &managedEvents{}
	fixture := newManagedFixture(log)

	result := verify.RunManagedVerify(t.Context(), managedRequest(), fixture.deps)
	if result.Code != verify.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if fixture.freezer.calls != 1 {
		t.Fatalf("freezes = %d, want 1", fixture.freezer.calls)
	}
	if !slices.Equal(fixture.freezer.archived, []string{tree}) {
		t.Fatalf("archived = %v, want the candidate %q", fixture.freezer.archived, tree)
	}
	if len(fixture.executor.requests) != 1 {
		t.Fatalf("executions = %d, want 1", len(fixture.executor.requests))
	}
	request := fixture.executor.requests[0]
	if request.Dir != fixture.freezer.candidate {
		t.Errorf("the verifier ran in %q, want the frozen candidate %q", request.Dir, fixture.freezer.candidate)
	}
	if request.Tree != tree || request.Base != verifierBase {
		t.Errorf("the verifier bound tree/base %q/%q, want %q/%q", request.Tree, request.Base, tree, verifierBase)
	}
	if !slices.Equal(fixture.freezer.discarded, []string{fixture.freezer.run}) {
		t.Errorf("discarded = %v, want the whole run %q", fixture.freezer.discarded, fixture.freezer.run)
	}
}

// TestManagedVerifyFailsClosedWhenTheSnapshotMoves: a snapshot that cannot be
// taken before execution starts nothing, and one that changed after execution
// seals no evidence and terminates its grant.
func TestManagedVerifyFailsClosedWhenTheSnapshotMoves(t *testing.T) {
	cases := map[string]struct {
		configure  func(*fakeSnapshotter)
		executions int
		issues     int
		failures   int
	}{
		"before unavailable": {
			configure: func(s *fakeSnapshotter) { s.errAt = 1; s.err = errors.New("cannot read the snapshot") },
		},
		"before empty": {
			configure: func(s *fakeSnapshotter) { s.digests = []string{"", snapshotBefore} },
		},
		"after changed": {
			configure: func(s *fakeSnapshotter) {
				s.digests = []string{snapshotBefore, "b2b3b4b5b6b7b8b9b0b1b2b3b4b5b6b7b8b9b0b1b2b3b4b5b6b7b8b9b0b1"}
			},
			executions: 1,
			issues:     1,
			failures:   1,
		},
		"after empty": {
			configure:  func(s *fakeSnapshotter) { s.digests = []string{snapshotBefore, ""} },
			executions: 1,
			issues:     1,
			failures:   1,
		},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			log := &managedEvents{}
			fixture := newManagedFixture(log)
			testCase.configure(fixture.snapshots)

			result := verify.RunManagedVerify(t.Context(), managedRequest(), fixture.deps)
			if result.Code != verify.CodeBlocked {
				t.Fatalf("a moving snapshot produced %+v", result)
			}
			if result.Sealed != nil || result.Managed || result.Fresh {
				t.Fatal("a moving snapshot sealed evidence")
			}
			if fixture.executor.calls != testCase.executions {
				t.Fatalf("executions = %d, want %d", fixture.executor.calls, testCase.executions)
			}
			if len(fixture.grants.issued) != testCase.issues {
				t.Fatalf("issued grants = %d, want %d", len(fixture.grants.issued), testCase.issues)
			}
			if fixture.grants.failures != testCase.failures {
				t.Fatalf("grant failures = %d, want %d", fixture.grants.failures, testCase.failures)
			}
			if fixture.grants.completes != 0 {
				t.Fatalf("completions = %d, want 0", fixture.grants.completes)
			}
		})
	}
}

// TestEveryManagedVerifyRunGetsItsOwnGrantAndSession: two runs never share
// authority. Each issues a distinct grant, creates and binds its own fresh
// session, and only then runs the verifier.
func TestEveryManagedVerifyRunGetsItsOwnGrantAndSession(t *testing.T) {
	log := &managedEvents{}
	fixture := newManagedFixture(log)

	first := verify.RunManagedVerify(t.Context(), managedRequest(), fixture.deps)
	second := verify.RunManagedVerify(t.Context(), managedRequest(), fixture.deps)
	if first.Code != verify.CodeOK || second.Code != verify.CodeOK {
		t.Fatalf("codes = %d/%d, want 0/0 (%s / %s)", first.Code, second.Code, first.Message, second.Message)
	}
	if len(fixture.grants.issued) != 2 || fixture.sessions.calls != 2 || fixture.executor.calls != 2 {
		t.Fatalf("grants/sessions/executions = %d/%d/%d, want 2/2/2", len(fixture.grants.issued), fixture.sessions.calls, fixture.executor.calls)
	}
	if fixture.grants.issued[0].ID == fixture.grants.issued[1].ID {
		t.Fatal("two runs reused the same grant")
	}
	if fixture.sessions.created[0].ID == fixture.sessions.created[1].ID {
		t.Fatal("two runs reused the same session")
	}
	if first.Sealed == nil || second.Sealed == nil {
		t.Fatal("a completed run sealed no runtime receipt")
	}
	if first.Sealed.Grant == second.Sealed.Grant || first.Sealed.Session == second.Sealed.Session {
		t.Fatal("two runs sealed the same grant or session")
	}
	if len(fixture.grants.bindings) != 2 || fixture.grants.bindings[0] != fixture.grants.bindings[1] {
		t.Fatal("the two runs did not bind the same canonical verifier input")
	}
	if fixture.grants.bound[fixture.grants.issued[0].ID.String()].Session != fixture.sessions.created[0].ID {
		t.Fatal("the grant did not bind the fresh session it ran with")
	}

	want := []string{"freeze", "snapshot", "issue", "claim", "create-session", "bind", "execute", "snapshot", "complete", "discard"}
	if got := log.snapshot(); len(got) != 2*len(want) || !slices.Equal(got[:len(want)], want) {
		t.Fatalf("first run order = %v, want %v", got, want)
	}
}

// TestVerifierBindingBindsEveryCanonicalField: the binding names the candidate
// tree and base, the runtime, the model, the verifier identity, and the fixed
// verifier role; changing any of them changes the digest.
func TestVerifierBindingBindsEveryCanonicalField(t *testing.T) {
	base := verify.VerifierBindingFields{
		Tree:     tree,
		Base:     verifierBase,
		Runtime:  "opencode",
		Model:    "opencode/model",
		Verifier: "odachi-verify",
	}
	baseline, err := verify.VerifierBinding(base)
	if err != nil {
		t.Fatalf("VerifierBinding: %v", err)
	}

	variants := map[string]verify.VerifierBindingFields{
		"tree":     {Tree: "c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3", Base: base.Base, Runtime: base.Runtime, Model: base.Model, Verifier: base.Verifier},
		"base":     {Tree: base.Tree, Base: "d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4", Runtime: base.Runtime, Model: base.Model, Verifier: base.Verifier},
		"runtime":  {Tree: base.Tree, Base: base.Base, Runtime: "another-runtime", Model: base.Model, Verifier: base.Verifier},
		"model":    {Tree: base.Tree, Base: base.Base, Runtime: base.Runtime, Model: "opencode/other-model", Verifier: base.Verifier},
		"verifier": {Tree: base.Tree, Base: base.Base, Runtime: base.Runtime, Model: base.Model, Verifier: "another-verifier"},
	}
	for name, fields := range variants {
		t.Run(name, func(t *testing.T) {
			got, err := verify.VerifierBinding(fields)
			if err != nil {
				t.Fatalf("VerifierBinding: %v", err)
			}
			if got == baseline {
				t.Fatalf("changing %s did not change the binding", name)
			}
		})
	}

	if verify.VerifierRole != "verifier" {
		t.Fatalf("VerifierRole = %q, want verifier", verify.VerifierRole)
	}
}

// TestVerifierBindingRefusesAnIncompleteCandidate: an unbound verifier is not a
// canonical one, so a missing tree, base, runtime, model or verifier is refused
// before a grant can name it.
func TestVerifierBindingRefusesAnIncompleteCandidate(t *testing.T) {
	base := verify.VerifierBindingFields{
		Tree:     tree,
		Base:     verifierBase,
		Runtime:  "opencode",
		Model:    "opencode/model",
		Verifier: "odachi-verify",
	}
	cases := map[string]verify.VerifierBindingFields{
		"tree":     {Base: base.Base, Runtime: base.Runtime, Model: base.Model, Verifier: base.Verifier},
		"base":     {Tree: base.Tree, Runtime: base.Runtime, Model: base.Model, Verifier: base.Verifier},
		"runtime":  {Tree: base.Tree, Base: base.Base, Model: base.Model, Verifier: base.Verifier},
		"model":    {Tree: base.Tree, Base: base.Base, Runtime: base.Runtime, Verifier: base.Verifier},
		"verifier": {Tree: base.Tree, Base: base.Base, Runtime: base.Runtime, Model: base.Model},
	}
	for name, fields := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := verify.VerifierBinding(fields); !errors.Is(err, domaingrant.ErrInvalidBinding) {
				t.Fatalf("VerifierBinding error = %v, want ErrInvalidBinding", err)
			}
		})
	}
}

// TestManagedVerifyExecutionExposesNoResumeInput: the runtime surface has one
// operation that takes no session identity, and the execution request carries
// no resume, continue, fork or caller-supplied parent session.
func TestManagedVerifyExecutionExposesNoResumeInput(t *testing.T) {
	typ := reflect.TypeFor[verify.VerifierSessions]()
	if typ.NumMethod() != 1 {
		t.Fatalf("VerifierSessions has %d methods, want 1", typ.NumMethod())
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

	request := reflect.TypeFor[verify.VerifierExecutionRequest]()
	for i := range request.NumField() {
		switch request.Field(i).Name {
		case "Resume", "Continue", "Fork", "ParentID", "SessionID":
			t.Fatalf("the execution request exposes caller-supplied %s", request.Field(i).Name)
		}
	}
	field, ok := request.FieldByName("Session")
	if !ok {
		t.Fatal("the execution request does not carry the fresh bound session")
	}
	if field.Type != reflect.TypeFor[domaingrant.SessionID]() {
		t.Fatalf("Session field type = %s, want the grant session identity", field.Type)
	}
}

// TestManagedVerifySealsToolDerivedRuntimeFacts: the sealed receipt names the
// grant, session, runtime, both models, role, candidate tree/base, snapshot
// digest, both timestamps and the outcome that the tool observed.
func TestManagedVerifySealsToolDerivedRuntimeFacts(t *testing.T) {
	log := &managedEvents{}
	fixture := newManagedFixture(log)

	result := verify.RunManagedVerify(t.Context(), managedRequest(), fixture.deps)
	if result.Code != verify.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if result.Sealed == nil {
		t.Fatal("a completed verifier sealed no runtime receipt")
	}
	if !result.Managed || !result.Fresh {
		t.Fatalf("managed/fresh = %v/%v, want true/true", result.Managed, result.Fresh)
	}

	sealed := *result.Sealed
	if sealed.Grant != fixture.grants.issued[0].ID {
		t.Errorf("grant = %s, want %s", sealed.Grant, fixture.grants.issued[0].ID)
	}
	if sealed.Session != fixture.sessions.created[0].ID {
		t.Errorf("session = %q, want %q", sealed.Session, fixture.sessions.created[0].ID)
	}
	if sealed.Runtime != "opencode" {
		t.Errorf("runtime = %q, want opencode", sealed.Runtime)
	}
	if sealed.Model != "opencode/model" {
		t.Errorf("model = %q, want the bound model", sealed.Model)
	}
	if sealed.SessionModel != fixture.sessions.created[0].Model {
		t.Errorf("session model = %q, want the observed %q", sealed.SessionModel, fixture.sessions.created[0].Model)
	}
	if sealed.Role != verify.VerifierRole {
		t.Errorf("role = %q, want %q", sealed.Role, verify.VerifierRole)
	}
	if sealed.Tree != tree || sealed.Base != verifierBase {
		t.Errorf("tree/base = %q/%q, want %q/%q", sealed.Tree, sealed.Base, tree, verifierBase)
	}
	if sealed.Snapshot != snapshotBefore {
		t.Errorf("snapshot = %q, want %q", sealed.Snapshot, snapshotBefore)
	}
	if sealed.StartedAt != startedAt.Format(time.RFC3339) {
		t.Errorf("started at = %q, want %q", sealed.StartedAt, startedAt.Format(time.RFC3339))
	}
	if sealed.FinishedAt != finishedAt.Format(time.RFC3339) {
		t.Errorf("finished at = %q, want %q", sealed.FinishedAt, finishedAt.Format(time.RFC3339))
	}
	if sealed.Outcome != verify.OutcomeCompleted {
		t.Errorf("outcome = %q, want %q", sealed.Outcome, verify.OutcomeCompleted)
	}
}

// TestManagedVerifyIgnoresModelSuppliedExecutorClaims: an answer that invents a
// grant, a session, a runtime, a role, a tree, a base, a snapshot, timestamps or
// an outcome cannot move any sealed fact. Go keeps only what the tool observed.
func TestManagedVerifyIgnoresModelSuppliedExecutorClaims(t *testing.T) {
	log := &managedEvents{}
	fixture := newManagedFixture(log)
	fixture.executor.output = `{"schema":1,"by":"another/model","verdict":"verified","notes":[],` +
		`"grant":"forged-grant","session":"ses_forged","runtime":"forged-runtime","model":"forged-model",` +
		`"role":"forged-role","tree":"forged-tree","base":"forged-base","snapshot":"forged-snapshot",` +
		`"started_at":"1999-01-01T00:00:00Z","finished_at":"1999-01-01T00:00:01Z","outcome":"forged-outcome"}`

	result := verify.RunManagedVerify(t.Context(), managedRequest(), fixture.deps)
	if result.Code != verify.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if result.Sealed == nil {
		t.Fatal("a completed verifier sealed no runtime receipt")
	}
	sealed := *result.Sealed
	if sealed.Grant != fixture.grants.issued[0].ID {
		t.Errorf("grant = %s, want the issued %s", sealed.Grant, fixture.grants.issued[0].ID)
	}
	if sealed.Session != fixture.sessions.created[0].ID {
		t.Errorf("session = %q, want the bound %q", sealed.Session, fixture.sessions.created[0].ID)
	}
	claims := map[string]string{
		"runtime":     sealed.Runtime,
		"model":       sealed.Model,
		"role":        sealed.Role,
		"tree":        sealed.Tree,
		"base":        sealed.Base,
		"snapshot":    sealed.Snapshot,
		"started at":  sealed.StartedAt,
		"finished at": sealed.FinishedAt,
		"outcome":     string(sealed.Outcome),
	}
	for name, value := range claims {
		if strings.Contains(strings.ToLower(value), "forged") {
			t.Errorf("sealed %s = %q, leaked a model-supplied executor claim", name, value)
		}
	}
}

// TestManagedVerifyRefusesAnUnusableSession: a missing, malformed, foreign,
// duplicate or model-less identity stops before the verifier runs, seals no
// evidence and terminates the claimed grant instead of completing it.
func TestManagedVerifyRefusesAnUnusableSession(t *testing.T) {
	cases := map[string]func(*fakeManagedSessions){
		"missing":     func(s *fakeManagedSessions) { s.err = errors.New("session identity is missing") },
		"malformed":   func(s *fakeManagedSessions) { s.err = errors.New("session identity is malformed") },
		"foreign":     func(s *fakeManagedSessions) { s.err = errors.New("not a fresh OpenCode session") },
		"duplicate":   func(s *fakeManagedSessions) { s.err = errors.New("duplicate session identity") },
		"empty id":    func(s *fakeManagedSessions) { s.responses = []verify.VerifierSession{{Model: "opencode/model"}} },
		"empty model": func(s *fakeManagedSessions) { s.responses = []verify.VerifierSession{{ID: "ses_verify_one"}} },
		"malformed grammar": func(s *fakeManagedSessions) {
			s.responses = []verify.VerifierSession{{ID: "ses verify", Model: "opencode/model"}}
		},
	}
	for name, configure := range cases {
		t.Run(name, func(t *testing.T) {
			log := &managedEvents{}
			fixture := newManagedFixture(log)
			configure(fixture.sessions)

			result := verify.RunManagedVerify(t.Context(), managedRequest(), fixture.deps)
			if result.Code != verify.CodeBlocked {
				t.Fatalf("an unusable session produced %+v", result)
			}
			if !strings.Contains(strings.ToLower(result.Message), "session") {
				t.Errorf("message %q does not name the unusable session", result.Message)
			}
			if result.Sealed != nil || result.Managed || result.Fresh {
				t.Fatal("an unusable session sealed evidence")
			}
			if fixture.executor.calls != 0 {
				t.Fatalf("executions = %d, want 0", fixture.executor.calls)
			}
			if fixture.grants.completes != 0 {
				t.Fatalf("completions = %d, want 0", fixture.grants.completes)
			}
			if fixture.grants.failures != 1 {
				t.Fatalf("grant failures = %d, want 1", fixture.grants.failures)
			}
		})
	}
}

// TestManagedVerifyRefusesANonZeroVerifierExit: a verifier that ran and failed
// seals no evidence and never completes its grant.
func TestManagedVerifyRefusesANonZeroVerifierExit(t *testing.T) {
	log := &managedEvents{}
	fixture := newManagedFixture(log)
	fixture.executor.exitCode = 2
	fixture.executor.output = "the verifier exploded"

	result := verify.RunManagedVerify(t.Context(), managedRequest(), fixture.deps)
	if result.Code != verify.CodeBlocked {
		t.Fatalf("a non-zero verifier produced %+v", result)
	}
	if result.Sealed != nil || result.Managed || result.Fresh {
		t.Fatal("a non-zero verifier sealed evidence")
	}
	if result.Output != fixture.executor.output || result.ExitCode != 2 {
		t.Fatalf("output/exit = %q/%d, want %q/2", result.Output, result.ExitCode, fixture.executor.output)
	}
	if !strings.Contains(result.Message, "2") {
		t.Errorf("message %q does not name the exit code", result.Message)
	}
	if fixture.grants.completes != 0 || fixture.grants.failures != 1 {
		t.Fatalf("completions/failures = %d/%d, want 0/1", fixture.grants.completes, fixture.grants.failures)
	}
}

// TestManagedVerifyRefusesAFreezeFailureBeforeAnyAuthority: a tree that cannot
// be frozen starts no verifier, creates no session and seals no evidence.
func TestManagedVerifyRefusesAFreezeFailureBeforeAnyAuthority(t *testing.T) {
	log := &managedEvents{}
	fixture := newManagedFixture(log)
	fixture.freezer.err = errors.New("git archive failed")

	result := verify.RunManagedVerify(t.Context(), managedRequest(), fixture.deps)
	if result.Code != verify.CodeBlocked {
		t.Fatalf("an unfreezable candidate produced %+v", result)
	}
	if !strings.Contains(strings.ToLower(result.Message), "freeze") {
		t.Errorf("message %q does not name the failed freeze", result.Message)
	}
	if len(fixture.grants.issued) != 0 || fixture.sessions.calls != 0 || fixture.executor.calls != 0 {
		t.Fatal("an unfreezable candidate reached a managed boundary")
	}
	if result.Sealed != nil || result.Managed || result.Fresh {
		t.Fatal("an unfreezable candidate sealed evidence")
	}
}

// TestManagedVerifyRefusesMissingBoundaries: no freezer, grant service, fresh
// runtime, executor, snapshot reader or clock means no verification.
func TestManagedVerifyRefusesMissingBoundaries(t *testing.T) {
	cases := map[string]func(*verify.ManagedVerifyDeps){
		"freezer":   func(d *verify.ManagedVerifyDeps) { d.Freezer = nil },
		"grants":    func(d *verify.ManagedVerifyDeps) { d.Grants = nil },
		"sessions":  func(d *verify.ManagedVerifyDeps) { d.Sessions = nil },
		"executor":  func(d *verify.ManagedVerifyDeps) { d.Executor = nil },
		"snapshots": func(d *verify.ManagedVerifyDeps) { d.Snapshots = nil },
		"clock":     func(d *verify.ManagedVerifyDeps) { d.Clock = nil },
	}
	for name, clear := range cases {
		t.Run(name, func(t *testing.T) {
			log := &managedEvents{}
			fixture := newManagedFixture(log)
			deps := fixture.deps
			clear(&deps)

			result := verify.RunManagedVerify(t.Context(), managedRequest(), deps)
			if result.Code != verify.CodeBlocked {
				t.Fatalf("an unwired verifier produced %+v", result)
			}
			if fixture.freezer.calls != 0 || fixture.sessions.calls != 0 || fixture.executor.calls != 0 {
				t.Fatal("an unwired verifier reached a boundary")
			}
			if result.Sealed != nil || result.Managed || result.Fresh {
				t.Fatal("an unwired verifier sealed evidence")
			}
		})
	}
}

// TestLegacyVerifyClaimsNoFreshnessOrSealedRuntime: the unmanaged `--exec` path
// runs the command and the prompt without a grant or a session, and its result
// is explicitly unmanaged with no sealed runtime receipt.
func TestLegacyVerifyClaimsNoFreshnessOrSealedRuntime(t *testing.T) {
	legacy := &fakeLegacyVerifier{output: "legacy-output"}

	result := verify.RunUnmanagedVerify(t.Context(), "agent --prompt", "the prompt", legacy)
	if result.Code != verify.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if result.Managed || result.Fresh || result.Sealed != nil {
		t.Fatal("the unmanaged path claimed freshness or a sealed runtime")
	}
	if legacy.calls != 1 || legacy.command != "agent --prompt" || legacy.prompt != "the prompt" {
		t.Fatalf("legacy run = %d %q %q, want one run of the command and prompt", legacy.calls, legacy.command, legacy.prompt)
	}
	if result.Output != "legacy-output" {
		t.Fatalf("output = %q, want legacy-output", result.Output)
	}
}

// TestLegacyVerifyNonZeroExitClaimsNoFreshness: a failed unmanaged run reports
// the output and still claims no authority.
func TestLegacyVerifyNonZeroExitClaimsNoFreshness(t *testing.T) {
	legacy := &fakeLegacyVerifier{output: "legacy-exploded", exitCode: 3}

	result := verify.RunUnmanagedVerify(t.Context(), "agent", "the prompt", legacy)
	if result.Code != verify.CodeBlocked {
		t.Fatalf("a non-zero legacy exit produced %+v", result)
	}
	if result.Managed || result.Fresh || result.Sealed != nil {
		t.Fatal("the unmanaged path claimed freshness or a sealed runtime")
	}
	if result.Output != "legacy-exploded" || result.ExitCode != 3 {
		t.Fatalf("output/exit = %q/%d, want legacy-exploded/3", result.Output, result.ExitCode)
	}
	if !strings.Contains(result.Message, "3") {
		t.Fatalf("message %q does not name the exit code", result.Message)
	}
}
