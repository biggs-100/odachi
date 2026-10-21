package profile_test

import (
	"reflect"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/profile"
)

// TestNextIsEarnedAndNeverAutomatic pins the graduation rule of Contract IV:
// senior is proposed only from novato, and only with 3 features that carry a
// done step and 5 answered vision questions. Craft decisions are shown but do
// not buy the graduation.
func TestNextIsEarnedAndNeverAutomatic(t *testing.T) {
	cases := []struct {
		name    string
		current string
		ev      profile.Evidence
		wantOK  bool
	}{
		{"exactly the threshold", "novato", profile.Evidence{Features: 3, VisionAnswers: 5}, true},
		{"above the threshold", "novato", profile.Evidence{Features: 7, VisionAnswers: 9, CraftDecisions: 4}, true},
		{"one feature short", "novato", profile.Evidence{Features: 2, VisionAnswers: 5}, false},
		{"one answer short", "novato", profile.Evidence{Features: 3, VisionAnswers: 4}, false},
		{"craft decisions are not vision", "novato", profile.Evidence{Features: 3, VisionAnswers: 0, CraftDecisions: 20}, false},
		{"already senior", "senior", profile.Evidence{Features: 9, VisionAnswers: 9}, false},
		{"ci never graduates", "ci", profile.Evidence{Features: 9, VisionAnswers: 9}, false},
		{"a custom profile is not gradable", "mi-perfil", profile.Evidence{Features: 9, VisionAnswers: 9}, false},
		{"nothing at all", "novato", profile.Evidence{}, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			name, ok := profile.Next(c.current, c.ev)
			if ok != c.wantOK {
				t.Fatalf("Next(%q, %+v) ok = %v, want %v", c.current, c.ev, ok, c.wantOK)
			}
			if ok && name != "senior" {
				t.Errorf("Next() = %q, want senior", name)
			}
			if !ok && name != "" {
				t.Errorf("Next() = %q with ok = false, want the empty string", name)
			}
		})
	}
}

// TestApplyReportsEveryChangedField: the human reads exactly what the profile
// change does, field by field, in the order .odachi/profile.toml uses.
func TestApplyReportsEveryChangedField(t *testing.T) {
	current := resolved(t, "novato")
	target := resolved(t, "senior")

	got, changes := profile.Apply(current, target)

	if !reflect.DeepEqual(got, target) {
		t.Errorf("Apply returned %+v, want the target profile %+v", got, target)
	}

	want := []profile.Change{
		{Field: "name", From: "novato", To: "senior"},
		{Field: "extends", From: "", To: "novato"},
		{Field: "brief_source", From: "interview", To: "repo"},
		{Field: "questions", From: "one-by-one", To: "batched"},
		{Field: "assumptions", From: "never", To: "declare"},
		{Field: "output", From: "line", To: "detail"},
		{Field: "proof", From: "command", To: "suite-or-ci"},
		{Field: "explain", From: "why", To: "none"},
	}
	if !reflect.DeepEqual(changes, want) {
		t.Fatalf("Apply changes =\n%+v\nwant\n%+v", changes, want)
	}
}

// TestApplyListsOnlyWhatChanges: applying the profile you already have writes
// nothing and says nothing.
func TestApplyListsOnlyWhatChanges(t *testing.T) {
	current := resolved(t, "senior")

	got, changes := profile.Apply(current, current)
	if len(changes) != 0 {
		t.Errorf("Apply(senior, senior) reported %d changes: %+v", len(changes), changes)
	}
	if !reflect.DeepEqual(got, current) {
		t.Errorf("Apply(senior, senior) = %+v, want the same profile", got)
	}

	// A hand-edited profile that only differs in one field reports one field.
	edited := current
	edited.Output = "detail"
	edited.Explain = "why"
	_, changes = profile.Apply(current, edited)
	want := []profile.Change{{Field: "explain", From: "none", To: "why"}}
	if !reflect.DeepEqual(changes, want) {
		t.Errorf("changes = %+v, want %+v", changes, want)
	}
}

// resolved resolves a built-in profile, so the test uses the shipped values
// instead of a copy that can drift.
func resolved(t *testing.T, name string) profile.Profile {
	t.Helper()
	p, err := profile.Resolve(profile.Builtin(), name)
	if err != nil {
		t.Fatalf("Resolve(%q): %v", name, err)
	}
	return p
}
