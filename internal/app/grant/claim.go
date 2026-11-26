package grant

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"

	domain "github.com/biggs-100/odachi/internal/domain/grant"
)

// claimedSuffix names the permanent consumed-authority marker.
const claimedSuffix = ".claimed"

// markClaimed publishes the permanent consumed-authority marker for id. The
// caller has already validated the issued record; this operation only decides
// who wins the one-time claim.
//
// The marker is an empty regular file created directly with O_EXCL. Its name is
// the authority, so it never goes through a staging record: a reader that sees
// the name sees a complete tombstone, and a crash leaves either no marker or the
// marker. It is never removed, so no later claim can take it over.
func (store *FileStore) markClaimed(ctx context.Context, id domain.ID) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := id.Validate(); err != nil {
		return err
	}
	root, err := store.open()
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()
	name, err := store.recordPath(id, claimedSuffix)
	if err != nil {
		return err
	}

	claimed, err := recordExists(ctx, root, name)
	if err != nil {
		return err
	}
	if claimed {
		return fmt.Errorf("%w: %s", domain.ErrAlreadyClaimed, id)
	}
	// A staging record beside the marker is not authority and must not be
	// overwritten or mistaken for a completed claim.
	staged, err := recordExists(ctx, root, name+stagingSuffix)
	if err != nil {
		return err
	}
	if staged {
		return stagingConflict(name + stagingSuffix)
	}

	if err := createClaimMarker(ctx, root, name); err != nil {
		// Losing the O_EXCL race is the expected concurrent outcome; re-read to
		// report either the winner's marker or a staging tombstone that needs
		// recovery, never a bare filesystem error.
		if !errors.Is(err, fs.ErrExist) {
			return err
		}
		return store.classifyClaimCollision(ctx, root, id, name)
	}
	return nil
}

// classifyClaimCollision explains a refused claim-marker create.
func (store *FileStore) classifyClaimCollision(ctx context.Context, root *os.Root, id domain.ID, name string) error {
	claimed, err := recordExists(ctx, root, name)
	if err != nil {
		return err
	}
	if claimed {
		return fmt.Errorf("%w: %s", domain.ErrAlreadyClaimed, id)
	}
	staged, err := recordExists(ctx, root, name+stagingSuffix)
	if err != nil {
		return err
	}
	if staged {
		return stagingConflict(name + stagingSuffix)
	}
	return fmt.Errorf("%w: claim marker creation refused without an existing marker", ErrStorageFailure)
}

// createClaimMarker creates the final marker name in one no-replace step. The
// empty file is complete the moment it exists. It is never removed on an error
// path: once created, it is consumed authority even if this process dies before
// the directory sync.
func createClaimMarker(ctx context.Context, root *os.Root, name string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	file, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("%w: create claim marker: %w", ErrStorageFailure, err)
	}
	closed := false
	defer func() {
		if !closed {
			_ = file.Close()
		}
	}()
	// The marker exists as soon as the create returns. Make it durable even if
	// the caller cancels, and never delete it: a cancellation must not hand the
	// same authority to a later claimant.
	commit := context.WithoutCancel(ctx)
	if err := file.Sync(); err != nil {
		return fmt.Errorf("%w: sync claim marker: %w", ErrStorageFailure, err)
	}
	closed = true
	if err := file.Close(); err != nil {
		return fmt.Errorf("%w: close claim marker: %w", ErrStorageFailure, err)
	}
	if err := syncParentDirectory(commit, root, name); err != nil {
		return fmt.Errorf("%w: sync claim marker directory: %w", ErrStorageFailure, err)
	}
	return nil
}
