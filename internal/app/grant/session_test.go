package grant

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"

	domain "github.com/biggs-100/odachi/internal/domain/grant"
)

func TestBindSessionBindsOneSessionExactlyOnce(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store)
	issued := mustIssue(t, service, domain.Binding{40})
	mustClaim(t, service, issued.ID)

	bound, err := service.BindSession(t.Context(), issued.ID, domain.SessionID("ses_fresh_001"))
	if err != nil {
		t.Fatalf("BindSession() error = %v", err)
	}
	if bound.State != domain.StateBound || bound.Session != "ses_fresh_001" {
		t.Fatalf("bound grant = %+v, want bound to ses_fresh_001", bound)
	}
	if bound.ID != issued.ID || bound.Binding != issued.Binding {
		t.Fatalf("bound identity = %v/%v, want %v/%v", bound.ID, bound.Binding, issued.ID, issued.Binding)
	}

	data, err := os.ReadFile(grantFile(t, store, issued.ID, sessionSuffix))
	if err != nil {
		t.Fatalf("ReadFile(session record) error = %v", err)
	}
	var record sessionRecord
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatalf("Unmarshal(session record) error = %v", err)
	}
	if record.Version != recordVersion || record.Session != "ses_fresh_001" {
		t.Fatalf("session record = %+v, want version %d and ses_fresh_001", record, recordVersion)
	}
	assertNamespaceHolds(t, store,
		issued.ID.String()+issuedSuffix,
		issued.ID.String()+claimedSuffix,
		issued.ID.String()+sessionSuffix,
	)

	for _, session := range []domain.SessionID{"ses_other_002", "ses_fresh_001"} {
		if _, err := service.BindSession(t.Context(), issued.ID, session); !errors.Is(err, domain.ErrAlreadyBound) {
			t.Fatalf("BindSession(%q) error = %v, want ErrAlreadyBound", session, err)
		}
	}
	if err := store.writeSessionRecord(t.Context(), issued.ID, domain.SessionID("ses_direct_other")); !errors.Is(err, errStoreConflict) {
		t.Fatalf("writeSessionRecord(second session) error = %v, want errStoreConflict", err)
	}

	after, err := os.ReadFile(grantFile(t, store, issued.ID, sessionSuffix))
	if err != nil {
		t.Fatalf("ReadFile(session record after refusal) error = %v", err)
	}
	if string(after) != string(data) {
		t.Fatalf("session record = %q, want the original %q untouched", after, data)
	}
}

func TestBindSessionRequiresAClaimedValidSession(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store)
	issued := mustIssue(t, service, domain.Binding{41})

	if _, err := service.BindSession(t.Context(), issued.ID, domain.SessionID("ses_early")); !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("BindSession(before claim) error = %v, want ErrInvalidTransition", err)
	}
	assertNoSessionMarker(t, store, issued.ID)

	mustClaim(t, service, issued.ID)
	for _, session := range []domain.SessionID{"", "has space", "sésion", "with/slash", domain.SessionID(strings.Repeat("a", 257))} {
		if _, err := service.BindSession(t.Context(), issued.ID, session); !errors.Is(err, domain.ErrInvalidSession) {
			t.Fatalf("BindSession(%q) error = %v, want ErrInvalidSession", session, err)
		}
	}
	assertNoSessionMarker(t, store, issued.ID)
}

func TestBindSessionRefusesTerminalGrant(t *testing.T) {
	for _, test := range []struct {
		name     string
		terminal func(*testing.T, *Service, domain.ID)
	}{
		{
			name: "completed",
			terminal: func(t *testing.T, service *Service, id domain.ID) {
				t.Helper()
				if _, err := service.Complete(t.Context(), id); err != nil {
					t.Fatalf("Complete() error = %v", err)
				}
			},
		},
		{
			name: "failed",
			terminal: func(t *testing.T, service *Service, id domain.ID) {
				t.Helper()
				if _, err := service.Fail(t.Context(), id); err != nil {
					t.Fatalf("Fail() error = %v", err)
				}
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := newTestStore(t)
			service := NewService(store)
			issued := mustIssue(t, service, domain.Binding{42})
			mustClaim(t, service, issued.ID)
			if _, err := service.BindSession(t.Context(), issued.ID, domain.SessionID("ses_first")); err != nil {
				t.Fatalf("BindSession() error = %v", err)
			}
			test.terminal(t, service, issued.ID)

			if _, err := service.BindSession(t.Context(), issued.ID, domain.SessionID("ses_second")); !errors.Is(err, domain.ErrTerminal) {
				t.Fatalf("BindSession(after %s) error = %v, want ErrTerminal", test.name, err)
			}
		})
	}
}

