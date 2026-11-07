# the-smallest-change

## What must exist
- before the code, the change names what it reuses, what it adds and what it deletes, and the PR carries the same list to the reviewer: an abstraction needs a second caller, a pattern waits for three uses for the same reason, and a wrong abstraction is inlined back before it is extracted again

## Done when
- `shape` names the smallest change (reuse, add, delete; an abstraction needs a second caller), the rule of three for the same reason, the props test, and the inline-first recovery
- `pull-request` says every PR names its new files, dependencies and abstractions with their reason and the code it reuses, and the reviewer asks what to delete, what was reused and which requirement each line serves
- tests in `skills/embed_test.go` pin both, observed red before the text lands
- `odachi skills --check` is green and the bodies stay under the sixty-line cap
- `odachi init` leaves the installed copy equal to the tree
- the changelog's Unreleased section names the rules
- `go test ./...` is green

## Not included (first version)
- no gate that counts new files or dependencies: that is the next feature
- no change to `commits`, `tests`, `triage` or `slices`: `triage` keeps its deletion rule for bugs
- no new verb and no new flag
- no change to the voice or the interview's questions
- no change to the stored proofs: never touched
- no change to `odachi role` or to the skill it renders
- no change to the design canvas, its token or the push plugin
- no change to the artboard format
- no change to the release workflow
- no change to the skills the feature does not touch: commits, tests, triage, slices, writing, design and ui-craft

## Decisions taken
- the rule lives in `shape` because the decision is taken before the brief; `triage` keeps its deletion rule for bugs
- "same reason" is the test, not the count: three copies that one decision would change at once
- the PR carries the list because a reviewer cannot reconstruct intent from the diff
- the recovery is inline-first, because a wrong abstraction costs more than the duplication it replaced
