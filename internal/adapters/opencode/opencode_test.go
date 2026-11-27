package opencode

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	grant "github.com/biggs-100/odachi/internal/domain/grant"
)

// fakeRunner is the injected runtime edge: it records every request and replays
// canned output. Ordinary tests never spawn OpenCode or call a live model.
type fakeRunner struct {
	mu        sync.Mutex
	requests  []Request
	responses []fakeResponse
	calls     int
}

type fakeResponse struct {
	out []byte
	err error
}

func (fake *fakeRunner) Run(_ context.Context, request Request) ([]byte, error) {
	fake.mu.Lock()
	defer fake.mu.Unlock()
	fake.requests = append(fake.requests, request)
	var response fakeResponse
	if fake.calls < len(fake.responses) {
		response = fake.responses[fake.calls]
	}
	fake.calls++
	return response.out, response.err
}

func (fake *fakeRunner) recorded() []Request {
	fake.mu.Lock()
	defer fake.mu.Unlock()
	return slices.Clone(fake.requests)
}

func sessionJSON(id string) []byte {
	return []byte(`{"data":{"id":` + strconv.Quote(id) + `,"model":{"providerID":"anthropic","id":"claude-sonnet-4"}}}`)
}

func TestCreateFreshSessionReturnsObservedIdentityAndModel(t *testing.T) {
	runner := &fakeRunner{responses: []fakeResponse{{out: sessionJSON("ses_fresh_001")}}}
	adapter := New(runner)

	session, err := adapter.CreateFreshSession(t.Context())
	if err != nil {
		t.Fatalf("CreateFreshSession() error = %v", err)
	}
	if session.ID != grant.SessionID("ses_fresh_001") {
		t.Fatalf("session.ID = %q, want ses_fresh_001", session.ID)
	}
	if session.Model.String() != "anthropic/claude-sonnet-4" {
		t.Fatalf("session.Model = %q, want anthropic/claude-sonnet-4", session.Model)
	}

	requests := runner.recorded()
	if len(requests) != 1 {
		t.Fatalf("runner calls = %d, want 1", len(requests))
	}
	request := requests[0]
	if request.Method != "POST" || request.Path != "/api/session" {
		t.Fatalf("request = %s %s, want POST /api/session", request.Method, request.Path)
	}
	if request.Body != "{}" {
		t.Fatalf("request body = %q, want a fixed empty object", request.Body)
	}
	for _, forbidden := range []string{"resume", "continue", "fork", "parentID", "session", `"id"`} {
		if strings.Contains(request.Body, forbidden) {
			t.Fatalf("fresh request body carries %q: %s", forbidden, request.Body)
		}
	}
}

func TestCreateFreshSessionIssuesANewSessionPerInvocation(t *testing.T) {
	runner := &fakeRunner{responses: []fakeResponse{
		{out: sessionJSON("ses_first")},
		{out: sessionJSON("ses_second")},
	}}
	adapter := New(runner)

	first, err := adapter.CreateFreshSession(t.Context())
	if err != nil {
		t.Fatalf("first CreateFreshSession() error = %v", err)
	}
	second, err := adapter.CreateFreshSession(t.Context())
	if err != nil {
		t.Fatalf("second CreateFreshSession() error = %v", err)
	}
	if first.ID == second.ID {
		t.Fatalf("both invocations reported session %q", first.ID)
	}
	if got := len(runner.recorded()); got != 2 {
		t.Fatalf("runner calls = %d, want 2", got)
	}
}

func TestCreateFreshSessionFailsClosedOnMissingSession(t *testing.T) {
	cases := map[string][]byte{
		"no envelope": []byte(`{}`),
		"null data":   []byte(`{"data":null}`),
		"no id":       []byte(`{"data":{"model":{"providerID":"p","id":"m"}}}`),
		"null id":     []byte(`{"data":{"id":null,"model":{"providerID":"p","id":"m"}}}`),
		"empty id":    []byte(`{"data":{"id":"","model":{"providerID":"p","id":"m"}}}`),
	}
	for name, out := range cases {
		t.Run(name, func(t *testing.T) {
			adapter := New(&fakeRunner{responses: []fakeResponse{{out: out}}})
			if _, err := adapter.CreateFreshSession(t.Context()); !errors.Is(err, ErrMissingSession) {
				t.Fatalf("CreateFreshSession() error = %v, want ErrMissingSession", err)
			}
		})
	}
}

