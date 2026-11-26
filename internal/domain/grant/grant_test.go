package grant_test

import (
	"errors"
	"strings"
	"testing"

	domain "github.com/biggs-100/odachi/internal/domain/grant"
)

func TestGrantIDsAreOpaque256BitRandomValues(t *testing.T) {
	first, err := domain.NewID()
	if err != nil {
		t.Fatalf("NewID() error = %v", err)
	}
	second, err := domain.NewID()
	if err != nil {
		t.Fatalf("NewID() error = %v", err)
	}
	if first == second {
		t.Fatal("NewID() returned the same grant twice")
	}
	if len(first.String()) != 43 {
		t.Fatalf("encoded ID length = %d, want 43", len(first.String()))
	}
	parsed, err := domain.ParseID(first.String())
	if err != nil {
		t.Fatalf("ParseID() error = %v", err)
	}
	if parsed != first {
		t.Fatalf("ParseID() = %v, want %v", parsed, first)
	}
}

func TestParseIDRejectsNonCanonicalAndUnsafeValues(t *testing.T) {
	for _, value := range []string{
		"",
		"../grant",
		strings.Repeat("A", 42),
		strings.Repeat("A", 44),
		strings.Repeat("A", 42) + "B",
		strings.Repeat("A", 42) + "/",
	} {
		t.Run(value, func(t *testing.T) {
			if _, err := domain.ParseID(value); !errors.Is(err, domain.ErrInvalidID) {
				t.Fatalf("ParseID(%q) error = %v, want ErrInvalidID", value, err)
			}
		})
	}
}

func TestBindingDigestIsCanonicalSHA256(t *testing.T) {
	value := strings.Repeat("ab", 32)
	binding, err := domain.ParseBinding(value)
	if err != nil {
		t.Fatalf("ParseBinding() error = %v", err)
	}
	if binding.String() != value {
		t.Fatalf("binding String() = %q, want %q", binding.String(), value)
	}
	for _, invalid := range []string{"", strings.Repeat("a", 63), strings.Repeat("A", 64), strings.Repeat("g", 64)} {
		if _, err := domain.ParseBinding(invalid); !errors.Is(err, domain.ErrInvalidBinding) {
			t.Fatalf("ParseBinding(%q) error = %v, want ErrInvalidBinding", invalid, err)
		}
	}
}

func TestGrantStateTransitionsAreOneWay(t *testing.T) {
	issued, err := domain.NewGrant(domain.Binding{1})
	if err != nil {
		t.Fatalf("NewGrant() error = %v", err)
	}
	if issued.State != domain.StateIssued {
		t.Fatalf("new grant state = %v, want issued", issued.State)
	}
	claimed, err := issued.Claim()
	if err != nil {
		t.Fatalf("Claim() error = %v", err)
	}
	if claimed.State != domain.StateClaimed {
		t.Fatalf("claimed state = %v, want claimed", claimed.State)
	}
	if _, err := claimed.Claim(); !errors.Is(err, domain.ErrAlreadyClaimed) {
		t.Fatalf("second Claim() error = %v, want ErrAlreadyClaimed", err)
	}
	bound, err := claimed.BindSession(domain.SessionID("ses_fresh_001"))
	if err != nil {
		t.Fatalf("BindSession() error = %v", err)
	}
	if bound.State != domain.StateBound || bound.Session != "ses_fresh_001" {
		t.Fatalf("bound grant = %+v", bound)
	}
	completed, err := bound.Complete()
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if completed.State != domain.StateCompleted {
		t.Fatalf("completed state = %v, want completed", completed.State)
	}
	if _, err := completed.Claim(); !errors.Is(err, domain.ErrAlreadyClaimed) {
		t.Fatalf("Claim() after completion error = %v, want ErrAlreadyClaimed", err)
	}
	if _, err := completed.Fail(); !errors.Is(err, domain.ErrTerminal) {
		t.Fatalf("Fail() after completion error = %v, want ErrTerminal", err)
	}
}

