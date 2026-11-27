package review_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/review"
	domaingrant "github.com/biggs-100/odachi/internal/domain/grant"
)

// baseTree is the base a managed review binds beside the candidate tree.
const baseTree = "b4c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9"

// reviewEvents is the ordered record of what the fake boundaries observed. A
// managed review is sequential, so a plain slice is enough under -race.
type reviewEvents struct {
	events []string
}

func (log *reviewEvents) add(event string) { log.events = append(log.events, event) }

func (log *reviewEvents) snapshot() []string { return slices.Clone(log.events) }

// fakeReviewGrants is the one-shot authority boundary shared by every unit and
// draw. It keeps each consumed identity, so a replay is refused the way the
// durable service refuses it.
type fakeReviewGrants struct {
	log       *reviewEvents
	issued    []domaingrant.Grant
	claimed   map[string]domaingrant.Grant
	bound     map[string]domaingrant.Grant
	last      domaingrant.Grant
	bindings  []domaingrant.Binding
	issueErr  error
	claimErr  error
	bindErr   error
	complete  error
	failErr   error
	replay    bool
	completes int
	failures  int
	// completeGrant, when set, is what Complete reports instead of the bound
	// grant, so a service that names another authority can be refused.
	completeGrant *domaingrant.Grant
}

func newFakeReviewGrants(log *reviewEvents) *fakeReviewGrants {
	return &fakeReviewGrants{
		log:     log,
		claimed: make(map[string]domaingrant.Grant),
		bound:   make(map[string]domaingrant.Grant),
	}
}

func (fake *fakeReviewGrants) Issue(_ context.Context, binding domaingrant.Binding) (domaingrant.Grant, error) {
	fake.log.add("issue")
	fake.bindings = append(fake.bindings, binding)
	if fake.issueErr != nil {
		return domaingrant.Grant{}, fake.issueErr
	}
	if fake.replay && fake.last.ID != (domaingrant.ID{}) {
		// The store hands back an authority whose identity was already
		// consumed, now carrying the new binding so only the claim can
		// catch it.
		replayed := fake.last
		replayed.Binding = binding
		return replayed, nil
	}
	issued, err := domaingrant.NewGrant(binding)
	if err != nil {
		return domaingrant.Grant{}, err
	}
	fake.last = issued
	fake.issued = append(fake.issued, issued)
	return issued, nil
}

