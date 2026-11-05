# the-ledger-of-0205-is-paid

## What must exist
- the four rows `docs/review-debt.md` left open are paid: the pack report counts a backup only when one ran, the gates check reads the table instead of searching it, the missing-reference drift has its test, and a git that cannot answer `Head` is an error instead of an empty history

## Done when
- a skill whose `SKILL.md` is unchanged and whose only difference is a new reference is not reported as backed up, in a real run or a dry run
- `## Decision gates` is read as a table and `## References` as a list, so a path named only in prose is refused with the sentence that says what a row is
- the doctor's drift test carries the case where the source has a reference and the installed folder does not
- `Messages` tells "this repository has no commit yet" apart from "git could not answer", and the second returns the error
- each fix carries the test that would have caught it, observed red before the fix
- the changelog's Unreleased section names the four fixes
- `go test ./...` is green

## Not included (first version)
- no change to the curriculum or the contract: no file under `skills/` moves, and `--check` keeps the rule it has
- no change to the scaffold, the output contract or the documents the last feature wrote
- no change to the stored proofs: never touched
- no change to the voice or the interview's questions
- no change to `odachi role` or to the skill it renders
- no change to the design canvas, its token or the push plugin
- no change to the artboard format
- no change to the release workflow
- no new verb and no new flag

## Decisions taken
- each row is paid in the file it names, without a new abstraction
- the grammar is parsed where it is owned: the skill package reads the table and the list, and its refusal says what shape it wanted
- a test-only row lands with the others, so the ledger can close in one feature
