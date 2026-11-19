package profile_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/profile"
)

// specFieldOrder is the closed set from docs/odachi-bootstrap.md, Contract
// A. A profile changes ceremony, never rigor: there is no ninth field.
var specFieldOrder = []string{
	"Name",
	"Extends",
	"BriefSource",
	"Questions",
	"Assumptions",
	"Output",
	"Proof",
	"Explain",
}

// allowedEnumValues is Contract A read as a table: the only values the schema
// can hold, keyed by field name.
var allowedEnumValues = map[string][]string{
	"BriefSource": {"interview", "repo", "file"},
	"Questions":   {"one-by-one", "batched", "all"},
	"Assumptions": {"never", "declare", "prefer-safe"},
	"Output":      {"line", "detail", "json"},
	"Proof":       {"command", "suite-or-ci"},
	"Explain":     {"why", "none"},
}

func TestSchemaHasExactlyTheAllowedFields(t *testing.T) {
	typ := reflect.TypeOf(profile.Profile{})

	var got []string
	for i := 0; i < typ.NumField(); i++ {
		got = append(got, typ.Field(i).Name)
	}

	if !reflect.DeepEqual(got, specFieldOrder) {
		t.Fatalf("profile.Profile fields = %v, want %v; the profile schema is a closed set with exactly these eight fields (docs/odachi-bootstrap.md, Contract A and B)", got, specFieldOrder)
	}
}

func TestSchemaCannotExpressAGate(t *testing.T) {
	forbidden := []string{
		"gate", "allow", "skip", "disable", "bypass",
		"enforce", "require", "deny", "force",
	}

	typ := reflect.TypeOf(profile.Profile{})
	for i := 0; i < typ.NumField(); i++ {
		name := strings.ToLower(typ.Field(i).Name)
		for _, word := range forbidden {
			if strings.Contains(name, word) {
				t.Errorf("field %q contains %q: a profile may change ceremony, never rigor (docs/odachi-bootstrap.md, Contract B)", typ.Field(i).Name, word)
			}
		}
	}
}

func TestValidateAcceptsEmptyAndEveryAllowedValue(t *testing.T) {
	if err := profile.Validate(profile.Profile{Name: "vacio"}); err != nil {
		t.Fatalf("Validate with every enum field empty = %v, want nil (empty means inherited)", err)
	}

	for field, values := range allowedEnumValues {
		for _, value := range values {
			p := profile.Profile{Name: "prueba"}
			setField(&p, field, value)
			if err := profile.Validate(p); err != nil {
				t.Errorf("Validate with %s = %q returned %v, want nil", field, value, err)
			}
		}
	}
}

func TestValidateRejectsUnknownValue(t *testing.T) {
	p := profile.Profile{Name: "prueba", Output: "quiet"}

	err := profile.Validate(p)
	if err == nil {
		t.Fatal("Validate with output = \"quiet\" returned nil, want an error")
	}

	message := err.Error()
	if !strings.Contains(strings.ToLower(message), "output") {
		t.Errorf("error %q does not name the field %q", message, "output")
	}
	for _, allowed := range allowedEnumValues["Output"] {
		if !strings.Contains(message, allowed) {
			t.Errorf("error %q does not name the allowed value %q", message, allowed)
		}
	}
}

func TestResolveChildOverridesParent(t *testing.T) {
	known := map[string]profile.Profile{
		"novato": {Name: "novato", BriefSource: "interview", Output: "line"},
		"senior": {Name: "senior", Extends: "novato", BriefSource: "repo", Output: "detail"},
	}

	got, err := profile.Resolve(known, "senior")
	if err != nil {
		t.Fatalf("Resolve(senior) = %v, want nil", err)
	}
	if got.Name != "senior" {
		t.Errorf("Name = %q, want %q", got.Name, "senior")
	}
	if got.BriefSource != "repo" {
		t.Errorf("BriefSource = %q, want %q (child wins over parent)", got.BriefSource, "repo")
	}
	if got.Output != "detail" {
		t.Errorf("Output = %q, want %q (child wins over parent)", got.Output, "detail")
	}
}