func TestCompleteRequiresABoundSession(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store)
	issued := mustIssue(t, service, domain.Binding{43})

	if _, err := service.Complete(t.Context(), issued.ID); !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("Complete(before claim) error = %v, want ErrInvalidTransition", err)
	}
	mustClaim(t, service, issued.ID)
	if _, err := service.Complete(t.Context(), issued.ID); !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("Complete(claimed) error = %v, want ErrInvalidTransition", err)
	}
	assertNoTerminalMarker(t, store, issued.ID)

	if _, err := service.BindSession(t.Context(), issued.ID, domain.SessionID("ses_bound_complete")); err != nil {
		t.Fatalf("BindSession() error = %v", err)
	}
	completed, err := service.Complete(t.Context(), issued.ID)
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if completed.State != domain.StateCompleted || completed.Session != "ses_bound_complete" {
		t.Fatalf("completed grant = %+v, want completed with the bound session", completed)
	}

	data, err := os.ReadFile(grantFile(t, store, issued.ID, terminalSuffix))
	if err != nil {
		t.Fatalf("ReadFile(terminal record) error = %v", err)
	}
	var record terminalRecord
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatalf("Unmarshal(terminal record) error = %v", err)
	}
	if record.Version != recordVersion || record.State != domain.StateCompleted.String() {
		t.Fatalf("terminal record = %+v, want completed", record)
	}
	assertNamespaceHolds(t, store,
		issued.ID.String()+issuedSuffix,
		issued.ID.String()+claimedSuffix,
		issued.ID.String()+sessionSuffix,
		issued.ID.String()+terminalSuffix,
	)

	if _, err := service.Fail(t.Context(), issued.ID); !errors.Is(err, domain.ErrTerminal) {
		t.Fatalf("Fail(after completion) error = %v, want ErrTerminal", err)
	}
	if _, err := service.Complete(t.Context(), issued.ID); !errors.Is(err, domain.ErrTerminal) {
		t.Fatalf("Complete(after completion) error = %v, want ErrTerminal", err)
	}
}

func TestFailTerminatesClaimedOrBoundGrant(t *testing.T) {
	t.Run("claimed", func(t *testing.T) {
		store := newTestStore(t)
		service := NewService(store)
		issued := mustIssue(t, service, domain.Binding{44})
		mustClaim(t, service, issued.ID)

		failed, err := service.Fail(t.Context(), issued.ID)
		if err != nil {
			t.Fatalf("Fail() error = %v", err)
		}
		if failed.State != domain.StateFailed || failed.Session != "" {
			t.Fatalf("failed grant = %+v, want failed without a session", failed)
		}
		assertTerminalState(t, store, issued.ID, domain.StateFailed)
		assertNamespaceHolds(t, store,
			issued.ID.String()+issuedSuffix,
			issued.ID.String()+claimedSuffix,
			issued.ID.String()+terminalSuffix,
		)

		if _, err := service.Fail(t.Context(), issued.ID); !errors.Is(err, domain.ErrTerminal) {
			t.Fatalf("second Fail() error = %v, want ErrTerminal", err)
		}
		if _, err := service.Complete(t.Context(), issued.ID); !errors.Is(err, domain.ErrTerminal) {
			t.Fatalf("Complete(after failure) error = %v, want ErrTerminal", err)
		}
	})

	t.Run("bound", func(t *testing.T) {
		store := newTestStore(t)
		service := NewService(store)
		issued := mustIssue(t, service, domain.Binding{45})
		mustClaim(t, service, issued.ID)
		if _, err := service.BindSession(t.Context(), issued.ID, domain.SessionID("ses_fail_bound")); err != nil {
			t.Fatalf("BindSession() error = %v", err)
		}

		failed, err := service.Fail(t.Context(), issued.ID)
		if err != nil {
			t.Fatalf("Fail() error = %v", err)
		}
		if failed.State != domain.StateFailed || failed.Session != "ses_fail_bound" {
			t.Fatalf("failed grant = %+v, want failed preserving ses_fail_bound", failed)
		}
		assertTerminalState(t, store, issued.ID, domain.StateFailed)
	})
}

