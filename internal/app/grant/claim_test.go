package grant

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"testing"

	domain "github.com/biggs-100/odachi/internal/domain/grant"
)

func TestIssuePersistsCompleteIssuedRecord(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store)
	binding := domain.Binding{1}

	issued, err := service.Issue(t.Context(), binding)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	if issued.State != domain.StateIssued || issued.Session != "" {
		t.Fatalf("issued grant = %+v, want issued with no session", issued)
	}
	if issued.Binding != binding {
		t.Fatalf("issued binding = %v, want %v", issued.Binding, binding)
	}
	if err := issued.ID.Validate(); err != nil {
		t.Fatalf("issued ID is invalid: %v", err)
	}

	data, err := os.ReadFile(grantFile(t, store, issued.ID, issuedSuffix))
	if err != nil {
		t.Fatalf("ReadFile(issued record) error = %v", err)
	}
	var record grantRecord
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatalf("Unmarshal(issued record) error = %v", err)
	}
	if record.Version != recordVersion {
		t.Fatalf("record version = %d, want %d", record.Version, recordVersion)
	}
	if record.ID != issued.ID.String() {
		t.Fatalf("record ID = %q, want %q", record.ID, issued.ID.String())
	}
	if record.Binding != binding.String() {
		t.Fatalf("record binding = %q, want %q", record.Binding, binding.String())
	}
	assertNamespaceHolds(t, store, issued.ID.String()+issuedSuffix)
}

func TestIssueRejectsInvalidBinding(t *testing.T) {
	store := newTestStore(t)
	if _, err := NewService(store).Issue(t.Context(), domain.Binding{}); !errors.Is(err, domain.ErrInvalidBinding) {
		t.Fatalf("Issue(zero binding) error = %v, want ErrInvalidBinding", err)
	}
	assertNamespaceEmpty(t, store)
}

func TestClaimConsumesIssuedAuthorityExactlyOnce(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store)
	issued := mustIssue(t, service, domain.Binding{2})

	claimed, err := service.Claim(t.Context(), issued.ID)
	if err != nil {
		t.Fatalf("Claim() error = %v", err)
	}
	if claimed.State != domain.StateClaimed || claimed.Session != "" {
		t.Fatalf("claimed grant = %+v, want claimed with no session", claimed)
	}
	if claimed.Binding != issued.Binding {
		t.Fatalf("claimed binding = %v, want %v", claimed.Binding, issued.Binding)
	}
	if _, err := os.Stat(grantFile(t, store, issued.ID, claimedSuffix)); err != nil {
		t.Fatalf("Stat(claim marker) error = %v", err)
	}
	assertNamespaceHolds(t, store, issued.ID.String()+issuedSuffix, issued.ID.String()+claimedSuffix)

	if _, err := service.Claim(t.Context(), issued.ID); !errors.Is(err, domain.ErrAlreadyClaimed) {
		t.Fatalf("second Claim() error = %v, want ErrAlreadyClaimed", err)
	}
}

func TestClaimRefusesMissingGrant(t *testing.T) {
	store := newTestStore(t)
	id, err := domain.NewID()
	if err != nil {
		t.Fatalf("NewID() error = %v", err)
	}
	if _, err := NewService(store).Claim(t.Context(), id); !errors.Is(err, ErrGrantNotFound) {
		t.Fatalf("Claim(missing) error = %v, want ErrGrantNotFound", err)
	}
	assertNamespaceEmpty(t, store)
}

func TestClaimRefusesCorruptIssuedRecord(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store)
	issued := mustIssue(t, service, domain.Binding{3})
	if err := os.WriteFile(grantFile(t, store, issued.ID, issuedSuffix), []byte(`{"version":`), 0o600); err != nil {
		t.Fatalf("WriteFile(corrupt record) error = %v", err)
	}
	if _, err := service.Claim(t.Context(), issued.ID); !errors.Is(err, ErrCorruptRecord) {
		t.Fatalf("Claim(corrupt) error = %v, want ErrCorruptRecord", err)
	}
	assertNoClaimMarker(t, store, issued.ID)
}

