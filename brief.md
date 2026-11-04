# the-review-names-the-model

## What must exist
- the review names the model it can afford, refuses a model the host does not serve before spending the call, and never switches by itself

## Done when
- before a reviewer starts, a lens model that `opencode models` does not list -while it lists that provider- is refused, naming the lens, the model, its source, the session's model and the host's models, with the `-free` names marked as the host's convention
- the refusal names `--model`, `odd/review.toml` and `ODACHI_MODEL`, and says the tool does not switch by itself
- a reviewer that fails after the preflight (could not run, timeout, non-zero exit) names the same options
- when the host cannot list its models, the run starts and the message keeps what it knows
- `odd/review.toml` names `opencode-go/deepseek-v4.1-flash` for the four lenses
- a verdict by the builder's own model is still named as a second reading, not refused (ADR 0048)
- README says the self-review is named, not refused, and that the preflight names the options
- `go test ./...` is green

## Not included (first version)
- no change to what proof/ and odd/tasks are: never touched
- no automatic model switch and no fallback: the tool names, the human chooses
- no change to the timeout, the prompt size or the store's retention
- no change to how a verdict is stored, merged or refuted
- no new verb and no new flag

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: why a preflight is cheaper than a timeout, and why naming options is not the same as choosing one.
