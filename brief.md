# review-on-by-default

## What must exist
- a new project is born with the review on: `init` writes `odd/review.toml` with `mode = "on"` and never replaces one

## Done when
- `odachi init` writes `odd/review.toml` with `mode = "on"` when the project has none, through the writer `review mode on` uses
- a project that already has the file keeps it byte for byte, and the run says it kept it
- `init --json` reports the file it wrote or kept
- `odachi check` on a project with that file asks for the verdict the mode owes
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- Unreleased names the change: a contract change, minor (0254)

## Not included (first version)
- no change to a project without the file: absence keeps meaning optional there - back when the human asks for a global default
- no model pinned by init: the reviewer uses the session's model - back when a project needs [lenses] at birth
- no change to the installer, the workflow, the canvas or the proofs - back when one of those paths changes
- no change to what `review mode` means or writes - back when the switch has to mean something else

## Decisions taken
- The default is a file, not an absence: a project that never wrote one keeps the behaviour it had.
- `init` never overwrites the file: an existing one is the human's decision.
- A new project with no reachable model closes by configuring one or by `review waive`; the gate names both.
- It is a contract change: the step a new project follows at a close now includes a verdict.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: whether the review on from the first commit catches what a project only learns later.