func TestClaimRefusesForeignIssuedRecord(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store)
	issued := mustIssue(t, service, domain.Binding{4})
	other, err := domain.NewID()
	if err != nil {
		t.Fatalf("NewID() error = %v", err)
	}
	foreign, err := json.Marshal(grantRecord{Version: recordVersion, ID: other.String(), Binding: issued.Binding.String()})
	if err != nil {
		t.Fatalf("Marshal(foreign record) error = %v", err)
	}
	if err := os.WriteFile(grantFile(t, store, issued.ID, issuedSuffix), foreign, 0o600); err != nil {
		t.Fatalf("WriteFile(foreign record) error = %v", err)
	}
	if _, err := service.Claim(t.Context(), issued.ID); !errors.Is(err, ErrForeignRecord) {
		t.Fatalf("Claim(foreign) error = %v, want ErrForeignRecord", err)
	}
	assertNoClaimMarker(t, store, issued.ID)
}

func TestClaimRefusesInvalidIssuedRecordVersion(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store)
	issued := mustIssue(t, service, domain.Binding{5})
	invalid, err := json.Marshal(grantRecord{Version: recordVersion + 1, ID: issued.ID.String(), Binding: issued.Binding.String()})
	if err != nil {
		t.Fatalf("Marshal(invalid record) error = %v", err)
	}
	if err := os.WriteFile(grantFile(t, store, issued.ID, issuedSuffix), invalid, 0o600); err != nil {
		t.Fatalf("WriteFile(invalid record) error = %v", err)
	}
	if _, err := service.Claim(t.Context(), issued.ID); !errors.Is(err, ErrCorruptRecord) {
		t.Fatalf("Claim(invalid version) error = %v, want ErrCorruptRecord", err)
	}
	assertNoClaimMarker(t, store, issued.ID)
}

func TestConcurrentClaimsProduceExactlyOneWinner(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store)
	issued := mustIssue(t, service, domain.Binding{6})

	const workers = 16
	start := make(chan struct{})
	results := make(chan error, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			<-start
			_, err := service.Claim(t.Context(), issued.ID)
			results <- err
		})
	}
	close(start)
	wg.Wait()
	close(results)

	winners := 0
	for err := range results {
		if err == nil {
			winners++
			continue
		}
		if !errors.Is(err, domain.ErrAlreadyClaimed) {
			t.Fatalf("losing Claim() error = %v, want ErrAlreadyClaimed", err)
		}
	}
	if winners != 1 {
		t.Fatalf("claim winners = %d, want exactly 1", winners)
	}
	if _, err := os.Stat(grantFile(t, store, issued.ID, claimedSuffix)); err != nil {
		t.Fatalf("Stat(claim marker) error = %v", err)
	}
	assertNamespaceHolds(t, store, issued.ID.String()+issuedSuffix, issued.ID.String()+claimedSuffix)
	if _, err := service.Claim(t.Context(), issued.ID); !errors.Is(err, domain.ErrAlreadyClaimed) {
		t.Fatalf("Claim() after the race error = %v, want ErrAlreadyClaimed", err)
	}
}

func TestClaimRefusalSurvivesRestart(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store)
	issued := mustIssue(t, service, domain.Binding{7})
	if _, err := service.Claim(t.Context(), issued.ID); err != nil {
		t.Fatalf("Claim() error = %v", err)
	}

	// A new store and service over the same root share no in-memory state with
	// the first writer: this is the restart that must still refuse replay.
	restarted, err := NewFileStore(store.root)
	if err != nil {
		t.Fatalf("NewFileStore(restart) error = %v", err)
	}
	if _, err := NewService(restarted).Claim(t.Context(), issued.ID); !errors.Is(err, domain.ErrAlreadyClaimed) {
		t.Fatalf("Claim(after restart) error = %v, want ErrAlreadyClaimed", err)
	}
	if _, err := os.Stat(grantFile(t, restarted, issued.ID, claimedSuffix)); err != nil {
		t.Fatalf("claim marker did not survive restart: %v", err)
	}
}