func TestFailAcceptsAnInterruptedSessionPublication(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store)
	issued := mustIssue(t, service, domain.Binding{46})
	mustClaim(t, service, issued.ID)
	writeLifecycle(t, store, issued.ID, sessionSuffix+stagingSuffix, "{")

	if _, err := service.BindSession(t.Context(), issued.ID, domain.SessionID("ses_replay")); !errors.Is(err, domain.ErrAlreadyBound) {
		t.Fatalf("BindSession(pending session) error = %v, want ErrAlreadyBound", err)
	}
	if _, err := service.Complete(t.Context(), issued.ID); !errors.Is(err, ErrPendingSession) {
		t.Fatalf("Complete(pending session) error = %v, want ErrPendingSession", err)
	}

	failed, err := service.Fail(t.Context(), issued.ID)
	if err != nil {
		t.Fatalf("Fail() error = %v", err)
	}
	if failed.State != domain.StateFailed || failed.Session != "" {
		t.Fatalf("failed grant = %+v, want failed without a session", failed)
	}
	loaded, err := service.Load(t.Context(), issued.ID)
	if err != nil {
		t.Fatalf("Load() after failure error = %v", err)
	}
	if loaded != failed {
		t.Fatalf("Load() = %+v, want %+v", loaded, failed)
	}
}

func TestSessionAndTerminalRecordsSurviveRestart(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store)
	issued := mustIssue(t, service, domain.Binding{47})
	mustClaim(t, service, issued.ID)
	if _, err := service.BindSession(t.Context(), issued.ID, domain.SessionID("ses_restart")); err != nil {
		t.Fatalf("BindSession() error = %v", err)
	}
	if _, err := service.Complete(t.Context(), issued.ID); err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	restarted, err := NewFileStore(store.root)
	if err != nil {
		t.Fatalf("NewFileStore(restart) error = %v", err)
	}
	reopened := NewService(restarted)
	loaded, err := reopened.Load(t.Context(), issued.ID)
	if err != nil {
		t.Fatalf("Load(after restart) error = %v", err)
	}
	if loaded.State != domain.StateCompleted || loaded.Session != "ses_restart" {
		t.Fatalf("Load(after restart) = %+v, want completed with ses_restart", loaded)
	}

	if _, err := reopened.BindSession(t.Context(), issued.ID, domain.SessionID("ses_replay")); !errors.Is(err, domain.ErrTerminal) {
		t.Fatalf("BindSession(after restart) error = %v, want ErrTerminal", err)
	}
	if _, err := reopened.Fail(t.Context(), issued.ID); !errors.Is(err, domain.ErrTerminal) {
		t.Fatalf("Fail(after restart) error = %v, want ErrTerminal", err)
	}
	if _, err := reopened.Complete(t.Context(), issued.ID); !errors.Is(err, domain.ErrTerminal) {
		t.Fatalf("Complete(after restart) error = %v, want ErrTerminal", err)
	}
	if _, err := os.Stat(grantFile(t, restarted, issued.ID, terminalSuffix)); err != nil {
		t.Fatalf("terminal record did not survive restart: %v", err)
	}
	if _, err := os.Stat(grantFile(t, restarted, issued.ID, sessionSuffix)); err != nil {
		t.Fatalf("session record did not survive restart: %v", err)
	}
}

