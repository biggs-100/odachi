package grant

import (
	"context"
	"errors"
	"fmt"
	"os"

	domain "github.com/biggs-100/odachi/internal/domain/grant"
)

const (
	// sessionSuffix names the immutable record that binds one runtime session.
	sessionSuffix = ".session"
	// terminalSuffix names the immutable record that concludes a grant.
	terminalSuffix = ".terminal"
)

var (
	// ErrPendingSession reports a session publication that started but never
	// finished. It is never authority and is never mistaken for a bound session.
	ErrPendingSession = errors.New("grant session publication is pending")
	// ErrPendingTerminal reports a terminal publication that started but never
	// finished. It consumes authority and dominates any pending session.
	ErrPendingTerminal = errors.New("grant terminal publication is pending")
)

// sessionRecord is the immutable envelope that binds a grant to one session.
// It carries no grant identity because the record's name is derived from it.
type sessionRecord struct {
	Version int              `json:"version"`
	Session domain.SessionID `json:"session"`
}

// terminalRecord is the immutable envelope that concludes a grant.
type terminalRecord struct {
	Version int    `json:"version"`
	State   string `json:"state"`
}

// lifecycleSnapshot is one read of the claim, session and terminal records of a
// single grant. Pending fields record a staging file that was never published.
type lifecycleSnapshot struct {
	claimed            bool
	hasSession         bool
	session            domain.SessionID
	hasSessionPending  bool
	hasTerminal        bool
	terminalState      domain.State
	hasTerminalPending bool
}

// Load reconstructs the full lifecycle of a persisted grant from its durable
// records. A completed or failed terminal record dominates every session and
// claim record, so the reconstructed outcome does not depend on the order the
// records are read or on which transition a crash interrupted. A staged
// terminal record fails closed and dominates a staged session record; a staged
// session record without a terminal outcome stays distinguishable and is never
// reported as a bound session. Anything missing, corrupt, foreign,
// unsupported-version or non-regular fails closed.
//
// Session binding, Complete, Fail and terminal-race linearization are the
// caller's concern; this reconstruction only reports the durable state it
// observes.
func (store *FileStore) Load(ctx context.Context, id domain.ID) (domain.Grant, error) {
	if err := id.Validate(); err != nil {
		return domain.Grant{}, err
	}
	root, err := store.open()
	if err != nil {
		return domain.Grant{}, err
	}
	defer func() { _ = root.Close() }()

	issued, err := store.readIssuedRecord(ctx, root, id)
	if err != nil {
		return domain.Grant{}, err
	}
	snapshot, err := store.readLifecycle(ctx, root, id)
	if err != nil {
		return domain.Grant{}, corruptRecord(err)
	}
	return reconstructGrant(issued, snapshot)
}

// readLifecycle reads the claim, session and terminal records exactly once. A
// published record hides its own staging file: a crash after the no-replace
// publication but before the staging cleanup leaves both names, and the
// published record is the truth.
func (store *FileStore) readLifecycle(ctx context.Context, root *os.Root, id domain.ID) (lifecycleSnapshot, error) {
	var snapshot lifecycleSnapshot

	claimed, err := store.readPresence(ctx, root, id, claimedSuffix)
	if err != nil {
		return lifecycleSnapshot{}, err
	}
	snapshot.claimed = claimed

	hasSession, err := store.readPresence(ctx, root, id, sessionSuffix)
	if err != nil {
		return lifecycleSnapshot{}, err
	}
	if hasSession {
		name, err := store.recordPath(id, sessionSuffix)
		if err != nil {
			return lifecycleSnapshot{}, err
		}
		var record sessionRecord
		if err := readRecord(ctx, root, name, &record); err != nil {
			return lifecycleSnapshot{}, err
		}
		if record.Version != recordVersion {
			return lifecycleSnapshot{}, fmt.Errorf("unsupported session record version %d", record.Version)
		}
		snapshot.hasSession = true
		snapshot.session = record.Session
	} else {
		snapshot.hasSessionPending, err = store.readPresence(ctx, root, id, sessionSuffix+stagingSuffix)
		if err != nil {
			return lifecycleSnapshot{}, err
		}
	}

	hasTerminal, err := store.readPresence(ctx, root, id, terminalSuffix)
	if err != nil {
		return lifecycleSnapshot{}, err
	}
	if hasTerminal {
		name, err := store.recordPath(id, terminalSuffix)
		if err != nil {
			return lifecycleSnapshot{}, err
		}
		var record terminalRecord
		if err := readRecord(ctx, root, name, &record); err != nil {
			return lifecycleSnapshot{}, err
		}
		if record.Version != recordVersion {
			return lifecycleSnapshot{}, fmt.Errorf("unsupported terminal record version %d", record.Version)
		}
		state, err := terminalStateFromRecord(record)
		if err != nil {
			return lifecycleSnapshot{}, err
		}
		snapshot.hasTerminal = true
		snapshot.terminalState = state
	} else {
		snapshot.hasTerminalPending, err = store.readPresence(ctx, root, id, terminalSuffix+stagingSuffix)
		if err != nil {
			return lifecycleSnapshot{}, err
		}
	}

	return snapshot, nil
}

