package grant

import (
	"context"

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