func TestResolveInheritsFromParent(t *testing.T) {
	known := map[string]profile.Profile{
		"novato": {Name: "novato", Assumptions: "never", Explain: "why"},
		"senior": {Name: "senior", Extends: "novato", Output: "detail"},
	}

	got, err := profile.Resolve(known, "senior")
	if err != nil {
		t.Fatalf("Resolve(senior) = %v, want nil", err)
	}
	if got.Assumptions != "never" {
		t.Errorf("Assumptions = %q, want %q (inherited)", got.Assumptions, "never")
	}
	if got.Explain != "why" {
		t.Errorf("Explain = %q, want %q (inherited)", got.Explain, "why")
	}
}

func TestResolveUnknownNameIsAnError(t *testing.T) {
	_, err := profile.Resolve(profile.Builtin(), "super-senior")
	if err == nil {
		t.Fatal("Resolve(super-senior) returned nil error, want an error naming the unknown profile")
	}
	if !strings.Contains(err.Error(), "super-senior") {
		t.Errorf("error %q does not name the unknown profile %q", err, "super-senior")
	}
}

func TestResolveTwoNodeCycleIsAnError(t *testing.T) {
	known := map[string]profile.Profile{
		"a": {Name: "a", Extends: "b"},
		"b": {Name: "b", Extends: "a"},
	}

	_, err := profile.Resolve(known, "a")
	if err == nil {
		t.Fatal("Resolve over a two-node cycle returned nil error, want an error naming the cycle")
	}
	if !strings.Contains(err.Error(), "ciclo") && !strings.Contains(err.Error(), "cycle") {
		t.Errorf("error %q does not name the cycle", err)
	}
}

func TestInfer(t *testing.T) {
	tests := []struct {
		name  string
		facts profile.Facts
		want  string
	}{
		{"machine", profile.Facts{Machine: true}, "ci"},
		{"machine wins over brief", profile.Facts{Machine: true, HasBrief: true}, "ci"},
		{"has brief", profile.Facts{HasBrief: true}, "senior"},
		{"nothing", profile.Facts{}, "novato"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := profile.Infer(tc.facts); got != tc.want {
				t.Fatalf("Infer(%+v) = %q, want %q", tc.facts, got, tc.want)
			}
		})
	}
}

func TestBuiltinChain(t *testing.T) {
	builtin := profile.Builtin()

	extends := map[string]string{
		"novato": "",
		"senior": "novato",
		"ci":     "senior",
	}
	for name, want := range extends {
		p, ok := builtin[name]
		if !ok {
			t.Fatalf("Builtin() has no %q profile", name)
		}
		if p.Extends != want {
			t.Errorf("Builtin()[%q].Extends = %q, want %q", name, p.Extends, want)
		}
		if err := profile.Validate(p); err != nil {
			t.Errorf("Builtin()[%q] does not validate: %v", name, err)
		}
	}
}

func TestBuiltinCIResolves(t *testing.T) {
	got, err := profile.Resolve(profile.Builtin(), "ci")
	if err != nil {
		t.Fatalf("Resolve(ci) = %v, want nil", err)
	}

	want := map[string]string{
		"BriefSource": "file",
		"Questions":   "all",
		"Assumptions": "never",
		"Output":      "json",
		"Explain":     "none",
	}
	for field, value := range want {
		if have := fieldValue(got, field); have != value {
			t.Errorf("Resolve(ci).%s = %q, want %q", field, have, value)
		}
	}
	if err := profile.Validate(got); err != nil {
		t.Errorf("Resolve(ci) does not validate: %v", err)
	}
}

