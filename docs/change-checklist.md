# Change checklist

Run in this order. A later layer never substitutes for an earlier one: a green
build with no focused test is not verification.

1. Focused package tests — `go test ./internal/<pkg>/...`
2. Full suite — `go test ./...`
3. Build — `go build ./...`
4. Vet — `go vet ./...`
5. Behaviour probe — run the command a human would run and read the output.

## Rules

- Every new domain package needs a test that fails without the implementation
  (observe RED first, then GREEN).
- Every user-facing message must name a command that resolves the block. A
  blocked result without `Next` is a defect, not a style issue.
- Tests and docs travel with the behavior; a work unit of about 400 authored
  lines is the review budget.
- The full suite runs once per tree: `odachi check --run "go test ./..."` is
  that run. Do not run `go test ./...` first and then again inside check: the
  proof is the same evidence, and the second run is pure cost.
- The domain stays stdlib-only. `internal/archtest` enforces it; do not edit the
  guard to make a change fit.