func (fake *fakeReviewGrants) Claim(_ context.Context, id domaingrant.ID) (domaingrant.Grant, error) {
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

func (fake *fakeReviewGrants) BindSession(_ context.Context, id domaingrant.ID, session domaingrant.SessionID) (domaingrant.Grant, error) {
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

func (fake *fakeReviewGrants) Complete(_ context.Context, id domaingrant.ID) (domaingrant.Grant, error) {
	fake.log.add("complete")
	fake.completes++
	if fake.complete != nil {
		return domaingrant.Grant{}, fake.complete
	}
	if fake.completeGrant != nil {
		return *fake.completeGrant, nil
	}
	bound, ok := fake.bound[id.String()]
	if !ok {
		return domaingrant.Grant{}, domaingrant.ErrInvalidTransition
	}
	return bound.Complete()
}

func (fake *fakeReviewGrants) Fail(_ context.Context, id domaingrant.ID) (domaingrant.Grant, error) {
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

// fakeReviewSessions is the fresh runtime boundary. It can answer a canned
// session, refuse a duplicate itself, or hand the same identity back twice so
// the review boundary's own duplicate guard is exercised.
type fakeReviewSessions struct {
	log            *reviewEvents
	calls          int
	seen           map[domaingrant.SessionID]bool
	allowDuplicate bool
	responses      []review.ReviewSession
	err            error
	created        []review.ReviewSession
}

func newFakeReviewSessions(log *reviewEvents) *fakeReviewSessions {
	return &fakeReviewSessions{log: log, seen: make(map[domaingrant.SessionID]bool)}
}

func (fake *fakeReviewSessions) CreateFreshSession(_ context.Context) (review.ReviewSession, error) {
	fake.log.add("create-session")
	fake.calls++
	if fake.err != nil {
		return review.ReviewSession{}, fake.err
	}
	session := review.ReviewSession{
		ID:    domaingrant.SessionID("ses_review_" + strconv.Itoa(fake.calls)),
		Model: "opencode/observed-model",
	}
	if fake.calls-1 < len(fake.responses) {
		session = fake.responses[fake.calls-1]
	}
	if session.ID != "" {
		if fake.seen[session.ID] && !fake.allowDuplicate {
			return review.ReviewSession{}, errors.New("duplicate session identity")
		}
		fake.seen[session.ID] = true
	}
	fake.created = append(fake.created, session)
	return session, nil
}

// fakeReviewExecutor records the unit and draw it was handed. It can fail or
// exit non-zero at one selected call.
type fakeReviewExecutor struct {
	log      *reviewEvents
	calls    int
	requests []review.ReviewExecutionRequest
	errAt    int
	err      error
	exitAt   int
	exitCode int
	output   string
}

func (fake *fakeReviewExecutor) Execute(_ context.Context, request review.ReviewExecutionRequest) (review.ReviewExecution, error) {
	fake.log.add("execute")
	fake.calls++
	fake.requests = append(fake.requests, request)
	if fake.errAt == fake.calls {
		return review.ReviewExecution{Output: fake.output, ExitCode: fake.exitCode}, fake.err
	}
	if fake.exitAt == fake.calls {
		return review.ReviewExecution{Output: fake.output, ExitCode: fake.exitCode}, nil
	}
	return review.ReviewExecution{Output: fake.output}, nil
}

// fakeReviewFreezer is the one freeze a managed review is allowed. It names the
// run and the candidate inside it, and records the discard.
type fakeReviewFreezer struct {
	log       *reviewEvents
	calls     int
	archived  []string
	run       string
	candidate string
	err       error
	discarded []string
}

func (fake *fakeReviewFreezer) ArchiveRun(_ string, tree string) (string, string, error) {
	fake.log.add("freeze")
	fake.calls++
	fake.archived = append(fake.archived, tree)
	if fake.err != nil {
		return "", "", fake.err
	}
	return fake.run, fake.candidate, nil
}

func (fake *fakeReviewFreezer) Discard(dir string) error {
	fake.log.add("discard")
	fake.discarded = append(fake.discarded, dir)
	return nil
}

// fakeLegacyReviewExecutor is the unmanaged `--exec` boundary: a command and a
// prompt, with no grant and no session.
type fakeLegacyReviewExecutor struct {
	calls    int
	command  string
	prompts  []string
	output   string
	exitCode int
	err      error
}

func (fake *fakeLegacyReviewExecutor) Run(_ context.Context, command, prompt string) (review.ReviewExecution, error) {
	fake.calls++
	fake.command = command
	fake.prompts = append(fake.prompts, prompt)
	if fake.err != nil {
		return review.ReviewExecution{Output: fake.output, ExitCode: fake.exitCode}, fake.err
	}
	return review.ReviewExecution{Output: fake.output, ExitCode: fake.exitCode}, nil
}

// reviewDeps wires the four fake boundaries over one event log.
func reviewDeps(log *reviewEvents) (review.ManagedReviewDeps, *fakeReviewFreezer, *fakeReviewGrants, *fakeReviewSessions, *fakeReviewExecutor) {
	freezer := &fakeReviewFreezer{
		log:       log,
		run:       "/temp/odachi-review-run",
		candidate: "/temp/odachi-review-run/candidate",
	}
	grants := newFakeReviewGrants(log)
	sessions := newFakeReviewSessions(log)
	executor := &fakeReviewExecutor{log: log, output: "reviewed"}
	return review.ManagedReviewDeps{
		Freezer:  freezer,
		Grants:   grants,
		Sessions: sessions,
		Executor: executor,
	}, freezer, grants, sessions, executor
}

// managedUnit is one reviewer process: a lens set over one part of one draw.
func managedUnit(lens, model string, draw, draws, part, parts int) review.ManagedUnit {
	return review.ManagedUnit{
		Lenses: []string{lens},
		Model:  model,
		Draw:   draw,
		Draws:  draws,
		Part:   part,
		Parts:  parts,
		Prompt: "# review " + lens + " draw " + strconv.Itoa(draw),
	}
}

// managedPlan is one review of currentTree against baseTree on one runtime.
func managedPlan(units ...review.ManagedUnit) review.ManagedReview {
	return review.ManagedReview{
		Dir:     "/proyecto",
		Tree:    currentTree,
		Base:    baseTree,
		Runtime: "opencode",
		Units:   units,
	}
}

// TestManagedReviewFreezesOneTreeForEveryUnitAndDraw: one freeze, many reviewer
// processes. Every lens, part and draw reads the same frozen evidence, and the
// run is discarded once at the end.
func TestManagedReviewFreezesOneTreeForEveryUnitAndDraw(t *testing.T) {
	log := &reviewEvents{}
	deps, freezer, _, _, executor := reviewDeps(log)
	plan := managedPlan(
		managedUnit("reliability", "opencode/model-a", 1, 2, 1, 1),
		managedUnit("risk", "opencode/model-b", 1, 2, 1, 2),
		managedUnit("risk", "opencode/model-b", 2, 2, 2, 2),
	)

	result := review.RunManagedReview(t.Context(), plan, deps)
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if freezer.calls != 1 {
		t.Fatalf("freezes = %d, want 1 for the whole review", freezer.calls)
	}
	if len(freezer.archived) != 1 || freezer.archived[0] != currentTree {
		t.Fatalf("archived = %v, want the candidate %q", freezer.archived, currentTree)
	}
	if len(executor.requests) != 3 {
		t.Fatalf("executions = %d, want 3", len(executor.requests))
	}
	for i, request := range executor.requests {
		if request.Dir != freezer.candidate {
			t.Errorf("execution %d ran in %q, want the frozen candidate %q", i, request.Dir, freezer.candidate)
		}
		if request.Tree != currentTree || request.Base != baseTree {
			t.Errorf("execution %d bound tree/base %q/%q, want %q/%q", i, request.Tree, request.Base, currentTree, baseTree)
		}
	}
	if !slices.Equal(freezer.discarded, []string{freezer.run}) {
		t.Errorf("discarded = %v, want %q", freezer.discarded, freezer.run)
	}
	if result.Frozen != freezer.candidate || result.Run != freezer.run {
		t.Errorf("result frozen/run = %q/%q, want %q/%q", result.Frozen, result.Run, freezer.candidate, freezer.run)
	}
}

// TestManagedReviewGivesEveryUnitAndDrawItsOwnGrantAndSession: no reviewer
// process reuses authority. Every unit and draw issues a distinct grant, creates
// and binds its own fresh session, and runs only after the bind.
func TestManagedReviewGivesEveryUnitAndDrawItsOwnGrantAndSession(t *testing.T) {
	log := &reviewEvents{}
	deps, _, grants, sessions, executor := reviewDeps(log)
	units := []review.ManagedUnit{
		managedUnit("reliability", "opencode/model-a", 1, 2, 1, 1),
		managedUnit("risk", "opencode/model-b", 1, 2, 1, 2),
		managedUnit("risk", "opencode/model-b", 2, 2, 2, 2),
	}
	plan := managedPlan(units...)

	result := review.RunManagedReview(t.Context(), plan, deps)
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if len(result.Attempts) != 3 {
		t.Fatalf("attempts = %d, want 3", len(result.Attempts))
	}
	if sessions.calls != 3 || executor.calls != 3 || len(grants.issued) != 3 {
		t.Fatalf("sessions/executions/grants = %d/%d/%d, want 3/3/3", sessions.calls, executor.calls, len(grants.issued))
	}

	grantIDs := make(map[domaingrant.ID]bool, 3)
	sessionIDs := make(map[domaingrant.SessionID]bool, 3)
	bindings := make(map[domaingrant.Binding]bool, 3)
	for i := range 3 {
		if grantIDs[grants.issued[i].ID] {
			t.Fatalf("grant %s was reused", grants.issued[i].ID)
		}
		grantIDs[grants.issued[i].ID] = true
		if sessionIDs[sessions.created[i].ID] {
			t.Fatalf("session %q was reused", sessions.created[i].ID)
		}
		sessionIDs[sessions.created[i].ID] = true
		if bindings[grants.bindings[i]] {
			t.Fatalf("binding %s was reused", grants.bindings[i])
		}
		bindings[grants.bindings[i]] = true

		request := executor.requests[i]
		if request.Session != sessions.created[i].ID {
			t.Errorf("execution %d session = %q, want the bound %q", i, request.Session, sessions.created[i].ID)
		}
		if request.Model != units[i].Model {
			t.Errorf("execution %d model = %q, want the bound %q", i, request.Model, units[i].Model)
		}
		if request.SessionModel != sessions.created[i].Model {
			t.Errorf("execution %d observed model = %q, want %q", i, request.SessionModel, sessions.created[i].Model)
		}
		if result.Attempts[i].Grant != grants.issued[i].ID {
			t.Errorf("attempt %d grant = %s, want %s", i, result.Attempts[i].Grant, grants.issued[i].ID)
		}
		if result.Attempts[i].Session != sessions.created[i].ID {
			t.Errorf("attempt %d session = %q, want %q", i, result.Attempts[i].Session, sessions.created[i].ID)
		}
		if result.Attempts[i].Draw != units[i].Draw || result.Attempts[i].Part != units[i].Part {
			t.Errorf("attempt %d draw/part = %d/%d, want %d/%d", i, result.Attempts[i].Draw, result.Attempts[i].Part, units[i].Draw, units[i].Part)
		}
	}

	// Each process consumes its authority before the reviewer runs and completes
	// it after: the order is the contract that a refusal cannot start a reviewer.
	want := []string{"freeze"}
	for range 3 {
		want = append(want, "issue", "claim", "create-session", "bind", "execute", "complete")
	}
	want = append(want, "discard")
	if got := log.snapshot(); !slices.Equal(got, want) {
		t.Fatalf("boundary order = %v, want %v", got, want)
	}
}

// TestManagedReviewBindsTheCanonicalCandidateAndUnit: each grant names the tree,
// base, role, runtime, model, lenses, draw and part of the process that consumes
// it.
func TestManagedReviewBindsTheCanonicalCandidateAndUnit(t *testing.T) {
	log := &reviewEvents{}
	deps, _, grants, _, _ := reviewDeps(log)
	units := []review.ManagedUnit{
		managedUnit("reliability", "opencode/model-a", 1, 2, 1, 1),
		managedUnit("risk", "opencode/model-b", 2, 2, 2, 2),
	}

	if result := review.RunManagedReview(t.Context(), managedPlan(units...), deps); result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	for i, unit := range units {
		want, err := review.ReviewBinding(review.ReviewBindingFields{
			Tree:    currentTree,
			Base:    baseTree,
			Runtime: "opencode",
			Model:   unit.Model,
			Lenses:  unit.Lenses,
			Draw:    unit.Draw,
			Draws:   unit.Draws,
			Part:    unit.Part,
			Parts:   unit.Parts,
		})
		if err != nil {
			t.Fatalf("ReviewBinding: %v", err)
		}
		if grants.bindings[i] != want {
			t.Errorf("unit %d binding = %s, want %s", i, grants.bindings[i], want)
		}
	}
	if review.ReviewRole != "reviewer" {
		t.Fatalf("ReviewRole = %q, want reviewer", review.ReviewRole)
	}
}

// TestReviewBindingChangesWithEveryCanonicalField: no canonical field is
// silently dropped, and lenses are a set: reordering them keeps the digest.
func TestReviewBindingChangesWithEveryCanonicalField(t *testing.T) {
	base := review.ReviewBindingFields{
		Tree:    currentTree,
		Base:    baseTree,
		Runtime: "opencode",
		Model:   "opencode/model",
		Lenses:  []string{"risk", "reliability"},
		Draw:    1,
		Draws:   2,
		Part:    1,
		Parts:   2,
	}
	baseline, err := review.ReviewBinding(base)
	if err != nil {
		t.Fatalf("ReviewBinding: %v", err)
	}

	change := func(mutate func(*review.ReviewBindingFields)) review.ReviewBindingFields {
		fields := base
		fields.Lenses = slices.Clone(base.Lenses)
		mutate(&fields)
		return fields
	}
	variants := map[string]review.ReviewBindingFields{
		"tree":    change(func(f *review.ReviewBindingFields) { f.Tree = otherTree }),
		"base":    change(func(f *review.ReviewBindingFields) { f.Base = "c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3" }),
		"runtime": change(func(f *review.ReviewBindingFields) { f.Runtime = "another-runtime" }),
		"model":   change(func(f *review.ReviewBindingFields) { f.Model = "opencode/other-model" }),
		"lenses":  change(func(f *review.ReviewBindingFields) { f.Lenses = []string{"risk"} }),
		"draw":    change(func(f *review.ReviewBindingFields) { f.Draw = 2 }),
		"draws":   change(func(f *review.ReviewBindingFields) { f.Draws = 3 }),
		"part":    change(func(f *review.ReviewBindingFields) { f.Part = 2 }),
		"parts":   change(func(f *review.ReviewBindingFields) { f.Parts = 3 }),
	}
	for name, fields := range variants {
		t.Run(name, func(t *testing.T) {
			got, err := review.ReviewBinding(fields)
			if err != nil {
				t.Fatalf("ReviewBinding: %v", err)
			}
			if got == baseline {
				t.Fatalf("changing %s did not change the binding", name)
			}
		})
	}

	reordered := change(func(f *review.ReviewBindingFields) {
		f.Lenses = []string{"reliability", "risk"}
	})
	got, err := review.ReviewBinding(reordered)
	if err != nil {
		t.Fatalf("ReviewBinding: %v", err)
	}
	if got != baseline {
		t.Fatal("reordering the lens set changed the binding")
	}
}

// TestReviewBindingRefusesAnIncompleteCandidate: an unbound review is not a
// canonical one, so a missing tree, base, runtime, model, lens, draw or part is
// refused before a grant can name it.
func TestReviewBindingRefusesAnIncompleteCandidate(t *testing.T) {
	base := review.ReviewBindingFields{
		Tree:    currentTree,
		Base:    baseTree,
		Runtime: "opencode",
		Model:   "opencode/model",
		Lenses:  []string{"risk"},
		Draw:    1,
		Draws:   1,
		Part:    1,
		Parts:   1,
	}
	cases := map[string]review.ReviewBindingFields{
		"tree":       func() review.ReviewBindingFields { f := base; f.Tree = ""; return f }(),
		"base":       func() review.ReviewBindingFields { f := base; f.Base = ""; return f }(),
		"runtime":    func() review.ReviewBindingFields { f := base; f.Runtime = ""; return f }(),
		"model":      func() review.ReviewBindingFields { f := base; f.Model = ""; return f }(),
		"no lens":    func() review.ReviewBindingFields { f := base; f.Lenses = nil; return f }(),
		"empty lens": func() review.ReviewBindingFields { f := base; f.Lenses = []string{""}; return f }(),
		"draw":       func() review.ReviewBindingFields { f := base; f.Draw = 0; return f }(),
		"draws":      func() review.ReviewBindingFields { f := base; f.Draws = 0; return f }(),
		"part":       func() review.ReviewBindingFields { f := base; f.Part = 0; return f }(),
		"parts":      func() review.ReviewBindingFields { f := base; f.Parts = 0; return f }(),
	}
	for name, fields := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := review.ReviewBinding(fields); !errors.Is(err, domaingrant.ErrInvalidBinding) {
				t.Fatalf("ReviewBinding error = %v, want ErrInvalidBinding", err)
			}
		})
	}
}

