package grant

import (
	"errors"
	"os"
	"testing"

	domain "github.com/biggs-100/odachi/internal/domain/grant"
)

const (
	sessionFixture   = `{"version":1,"session":"ses_bound_001"}`
	completedFixture = `{"version":1,"state":"completed"}`
	failedFixture    = `{"version":1,"state":"failed"}`
)

func TestLoadReconstructsIssuedGrant(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store)
	issued := mustIssue(t, service, domain.Binding{20})

	loaded, err := service.Load(t.Context(), issued.ID)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.State != domain.StateIssued || loaded.Session != "" {
		t.Fatalf("Load() = %+v, want issued with no session", loaded)
	}
	if loaded.ID != issued.ID || loaded.Binding != issued.Binding {
		t.Fatalf("Load() identity = %v/%v, want %v/%v", loaded.ID, loaded.Binding, issued.ID, issued.Binding)
	}
}

func TestLoadReconstructsClaimedGrant(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store)
	issued := mustIssue(t, service, domain.Binding{21})
	if _, err := service.Claim(t.Context(), issued.ID); err != nil {
		t.Fatalf("Claim() error = %v", err)
	}

	loaded, err := service.Load(t.Context(), issued.ID)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.State != domain.StateClaimed || loaded.Session != "" {
		t.Fatalf("Load() = %+v, want claimed with no session", loaded)
	}
}

func TestLoadReconstructsBoundSession(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store)
	issued := mustIssue(t, service, domain.Binding{22})
	if _, err := service.Claim(t.Context(), issued.ID); err != nil {
		t.Fatalf("Claim() error = %v", err)
	}
	writeLifecycle(t, store, issued.ID, ".session", sessionFixture)

	loaded, err := service.Load(t.Context(), issued.ID)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.State != domain.StateBound || loaded.Session != "ses_bound_001" {
		t.Fatalf("Load() = %+v, want bound to ses_bound_001", loaded)
	}
}

func TestLoadReconstructsTerminalStates(t *testing.T) {
	for _, test := range []struct {
		name  string
		body  string
		state domain.State
	}{
		{name: "completed", body: completedFixture, state: domain.StateCompleted},
		{name: "failed", body: failedFixture, state: domain.StateFailed},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := newTestStore(t)
			service := NewService(store)
			issued := mustIssue(t, service, domain.Binding{23})
			if _, err := service.Claim(t.Context(), issued.ID); err != nil {
				t.Fatalf("Claim() error = %v", err)
			}
			writeLifecycle(t, store, issued.ID, ".session", sessionFixture)
			writeLifecycle(t, store, issued.ID, ".terminal", test.body)

			loaded, err := service.Load(t.Context(), issued.ID)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if loaded.State != test.state || loaded.Session != "ses_bound_001" {
				t.Fatalf("Load() = %+v, want %s bound to ses_bound_001", loaded, test.state)
			}
		})
	}
}

// TestLoadTerminalDominatesEverySessionAndClaimRecord encodes the crash window
// where the claim, the session and the terminal record are all already durable:
// the terminal outcome must win no matter which record a reader happens to
// observe first.
func TestLoadTerminalDominatesEverySessionAndClaimRecord(t *testing.T) {
	for _, test := range []struct {
		name  string
		body  string
		state domain.State
	}{
		{name: "completed", body: completedFixture, state: domain.StateCompleted},
		{name: "failed", body: failedFixture, state: domain.StateFailed},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := newTestStore(t)
			service := NewService(store)
			issued := mustIssue(t, service, domain.Binding{24})
			if _, err := service.Claim(t.Context(), issued.ID); err != nil {
				t.Fatalf("Claim() error = %v", err)
			}
			writeLifecycle(t, store, issued.ID, ".session", sessionFixture)
			writeLifecycle(t, store, issued.ID, ".terminal", test.body)

			loaded, err := store.Load(t.Context(), issued.ID)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if loaded.State != test.state {
				t.Fatalf("Load() state = %v, want terminal %v", loaded.State, test.state)
			}
			if loaded.Session != "ses_bound_001" {
				t.Fatalf("Load() session = %q, want preserved ses_bound_001", loaded.Session)
			}
		})
	}
}

