package grant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
)

const (
	// stagingSuffix marks a record whose bytes are not yet durable and published.
	// A reader that finds this name knows the publication did not finish; it is
	// never a record.
	stagingSuffix = ".tmp"
	// maxRecordSize bounds every record this package writes or reads, so a
	// damaged or hostile file cannot make a reader allocate without limit.
	maxRecordSize = 64 << 10
)

var (
	// ErrStorageFailure is the single typed failure for storage that cannot be
	// written, synced or published. It always wraps the underlying cause.
	ErrStorageFailure = errors.New("grant storage failure")
	// ErrCorruptRecord is the single typed failure for a record that exists but
	// cannot be trusted: truncated, oversized, not a regular file, or carrying
	// fields this package does not know.
	ErrCorruptRecord = errors.New("grant record is corrupt")
)

// publisher is the two durability steps a publication owes: the bytes reach the
// disk before the name appears, and the directory entry reaches the disk after.
// They are fields so a test can observe the order; production always passes the
// real pair.
type publisher struct {
	syncFile   func(*os.File) error
	syncParent func(context.Context, *os.Root, string) error
}

// publishRecord writes payload to name atomically and durably. The bytes go to a
// staging file in the same directory, are flushed, and only then does name appear
// in one filesystem step. A concurrent reader therefore finds no record or the
// whole record, and an existing record is never replaced: the previous complete
// record survives a failed publication untouched.
//
// A staging file is removed on every path; a removal that fails after the record
// is published is reported as ErrStorageFailure instead of leaving a temporary
// file behind silently. A failure after the record is visible (the directory
// sync) reports ErrStorageFailure and keeps the complete record, because deleting
// something a reader may already have seen would be a second unproven mutation.
func publishRecord(ctx context.Context, root *os.Root, name string, payload []byte) error {
	return publishRecordWith(ctx, root, name, payload, publisher{
		syncFile:   (*os.File).Sync,
		syncParent: syncParentDirectory,
	})
}

// publishRecordWith is publishRecord with its durability steps supplied.
func publishRecordWith(ctx context.Context, root *os.Root, name string, payload []byte, syncs publisher) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if syncs.syncFile == nil || syncs.syncParent == nil {
		return fmt.Errorf("%w: publication durability is required", ErrStorageFailure)
	}
	if len(payload) == 0 {
		return fmt.Errorf("%w: refusing to publish an empty record", ErrStorageFailure)
	}
	if len(payload) > maxRecordSize {
		return fmt.Errorf("%w: record is %d bytes, the limit is %d", ErrStorageFailure, len(payload), maxRecordSize)
	}

	staging := name + stagingSuffix
	file, err := root.OpenFile(staging, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("%w: create staging record: %w", ErrStorageFailure, err)
	}
	// The staging file is never a record: whatever path this call takes, the name
	// is removed. On success the removal is made durable by the directory sync.
	defer func() {
		_ = file.Close()
		_ = root.Remove(staging)
	}()

	written, err := file.Write(payload)
	if err != nil {
		return fmt.Errorf("%w: write staging record: %w", ErrStorageFailure, err)
	}
	if written != len(payload) {
		return fmt.Errorf("%w: write staging record: %w", ErrStorageFailure, io.ErrShortWrite)
	}
	if err := syncs.syncFile(file); err != nil {
		return fmt.Errorf("%w: sync staging record: %w", ErrStorageFailure, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("%w: close staging record: %w", ErrStorageFailure, err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	// The hard link is the atomic, no-replace publication: a reader either sees no
	// name or the complete, already flushed inode.
	if err := root.Link(staging, name); err != nil {
		return fmt.Errorf("%w: publish record: %w", ErrStorageFailure, err)
	}
	if err := root.Remove(staging); err != nil {
		return fmt.Errorf("%w: remove staging record after publication: %w", ErrStorageFailure, err)
	}
	if err := syncs.syncParent(ctx, root, name); err != nil {
		return fmt.Errorf("%w: sync published record directory: %w", ErrStorageFailure, err)
	}
	return nil
}

// readRecord decodes one complete record. Absence is reported unwrapped as
// fs.ErrNotExist, so a caller can tell "no record yet" from a damaged one; every
// other failure is ErrCorruptRecord. Only a regular file of bounded size that
// holds exactly one known envelope is accepted.
func readRecord(ctx context.Context, root *os.Root, name string, destination any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	file, err := root.Open(name)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	info, err := file.Stat()
	if err != nil {
		return corruptRecord(fmt.Errorf("stat record: %w", err))
	}
	if !info.Mode().IsRegular() {
		return corruptRecord(fmt.Errorf("record %q is not a regular file", name))
	}
	if info.Size() > maxRecordSize {
		return corruptRecord(fmt.Errorf("record %q is %d bytes, the limit is %d", name, info.Size(), maxRecordSize))
	}

	decoder := json.NewDecoder(io.LimitReader(file, maxRecordSize+1))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return corruptRecord(fmt.Errorf("decode record %q: %w", name, err))
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return corruptRecord(fmt.Errorf("record %q carries trailing content", name))
		}
		return corruptRecord(fmt.Errorf("record %q carries trailing content: %w", name, err))
	}
	return nil
}

// recordExists reports whether name is a published record. A missing record is
// false with no error; anything that exists but is not a regular file is a
// damaged store, never an absence.
func recordExists(ctx context.Context, root *os.Root, name string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	info, err := root.Lstat(name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	if !info.Mode().IsRegular() {
		return false, corruptRecord(fmt.Errorf("record %q is not a regular file", name))
	}
	return true, nil
}

// corruptRecord marks a failure as untrustworthy content while keeping the cause
// matchable.
func corruptRecord(err error) error {
	return fmt.Errorf("%w: %w", ErrCorruptRecord, err)
}
