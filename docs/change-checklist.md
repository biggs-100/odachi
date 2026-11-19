# Change checklist

Run in this order. A later layer never substitutes for an earlier one: a green
build with no focused test is not verification.

1. Focused package tests — `go test ./internal/<pkg>/...`
2. Full suite — `go test ./...`
3. Build — `go build ./...`
4. Vet — `go vet ./...`
5. Behaviour probe — run the command a human would run and read the output.
6. Review debt — at close, copy into `docs/review-debt.md` the potent rows the
   review store still holds about unchanged bytes, and drop the rows whose bytes
   changed. The store is per machine and ignored; that file travels.

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

## The review budget

A candidate's authored lines are every added line plus every deleted line of the
unified diff from the feature's base to the tree that closes it. The headers that
name a file and mark a hunk are not lines, and neither are the context lines a hunk
carries; an empty diff, a rename and a mode change count zero. Nothing under
`proof/` counts, because the proofs are the tool's own evidence, and no path the
repository ignores counts either, because git never puts one in the diff. A file a
command generates does not count: a feature that only reruns the generator wrote no
line. That list is closed and lives in `generatedPaths`
(`internal/app/check/lines.go`), one exact file per entry with the command that owns
it -today the project's registry `SKILLS.md`, owned by `odachi skills --write`, and
the registry beside the skills a person writes, `.agents/skills/SKILLS.md`, owned by
`odachi skill new --scope project`-. A folder or a suffix is not an entry: the
human's own skill in `.agents/skills/<name>/SKILL.md` is authored work and is
charged.
