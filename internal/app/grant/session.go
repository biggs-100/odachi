package grant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"

	domain "github.com/biggs-100/odachi/internal/domain/grant"
)

// errStoreConflict reports that another writer already owns the immutable
// lifecycle name this attempt was creating, or that a staging file for it is in
// flight. The caller maps it onto the domain refusal that matches the record: an
// existing session is ErrAlreadyBound and an existing terminal outcome is
// ErrTerminal. It is never authority on its own.
var errStoreConflict = errors.New("grant lifecycle record already published")

// writeSessionRecord publishes the immutable session binding for id. The name is
// no-replace, so the first published session is permanent and no later attempt
// can take it over.
func (store *FileStore) writeSessionRecord(ctx context.Context, id domain.ID, session domain.SessionID) error {
	return store.writeSessionRecordWith(ctx, id, session, defaultPublisher())
}

// writeSessionRecordWith is writeSessionRecord with its durability steps
// supplied, so a test can make publication fail after the marker is created.
func (store *FileStore) writeSessionRecordWith(ctx context.Context, id domain.ID, session domain.SessionID, syncs publisher) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := id.Validate(); err != nil {
		return err
	}
	record := sessionRecord{Version: recordVersion, Session: session}
	payload, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("%w: encode session record: %w", ErrStorageFailure, err)
	}
	if hook := store.hooks.beforeSession; hook != nil {
		hook()
	}
	return store.publishLifecycle(ctx, id, sessionSuffix, payload, syncs)
}

// writeTerminalRecord publishes the immutable terminal outcome for id. Only a
// completed or failed grant can own a terminal record, and the first published
// outcome wins permanently.
func (store *FileStore) writeTerminalRecord(ctx context.Context, id domain.ID, state domain.State) error {
	return store.writeTerminalRecordWith(ctx, id, state, defaultPublisher())
}

// writeTerminalRecordWith is writeTerminalRecord with its durability steps
// supplied, so a test can make publication fail after the marker is created.
func (store *FileStore) writeTerminalRecordWith(ctx context.Context, id domain.ID, state domain.State, syncs publisher) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := id.Validate(); err != nil {
		return err
	}
	switch state {
	case domain.StateCompleted, domain.StateFailed:
	default:
		return fmt.Errorf("%w: %q is not a terminal outcome", ErrStorageFailure, state)
	}
	record := terminalRecord{Version: recordVersion, State: state.String()}
	payload, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("%w: encode terminal record: %w", ErrStorageFailure, err)
	}
	if hook := store.hooks.beforeTerminal; hook != nil {
		hook()
	}
	return store.publishLifecycle(ctx, id, terminalSuffix, payload, syncs)
}

// publishLifecycle publishes one lifecycle record with the same complete-or-
// absent durability as the issued record. Any name collision is errStoreConflict:
// a published marker and an in-flight or stale staging file both mean this
// attempt must not create authority, and neither path deletes the existing name.
// The same mapping matches reconstruction, where a staged lifecycle record fails
// closed instead of being reported as a completed transition.
func (store *FileStore) publishLifecycle(ctx context.Context, id domain.ID, suffix string, payload []byte, syncs publisher) error {
	root, err := store.open()
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()
	name, err := store.recordPath(id, suffix)
	if err != nil {
		return err
	}
	if err := publishRecordWith(ctx, root, name, payload, syncs); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return errStoreConflict
		}
		return err
	}
	return nil
}

// defaultPublisher is the production durability pair: the record's bytes reach
// the disk before its name appears, and its directory entry reaches the disk
// after.
func defaultPublisher() publisher {
	return publisher{syncFile: (*os.File).Sync, syncParent: syncParentDirectory}
}
