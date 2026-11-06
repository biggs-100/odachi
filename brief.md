# the-gate-fits-the-suite

## What must exist
- the gate's default bound fits the suite it actually runs: `go test ./...` takes about 700 seconds on this machine (686 for cmd/odachi under contention, 207 alone), so 600 kills a green suite and the default moves to 900

## Done when
- `check.DefaultTimeout` is 900, and `TestTheDefaultTimeoutFitsARealSuite` pins it and the seconds the shell receives
- the comment names the measurement that decided 900, and the help text and `docs/odachi-bootstrap.md` say 900
- the changelog's Unreleased section names the bound
- `odachi check --run "go test ./..."` is green with no `--timeout` on this machine
- `go test ./...` is green

## Not included (first version)
- no change to the gate's verdicts, reasons, exit codes or the proof format
- no change to `--timeout` semantics: it still tightens the bound
- no change to `review.ExecTimeoutSeconds`
- no new verb and no new flag
- no change to the suite's speed, its `-p` or its isolation: the contention is recorded, not fixed
- no change to the voice or the interview's questions
- no change to the stored proofs: never touched
- no change to `odachi role` or to the skill it renders
- no change to the design canvas, its token or the push plugin
- no change to the artboard format
- no change to the release workflow
- no change to the skills: none is part of this feature

## Decisions taken
- 900, not 1200: 686 seconds measured plus about a third of margin; a real hang still fails inside fifteen minutes
- the number lives in the constant, the test and the comment, so the next measurement has one place to land
- the contention (686 seconds under ./... against 207 alone) is recorded in the next session, not fixed here: making the suite faster is another feature
- the gate itself is unchanged: only the default bound moves, and `--timeout` still wins
