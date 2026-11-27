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

const (
	// recordVersion is the envelope version every grant record carries.
	recordVersion = 1
	// issuedSuffix names the immutable issued record of a grant.
	issuedSuffix = ".json"
)

var (
	// ErrGrantNotFound reports an absent issued record. It is a refusal, never
	// an empty grant.
	ErrGrantNotFound = errors.New("grant record not found")
	// ErrGrantExists reports a collision with an already published issued record.
	ErrGrantExists = errors.New("grant record already exists")
	// ErrForeignRecord reports a record whose identity does not match the grant
	// it was looked up as.
	ErrForeignRecord = errors.New("grant record belongs to another grant")
	// ErrStagingConflict reports a staging record that must be recovered by an
	// operator. It is never authority and it is never overwritten.
	ErrStagingConflict = errors.New("grant record staging conflict")
)

// grantRecord is the immutable issued envelope. It carries no lifecycle state:
// later records prove claim, session and terminal transitions.
type grantRecord struct {
	Version int    `json:"version"`
	ID      string `json:"id"`
	Binding string `json:"binding"`
}

// createRecord publishes the complete issued record without replacing an
// existing one. The bytes are published through the durable record primitives,
// so a reader sees either no record or the whole record.
func (store *FileStore) createRecord(ctx context.Context, issued domain.Grant) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if issued.State != domain.StateIssued {
		return fmt.Errorf("%w: only an issued grant can be persisted", ErrStorageFailure)
	}
	if err := issued.Validate(); err != nil {
		return err
	}
	record := grantRecord{
		Version: recordVersion,
		ID:      issued.ID.String(),
		Binding: issued.Binding.String(),
	}
	payload, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("%w: encode issued record: %w", ErrStorageFailure, err)
	}
	root, err := store.open()
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()
	name, err := store.recordPath(issued.ID, issuedSuffix)
	if err != nil {
		return err
	}
	if err := publishRecord(ctx, root, name, payload); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return store.classifyPublicationCollision(ctx, root, name)
		}
		return err
	}
	return nil
}

// loadIssued reconstructs only the issued state. Lifecycle reconstruction is a
// later step and is deliberately absent here: an issued record that has a claim
// marker still reconstructs as issued, and the marker alone decides the claim.
func (store *FileStore) loadIssued(ctx context.Context, id domain.ID) (domain.Grant, error) {
	if err := id.Validate(); err != nil {
		return domain.Grant{}, err
	}
	root, err := store.open()
	if err != nil {
		return domain.Grant{}, err
	}
	defer func() { _ = root.Close() }()
	name, err := store.recordPath(id, issuedSuffix)
	if err != nil {
		return domain.Grant{}, err
	}
	var record grantRecord
	if err := readRecord(ctx, root, name, &record); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return domain.Grant{}, err
		}
		// A staging record is not an issued record. Do not fall back to it and do
		// not report a plain absence: an interrupted publication needs recovery.
		staged, stageErr := recordExists(ctx, root, name+stagingSuffix)
		if stageErr != nil {
			return domain.Grant{}, stageErr
		}
		if staged {
			return domain.Grant{}, stagingConflict(name + stagingSuffix)
		}
		return domain.Grant{}, fmt.Errorf("%w: %s", ErrGrantNotFound, id)
	}
	return issuedFromRecord(id, record)
}

// issuedFromRecord turns one decoded envelope into the issued domain state and
// refuses anything that is not exactly the grant it was looked up as.
func issuedFromRecord(id domain.ID, record grantRecord) (domain.Grant, error) {
	if record.Version != recordVersion {
		return domain.Grant{}, corruptRecord(fmt.Errorf("unsupported issued record version %d", record.Version))
	}
	recordID, err := domain.ParseID(record.ID)
	if err != nil {
		return domain.Grant{}, corruptRecord(fmt.Errorf("invalid issued record identity: %w", err))
	}
	if recordID != id {
		return domain.Grant{}, fmt.Errorf("%w: record %s was read as %s", ErrForeignRecord, recordID, id)
	}
	binding, err := domain.ParseBinding(record.Binding)
	if err != nil {
		return domain.Grant{}, corruptRecord(fmt.Errorf("invalid issued record binding: %w", err))
	}
	issued := domain.Grant{ID: recordID, Binding: binding, State: domain.StateIssued}
	if err := issued.Validate(); err != nil {
		return domain.Grant{}, corruptRecord(err)
	}
	return issued, nil
}

// classifyPublicationCollision turns a no-replace refusal into the complete
// record that already owns the name, or into a typed staging conflict. It never
// treats a staging file as a record.
func (store *FileStore) classifyPublicationCollision(ctx context.Context, root *os.Root, name string) error {
	exists, err := recordExists(ctx, root, name)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("%w: %s", ErrGrantExists, name)
	}
	staged, err := recordExists(ctx, root, name+stagingSuffix)
	if err != nil {
		return err
	}
	if staged {
		return stagingConflict(name + stagingSuffix)
	}
	return fmt.Errorf("%w: publication refused without an existing or staged record", ErrStorageFailure)
}

// stagingConflict is the typed refusal for a staging record that was never
// published. The caller must inspect the named file and remove it deliberately;
// retrying would either overwrite evidence or mistake a tombstone for authority.
func stagingConflict(name string) error {
	return fmt.Errorf("%w: %w: %s needs operator recovery", ErrStorageFailure, ErrStagingConflict, name)
}
