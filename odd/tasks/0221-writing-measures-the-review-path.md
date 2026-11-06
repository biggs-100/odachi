# writing-measures-the-review-path
Base: 60c4b4e0123b2ab441a85aeb4bacaf776c0e37f3

## Requirements (verbatim, from the user)
R1 · "the writing pair gives the reviewer a path and the commands that measure it: what to review first, what is out of scope, the previous and next PR, the acceptance checklist, the changed-markdown command and the PR-size command, and a document that feels dense triggers the skill"

## Done when
- `the-doc.md` has a `## Commands` section with `git diff --name-only -- '*.md'`, and a review-path checklist naming what to review first, what is out of scope, the previous and next PR, one decision per section and the acceptance checklist
- `pull-request` names `gh pr view <PR> --json additions,deletions,changedFiles` next to the 400-line budget
- `writing`'s `## When` names a document that feels long, dense or hard to scan
- tests in `skills/embed_test.go` pin the three, observed red before the text lands
- `odachi skills --check` is green and the bodies stay under the sixty-line cap
- `odachi init` leaves the installed copy equal to the tree
- the changelog's Unreleased section names the rule
- `go test ./...` is green

## Steps
- [ ] T1 · the writing pair measures the review path   R1 · route:inline
- [ ] T2 · the changelog names the rule   R1 · route:inline

## History
L1 · "the writing pair gives the reviewer a path and the commands that measure it: what to review first, what is out of scope, the previous and next PR, the acceptance checklist, the changed-markdown command and the PR-size command, and a document that feels dense triggers the skill"