// TestManagedReviewRefusesAnUnusableSessionBeforeReviewerExecution: a missing,
// malformed, foreign, duplicate, empty or model-less identity stops before the
// reviewer runs and terminates the claimed grant instead of completing it.
func TestManagedReviewRefusesAnUnusableSessionBeforeReviewerExecution(t *testing.T) {
	cases := map[string]struct {
		units       int
		configure   func(*fakeReviewSessions)
		executions  int
		attempts    int
		completions int
	}{
		"missing": {
			units:     1,
			configure: func(s *fakeReviewSessions) { s.err = errors.New("session identity is missing") },
		},
		"malformed": {
			units:     1,
			configure: func(s *fakeReviewSessions) { s.err = errors.New("session identity is malformed") },
		},
		"foreign": {
			units:     1,
			configure: func(s *fakeReviewSessions) { s.err = errors.New("session identity is not a fresh OpenCode session") },
		},
		"empty id": {
			units:     1,
			configure: func(s *fakeReviewSessions) { s.responses = []review.ReviewSession{{Model: "opencode/model"}} },
		},
		"empty model": {
			units:     1,
			configure: func(s *fakeReviewSessions) { s.responses = []review.ReviewSession{{ID: "ses_review_one"}} },
		},
		"malformed grammar": {
			units: 1,
			configure: func(s *fakeReviewSessions) {
				s.responses = []review.ReviewSession{{ID: "ses review", Model: "opencode/model"}}
			},
		},
		"duplicate": {
			units: 2,
			configure: func(s *fakeReviewSessions) {
				s.allowDuplicate = true
				s.responses = []review.ReviewSession{
					{ID: "ses_review_dup", Model: "opencode/model"},
					{ID: "ses_review_dup", Model: "opencode/model"},
				}
			},
			executions:  1,
			attempts:    1,
			completions: 1,
		},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			log := &reviewEvents{}
			deps, _, grants, sessions, executor := reviewDeps(log)
			testCase.configure(sessions)
			units := make([]review.ManagedUnit, 0, testCase.units)
			for i := range testCase.units {
				units = append(units, managedUnit("reliability", "opencode/model", 1, 1, i+1, testCase.units))
			}

			result := review.RunManagedReview(t.Context(), managedPlan(units...), deps)
			if result.Code != review.CodeBlocked {
				t.Fatalf("an unusable session produced %+v", result)
			}
			if !strings.Contains(strings.ToLower(result.Message), "session") {
				t.Errorf("message %q does not name the unusable session", result.Message)
			}
			if executor.calls != testCase.executions {
				t.Fatalf("executions = %d, want %d", executor.calls, testCase.executions)
			}
			if len(result.Attempts) != testCase.attempts {
				t.Fatalf("attempts = %d, want %d", len(result.Attempts), testCase.attempts)
			}
			if grants.completes != testCase.completions {
				t.Fatalf("completions = %d, want %d", grants.completes, testCase.completions)
			}
			if grants.failures != 1 {
				t.Fatalf("grant failures = %d, want 1", grants.failures)
			}
			if testCase.attempts == 0 && (result.Managed || result.Fresh) {
				t.Fatal("a refused session claimed freshness")
			}
		})
	}
}