func TestClaimRefusesStaleClaimStagingRecord(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store)
	issued := mustIssue(t, service, domain.Binding{8})

	stale := []byte("interrupted claim")
	staging := grantFile(t, store, issued.ID, claimedSuffix+stagingSuffix)
	if err := os.WriteFile(staging, stale, 0o600); err != nil {
		t.Fatalf("WriteFile(stale staging) error = %v", err)
	}

	_, err := service.Claim(t.Context(), issued.ID)
	if !errors.Is(err, ErrStagingConflict) || !errors.Is(err, ErrStorageFailure) {
		t.Fatalf("Claim(stale staging) error = %v, want ErrStagingConflict wrapping ErrStorageFailure", err)
	}
	assertNoClaimMarker(t, store, issued.ID)
	got, readErr := os.ReadFile(staging)
	if readErr != nil {
		t.Fatalf("stale staging was removed: %v", readErr)
	}
	if string(got) != string(stale) {
		t.Fatalf("stale staging = %q, want it untouched at %q", got, stale)
	}
}

func TestClaimRefusesStaleIssuedStagingRecord(t *testing.T) {
	store := newTestStore(t)
	binding := domain.Binding{9}
	issued, err := domain.NewGrant(binding)
	if err != nil {
		t.Fatalf("NewGrant() error = %v", err)
	}
	record, err := json.Marshal(grantRecord{Version: recordVersion, ID: issued.ID.String(), Binding: binding.String()})
	if err != nil {
		t.Fatalf("Marshal(record) error = %v", err)
	}
	if err := os.WriteFile(grantFile(t, store, issued.ID, issuedSuffix+stagingSuffix), record, 0o600); err != nil {
		t.Fatalf("WriteFile(issued staging) error = %v", err)
	}

	_, err = NewService(store).Claim(t.Context(), issued.ID)
	if !errors.Is(err, ErrStagingConflict) {
		t.Fatalf("Claim(stale issued staging) error = %v, want ErrStagingConflict", err)
	}
	if _, statErr := os.Stat(grantFile(t, store, issued.ID, issuedSuffix)); !errors.Is(statErr, fs.ErrNotExist) {
		t.Fatalf("staging was mistaken for an issued record: %v", statErr)
	}
	assertNoClaimMarker(t, store, issued.ID)
}

func TestClaimRefusesClaimMarkerThatIsNotARegularFile(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store)
	issued := mustIssue(t, service, domain.Binding{10})
	if err := os.Mkdir(grantFile(t, store, issued.ID, claimedSuffix), 0o700); err != nil {
		t.Fatalf("Mkdir(claim marker) error = %v", err)
	}
	if _, err := service.Claim(t.Context(), issued.ID); !errors.Is(err, ErrCorruptRecord) {
		t.Fatalf("Claim(directory marker) error = %v, want ErrCorruptRecord", err)
	}
}

// newTestStore opens a durable store under a test-owned root.
func newTestStore(t *testing.T) *FileStore {
	t.Helper()
	store, err := NewFileStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewFileStore() error = %v", err)
	}
	return store
}

// mustIssue issues a grant whose durability is part of the test setup.
func mustIssue(t *testing.T, service *Service, binding domain.Binding) domain.Grant {
	t.Helper()
	issued, err := service.Issue(t.Context(), binding)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	return issued
}

// grantFile returns the on-disk path of one grant record.
func grantFile(t *testing.T, store *FileStore, id domain.ID, suffix string) string {
	t.Helper()
	return filepath.Join(store.root, namespaceName, id.String()+suffix)
}

func assertNamespaceHolds(t *testing.T, store *FileStore, want ...string) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(store.root, namespaceName))
	if err != nil {
		t.Fatalf("ReadDir(namespace) error = %v", err)
	}
	got := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		got[entry.Name()] = struct{}{}
	}
	for _, name := range want {
		if _, ok := got[name]; !ok {
			t.Fatalf("namespace is missing %q, holds %v", name, got)
		}
	}
	if len(got) != len(want) {
		t.Fatalf("namespace holds %v, want exactly %v", got, want)
	}
}

func assertNamespaceEmpty(t *testing.T, store *FileStore) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(store.root, namespaceName))
	if err != nil {
		t.Fatalf("ReadDir(namespace) error = %v", err)
	}
	if len(entries) != 0 {
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		t.Fatalf("namespace holds %v, want it empty", names)
	}
}

func assertNoClaimMarker(t *testing.T, store *FileStore, id domain.ID) {
	t.Helper()
	if _, err := os.Stat(grantFile(t, store, id, claimedSuffix)); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("claim marker exists after a refused claim: %v", err)
	}
}
