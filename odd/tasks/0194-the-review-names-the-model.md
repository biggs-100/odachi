# the-review-names-the-model
Base: e1d3e66ebfc3bccca8d5a9675fbbbc294be1d9ec

## Requirements (verbatim, from the user)
R1 · "the review names the model it can afford, refuses a model the host does not serve before spending the call, and never switches by itself"

## Done when
- before a reviewer starts, a lens model that `opencode models` does not list -while it lists that provider- is refused, naming the lens, the model, its source, the session's model and the host's models, with the `-free` names marked as the host's convention
- the refusal names `--model`, `odd/review.toml` and `ODACHI_MODEL`, and says the tool does not switch by itself
- a reviewer that fails after the preflight (could not run, timeout, non-zero exit) names the same options
- when the host cannot list its models, the run starts and the message keeps what it knows
- `odd/review.toml` names `opencode-go/deepseek-v4.1-flash` for the four lenses
- a verdict by the builder's own model is still named as a second reading, not refused (ADR 0048)
- README says the self-review is named, not refused, and that the preflight names the options
- `go test ./...` is green

## Steps
- [x] T1 · the host answers which models it serves, and the review can ask   R1 · route:inline · tree:e4da2b22f7d89a9aac71ac5cea0f929fab04255b
- [x] T2 · the review refuses a model the host serves a provider for but does not list, before the reviewer starts   R1 · route:inline · tree:b6045b518ed0cb1b30ba44f21b5436c2d8f25a8e
- [x] T3 · a failed reviewer names the model, the lens and the options   R1 · route:inline · tree:e0fd4aa87fb3d4364bb90c0002843e9d2c62b860
- [x] T4 · the project's four lenses name DeepSeek V4.1 Flash   R1 · route:inline · tree:920ac910011bd55d8e6409829a07c5791562e112
- [x] T5 · README says the self-review is named, not refused, and the preflight names the options   R1 · route:inline · tree:e4348a33324830c1bf7eb96b653b7762af32f4de

## History
L1 · "the review names the model it can afford, refuses a model the host does not serve before spending the call, and never switches by itself"
L2 · OFICIO · "a failure names the ways to switch even when the host says nothing" (porque the flag, the project's file and the variable do not depend on a catalog; a dead end is the one thing a failure must not be)