// TestManagedReviewRefusesAReplayedGrantBeforeASecondSession: a consumed grant
// cannot open a second reviewer, and the refusal claims no second session.
func TestManagedReviewRefusesAReplayedGrantBeforeASecondSession(t *testing.T) {
	log := &reviewEvents{}
	deps, _, grants, sessions, executor := reviewDeps(log)
	grants.replay = true
	plan := managedPlan(
		managedUnit("reliability", "opencode/model-a", 1, 1, 1, 1),
		managedUnit("risk", "opencode/model-b", 1, 1, 1, 2),
	)

	result := review.RunManagedReview(t.Context(), plan, deps)
	if result.Code != review.CodeBlocked {
		t.Fatalf("a replayed grant produced %+v", result)
	}
	if !strings.Contains(strings.ToLower(result.Message), "claim") {
		t.Errorf("message %q does not name the refused claim", result.Message)
	}
	if sessions.calls != 1 {
		t.Fatalf("session creations = %d, want 1", sessions.calls)
	}
	if executor.calls != 1 {
		t.Fatalf("executions = %d, want 1", executor.calls)
	}
	if len(result.Attempts) != 1 {
		t.Fatalf("attempts = %d, want the one that landed", len(result.Attempts))
	}
	if grants.failures != 0 {
		t.Fatalf("grant failures = %d, want 0: a grant this run never claimed is not its to fail", grants.failures)
	}
}

