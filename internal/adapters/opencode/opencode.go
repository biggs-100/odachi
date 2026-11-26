// Package opencode is the fresh-only OpenCode runtime edge. Every managed
// invocation creates exactly one new session through the documented V2 HTTP API
// and reports the observed session identity and model; it offers no resume,
// continue or fork operation, and no caller-supplied session value reaches
// OpenCode.
package opencode

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	grant "github.com/biggs-100/odachi/internal/domain/grant"
)

// Typed failures at the runtime boundary. A caller decides what each one means
// for the grant; the adapter never stores authority.
var (
	ErrCanceled         = errors.New("opencode: session creation was canceled")
	ErrTimeout          = errors.New("opencode: session creation timed out")
	ErrExit             = errors.New("opencode: opencode exited non-zero")
	ErrRequest          = errors.New("opencode: could not run opencode")
	ErrMalformedOutput  = errors.New("opencode: malformed session response")
	ErrMissingSession   = errors.New("opencode: session identity is missing")
	ErrMalformedSession = errors.New("opencode: session identity is malformed")
	ErrDuplicateSession = errors.New("opencode: duplicate session identity")
	ErrForeignSession   = errors.New("opencode: session identity is not a fresh OpenCode session")
	ErrMissingModel     = errors.New("opencode: observed model is missing")
	ErrMalformedModel   = errors.New("opencode: observed model is malformed")
)

const (
	// The fresh-session request is a constant. It has no identity field, so a
	// caller cannot inject a session, continue, fork or resume value: OpenCode
	// always mints a new identity for POST /api/session.
	createMethod = "POST"
	createPath   = "/api/session"
	createBody   = "{}"

	// maxSessionLength is the T1 domain's byte bound for a session identity.
	maxSessionLength = 256
	// maxModelPartLength bounds each observed model identifier.
	maxModelPartLength = 256
	// openCodeSessionPrefix is the documented pattern for an OpenCode session ID.
	openCodeSessionPrefix = "ses"
)

// Model is the identity OpenCode observed for the session's active model.
type Model struct {
	Provider string
	ID       string
}

// String spells the observed model the way a grant records it.
func (model Model) String() string {
	if model.Provider == "" {
		return model.ID
	}
	return model.Provider + "/" + model.ID
}

// Session is the fresh runtime fact the adapter reports to its caller.
type Session struct {
	ID    grant.SessionID
	Model Model
}

// Adapter creates fresh OpenCode sessions over the injected runner boundary.
type Adapter struct {
	runner Runner
	mu     sync.Mutex
	seen   map[string]struct{}
}

// New builds an adapter over the supplied runner. The runner is the only edge
// that reaches OpenCode, so tests inject a fake and never spawn a live model.
func New(runner Runner) *Adapter {
	return &Adapter{runner: runner, seen: make(map[string]struct{})}
}

// CreateFreshSession is the whole managed runtime surface. It takes no session,
// continue, fork or resume value: the request is a fixed POST /api/session with
// an empty body, so OpenCode must mint a new session identity every time. The
// reported identity and model are the observed response facts, validated before
// they are returned; a duplicate identity is refused instead of reported twice.
func (adapter *Adapter) CreateFreshSession(ctx context.Context) (Session, error) {
	if adapter.runner == nil {
		return Session{}, fmt.Errorf("%w: no runtime runner is configured", ErrRequest)
	}
	if err := ctx.Err(); err != nil {
		return Session{}, classifyContext(err)
	}

	out, err := adapter.runner.Run(ctx, Request{Method: createMethod, Path: createPath, Body: createBody})
	if err != nil {
		return Session{}, classifyRunError(ctx, err)
	}
	if err := ctx.Err(); err != nil {
		return Session{}, classifyContext(err)
	}

	session, err := parseSession(out)
	if err != nil {
		return Session{}, err
	}
	if err := adapter.record(session.ID); err != nil {
		return Session{}, err
	}
	return session, nil
}

// record refuses to report the same session identity twice. Each managed
// invocation must own a distinct fresh session; a repeated identity means the
// runtime did not create a new one.
func (adapter *Adapter) record(id grant.SessionID) error {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()
	if _, duplicate := adapter.seen[string(id)]; duplicate {
		return fmt.Errorf("%w: %s", ErrDuplicateSession, id)
	}
	adapter.seen[string(id)] = struct{}{}
	return nil
}

// classifyContext maps a canceled or expired context onto its typed failure.
func classifyContext(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%w: %w", ErrTimeout, err)
	}
	return fmt.Errorf("%w: %w", ErrCanceled, err)
}