func TestCreateFreshSessionFailsClosedOnMalformedSession(t *testing.T) {
	cases := map[string][]byte{
		"space":    sessionJSON("ses bad"),
		"newline":  sessionJSON("ses\n"),
		"unicode":  sessionJSON("sesión"),
		"too long": sessionJSON(strings.Repeat("s", 257)),
		"slash":    sessionJSON("ses/child"),
	}
	for name, out := range cases {
		t.Run(name, func(t *testing.T) {
			adapter := New(&fakeRunner{responses: []fakeResponse{{out: out}}})
			if _, err := adapter.CreateFreshSession(t.Context()); !errors.Is(err, ErrMalformedSession) {
				t.Fatalf("CreateFreshSession() error = %v, want ErrMalformedSession", err)
			}
		})
	}
}

func TestCreateFreshSessionFailsClosedOnMalformedOutput(t *testing.T) {
	cases := map[string][]byte{
		"not json":      []byte(`not-json`),
		"truncated":     []byte(`{"data":`),
		"trailing data": []byte(`{"data":{"id":"ses_a"}}{"data":{}}`),
		"duplicate key": []byte(`{"data":{"id":"ses_a","id":"ses_b","model":{"providerID":"p","id":"m"}}}`),
	}
	for name, out := range cases {
		t.Run(name, func(t *testing.T) {
			adapter := New(&fakeRunner{responses: []fakeResponse{{out: out}}})
			if _, err := adapter.CreateFreshSession(t.Context()); !errors.Is(err, ErrMalformedOutput) {
				t.Fatalf("CreateFreshSession() error = %v, want ErrMalformedOutput", err)
			}
		})
	}
}

func TestCreateFreshSessionFailsClosedOnForeignSession(t *testing.T) {
	cases := map[string][]byte{
		"not opencode": []byte(`{"data":{"id":"abc123","model":{"providerID":"p","id":"m"}}}`),
		"child":        []byte(`{"data":{"id":"ses_child","parentID":"ses_parent","model":{"providerID":"p","id":"m"}}}`),
		"fork":         []byte(`{"data":{"id":"ses_fork","fork":{"sessionID":"ses_parent","boundary":{}},"model":{"providerID":"p","id":"m"}}}`),
	}
	for name, out := range cases {
		t.Run(name, func(t *testing.T) {
			adapter := New(&fakeRunner{responses: []fakeResponse{{out: out}}})
			if _, err := adapter.CreateFreshSession(t.Context()); !errors.Is(err, ErrForeignSession) {
				t.Fatalf("CreateFreshSession() error = %v, want ErrForeignSession", err)
			}
		})
	}
}

func TestCreateFreshSessionFailsClosedOnDuplicateSessionIdentity(t *testing.T) {
	runner := &fakeRunner{responses: []fakeResponse{
		{out: sessionJSON("ses_reused")},
		{out: sessionJSON("ses_reused")},
	}}
	adapter := New(runner)

	if _, err := adapter.CreateFreshSession(t.Context()); err != nil {
		t.Fatalf("first CreateFreshSession() error = %v", err)
	}
	if _, err := adapter.CreateFreshSession(t.Context()); !errors.Is(err, ErrDuplicateSession) {
		t.Fatalf("second CreateFreshSession() error = %v, want ErrDuplicateSession", err)
	}
}