// TestManagedReviewKeepsTheAttemptWhenTheReviewerFails: after the fresh session
// is bound, a reviewer that fails or exits non-zero is real evidence of a fresh
// attempt, but it is blocked, failed and never completed.
func TestManagedReviewKeepsTheAttemptWhenTheReviewerFails(t *testing.T) {
	cases := map[string]struct {
		configure func(*fakeReviewExecutor)
		code      string
	}{
		"error": {
			configure: func(e *fakeReviewExecutor) {
				e.errAt = 1
				e.err = errors.New("reviewer crashed")
				e.output = "half an answer"
			},
		},
		"non-zero exit": {
			configure: func(e *fakeReviewExecutor) { e.exitAt = 1; e.exitCode = 2; e.output = "reviewer exploded" },
			code:      "2",
		},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			log := &reviewEvents{}
			deps, _, grants, sessions, executor := reviewDeps(log)
			testCase.configure(executor)

			result := review.RunManagedReview(t.Context(), managedPlan(managedUnit("risk", "opencode/model", 1, 1, 1, 1)), deps)
			if result.Code != review.CodeBlocked {
				t.Fatalf("a failed reviewer produced %+v", result)
			}
			if !result.Managed || !result.Fresh {
				t.Fatalf("managed/fresh = %v/%v, want true/true after a bind", result.Managed, result.Fresh)
			}
			if len(result.Attempts) != 1 {
				t.Fatalf("attempts = %d, want the bound attempt", len(result.Attempts))
			}
			if result.Attempts[0].Session != sessions.created[0].ID {
				t.Fatalf("attempt session = %q, want %q", result.Attempts[0].Session, sessions.created[0].ID)
			}
			if grants.completes != 0 {
				t.Fatalf("completions = %d, want 0", grants.completes)
			}
			if grants.failures != 1 {
				t.Fatalf("grant failures = %d, want 1", grants.failures)
			}
			if result.Output != executor.output {
				t.Errorf("output = %q, want %q", result.Output, executor.output)
			}
			if testCase.code != "" && !strings.Contains(result.Message, testCase.code) {
				t.Errorf("message %q does not name exit code %s", result.Message, testCase.code)
			}
			if !strings.Contains(result.Message, "risk") {
				t.Errorf("message %q does not name the failed lens", result.Message)
			}
		})
	}
}

