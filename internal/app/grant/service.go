package grant

import (
	"context"
	"errors"
	"fmt"

	domain "github.com/biggs-100/odachi/internal/domain/grant"
)

// Service coordinates domain grant transitions with durable grant records.
type Service struct {
	store *FileStore
}

// NewService returns a service that persists grants through store.
func NewService(store *FileStore) *Service {
	return &Service{store: store}
}

// Load reconstructs the full lifecycle of a persisted grant from its durable
// records with terminal dominance. It reads only; session binding, Complete and
// Fail are performed by their own operations.
func (service *Service) Load(ctx context.Context, id domain.ID) (domain.Grant, error) {
	return service.store.Load(ctx, id)
}

// Issue creates a fresh unclaimed grant and persists its complete issued record.
// A persistence failure returns no grant, so a caller never holds authority the
// store could not durably record.
func (service *Service) Issue(ctx context.Context, binding domain.Binding) (domain.Grant, error) {
	if err := ctx.Err(); err != nil {
		return domain.Grant{}, err
	}
	issued, err := domain.NewGrant(binding)
	if err != nil {
		return domain.Grant{}, err
	}
	if err := service.store.createRecord(ctx, issued); err != nil {
		return domain.Grant{}, err
	}
	return issued, nil
}

// Claim atomically consumes a grant's issued authority exactly once. The issued
// record is read and validated first; the permanent marker is then published in
// one no-replace step. A grant that is already claimed, missing, corrupt or
// foreign returns no grant.
func (service *Service) Claim(ctx context.Context, id domain.ID) (domain.Grant, error) {
	if err := id.Validate(); err != nil {
		return domain.Grant{}, err
	}
	issued, err := service.store.loadIssued(ctx, id)
	if err != nil {
		return domain.Grant{}, err
	}
	claimed, err := issued.Claim()
	if err != nil {
		return domain.Grant{}, err
	}
	if err := service.store.markClaimed(ctx, id); err != nil {
		return domain.Grant{}, err
	}
	return claimed, nil
}

// loadLifecycle reconstructs a grant and maps the two typed pending-publication
// refusals Load reports onto the caller's policy for that stage. A staged session
// is not a bound session and a staged terminal has no known outcome, so both are
// fail-closed at the call site that names them.
func (service *Service) loadLifecycle(ctx context.Context, id domain.ID, pendingSession, pendingTerminal error) (domain.Grant, error) {
	current, err := service.store.Load(ctx, id)
	if err != nil {
		if pendingSession != nil && errors.Is(err, ErrPendingSession) {
			return domain.Grant{}, pendingSession
		}
		if pendingTerminal != nil && errors.Is(err, ErrPendingTerminal) {
			return domain.Grant{}, pendingTerminal
		}
	}
	return current, err
}

// BindSession publishes the single runtime session bound to a claimed grant. The
// session marker is immutable and no-replace, so a second or different session is
// refused. After the marker is durable the lifecycle is reloaded: that terminal
// observation is the operation's linearization point, so a Fail that won the race
// is returned instead of a stale bound grant.
func (service *Service) BindSession(ctx context.Context, id domain.ID, session domain.SessionID) (domain.Grant, error) {
	if err := id.Validate(); err != nil {
		return domain.Grant{}, err
	}
	current, err := service.loadLifecycle(ctx, id, domain.ErrAlreadyBound, domain.ErrTerminal)
	if err != nil {
		return domain.Grant{}, err
	}
	if _, err := current.BindSession(session); err != nil {
		return domain.Grant{}, err
	}
	if err := service.store.writeSessionRecord(ctx, id, session); err != nil {
		if errors.Is(err, errStoreConflict) {
			return service.bindSessionConflict(ctx, id)
		}
		return domain.Grant{}, fmt.Errorf("bind grant session: %w", err)
	}

	current, err = service.loadLifecycle(ctx, id, domain.ErrAlreadyBound, domain.ErrTerminal)
	if err != nil {
		return domain.Grant{}, fmt.Errorf("reload grant after session bind: %w", err)
	}
	switch current.State {
	case domain.StateCompleted, domain.StateFailed:
		return current, domain.ErrTerminal
	case domain.StateBound:
		if current.Session == session {
			return current, nil
		}
	}
	return domain.Grant{}, fmt.Errorf("%w: session binding was not reconstructed", ErrStorageFailure)
}

// bindSessionConflict explains a session marker that was already published: the
// durable outcome decides whether this bind lost to a terminal transition or to
// another session.
func (service *Service) bindSessionConflict(ctx context.Context, id domain.ID) (domain.Grant, error) {
	current, err := service.loadLifecycle(ctx, id, domain.ErrAlreadyBound, domain.ErrTerminal)
	if err != nil {
		return domain.Grant{}, fmt.Errorf("reload grant after session conflict: %w", err)
	}
	switch current.State {
	case domain.StateCompleted, domain.StateFailed:
		return current, domain.ErrTerminal
	case domain.StateBound:
		return domain.Grant{}, domain.ErrAlreadyBound
	default:
		return domain.Grant{}, errStoreConflict
	}
}

// Complete publishes the durable completed terminal record of a bound grant.
// Only a bound grant can complete, and the no-replace terminal record means a
// concurrent or later terminal transition loses without replacing the winner.
func (service *Service) Complete(ctx context.Context, id domain.ID) (domain.Grant, error) {
	if err := id.Validate(); err != nil {
		return domain.Grant{}, err
	}
	current, err := service.loadLifecycle(ctx, id, nil, domain.ErrTerminal)
	if err != nil {
		return domain.Grant{}, err
	}
	completed, err := current.Complete()
	if err != nil {
		return domain.Grant{}, err
	}
	if err := service.store.writeTerminalRecord(ctx, id, domain.StateCompleted); err != nil {
		if errors.Is(err, errStoreConflict) {
			return domain.Grant{}, domain.ErrTerminal
		}
		return domain.Grant{}, fmt.Errorf("complete grant: %w", err)
	}
	return completed, nil
}

// Fail publishes the durable failed terminal record of a claimed or bound grant.
// An interrupted session publication proves the grant was claimed, so failure is
// still a valid transition; the permanent terminal record dominates the staged
// session. The post-publication reload returns the session that won the race.
func (service *Service) Fail(ctx context.Context, id domain.ID) (domain.Grant, error) {
	if err := id.Validate(); err != nil {
		return domain.Grant{}, err
	}
	current, err := service.store.Load(ctx, id)
	switch {
	case errors.Is(err, ErrPendingTerminal):
		return domain.Grant{}, domain.ErrTerminal
	case errors.Is(err, ErrPendingSession):
		// Claimed, with a session publication that never finished. The terminal
		// marker is still a valid and dominant outcome for that grant.
	case err != nil:
		return domain.Grant{}, err
	default:
		if _, failErr := current.Fail(); failErr != nil {
			return domain.Grant{}, failErr
		}
	}
	if err := service.store.writeTerminalRecord(ctx, id, domain.StateFailed); err != nil {
		if errors.Is(err, errStoreConflict) {
			return domain.Grant{}, domain.ErrTerminal
		}
		return domain.Grant{}, fmt.Errorf("fail grant: %w", err)
	}

	current, err = service.loadLifecycle(ctx, id, nil, domain.ErrTerminal)
	if err != nil {
		return domain.Grant{}, fmt.Errorf("reload grant after failure: %w", err)
	}
	if current.State != domain.StateFailed {
		return domain.Grant{}, fmt.Errorf("%w: failure was not reconstructed", ErrStorageFailure)
	}
	return current, nil
}
