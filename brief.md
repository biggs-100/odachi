# the-verifier-runs-focused-commands

## What must exist
- the delegation policy names the verifier's commands: the focused package the change touches, never a second full gate, because a full suite beside the gate competes for the machine and makes the measurement wrong

## Done when
- `slices/references/delegation.md` says the request carries the focused package or packages, never a second `go test ./...`, and says why
- a test in `skills/embed_test.go` pins the rule, observed red before the text lands
- `odachi skills --check` is green and the slices body stays within the sixty-line cap
- `odachi init` leaves the installed copy equal to the tree
- the changelog's Unreleased section names the rule
- `go test ./...` is green

## Not included (first version)
- no change to the verifier subagent, its permissions or its prompt
- no change to `tests` or to `drive`: the test discipline and the route already own their part
- no change to the gate or to its timeout
- no new verb and no new flag
- no change to the voice or the interview's questions
- no change to the stored proofs: never touched
- no change to `odachi role` or to the skill it renders
- no change to the design canvas, its token or the push plugin
- no change to the artboard format
- no change to the release workflow
- no change to the skills the feature does not touch: tests, drive, writing, pull-request, commits and triage
- the `all:*` pin is another feature

## Decisions taken
- the rule lives in `slices/references/delegation.md` because that file owns the handoff; `tests` owns what a red means, not what a subagent runs
- the pin is a string in the existing delegation test, because a skill is its folder (0205) and the reference travels with it
- the reason is measured, not stylistic: a full suite beside the gate competes for the machine and made 686 seconds out of a 207-second package (0217)
- the verifier stays read-only: the rule changes the commands it receives, not what it may do