// TestManagedReviewRefusesACompletionThatNamesAnotherGrant: a completed answer
// must be this process's own grant in its terminal state. A service that names
// another authority leaves the unit blocked and the grant failed.
func TestManagedReviewRefusesACompletionThatNamesAnotherGrant(t *testing.T) {
	log := &reviewEvents{}
	deps, _, grants, sessions, executor := reviewDeps(log)
	binding, err := review.ReviewBinding(review.ReviewBindingFields{
		Tree:    currentTree,
		Base:    baseTree,
		Runtime: "opencode",
		Model:   "opencode/model",
		Lenses:  []string{"reliability"},
		Draw:    1,
		Draws:   1,
		Part:    1,
		Parts:   1,
	})
	if err != nil {
		t.Fatalf("ReviewBinding: %v", err)
	}
	stray, err := domaingrant.NewGrant(binding)
	if err != nil {
		t.Fatalf("NewGrant: %v", err)
	}
	grants.completeGrant = &stray

	result := review.RunManagedReview(t.Context(), managedPlan(managedUnit("reliability", "opencode/model", 1, 1, 1, 1)), deps)
	if result.Code != review.CodeBlocked {
		t.Fatalf("a foreign completion produced %+v", result)
	}
	if !strings.Contains(strings.ToLower(result.Message), "complete") {
		t.Errorf("message %q does not name the refused completion", result.Message)
	}
	if executor.calls != 1 || sessions.calls != 1 {
		t.Fatalf("executions/sessions = %d/%d, want 1/1", executor.calls, sessions.calls)
	}
	if len(result.Attempts) != 1 {
		t.Fatalf("attempts = %d, want the bound attempt", len(result.Attempts))
	}
	if grants.failures != 1 {
		t.Fatalf("grant failures = %d, want 1", grants.failures)
	}
}