// TestBindSessionObservesTerminalThatWinsBeforeSessionPublication pauses the
// session marker after the caller validated the transition but before it is
// created, lets Fail publish and return, and then proves BindSession's
// post-publication reload observed the permanent terminal outcome instead of
// returning a stale bound grant.
func TestBindSessionObservesTerminalThatWinsBeforeSessionPublication(t *testing.T) {
	store := newTestStore(t)
	entered := make(chan struct{}, 1)
	release := make(chan struct{})
	releaseOnce := sync.OnceFunc(func() { close(release) })
	t.Cleanup(releaseOnce)
	store.hooks.beforeSession = func() {
		entered <- struct{}{}
		<-release
	}
	service := NewService(store)
	issued := mustIssue(t, service, domain.Binding{48})
	mustClaim(t, service, issued.ID)

	type bindResult struct {
		grant domain.Grant
		err   error
	}
	result := make(chan bindResult, 1)
	go func() {
		bound, bindErr := service.BindSession(t.Context(), issued.ID, domain.SessionID("ses_after_failure"))
		result <- bindResult{grant: bound, err: bindErr}
	}()
	<-entered

	failed, err := service.Fail(t.Context(), issued.ID)
	if err != nil {
		t.Fatalf("Fail() error = %v", err)
	}
	releaseOnce()
	got := <-result

	if !errors.Is(got.err, domain.ErrTerminal) {
		t.Fatalf("BindSession() error = %v, want ErrTerminal", got.err)
	}
	if got.grant.State != domain.StateFailed {
		t.Fatalf("BindSession() state = %v, want failed", got.grant.State)
	}
	if got.grant.Session != "ses_after_failure" {
		t.Fatalf("BindSession() session = %q, want the published ses_after_failure", got.grant.Session)
	}
	if got.grant.ID != failed.ID || got.grant.Binding != failed.Binding {
		t.Fatalf("BindSession() grant = %+v, want identity of %+v", got.grant, failed)
	}

	restarted, err := NewFileStore(store.root)
	if err != nil {
		t.Fatalf("NewFileStore(restart) error = %v", err)
	}
	loaded, err := restarted.Load(t.Context(), issued.ID)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded != got.grant {
		t.Fatalf("Load() = %+v, want %+v", loaded, got.grant)
	}

	if _, err := service.BindSession(t.Context(), issued.ID, domain.SessionID("ses_replay")); !errors.Is(err, domain.ErrTerminal) {
		t.Fatalf("replayed BindSession() error = %v, want ErrTerminal", err)
	}
	if _, err := service.Fail(t.Context(), issued.ID); !errors.Is(err, domain.ErrTerminal) {
		t.Fatalf("replayed Fail() error = %v, want ErrTerminal", err)
	}
	if _, err := service.Complete(t.Context(), issued.ID); !errors.Is(err, domain.ErrTerminal) {
		t.Fatalf("Complete(after failure) error = %v, want ErrTerminal", err)
	}
}

// TestFailReturnsSessionThatWonBeforeTerminalPublication pauses Fail after it
// validated the transition but before the terminal marker exists, lets
// BindSession publish and return, and then proves Fail's post-publication reload
// returned the session that won the race.
func TestFailReturnsSessionThatWonBeforeTerminalPublication(t *testing.T) {
	store := newTestStore(t)
	entered := make(chan struct{}, 1)
	release := make(chan struct{})
	releaseOnce := sync.OnceFunc(func() { close(release) })
	t.Cleanup(releaseOnce)
	store.hooks.beforeTerminal = func() {
		entered <- struct{}{}
		<-release
	}
	service := NewService(store)
	issued := mustIssue(t, service, domain.Binding{49})
	mustClaim(t, service, issued.ID)

	type failResult struct {
		grant domain.Grant
		err   error
	}
	result := make(chan failResult, 1)
	go func() {
		failed, failErr := service.Fail(t.Context(), issued.ID)
		result <- failResult{grant: failed, err: failErr}
	}()
	<-entered

	bound, err := service.BindSession(t.Context(), issued.ID, domain.SessionID("ses_fail_race"))
	if err != nil {
		t.Fatalf("BindSession() error = %v", err)
	}
	releaseOnce()
	got := <-result

	if got.err != nil {
		t.Fatalf("Fail() error = %v", got.err)
	}
	if got.grant.State != domain.StateFailed {
		t.Fatalf("Fail() state = %v, want failed", got.grant.State)
	}
	if got.grant.Session != bound.Session {
		t.Fatalf("Fail() session = %q, want race winner %q", got.grant.Session, bound.Session)
	}

	restarted, err := NewFileStore(store.root)
	if err != nil {
		t.Fatalf("NewFileStore(restart) error = %v", err)
	}
	loaded, err := restarted.Load(t.Context(), issued.ID)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded != got.grant {
		t.Fatalf("Load() = %+v, want %+v", loaded, got.grant)
	}
}

func TestOnlyOneTerminalOutcomeWins(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store)
	issued := mustIssue(t, service, domain.Binding{50})
	mustClaim(t, service, issued.ID)
	if _, err := service.BindSession(t.Context(), issued.ID, domain.SessionID("ses_terminal_race")); err != nil {
		t.Fatalf("BindSession() error = %v", err)
	}

	const contenders = 16
	start := make(chan struct{})
	results := make(chan error, contenders*2)
	var wait sync.WaitGroup
	for range contenders {
		wait.Go(func() {
			<-start
			_, err := service.Complete(t.Context(), issued.ID)
			results <- err
		})
		wait.Go(func() {
			<-start
			_, err := service.Fail(t.Context(), issued.ID)
			results <- err
		})
	}
	close(start)
	wait.Wait()
	close(results)

	winners := 0
	for err := range results {
		if err == nil {
			winners++
			continue
		}
		if !errors.Is(err, domain.ErrTerminal) {
			t.Fatalf("terminal contender error = %v, want ErrTerminal", err)
		}
	}
	if winners != 1 {
		t.Fatalf("terminal winners = %d, want exactly 1", winners)
	}

	loaded, err := service.Load(t.Context(), issued.ID)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.State != domain.StateCompleted && loaded.State != domain.StateFailed {
		t.Fatalf("Load() state = %v, want one terminal outcome", loaded.State)
	}
}