func TestCreateFreshSessionFailsClosedOnMissingOrMalformedModel(t *testing.T) {
	cases := map[string]struct {
		out  []byte
		want error
	}{
		"missing model":    {[]byte(`{"data":{"id":"ses_001"}}`), ErrMissingModel},
		"missing provider": {[]byte(`{"data":{"id":"ses_002","model":{"id":"m"}}}`), ErrMalformedModel},
		"missing id":       {[]byte(`{"data":{"id":"ses_003","model":{"providerID":"p"}}}`), ErrMalformedModel},
		"unsafe provider":  {[]byte(`{"data":{"id":"ses_004","model":{"providerID":"p p","id":"m"}}}`), ErrMalformedModel},
		"empty model id":   {[]byte(`{"data":{"id":"ses_005","model":{"providerID":"p","id":""}}}`), ErrMalformedModel},
	}
	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			adapter := New(&fakeRunner{responses: []fakeResponse{{out: test.out}}})
			if _, err := adapter.CreateFreshSession(t.Context()); !errors.Is(err, test.want) {
				t.Fatalf("CreateFreshSession() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestCreateFreshSessionReportsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	adapter := New(&fakeRunner{responses: []fakeResponse{{out: sessionJSON("ses_canceled")}}})

	_, err := adapter.CreateFreshSession(ctx)
	if !errors.Is(err, ErrCanceled) {
		t.Fatalf("CreateFreshSession() error = %v, want ErrCanceled", err)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("CreateFreshSession() error = %v, want it to wrap context.Canceled", err)
	}
}

func TestCreateFreshSessionReportsTimeout(t *testing.T) {
	ctx, cancel := context.WithDeadline(t.Context(), time.Unix(0, 0))
	defer cancel()
	adapter := New(&fakeRunner{responses: []fakeResponse{{out: sessionJSON("ses_timeout")}}})

	_, err := adapter.CreateFreshSession(ctx)
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("CreateFreshSession() error = %v, want ErrTimeout", err)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("CreateFreshSession() error = %v, want it to wrap context.DeadlineExceeded", err)
	}
}

func TestCreateFreshSessionReportsNonZeroExit(t *testing.T) {
	adapter := New(&fakeRunner{responses: []fakeResponse{{err: &ExitError{Code: 7, Stderr: "boom"}}}})

	_, err := adapter.CreateFreshSession(t.Context())
	if !errors.Is(err, ErrExit) {
		t.Fatalf("CreateFreshSession() error = %v, want ErrExit", err)
	}
	exit, ok := errors.AsType[*ExitError](err)
	if !ok {
		t.Fatalf("CreateFreshSession() error = %v, want a *ExitError", err)
	}
	if exit.Code != 7 {
		t.Fatalf("exit code = %d, want 7", exit.Code)
	}
}

func TestCreateFreshSessionFailsClosedWithoutARunner(t *testing.T) {
	adapter := New(nil)
	if _, err := adapter.CreateFreshSession(t.Context()); !errors.Is(err, ErrRequest) {
		t.Fatalf("CreateFreshSession() error = %v, want ErrRequest", err)
	}
}

func TestExecRunnerBuildsOpenCodeAPIArgv(t *testing.T) {
	runner := ExecRunner{Binary: "opencode"}
	args := runner.args(Request{Method: "POST", Path: "/api/session", Body: "{}"})
	want := []string{"api", "post", "/api/session", "--data", "{}"}
	if !slices.Equal(args, want) {
		t.Fatalf("args = %q, want %q", args, want)
	}
	for _, forbidden := range []string{"resume", "continue", "fork", "--session"} {
		if slices.Contains(args, forbidden) {
			t.Fatalf("argv carries %q: %q", forbidden, args)
		}
	}
}

func TestExecRunnerUsesOpenCodeBinaryByDefault(t *testing.T) {
	if got := (ExecRunner{}).binary(); got != "opencode" {
		t.Fatalf("binary() = %q, want opencode", got)
	}
	if got := (ExecRunner{Binary: "custom-open"}).binary(); got != "custom-open" {
		t.Fatalf("binary() = %q, want custom-open", got)
	}
}

// TestSessionGrammarMatchesGrantDomain pins the adapter's mirrored grammar to
// the T1 domain: the adapter accepts exactly the identities Grant.BindSession
// accepts. If either grammar changes, this test fails before a session reaches
// the grant.
func TestSessionGrammarMatchesGrantDomain(t *testing.T) {
	binding, err := grant.ParseBinding(strings.Repeat("a", 64))
	if err != nil {
		t.Fatalf("ParseBinding() error = %v", err)
	}
	candidates := []grant.SessionID{
		"ses_fresh_001",
		"ses-abc.DEF:123",
		"ABC123",
		"",
		" ses",
		"ses ",
		"ses\n",
		grant.SessionID(strings.Repeat("s", 257)),
		"with/slash",
		"sesión",
		"ses_",
		"ses.a:b-c_d",
	}
	for _, candidate := range candidates {
		issued, err := grant.NewGrant(binding)
		if err != nil {
			t.Fatalf("NewGrant() error = %v", err)
		}
		claimed, err := issued.Claim()
		if err != nil {
			t.Fatalf("Claim() error = %v", err)
		}
		_, domainErr := claimed.BindSession(candidate)
		domainOK := domainErr == nil
		if adapterOK := validSessionID(candidate); adapterOK != domainOK {
			t.Errorf("validSessionID(%q) = %v, grant.BindSession = %v", candidate, adapterOK, domainOK)
		}
	}
}