// TestManagedReviewRefusesAFreezeFailureBeforeAnyAuthority: a tree that cannot
// be frozen starts no reviewer and creates no session.
func TestManagedReviewRefusesAFreezeFailureBeforeAnyAuthority(t *testing.T) {
	log := &reviewEvents{}
	deps, freezer, grants, sessions, executor := reviewDeps(log)
	freezer.err = errors.New("git archive failed")

	result := review.RunManagedReview(t.Context(), managedPlan(managedUnit("risk", "opencode/model", 1, 1, 1, 1)), deps)
	if result.Code != review.CodeBlocked {
		t.Fatalf("an unfreezable candidate produced %+v", result)
	}
	if !strings.Contains(strings.ToLower(result.Message), "freeze") {
		t.Errorf("message %q does not name the failed freeze", result.Message)
	}
	if len(grants.bindings) != 0 || sessions.calls != 0 || executor.calls != 0 {
		t.Fatal("an unfreezable candidate reached a managed boundary")
	}
	if result.Managed || result.Fresh || len(result.Attempts) != 0 {
		t.Fatal("an unfreezable candidate claimed freshness")
	}
}

// TestManagedReviewRefusesMissingBoundaries: no freezer, no grant service, no
// fresh runtime and no executor means no review.
func TestManagedReviewRefusesMissingBoundaries(t *testing.T) {
	cases := map[string]func(*review.ManagedReviewDeps){
		"freezer":  func(d *review.ManagedReviewDeps) { d.Freezer = nil },
		"grants":   func(d *review.ManagedReviewDeps) { d.Grants = nil },
		"sessions": func(d *review.ManagedReviewDeps) { d.Sessions = nil },
		"executor": func(d *review.ManagedReviewDeps) { d.Executor = nil },
	}
	for name, clear := range cases {
		t.Run(name, func(t *testing.T) {
			log := &reviewEvents{}
			deps, freezer, _, sessions, executor := reviewDeps(log)
			clear(&deps)

			result := review.RunManagedReview(t.Context(), managedPlan(managedUnit("risk", "opencode/model", 1, 1, 1, 1)), deps)
			if result.Code != review.CodeBlocked {
				t.Fatalf("an unwired review produced %+v", result)
			}
			if freezer.calls != 0 || sessions.calls != 0 || executor.calls != 0 {
				t.Fatal("an unwired review reached a boundary")
			}
			if result.Managed || result.Fresh || len(result.Attempts) != 0 {
				t.Fatal("an unwired review claimed freshness")
			}
		})
	}
}

