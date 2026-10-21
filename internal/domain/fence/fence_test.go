package fence_test

import (
	"testing"

	"github.com/biggz/odachi/internal/domain/fence"
)

// ruleFor builds a single-rule fence with the given paths and deps.
func ruleFor(paths, deps []string) []fence.Rule {
	return []fence.Rule{{Not: "pagos", Paths: paths, Deps: deps}}
}

func TestSegmentGlobMatchesAnyPathSegment(t *testing.T) {
	if got := fence.Check(ruleFor([]string{"payment*"}, nil), []string{"internal/payments/stripe.go"}, nil); len(got) != 1 {
		t.Fatalf("pattern %q against %q: got %d violations, want 1", "payment*", "internal/payments/stripe.go", len(got))
	}
	if got := fence.Check(ruleFor([]string{"payment*"}, nil), []string{"internal/order.go"}, nil); len(got) != 0 {
		t.Fatalf("pattern %q against %q: got %d violations, want 0", "payment*", "internal/order.go", len(got))
	}
}

func TestDoubleStarMatchesAnyNumberOfSegments(t *testing.T) {
	if got := fence.Check(ruleFor([]string{"**/payments/**"}, nil), []string{"internal/payments/stripe.go"}, nil); len(got) != 1 {
		t.Fatalf("pattern %q against %q: got %d violations, want 1", "**/payments/**", "internal/payments/stripe.go", len(got))
	}
	if got := fence.Check(ruleFor([]string{"**/payments/**"}, nil), []string{"internal/orders/x.go"}, nil); len(got) != 0 {
		t.Fatalf("pattern %q against %q: got %d violations, want 0", "**/payments/**", "internal/orders/x.go", len(got))
	}
}

func TestDoubleStarPrefixWithSegmentGlob(t *testing.T) {
	if got := fence.Check(ruleFor([]string{"**/auth/patient*"}, nil), []string{"pkg/auth/patient.go"}, nil); len(got) != 1 {
		t.Fatalf("pattern %q against %q: got %d violations, want 1", "**/auth/patient*", "pkg/auth/patient.go", len(got))
	}
	if got := fence.Check(ruleFor([]string{"**/auth/patient*"}, nil), []string{"pkg/auth/user.go"}, nil); len(got) != 0 {
		t.Fatalf("pattern %q against %q: got %d violations, want 0", "**/auth/patient*", "pkg/auth/user.go", len(got))
	}
}

func TestTrailingDoubleStarCoversSubtree(t *testing.T) {
	if got := fence.Check(ruleFor([]string{"internal/**"}, nil), []string{"internal/a/b.go"}, nil); len(got) != 1 {
		t.Fatalf("pattern %q against %q: got %d violations, want 1", "internal/**", "internal/a/b.go", len(got))
	}
	if got := fence.Check(ruleFor([]string{"internal/**"}, nil), []string{"cmd/main.go"}, nil); len(got) != 0 {
		t.Fatalf("pattern %q against %q: got %d violations, want 0", "internal/**", "cmd/main.go", len(got))
	}
}

func TestMatchingIsCaseInsensitive(t *testing.T) {
	if got := fence.Check(ruleFor([]string{"Login*"}, nil), []string{"web/login.tsx"}, nil); len(got) != 1 {
		t.Fatalf("pattern %q against %q: got %d violations, want 1 (matching is case-insensitive)", "Login*", "web/login.tsx", len(got))
	}
	if got := fence.Check(ruleFor([]string{"payment*"}, nil), []string{"internal/Payments/Stripe.go"}, nil); len(got) != 1 {
		t.Fatalf("pattern %q against %q: got %d violations, want 1 (matching is case-insensitive)", "payment*", "internal/Payments/Stripe.go", len(got))
	}
}

func TestDepMatchesAnyDependencySegment(t *testing.T) {
	cases := []struct {
		pattern string
		dep     string
		want    int
	}{
		{"stripe", "github.com/stripe/stripe-go/v76", 1},
		{"stripe", "github.com/foo/bar", 0},
		{"mercadopago", "github.com/mercadopago/sdk-go", 1},
	}
	for _, tc := range cases {
		got := fence.Check(ruleFor(nil, []string{tc.pattern}), nil, []string{tc.dep})
		if len(got) != tc.want {
			t.Fatalf("dep pattern %q against %q: got %d violations, want %d", tc.pattern, tc.dep, len(got), tc.want)
		}
	}
}

func TestViolationCarriesRuleWhatAndKind(t *testing.T) {
	rules := []fence.Rule{{
		Not:   "pagos",
		Paths: []string{"payment*"},
		Deps:  []string{"stripe"},
	}}
	got := fence.Check(rules, []string{"internal/payments/stripe.go"}, []string{"github.com/stripe/stripe-go/v76"})

	if len(got) != 2 {
		t.Fatalf("got %d violations, want 2 (one path, one dep): %#v", len(got), got)
	}

	byKind := map[string]fence.Violation{}
	for _, v := range got {
		if _, seen := byKind[v.Kind]; seen {
			t.Fatalf("kind %q reported twice: %#v", v.Kind, got)
		}
		byKind[v.Kind] = v
	}

	pathViolation, ok := byKind[fence.KindPath]
	if !ok {
		t.Fatalf("missing violation of kind %q: %#v", fence.KindPath, got)
	}
	if pathViolation.Rule != "pagos" || pathViolation.What != "internal/payments/stripe.go" {
		t.Fatalf("path violation = %#v, want rule %q and what %q", pathViolation, "pagos", "internal/payments/stripe.go")
	}

	depViolation, ok := byKind[fence.KindDep]
	if !ok {
		t.Fatalf("missing violation of kind %q: %#v", fence.KindDep, got)
	}
	if depViolation.Rule != "pagos" || depViolation.What != "github.com/stripe/stripe-go/v76" {
		t.Fatalf("dep violation = %#v, want rule %q and what %q", depViolation, "pagos", "github.com/stripe/stripe-go/v76")
	}
}

func TestOneViolationPerRulePerOffendingPath(t *testing.T) {
	rules := []fence.Rule{
		{Not: "pagos", Paths: []string{"payment*"}},
		{Not: "todo lo interno", Paths: []string{"internal/**"}},
	}
	got := fence.Check(rules, []string{"internal/payments/stripe.go"}, nil)
	if len(got) != 2 {
		t.Fatalf("got %d violations, want 2 (one per rule): %#v", len(got), got)
	}
	if got[0].Rule == got[1].Rule {
		t.Fatalf("both violations name rule %q, want one per rule: %#v", got[0].Rule, got)
	}
}

func TestRuleWithoutMatchProducesNothing(t *testing.T) {
	got := fence.Check(ruleFor([]string{"payment*"}, []string{"stripe"}), []string{"internal/order.go"}, []string{"github.com/foo/bar"})
	if len(got) != 0 {
		t.Fatalf("got %d violations, want 0: %#v", len(got), got)
	}
}

func TestEmptyInputsProduceEmptyResult(t *testing.T) {
	if got := fence.Check(nil, nil, nil); len(got) != 0 {
		t.Fatalf("Check(nil, nil, nil) returned %d violations, want 0: %#v", len(got), got)
	}
	if got := fence.Check(ruleFor([]string{"payment*"}, nil), nil, nil); len(got) != 0 {
		t.Fatalf("Check with no changed paths returned %d violations, want 0: %#v", len(got), got)
	}
	if got := fence.Check(nil, []string{"internal/payments/stripe.go"}, nil); len(got) != 0 {
		t.Fatalf("Check with no rules returned %d violations, want 0: %#v", len(got), got)
	}
}