// classifyRunError maps a runner failure onto its typed failure.
func classifyRunError(ctx context.Context, err error) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return classifyContext(ctxErr)
	}
	if exit, ok := errors.AsType[*ExitError](err); ok {
		return fmt.Errorf("%w: %w", ErrExit, exit)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%w: %w", ErrTimeout, err)
	}
	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("%w: %w", ErrCanceled, err)
	}
	return fmt.Errorf("%w: %w", ErrRequest, err)
}

// responseEnvelope is the documented POST /api/session response shape:
// {"data": Session.Info}.
type responseEnvelope struct {
	Data *sessionPayload `json:"data"`
}

type sessionPayload struct {
	ID       *string   `json:"id"`
	ParentID *string   `json:"parentID"`
	Fork     *forkRef  `json:"fork"`
	Model    *modelRef `json:"model"`
}

type forkRef struct {
	SessionID *string `json:"sessionID"`
}

type modelRef struct {
	ProviderID *string `json:"providerID"`
	ID         *string `json:"id"`
	Variant    *string `json:"variant"`
}

// parseSession validates one response completely before any fact is returned.
// Missing, malformed, duplicate and foreign identities fail closed, in that
// order, and the observed model is required.
func parseSession(out []byte) (Session, error) {
	if err := checkStrictJSON(out); err != nil {
		return Session{}, err
	}

	var envelope responseEnvelope
	if err := json.Unmarshal(out, &envelope); err != nil {
		return Session{}, fmt.Errorf("%w: %w", ErrMalformedOutput, err)
	}
	if envelope.Data == nil || envelope.Data.ID == nil || *envelope.Data.ID == "" {
		return Session{}, ErrMissingSession
	}
	if envelope.Data.ParentID != nil || envelope.Data.Fork != nil {
		return Session{}, ErrForeignSession
	}

	id := grant.SessionID(*envelope.Data.ID)
	if !validSessionID(id) {
		return Session{}, ErrMalformedSession
	}
	if !strings.HasPrefix(string(id), openCodeSessionPrefix) {
		return Session{}, ErrForeignSession
	}

	model, err := parseModel(envelope.Data.Model)
	if err != nil {
		return Session{}, err
	}
	return Session{ID: id, Model: model}, nil
}

// parseModel validates the observed model as non-empty safe identifiers.
func parseModel(ref *modelRef) (Model, error) {
	if ref == nil {
		return Model{}, ErrMissingModel
	}
	if ref.ProviderID == nil || ref.ID == nil {
		return Model{}, ErrMalformedModel
	}
	provider, id := *ref.ProviderID, *ref.ID
	if !validIdentifier(provider) || !validIdentifier(id) {
		return Model{}, ErrMalformedModel
	}
	if ref.Variant != nil && !validIdentifier(*ref.Variant) {
		return Model{}, ErrMalformedModel
	}
	return Model{Provider: provider, ID: id}, nil
}

// checkStrictJSON rejects a response with duplicate object keys or trailing data.
// encoding/json silently keeps the last duplicate, which would let an ambiguous
// identity through; the token walk refuses it before decoding.
func checkStrictJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := walkJSON(decoder); err != nil {
		return fmt.Errorf("%w: %w", ErrMalformedOutput, err)
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: trailing data after the session response", ErrMalformedOutput)
	}
	return nil
}

func walkJSON(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := make(map[string]struct{})
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return errors.New("object key is not a string")
			}
			if _, duplicate := seen[key]; duplicate {
				return errors.New("duplicate object key")
			}
			seen[key] = struct{}{}
			if err := walkJSON(decoder); err != nil {
				return err
			}
		}
		_, err := decoder.Token()
		return err
	case '[':
		for decoder.More() {
			if err := walkJSON(decoder); err != nil {
				return err
			}
		}
		_, err := decoder.Token()
		return err
	default:
		return nil
	}
}

// validSessionID mirrors the T1 domain grammar. It exists so the adapter never
// reports an identity the grant domain would refuse; TestSessionGrammarMatches-
// GrantDomain pins the two grammars together through Grant.BindSession.
func validSessionID(id grant.SessionID) bool {
	if len(id) == 0 || len(id) > maxSessionLength {
		return false
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-', r == '_', r == '.', r == ':':
		default:
			return false
		}
	}
	return true
}

// validIdentifier bounds an observed model part to a safe ASCII identifier.
func validIdentifier(value string) bool {
	if len(value) == 0 || len(value) > maxModelPartLength {
		return false
	}
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-', r == '_', r == '.', r == ':', r == '/':
		default:
			return false
		}
	}
	return true
}
