# verification-is-proportional

## What must exist
- verification is proportional — a passive change is verified by reading it, a medium one keeps the writer's own verification, and only a named risk adds an independent verifier — and a red is compared against the base and the environment before it is blamed on the change

## Done when
- `tests` names the three tiers (passive, medium, high) and who verifies each one: the reader, the writer, `odachi-verify`
- `tests` says a red is run against the base first, and a failure that was already there is named instead of blamed on the change
- `tests` says a red that comes from the machine is named with its condition, and it is never reported as green and never silently skipped
- `slices/references/delegation.md` keeps "a named risk gets a verifier" and sends the reader to the three tiers, so the delegation policy and the test discipline agree
- a test in `skills/embed_test.go` pins the three tier names and the two comparisons, observed red before the text lands
- the body of `tests` stays within the sixty-line cap and `odachi skills --check` is green
- `odachi init` leaves the installed copy equal to the tree
- the changelog's Unreleased section names the rule
- `go test ./...` is green

## Not included (first version)
- no change to the six risks in `commits` or to their table: they are the named high-risk items already
- no new verb and no new flag: the gate does not measure the tiers
- no change to `odachi check` or to `odachi review` behavior
- no change to the skill contract or its format: only the text of two skills and one pinned test
- no change to the voice or the interview's questions
- no change to the stored proofs: never touched
- no change to `odachi role` or to the skill it renders
- no change to the design canvas, its token or the push plugin
- no change to the artboard format
- no change to the release workflow
- no readback after publishing, and no `assets/` in the skill contract: each is another feature

## Decisions taken
- three tiers, not a percentage: the tier is decided by what the change touches, and the `Risk:` line is the only thing that names a high item
- the base comparison is a command, not a memory: run the same test on the base, because "it already failed" is evidence and a hunch is not
- an environmental failure is named with the check that measures it (`odachi doctor` names the PATH) and never turned into a skip
- the rule lives in `tests` because it is the test discipline; `slices/references/delegation.md` keeps only who runs the verifier