// readPresence reports whether one record name is present. A name that exists
// but is not a regular file is a damaged store, never an absence.
func (store *FileStore) readPresence(ctx context.Context, root *os.Root, id domain.ID, suffix string) (bool, error) {
	name, err := store.recordPath(id, suffix)
	if err != nil {
		return false, err
	}
	return recordExists(ctx, root, name)
}

// terminalStateFromRecord accepts only the two terminal outcomes this package
// can reconstruct. Any other state string is corrupt.
func terminalStateFromRecord(record terminalRecord) (domain.State, error) {
	switch record.State {
	case domain.StateCompleted.String():
		return domain.StateCompleted, nil
	case domain.StateFailed.String():
		return domain.StateFailed, nil
	default:
		return domain.StateIssued, fmt.Errorf("unknown terminal state %q", record.State)
	}
}

// reconstructGrant applies the durable records to the issued grant in lifecycle
// order, with terminal dominance decided before any session or claim is trusted.
func reconstructGrant(issued domain.Grant, snapshot lifecycleSnapshot) (domain.Grant, error) {
	// A staged terminal record consumes authority and fails closed. Checking it
	// first is what makes it dominate a staged session publication.
	if snapshot.hasTerminalPending {
		if !snapshot.claimed {
			return domain.Grant{}, corruptRecord(fmt.Errorf("terminal staging without claim tombstone"))
		}
		return domain.Grant{}, corruptRecord(ErrPendingTerminal)
	}
	// A terminal outcome dominates every session and claim record: it is
	// reconstructed last so the terminal state wins regardless of read order.
	if snapshot.hasTerminal && !snapshot.claimed {
		return domain.Grant{}, corruptRecord(fmt.Errorf("terminal record without claim tombstone"))
	}
	// Without a terminal outcome, a staged session stays distinguishable and
	// fails closed instead of reconstructing a binding that never completed.
	if snapshot.hasSessionPending && !snapshot.hasSession && !snapshot.hasTerminal {
		if !snapshot.claimed {
			return domain.Grant{}, corruptRecord(fmt.Errorf("session staging without claim tombstone"))
		}
		return domain.Grant{}, corruptRecord(ErrPendingSession)
	}
	if snapshot.hasSession && !snapshot.claimed {
		return domain.Grant{}, corruptRecord(fmt.Errorf("session record without claim tombstone"))
	}

	current := issued
	if snapshot.claimed {
		claimed, err := current.Claim()
		if err != nil {
			return domain.Grant{}, corruptRecord(err)
		}
		current = claimed
	}
	if snapshot.hasSession {
		bound, err := current.BindSession(snapshot.session)
		if err != nil {
			return domain.Grant{}, corruptRecord(err)
		}
		current = bound
	}
	if snapshot.hasTerminal {
		var err error
		switch snapshot.terminalState {
		case domain.StateCompleted:
			current, err = current.Complete()
		case domain.StateFailed:
			current, err = current.Fail()
		default:
			return domain.Grant{}, corruptRecord(fmt.Errorf("unknown terminal state"))
		}
		if err != nil {
			return domain.Grant{}, corruptRecord(err)
		}
	}
	if err := current.Validate(); err != nil {
		return domain.Grant{}, corruptRecord(err)
	}
	return current, nil
}
