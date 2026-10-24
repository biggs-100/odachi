package hostmodel_test

import (
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/hostmodel"
)

// exported is a trimmed `opencode session export`: the info block is the whole
// contract, and the messages below it are there to prove they are not read.
const exported = `{
  "info": {
    "id": "ses_123",
    "model": {
      "id": "deepseek-v4.1-flash",
      "providerID": "opencode-go",
      "variant": "max"
    }
  },
  "messages": [{"role": "user"}]
}`

// TestAModelComesOutAsProviderNameAndVariant: the shape is the one the reviewer
// tool takes, so the placeholder never needs a second translation at the shell.
func TestAModelComesOutAsProviderNameAndVariant(t *testing.T) {
	const want = "opencode-go/deepseek-v4.1-flash#max"
	if got := hostmodel.ParseSession([]byte(exported)); got != want {
		t.Fatalf("ParseSession = %q, want %q", got, want)
	}
}

// TestAModelWithoutAVariantIsStillAModel: a provider with no variants must not
// leave a trailing #, which the reviewer would read as an empty variant.
func TestAModelWithoutAVariantIsStillAModel(t *testing.T) {
	doc := `{"info":{"model":{"id":"flash","providerID":"go"}}}`
	if got, want := hostmodel.ParseSession([]byte(doc)), "go/flash"; got != want {
		t.Fatalf("ParseSession = %q, want %q", got, want)
	}
}

// TestAnAnswerWithoutAModelIsNoModel: a half-answer is not an identity. The
// caller decides what a missing model means; inventing one here would hide it.
func TestAnAnswerWithoutAModelIsNoModel(t *testing.T) {
	for name, doc := range map[string]string{
		"no info":      `{"messages":[]}`,
		"no provider":  `{"info":{"model":{"id":"flash"}}}`,
		"no name":      `{"info":{"model":{"providerID":"go"}}}`,
		"not json":     `opencode: no such session`,
		"empty object": `{}`,
	} {
		t.Run(name, func(t *testing.T) {
			if got := hostmodel.ParseSession([]byte(doc)); got != "" {
				t.Fatalf("ParseSession = %q, want no model", got)
			}
		})
	}
}

// TestTheExportedModelWinsOverTheProbe: any host can set the variable, so the
// contract is a name and not a vendor. With it set, odachi does not start a CLI.
func TestTheExportedModelWinsOverTheProbe(t *testing.T) {
	t.Setenv(hostmodel.EnvModel, "other/host-model#low")
	t.Setenv(hostmodel.SessionEnv, "ses_unreachable")

	if got, want := hostmodel.Model(""), "other/host-model#low"; got != want {
		t.Fatalf("Model = %q, want %q", got, want)
	}
}

// TestNoHostNoModel: outside an agent that can answer, the answer is empty. No
// environment, no probe, no guess.
func TestNoHostNoModel(t *testing.T) {
	t.Setenv(hostmodel.EnvModel, "")
	t.Setenv(hostmodel.SessionEnv, "")

	if got := hostmodel.Model(""); got != "" {
		t.Fatalf("Model = %q, want no model", got)
	}
}