// TestLoadTerminalRecordDominatesPendingSessionStaging is the crash window where
// a terminal record was published after a session publication stalled: the
// terminal must dominate the pending session staging instead of failing closed.
func TestLoadTerminalRecordDominatesPendingSessionStaging(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store)
	issued := mustIssue(t, service, domain.Binding{25})
	if _, err := service.Claim(t.Context(), issued.ID); err != nil {
		t.Fatalf("Claim() error = %v", err)
	}
	writeLifecycle(t, store, issued.ID, ".session.tmp", "{")
	writeLifecycle(t, store, issued.ID, ".terminal", failedFixture)

	loaded, err := store.Load(t.Context(), issued.ID)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.State != domain.StateFailed || loaded.Session != "" {
		t.Fatalf("Load() = %+v, want failed without a session", loaded)
	}
}

func TestLoadRefusesPendingTerminalStaging(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store)
	issued := mustIssue(t, service, domain.Binding{26})
	if _, err := service.Claim(t.Context(), issued.ID); err != nil {
		t.Fatalf("Claim() error = %v", err)
	}
	writeLifecycle(t, store, issued.ID, ".terminal.tmp", "{")

	if _, err := service.Load(t.Context(), issued.ID); !errors.Is(err, ErrPendingTerminal) || !errors.Is(err, ErrCorruptRecord) {
		t.Fatalf("Load() error = %v, want ErrPendingTerminal wrapping ErrCorruptRecord", err)
	}
}

func TestLoadPendingTerminalStagingDominatesPendingSessionStaging(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store)
	issued := mustIssue(t, service, domain.Binding{27})
	if _, err := service.Claim(t.Context(), issued.ID); err != nil {
		t.Fatalf("Claim() error = %v", err)
	}
	writeLifecycle(t, store, issued.ID, ".session.tmp", "{")
	writeLifecycle(t, store, issued.ID, ".terminal.tmp", "{")

	_, err := service.Load(t.Context(), issued.ID)
	if !errors.Is(err, ErrPendingTerminal) {
		t.Fatalf("Load() error = %v, want ErrPendingTerminal to dominate", err)
	}
	if errors.Is(err, ErrPendingSession) {
		t.Fatalf("Load() error = %v, want the pending terminal to hide the pending session", err)
	}
}

// TestLoadRefusesPendingSessionStagingWithoutTerminal proves a stalled session
// publication is distinguishable and is never mistaken for a bound session.
func TestLoadRefusesPendingSessionStagingWithoutTerminal(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store)
	issued := mustIssue(t, service, domain.Binding{28})
	if _, err := service.Claim(t.Context(), issued.ID); err != nil {
		t.Fatalf("Claim() error = %v", err)
	}
	writeLifecycle(t, store, issued.ID, ".session.tmp", "{")

	loaded, err := service.Load(t.Context(), issued.ID)
	if !errors.Is(err, ErrPendingSession) || !errors.Is(err, ErrCorruptRecord) {
		t.Fatalf("Load() error = %v, want ErrPendingSession wrapping ErrCorruptRecord", err)
	}
	if loaded.State == domain.StateBound {
		t.Fatalf("Load() = %+v, want no bound session from pending staging", loaded)
	}
}

func TestLoadRefusesTerminalRecordWithoutClaimTombstone(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store)
	issued := mustIssue(t, service, domain.Binding{29})
	writeLifecycle(t, store, issued.ID, ".terminal", failedFixture)

	if _, err := service.Load(t.Context(), issued.ID); !errors.Is(err, ErrCorruptRecord) {
		t.Fatalf("Load() error = %v, want ErrCorruptRecord", err)
	}
}

func TestLoadRefusesSessionRecordWithoutClaimTombstone(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store)
	issued := mustIssue(t, service, domain.Binding{30})
	writeLifecycle(t, store, issued.ID, ".session", sessionFixture)

	if _, err := service.Load(t.Context(), issued.ID); !errors.Is(err, ErrCorruptRecord) {
		t.Fatalf("Load() error = %v, want ErrCorruptRecord", err)
	}
}