func TestFailureCanTerminateClaimedOrBoundGrant(t *testing.T) {
	for _, bind := range []bool{false, true} {
		t.Run(map[bool]string{false: "claimed", true: "bound"}[bind], func(t *testing.T) {
			issued, err := domain.NewGrant(domain.Binding{2})
			if err != nil {
				t.Fatalf("NewGrant() error = %v", err)
			}
			claimed, err := issued.Claim()
			if err != nil {
				t.Fatalf("Claim() error = %v", err)
			}
			current := claimed
			if bind {
				current, err = claimed.BindSession(domain.SessionID("ses_fresh_002"))
				if err != nil {
					t.Fatalf("BindSession() error = %v", err)
				}
			}
			failed, err := current.Fail()
			if err != nil {
				t.Fatalf("Fail() error = %v", err)
			}
			if failed.State != domain.StateFailed {
				t.Fatalf("failed state = %v, want failed", failed.State)
			}
			if _, err := failed.Complete(); !errors.Is(err, domain.ErrTerminal) {
				t.Fatalf("Complete() after failure error = %v, want ErrTerminal", err)
			}
		})
	}
}

func TestSessionIDValidationRejectsMissingAndMalformedIdentities(t *testing.T) {
	for _, session := range []domain.SessionID{"", " ses", "ses ", "ses\n", domain.SessionID(strings.Repeat("s", 257))} {
		issued, err := domain.NewGrant(domain.Binding{3})
		if err != nil {
			t.Fatalf("NewGrant() error = %v", err)
		}
		claimed, err := issued.Claim()
		if err != nil {
			t.Fatalf("Claim() error = %v", err)
		}
		if _, err := claimed.BindSession(session); !errors.Is(err, domain.ErrInvalidSession) {
			t.Fatalf("BindSession(%q) error = %v, want ErrInvalidSession", session, err)
		}
	}
}

func TestBindSessionRejectsUnicodeFormatControls(t *testing.T) {
	cases := []struct {
		name    string
		session domain.SessionID
	}{
		{"zero width space prefix", "\u200bses_fresh"},
		{"zero width space embedded", "ses\u200bfresh"},
		{"right to left override suffix", "ses_fresh\u202e"},
		{"left to right isolate prefix", "\u2066ses_fresh"},
		{"byte order mark prefix", "\ufeffses_fresh"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			issued, err := domain.NewGrant(domain.Binding{5})
			if err != nil {
				t.Fatalf("NewGrant() error = %v", err)
			}
			claimed, err := issued.Claim()
			if err != nil {
				t.Fatalf("Claim() error = %v", err)
			}
			if _, err := claimed.BindSession(tc.session); !errors.Is(err, domain.ErrInvalidSession) {
				t.Fatalf("BindSession(%q) error = %v, want ErrInvalidSession", tc.session, err)
			}
		})
	}
}

func TestBindSessionRejectsNonASCIISessions(t *testing.T) {
	cases := []struct {
		name    string
		session domain.SessionID
	}{
		{"accented letter", "ses_\u00e9"},
		{"emoji", "ses_\U0001f389"},
		{"fullwidth digit", "ses_\uff11"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			issued, err := domain.NewGrant(domain.Binding{6})
			if err != nil {
				t.Fatalf("NewGrant() error = %v", err)
			}
			claimed, err := issued.Claim()
			if err != nil {
				t.Fatalf("Claim() error = %v", err)
			}
			if _, err := claimed.BindSession(tc.session); !errors.Is(err, domain.ErrInvalidSession) {
				t.Fatalf("BindSession(%q) error = %v, want ErrInvalidSession", tc.session, err)
			}
		})
	}
}

func TestBindSessionAcceptsSafeASCIIGrammar(t *testing.T) {
	for _, session := range []domain.SessionID{"ses_fresh_001", "ses-abc.DEF:123", "ABC123"} {
		t.Run(string(session), func(t *testing.T) {
			issued, err := domain.NewGrant(domain.Binding{7})
			if err != nil {
				t.Fatalf("NewGrant() error = %v", err)
			}
			claimed, err := issued.Claim()
			if err != nil {
				t.Fatalf("Claim() error = %v", err)
			}
			bound, err := claimed.BindSession(session)
			if err != nil {
				t.Fatalf("BindSession(%q) error = %v, want nil", session, err)
			}
			if bound.State != domain.StateBound || bound.Session != session {
				t.Fatalf("bound grant = %+v, want state bound and session %q", bound, session)
			}
		})
	}
}