func TestBuiltinSeniorResolves(t *testing.T) {
	got, err := profile.Resolve(profile.Builtin(), "senior")
	if err != nil {
		t.Fatalf("Resolve(senior) = %v, want nil", err)
	}

	want := map[string]string{
		"BriefSource": "repo",
		"Questions":   "batched",
		"Assumptions": "declare",
		"Output":      "detail",
		"Explain":     "none",
	}
	for field, value := range want {
		if have := fieldValue(got, field); have != value {
			t.Errorf("Resolve(senior).%s = %q, want %q", field, have, value)
		}
	}
}

// setField assigns an enum value by field name.
func setField(p *profile.Profile, field, value string) {
	switch field {
	case "BriefSource":
		p.BriefSource = value
	case "Questions":
		p.Questions = value
	case "Assumptions":
		p.Assumptions = value
	case "Output":
		p.Output = value
	case "Proof":
		p.Proof = value
	case "Explain":
		p.Explain = value
	default:
		panic("unknown enum field: " + field)
	}
}

// fieldValue reads an enum value by field name.
func fieldValue(p profile.Profile, field string) string {
	switch field {
	case "BriefSource":
		return p.BriefSource
	case "Questions":
		return p.Questions
	case "Assumptions":
		return p.Assumptions
	case "Output":
		return p.Output
	case "Proof":
		return p.Proof
	case "Explain":
		return p.Explain
	default:
		panic("unknown enum field: " + field)
	}
}

// TestAxesNamesEveryFieldTheFileMayCarry: the editor renders the axes from here,
// so a field the table declares but this list forgets is a field no screen can
// reach. The order is the one odd/profile.toml prints (0336).
func TestAxesNamesEveryFieldTheFileMayCarry(t *testing.T) {
	axes := profile.Axes(profile.Builtin()["novato"])

	want := []string{"brief_source", "questions", "assumptions", "output", "proof", "explain"}
	if len(axes) != len(want) {
		t.Fatalf("Axes = %d fields, want %d: %v", len(axes), len(want), axes)
	}
	for i, axis := range axes {
		if axis.Name != want[i] {
			t.Errorf("Axes[%d] = %q, want %q", i, axis.Name, want[i])
		}
		if len(axis.Allowed) == 0 {
			t.Errorf("%s declares no allowed value, so a screen has nothing to offer", axis.Name)
		}
	}
	if axes[3].Value != "line" {
		t.Errorf("output = %q, want the value the profile carries", axes[3].Value)
	}
}

// TestSetChangesOneFieldAndLeavesTheRest: an editor changes one axis; a profile
// that came from a builtin stays what it was everywhere else (0336).
func TestSetChangesOneFieldAndLeavesTheRest(t *testing.T) {
	before := profile.Builtin()["novato"]

	after, err := profile.Set(before, "output", "detail")
	if err != nil {
		t.Fatalf("Set: %v", err)
	}
	if after.Output != "detail" {
		t.Errorf("output = %q, want detail", after.Output)
	}

	before.Output = "detail"
	if after != before {
		t.Errorf("Set changed another field: %+v", after)
	}
}

// TestSetRefusesAnAxisNobodyDeclares: a name the table does not carry is a typo,
// and the error lists the names that exist (0336).
func TestSetRefusesAnAxisNobodyDeclares(t *testing.T) {
	_, err := profile.Set(profile.Builtin()["novato"], "tone", "loud")
	if err == nil {
		t.Fatal("Set accepted an axis nobody declares")
	}
	for _, want := range []string{"tone", "output", "questions"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %v does not name %s", err, want)
		}
	}
}

// TestSetRefusesAValueTheAxisDoesNotAllow: the allowed values are the contract,
// and a screen must not be able to write one the parser rejects (0336).
func TestSetRefusesAValueTheAxisDoesNotAllow(t *testing.T) {
	_, err := profile.Set(profile.Builtin()["novato"], "output", "yaml")
	if err == nil {
		t.Fatal("Set accepted a value the axis does not allow")
	}
	for _, want := range []string{"output", "yaml", "line", "detail", "json"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %v does not name %s", err, want)
		}
	}
}
