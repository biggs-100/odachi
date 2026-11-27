// Package grant is the application layer of the one-shot runtime grant: it owns
// the clone-local namespace where a grant's records live, publishes one complete
// record in a single atomic step, and reads nothing a writer may still be
// finishing.
//
// The domain contract in internal/domain/grant decides what a grant is and which
// transitions are legal. This package only moves that state to and from disk, so
// the invariant it owns is narrower and heavier: a reader sees an old record or a
// new one, never an empty or half-written one.
package grant

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	domain "github.com/biggs-100/odachi/internal/domain/grant"
)

const (
	// namespaceName is the single directory this package owns beneath the
	// runtime store root. Nothing outside it is a grant.
	namespaceName = "grants"
	// storeMode keeps the namespace private to the clone that created it.
	storeMode = 0o700
)

// FileStore publishes and reads grant records beneath one runtime store root.
// The root is configuration; a record name is derived from a validated grant
// identity, never from text a caller supplies, so a hostile identifier cannot
// select a path.
type FileStore struct {
	root string
	// hooks are nil in production. A test sets one to pause a lifecycle
	// publication between the caller's validation and the marker's creation,
	// which makes a session-versus-terminal race deterministic without a sleep.
	hooks lifecycleHooks
}

// lifecycleHooks are the optional observation points a test may install on a
// FileStore. They run before the named lifecycle marker is created and never
// change what the operation writes.
type lifecycleHooks struct {
	beforeSession  func()
	beforeTerminal func()
}

// NewFileStore prepares a durable <runtime-store-root>/grants namespace. The
// caller supplies the clone-local runtime path, which may not exist yet; this
// package owns only the namespace it names.
func NewFileStore(root string) (*FileStore, error) {
	return newFileStore(root, ensureDurableDirectory)
}

// newFileStore is the constructor with the durability step supplied, so a test
// can observe which directories were synced and what a platform that cannot sync
// a directory does.
func newFileStore(root string, ensureDirectory func(context.Context, string) error) (*FileStore, error) {
	if strings.TrimSpace(root) == "" {
		return nil, fmt.Errorf("%w: empty runtime store root", ErrStorageFailure)
	}
	if ensureDirectory == nil {
		return nil, fmt.Errorf("%w: directory durability is required", ErrStorageFailure)
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("%w: resolve runtime store root: %w", ErrStorageFailure, err)
	}
	ctx := context.Background()
	if err := ensureDirectory(ctx, absolute); err != nil {
		return nil, fmt.Errorf("%w: prepare runtime store root: %w", ErrStorageFailure, err)
	}
	if err := ensureDirectory(ctx, filepath.Join(absolute, namespaceName)); err != nil {
		return nil, fmt.Errorf("%w: prepare grant namespace: %w", ErrStorageFailure, err)
	}
	store := &FileStore{root: absolute}
	// Prove the root can be traversed before a caller believes the store exists:
	// a permission failure here is a refusal, not a surprise on the first write.
	probe, err := store.open()
	if err != nil {
		return nil, err
	}
	if err := probe.Close(); err != nil {
		return nil, fmt.Errorf("%w: close runtime store root: %w", ErrStorageFailure, err)
	}
	return store, nil
}

// recordPath names one record inside the grant namespace. The identity is
// validated and the suffix is a fixed extension, so the joined path always stays
// under the namespace.
func (store *FileStore) recordPath(id domain.ID, suffix string) (string, error) {
	if err := id.Validate(); err != nil {
		return "", err
	}
	if !strings.HasPrefix(suffix, ".") || strings.ContainsAny(suffix, `/\`) {
		return "", fmt.Errorf("%w: invalid record suffix %q", ErrStorageFailure, suffix)
	}
	return filepath.Join(namespaceName, id.String()+suffix), nil
}

// open grants access to the store root as a rooted filesystem: every later path
// is resolved inside it, so no record name can climb out.
func (store *FileStore) open() (*os.Root, error) {
	root, err := os.OpenRoot(store.root)
	if err != nil {
		return nil, fmt.Errorf("%w: open runtime store root: %w", ErrStorageFailure, err)
	}
	return root, nil
}