func TestClaimRefusesInvalidGrant(t *testing.T) {
	id, err := domain.NewID()
	if err != nil {
		t.Fatalf("NewID() error = %v", err)
	}
	binding := domain.Binding{9}
	cases := []struct {
		name  string
		grant domain.Grant
		want  error
	}{
		{"issued with session", domain.Grant{ID: id, Binding: binding, State: domain.StateIssued, Session: "ses_x"}, domain.ErrInvalidSession},
		{"zero id", domain.Grant{Binding: binding, State: domain.StateIssued}, domain.ErrInvalidID},
		{"zero binding", domain.Grant{ID: id, State: domain.StateIssued}, domain.ErrInvalidBinding},
		{"unknown state", domain.Grant{ID: id, Binding: binding, State: domain.State(99)}, domain.ErrInvalidTransition},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			claimed, err := tc.grant.Claim()
			if !errors.Is(err, tc.want) {
				t.Fatalf("Claim() error = %v, want %v", err, tc.want)
			}
			if claimed != (domain.Grant{}) {
				t.Fatalf("Claim() returned %+v, want the zero grant", claimed)
			}
		})
	}
}

func TestGrantValidateAcceptsConsistentStates(t *testing.T) {
	id, err := domain.NewID()
	if err != nil {
		t.Fatalf("NewID() error = %v", err)
	}
	binding := domain.Binding{8}
	valid := []struct {
		name  string
		grant domain.Grant
	}{
		{"issued", domain.Grant{ID: id, Binding: binding, State: domain.StateIssued}},
		{"claimed", domain.Grant{ID: id, Binding: binding, State: domain.StateClaimed}},
		{"bound", domain.Grant{ID: id, Binding: binding, State: domain.StateBound, Session: "ses_ok"}},
		{"completed", domain.Grant{ID: id, Binding: binding, State: domain.StateCompleted, Session: "ses_ok"}},
		{"failed from claimed", domain.Grant{ID: id, Binding: binding, State: domain.StateFailed}},
		{"failed from bound", domain.Grant{ID: id, Binding: binding, State: domain.StateFailed, Session: "ses_ok"}},
	}
	for _, tc := range valid {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.grant.Validate(); err != nil {
				t.Fatalf("Validate() = %v, want nil", err)
			}
		})
	}
}

func TestGrantValidateRejectsInconsistentStates(t *testing.T) {
	id, err := domain.NewID()
	if err != nil {
		t.Fatalf("NewID() error = %v", err)
	}
	binding := domain.Binding{10}
	invalid := []struct {
		name  string
		grant domain.Grant
		want  error
	}{
		{"zero id", domain.Grant{Binding: binding, State: domain.StateIssued}, domain.ErrInvalidID},
		{"zero binding", domain.Grant{ID: id, State: domain.StateIssued}, domain.ErrInvalidBinding},
		{"issued with session", domain.Grant{ID: id, Binding: binding, State: domain.StateIssued, Session: "ses_x"}, domain.ErrInvalidSession},
		{"claimed with session", domain.Grant{ID: id, Binding: binding, State: domain.StateClaimed, Session: "ses_x"}, domain.ErrInvalidSession},
		{"bound without session", domain.Grant{ID: id, Binding: binding, State: domain.StateBound}, domain.ErrInvalidSession},
		{"completed without session", domain.Grant{ID: id, Binding: binding, State: domain.StateCompleted}, domain.ErrInvalidSession},
		{"bound with format control session", domain.Grant{ID: id, Binding: binding, State: domain.StateBound, Session: "\u202eSpoof"}, domain.ErrInvalidSession},
		{"unknown state", domain.Grant{ID: id, Binding: binding, State: domain.State(99)}, domain.ErrInvalidTransition},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.grant.Validate(); !errors.Is(err, tc.want) {
				t.Fatalf("Validate() = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestStateStringNamesEveryLifecycleState(t *testing.T) {
	cases := []struct {
		state domain.State
		want  string
	}{
		{domain.StateIssued, "issued"},
		{domain.StateClaimed, "claimed"},
		{domain.StateBound, "bound"},
		{domain.StateCompleted, "completed"},
		{domain.StateFailed, "failed"},
		{domain.State(99), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.state.String(); got != tc.want {
			t.Fatalf("State(%d).String() = %q, want %q", tc.state, got, tc.want)
		}
	}
}