func TestLoadRefusesMissingGrant(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store)
	missing, err := domain.NewID()
	if err != nil {
		t.Fatalf("NewID() error = %v", err)
	}
	if _, err := service.Load(t.Context(), missing); !errors.Is(err, ErrGrantNotFound) {
		t.Fatalf("Load(missing) error = %v, want ErrGrantNotFound", err)
	}
}

func TestLoadRefusesCorruptForeignOrUnsupportedIssuedRecord(t *testing.T) {
	for _, test := range []struct {
		name    string
		body    func(issued domain.Grant) string
		wantErr error
	}{
		{
			name:    "corrupt",
			body:    func(domain.Grant) string { return `{"version":` },
			wantErr: ErrCorruptRecord,
		},
		{
			name: "foreign identity",
			body: func(issued domain.Grant) string {
				other, err := domain.NewID()
				if err != nil {
					t.Fatalf("NewID() error = %v", err)
				}
				return `{"version":1,"id":"` + other.String() + `","binding":"` + issued.Binding.String() + `"}`
			},
			wantErr: ErrForeignRecord,
		},
		{
			name: "unsupported version",
			body: func(issued domain.Grant) string {
				return `{"version":2,"id":"` + issued.ID.String() + `","binding":"` + issued.Binding.String() + `"}`
			},
			wantErr: ErrCorruptRecord,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := newTestStore(t)
			service := NewService(store)
			issued := mustIssue(t, service, domain.Binding{31})
			writeLifecycle(t, store, issued.ID, ".json", test.body(issued))

			if _, err := service.Load(t.Context(), issued.ID); !errors.Is(err, test.wantErr) {
				t.Fatalf("Load() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestLoadRefusesMalformedLifecycleEnvelopes(t *testing.T) {
	for _, test := range []struct {
		name   string
		suffix string
		body   string
		dir    bool
	}{
		{name: "corrupt session", suffix: ".session", body: "not-json"},
		{name: "unsupported session version", suffix: ".session", body: `{"version":2,"session":"ses_x"}`},
		{name: "corrupt terminal", suffix: ".terminal", body: "not-json"},
		{name: "unsupported terminal version", suffix: ".terminal", body: `{"version":2,"state":"failed"}`},
		{name: "unknown terminal state", suffix: ".terminal", body: `{"version":1,"state":"issued"}`},
		{name: "session is a directory", suffix: ".session", dir: true},
		{name: "terminal is a directory", suffix: ".terminal", dir: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := newTestStore(t)
			service := NewService(store)
			issued := mustIssue(t, service, domain.Binding{32})
			if _, err := service.Claim(t.Context(), issued.ID); err != nil {
				t.Fatalf("Claim() error = %v", err)
			}
			if test.dir {
				mkdirLifecycle(t, store, issued.ID, test.suffix)
			} else {
				writeLifecycle(t, store, issued.ID, test.suffix, test.body)
			}

			if _, err := service.Load(t.Context(), issued.ID); !errors.Is(err, ErrCorruptRecord) {
				t.Fatalf("Load() error = %v, want ErrCorruptRecord", err)
			}
		})
	}
}

// writeLifecycle writes one lifecycle record or staging file as the later
// session and terminal writers will publish it.
func writeLifecycle(t *testing.T, store *FileStore, id domain.ID, suffix, body string) {
	t.Helper()
	if err := os.WriteFile(grantFile(t, store, id, suffix), []byte(body), 0o600); err != nil {
		t.Fatalf("write lifecycle %q: %v", suffix, err)
	}
}

// mkdirLifecycle creates a non-regular lifecycle path so the reader proves it
// refuses anything that is not a regular file.
func mkdirLifecycle(t *testing.T, store *FileStore, id domain.ID, suffix string) {
	t.Helper()
	if err := os.Mkdir(grantFile(t, store, id, suffix), 0o700); err != nil {
		t.Fatalf("mkdir lifecycle %q: %v", suffix, err)
	}
}