func TestSessionPublicationFailureAfterMarkerKeepsAuthority(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store)
	issued := mustIssue(t, service, domain.Binding{51})
	mustClaim(t, service, issued.ID)

	ctx, cancel := context.WithCancel(t.Context())
	err := store.writeSessionRecordWith(ctx, issued.ID, domain.SessionID("ses_durable"), publisher{
		syncFile: (*os.File).Sync,
		syncParent: func(context.Context, *os.Root, string) error {
			cancel()
			return ctx.Err()
		},
	})
	if !errors.Is(err, ErrStorageFailure) || !errors.Is(err, context.Canceled) {
		t.Fatalf("writeSessionRecordWith() error = %v, want storage failure wrapping context.Canceled", err)
	}
	if _, statErr := os.Stat(grantFile(t, store, issued.ID, sessionSuffix)); statErr != nil {
		t.Fatalf("session marker was removed after a post-create failure: %v", statErr)
	}
	if _, err := service.BindSession(t.Context(), issued.ID, domain.SessionID("ses_second")); !errors.Is(err, domain.ErrAlreadyBound) {
		t.Fatalf("BindSession(after post-create failure) error = %v, want ErrAlreadyBound", err)
	}
}

func TestTerminalPublicationFailureAfterMarkerKeepsAuthority(t *testing.T) {
	store := newTestStore(t)
	service := NewService(store)
	issued := mustIssue(t, service, domain.Binding{52})
	mustClaim(t, service, issued.ID)
	if _, err := service.BindSession(t.Context(), issued.ID, domain.SessionID("ses_terminal_durable")); err != nil {
		t.Fatalf("BindSession() error = %v", err)
	}

	failure := errors.New("directory flush refused")
	err := store.writeTerminalRecordWith(t.Context(), issued.ID, domain.StateFailed, publisher{
		syncFile:   (*os.File).Sync,
		syncParent: func(context.Context, *os.Root, string) error { return failure },
	})
	if !errors.Is(err, ErrStorageFailure) || !errors.Is(err, failure) {
		t.Fatalf("writeTerminalRecordWith() error = %v, want storage failure wrapping the directory failure", err)
	}
	if _, statErr := os.Stat(grantFile(t, store, issued.ID, terminalSuffix)); statErr != nil {
		t.Fatalf("terminal marker was removed after a post-create failure: %v", statErr)
	}

	loaded, err := service.Load(t.Context(), issued.ID)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.State != domain.StateFailed {
		t.Fatalf("Load() state = %v, want the durable failed outcome", loaded.State)
	}
	if _, err := service.Fail(t.Context(), issued.ID); !errors.Is(err, domain.ErrTerminal) {
		t.Fatalf("Fail(after post-create failure) error = %v, want ErrTerminal", err)
	}
}

func mustClaim(t *testing.T, service *Service, id domain.ID) {
	t.Helper()
	if _, err := service.Claim(t.Context(), id); err != nil {
		t.Fatalf("Claim() error = %v", err)
	}
}

func assertNoSessionMarker(t *testing.T, store *FileStore, id domain.ID) {
	t.Helper()
	if _, err := os.Stat(grantFile(t, store, id, sessionSuffix)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("session marker exists after a refused bind: %v", err)
	}
}

func assertNoTerminalMarker(t *testing.T, store *FileStore, id domain.ID) {
	t.Helper()
	if _, err := os.Stat(grantFile(t, store, id, terminalSuffix)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("terminal marker exists before a terminal transition: %v", err)
	}
}

func assertTerminalState(t *testing.T, store *FileStore, id domain.ID, want domain.State) {
	t.Helper()
	data, err := os.ReadFile(grantFile(t, store, id, terminalSuffix))
	if err != nil {
		t.Fatalf("ReadFile(terminal record) error = %v", err)
	}
	var record terminalRecord
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatalf("Unmarshal(terminal record) error = %v", err)
	}
	if record.Version != recordVersion || record.State != want.String() {
		t.Fatalf("terminal record = %+v, want %s", record, want)
	}
}
