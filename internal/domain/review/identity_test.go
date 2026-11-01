package review_test

import (
	"testing"

	"github.com/biggs-100/odachi/internal/domain/review"
)

// TestModelIdentityIgnoresTheVariant: a variant is an effort or a setting of the
// same model, so two values that differ only there are the same judge. Calling
// them different is how a review gets accepted as a second opinion it never was.
func TestModelIdentityIgnoresTheVariant(t *testing.T) {
	cases := map[string]string{
		"opencode-go/deepseek-v4.1-flash#high": "opencode-go/deepseek-v4.1-flash",
		"  OpenCode-Go/DeepSeek-v4.1-flash ":   "opencode-go/deepseek-v4.1-flash",
		"provider/model":                       "provider/model",
		"provider/model#":                      "provider/model",
		"":                                     "",
	}

	for value, want := range cases {
		if got := review.ModelIdentity(value); got != want {
			t.Errorf("ModelIdentity(%q) = %q, want %q", value, got, want)
		}
	}
}

// TestTheSameModelAtAnotherEffortIsTheSameJudge: the pair the gate compares is the
// identity, not the setting.
func TestTheSameModelAtAnotherEffortIsTheSameJudge(t *testing.T) {
	if review.ModelIdentity("provider/model#high") != review.ModelIdentity("provider/model") {
		t.Fatal("the same model at another effort reads as two judges")
	}
	if review.ModelIdentity("provider/model") == review.ModelIdentity("provider/other") {
		t.Fatal("two different models read as one judge")
	}
}