// TestManagedReviewExecutionExposesNoResumeInput: the runtime surface has one
// operation that takes no session identity, and the execution request carries no
// resume, continue, fork or caller-supplied parent session.
func TestManagedReviewExecutionExposesNoResumeInput(t *testing.T) {
	typ := reflect.TypeFor[review.ReviewSessions]()
	if typ.NumMethod() != 1 {
		t.Fatalf("ReviewSessions has %d methods, want 1", typ.NumMethod())
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

	request := reflect.TypeFor[review.ReviewExecutionRequest]()
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

// TestLegacyExecPathMakesNoFreshnessClaim: the unmanaged command path runs every
// unit without a grant or a session and says so.
func TestLegacyExecPathMakesNoFreshnessClaim(t *testing.T) {
	legacy := &fakeLegacyReviewExecutor{output: "legacy-output"}
	units := []review.ManagedUnit{
		managedUnit("reliability", "opencode/model-a", 1, 1, 1, 1),
		managedUnit("risk", "opencode/model-b", 1, 1, 1, 1),
	}

	result := review.RunUnmanagedReview(t.Context(), "agent --prompt", units, legacy)
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if result.Managed || result.Fresh || len(result.Attempts) != 0 {
		t.Fatal("the unmanaged path claimed freshness")
	}
	if legacy.calls != 2 {
		t.Fatalf("legacy runs = %d, want 2", legacy.calls)
	}
	if legacy.command != "agent --prompt" {
		t.Fatalf("legacy command = %q", legacy.command)
	}
	if len(legacy.prompts) != 2 || legacy.prompts[0] != units[0].Prompt || legacy.prompts[1] != units[1].Prompt {
		t.Fatalf("legacy prompts = %v, want each unit's prompt", legacy.prompts)
	}
	if result.Output != "legacy-output" {
		t.Fatalf("output = %q, want legacy-output", result.Output)
	}
}

// TestLegacyNonZeroExitIsBlockedWithoutFreshness: the unmanaged path reads the
// exit code, reports the output and still claims no authority.
func TestLegacyNonZeroExitIsBlockedWithoutFreshness(t *testing.T) {
	legacy := &fakeLegacyReviewExecutor{output: "legacy-exploded", exitCode: 3}

	result := review.RunUnmanagedReview(t.Context(), "agent", []review.ManagedUnit{
		managedUnit("risk", "opencode/model", 1, 1, 1, 1),
	}, legacy)
	if result.Code != review.CodeBlocked {
		t.Fatalf("a non-zero legacy exit produced %+v", result)
	}
	if result.Managed || result.Fresh || len(result.Attempts) != 0 {
		t.Fatal("the unmanaged path claimed freshness")
	}
	if result.Output != "legacy-exploded" {
		t.Fatalf("output = %q, want the legacy output", result.Output)
	}
	if !strings.Contains(result.Message, fmt.Sprintf("%d", 3)) {
		t.Fatalf("message %q does not name the exit code", result.Message)
	}
	if legacy.calls != 1 {
		t.Fatalf("legacy runs = %d, want 1", legacy.calls)
	}
}
