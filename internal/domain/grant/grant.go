// Package grant models one-shot authority for a managed runtime invocation.
package grant

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
)

const (
	idSize      = 32
	bindingSize = 32
	maxSession  = 256
)

var (
	ErrInvalidID         = errors.New("invalid grant ID")
	ErrInvalidBinding    = errors.New("invalid binding digest")
	ErrInvalidSession    = errors.New("invalid runtime session ID")
	ErrInvalidTransition = errors.New("invalid grant state transition")
	ErrAlreadyClaimed    = errors.New("grant already claimed")
	ErrAlreadyBound      = errors.New("grant already bound to a session")
	ErrTerminal          = errors.New("grant is terminal")
)

// ID is an opaque, canonical base64url-encoded 256-bit grant identity.
type ID [idSize]byte

// NewID returns a cryptographically random grant identity.
func NewID() (ID, error) {
	for {
		var id ID
		if _, err := rand.Read(id[:]); err != nil {
			return ID{}, fmt.Errorf("generate grant ID: %w", err)
		}
		if id.Validate() == nil {
			return id, nil
		}
	}
}

// ParseID accepts only the canonical encoding of a non-zero 256-bit ID.
func ParseID(value string) (ID, error) {
	var id ID
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(decoded) != idSize {
		return ID{}, ErrInvalidID
	}
	copy(id[:], decoded)
	if id.Validate() != nil || id.String() != value {
		return ID{}, ErrInvalidID
	}
	return id, nil
}

// Validate reports whether the ID can safely select a storage path.
func (id ID) Validate() error {
	if id == (ID{}) {
		return ErrInvalidID
	}
	return nil
}

func (id ID) String() string {
	return base64.RawURLEncoding.EncodeToString(id[:])
}

// Binding is the caller-supplied canonical SHA-256 digest of the bound work.
type Binding [bindingSize]byte

// ParseBinding accepts only canonical lowercase SHA-256 hex digests.
func ParseBinding(value string) (Binding, error) {
	var binding Binding
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != bindingSize {
		return Binding{}, ErrInvalidBinding
	}
	copy(binding[:], decoded)
	if binding.Validate() != nil || binding.String() != value {
		return Binding{}, ErrInvalidBinding
	}
	return binding, nil
}

// Validate reports whether a binding digest is present and canonical.
func (binding Binding) Validate() error {
	if binding == (Binding{}) {
		return ErrInvalidBinding
	}
	return nil
}

func (binding Binding) String() string {
	return hex.EncodeToString(binding[:])
}

// SessionID is an opaque runtime session identity supplied by a runtime adapter.
type SessionID string

// validate enforces a strict ASCII grammar so hostile Unicode such as
// format controls (Cf), bidirectional overrides, whitespace and non-ASCII
// lookalikes can never reach storage, a log line or a path.
func (session SessionID) validate() error {
	if len(session) == 0 || len(session) > maxSession {
		return ErrInvalidSession
	}
	for _, r := range session {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-', r == '_', r == '.', r == ':':
		default:
			return ErrInvalidSession
		}
	}
	return nil
}

// State is the monotonic lifecycle of a grant.
type State uint8

const (
	StateIssued State = iota
	StateClaimed
	StateBound
	StateCompleted
	StateFailed
)

func (state State) String() string {
	switch state {
	case StateIssued:
		return "issued"
	case StateClaimed:
		return "claimed"
	case StateBound:
		return "bound"
	case StateCompleted:
		return "completed"
	case StateFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// Grant is the domain state reconstructed from immutable storage records.
type Grant struct {
	ID      ID
	Binding Binding
	State   State
	Session SessionID
}

// NewGrant issues an unclaimed grant for a canonical binding digest.
func NewGrant(binding Binding) (Grant, error) {
	if err := binding.Validate(); err != nil {
		return Grant{}, err
	}
	id, err := NewID()
	if err != nil {
		return Grant{}, err
	}
	return Grant{ID: id, Binding: binding, State: StateIssued}, nil
}

// Validate checks the complete in-memory grant state.
func (grant Grant) Validate() error {
	if err := grant.ID.Validate(); err != nil {
		return err
	}
	if err := grant.Binding.Validate(); err != nil {
		return err
	}
	switch grant.State {
	case StateIssued, StateClaimed:
		if grant.Session != "" {
			return ErrInvalidSession
		}
	case StateBound, StateCompleted:
		if err := grant.Session.validate(); err != nil {
			return err
		}
	case StateFailed:
		if grant.Session != "" {
			if err := grant.Session.validate(); err != nil {
				return err
			}
		}
	default:
		return ErrInvalidTransition
	}
	return nil
}

// Claim consumes issued authority. Every later state refuses another claim.
// The grant is validated before and after the transition so an inconsistent
// grant can never become a claimed grant that is still invalid.
func (grant Grant) Claim() (Grant, error) {
	if err := grant.Validate(); err != nil {
		return Grant{}, err
	}
	if grant.State != StateIssued {
		return Grant{}, ErrAlreadyClaimed
	}
	next := grant
	next.State = StateClaimed
	if err := next.Validate(); err != nil {
		return Grant{}, err
	}
	return next, nil
}

// BindSession attaches the one fresh runtime session to a claimed grant.
// The grant and the session are validated before the transition, and the
// bound grant is validated again before it is returned.
func (grant Grant) BindSession(session SessionID) (Grant, error) {
	if err := grant.Validate(); err != nil {
		return Grant{}, err
	}
	if err := session.validate(); err != nil {
		return Grant{}, err
	}
	switch grant.State {
	case StateClaimed:
		next := grant
		next.Session = session
		next.State = StateBound
		if err := next.Validate(); err != nil {
			return Grant{}, err
		}
		return next, nil
	case StateBound:
		return Grant{}, ErrAlreadyBound
	case StateCompleted, StateFailed:
		return Grant{}, ErrTerminal
	default:
		return Grant{}, ErrInvalidTransition
	}
}

// Complete terminates a bound grant after successful runtime work. The grant
// is validated before and after the transition.
func (grant Grant) Complete() (Grant, error) {
	if err := grant.Validate(); err != nil {
		return Grant{}, err
	}
	switch grant.State {
	case StateBound:
		next := grant
		next.State = StateCompleted
		if err := next.Validate(); err != nil {
			return Grant{}, err
		}
		return next, nil
	case StateCompleted, StateFailed:
		return Grant{}, ErrTerminal
	default:
		return Grant{}, ErrInvalidTransition
	}
}

// Fail terminates a claimed or bound grant after unsuccessful runtime work.
// The grant is validated before and after the transition.
func (grant Grant) Fail() (Grant, error) {
	if err := grant.Validate(); err != nil {
		return Grant{}, err
	}
	switch grant.State {
	case StateClaimed, StateBound:
		next := grant
		next.State = StateFailed
		if err := next.Validate(); err != nil {
			return Grant{}, err
		}
		return next, nil
	case StateCompleted, StateFailed:
		return Grant{}, ErrTerminal
	default:
		return Grant{}, ErrInvalidTransition
	}
}
