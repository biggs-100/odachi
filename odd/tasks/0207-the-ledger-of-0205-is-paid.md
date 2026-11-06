# the-ledger-of-0205-is-paid
Base: 132f51788ebc58f20b8ff149faea1006854eb37d

## Requirements (verbatim, from the user)
R1 · "the four rows `docs/review-debt.md` left open are paid: the pack report counts a backup only when one ran, the gates check reads the table instead of searching it, the missing-reference drift has its test, and a git that cannot answer `Head` is an error instead of an empty history"

## Done when
- a skill whose `SKILL.md` is unchanged and whose only difference is a new reference is not reported as backed up, in a real run or a dry run
- `## Decision gates` is read as a table and `## References` as a list, so a path named only in prose is refused with the sentence that says what a row is
- the doctor's drift test carries the case where the source has a reference and the installed folder does not
- `Messages` tells "this repository has no commit yet" apart from "git could not answer", and the second returns the error
- each fix carries the test that would have caught it, observed red before the fix
- the changelog's Unreleased section names the four fixes
- `go test ./...` is green

## Steps
- [x] T1 · pack counts a backup only after one ran, and a new reference is not one   R1 · route:inline · tree:d5d6d1269b3d38774d7fe182bdc6198eb337ce63
- [x] T2 · the gates table and the references list are parsed, so prose is refused   R1 · route:inline · tree:68709ef0c1e5f6c25162786909378b67e82d39ad
- [ ] T3 · the missing half of the reference drift has its test   R1 · route:inline
- [ ] T4 · a git that cannot answer Head is an error, not an empty history   R1 · route:inline
- [ ] T5 · the changelog and the ledger record the four rows as paid   R1 · route:inline

## History
L1 · "the four rows `docs/review-debt.md` left open are paid: the pack report counts a backup only when one ran, the gates check reads the table instead of searching it, the missing-reference drift has its test, and a git that cannot answer `Head` is an error instead of an empty history"
