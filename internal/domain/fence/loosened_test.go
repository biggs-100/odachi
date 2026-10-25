package fence_test

import (
	"slices"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/fence"
)

// TestLoosenedNamesWhatDisappeared: the candidate may add rules and forbid more,
// but every rule the base had must survive with at least the paths and
// dependencies it had. Identity is the rule's own text: renaming it is removing
// it, because the name is what the human wrote.
func TestLoosenedNamesWhatDisappeared(t *testing.T) {
	base := []fence.Rule{
		{Not: "monorepo scopes", Paths: []string{"internal/tenants/*", "cmd/scopes/*"}, Deps: []string{"github.com/x/scopes"}},
		{Not: "signing", Paths: []string{"internal/sign/*"}},
	}

	cases := map[string]struct {
		candidate []fence.Rule
		want      []fence.Loosening
	}{
		"the same fence": {base, nil},
		"stricter": {[]fence.Rule{
			{Not: "monorepo scopes", Paths: []string{"internal/tenants/*", "cmd/scopes/*", "extra/*"}, Deps: []string{"github.com/x/scopes"}},
			{Not: "signing", Paths: []string{"internal/sign/*"}},
			{Not: "a new rule", Paths: []string{"x/*"}},
		}, nil},
		"a rule is gone": {
			[]fence.Rule{{Not: "signing", Paths: []string{"internal/sign/*"}}},
			[]fence.Loosening{{Rule: "monorepo scopes", Kind: fence.KindRule}},
		},
		"a path is gone": {
			[]fence.Rule{
				{Not: "monorepo scopes", Paths: []string{"internal/tenants/*"}, Deps: []string{"github.com/x/scopes"}},
				{Not: "signing", Paths: []string{"internal/sign/*"}},
			},
			[]fence.Loosening{{Rule: "monorepo scopes", What: "cmd/scopes/*", Kind: fence.KindPath}},
		},
		"a dependency is gone": {
			[]fence.Rule{
				{Not: "monorepo scopes", Paths: []string{"internal/tenants/*", "cmd/scopes/*"}},
				{Not: "signing", Paths: []string{"internal/sign/*"}},
			},
			[]fence.Loosening{{Rule: "monorepo scopes", What: "github.com/x/scopes", Kind: fence.KindDep}},
		},
		"a renamed rule is a gone rule": {
			[]fence.Rule{
				{Not: "monorepo", Paths: []string{"internal/tenants/*", "cmd/scopes/*"}, Deps: []string{"github.com/x/scopes"}},
				{Not: "signing", Paths: []string{"internal/sign/*"}},
			},
			[]fence.Loosening{{Rule: "monorepo scopes", Kind: fence.KindRule}},
		},
		"an empty fence": {
			nil,
			[]fence.Loosening{
				{Rule: "monorepo scopes", Kind: fence.KindRule},
				{Rule: "signing", Kind: fence.KindRule},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := fence.Loosened(base, tc.candidate)
			if !slices.Equal(got, tc.want) {
				t.Errorf("Loosened() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

// TestLoosenedWithNoBaseIsSilent: a ref with no fence has no policy to lose, so
// there is nothing to compare. The same applies to a candidate that keeps
// everything: silence is the pass.
func TestLoosenedWithNoBaseIsSilent(t *testing.T) {
	if got := fence.Loosened(nil, []fence.Rule{{Not: "x", Paths: []string{"x/*"}}}); len(got) != 0 {
		t.Errorf("Loosened(no base) = %+v, want none", got)
	}
}

// TestLoosenedReadsDuplicateRulesTogether: two rules with the same name can only
// add restrictions, so the candidate's paths are the union of both.
func TestLoosenedReadsDuplicateRulesTogether(t *testing.T) {
	base := []fence.Rule{{Not: "auth", Paths: []string{"internal/auth/*", "internal/token/*"}}}
	candidate := []fence.Rule{
		{Not: "auth", Paths: []string{"internal/auth/*"}},
		{Not: "auth", Paths: []string{"internal/token/*"}},
	}

	if got := fence.Loosened(base, candidate); len(got) != 0 {
		t.Errorf("Loosened() = %+v, want none: the two rows together forbid what the base forbids", got)
	}
}
