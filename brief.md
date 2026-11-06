# writing-measures-the-review-path

## What must exist
- the writing pair gives the reviewer a path and the commands that measure it: what to review first, what is out of scope, the previous and next PR, the acceptance checklist, the changed-markdown command and the PR-size command, and a document that feels dense triggers the skill

## Done when
- `the-doc.md` has a `## Commands` section with `git diff --name-only -- '*.md'`, and a review-path checklist naming what to review first, what is out of scope, the previous and next PR, one decision per section and the acceptance checklist
- `pull-request` names `gh pr view <PR> --json additions,deletions,changedFiles` next to the 400-line budget
- `writing`'s `## When` names a document that feels long, dense or hard to scan
- tests in `skills/embed_test.go` pin the three, observed red before the text lands
- `odachi skills --check` is green and the bodies stay under the sixty-line cap
- `odachi init` leaves the installed copy equal to the tree
- the changelog's Unreleased section names the rule
- `go test ./...` is green

## Not included (first version)
- no change to `the-comment.md`: its command belongs to `pull-request`, and the comment sends the reader there
- no change to `assets/doc.md`: the shape already carries the checklist
- no change to the other skills
- no new verb and no new flag
- no change to the voice or the interview's questions
- no change to the stored proofs: never touched
- no change to `odachi role` or to the skill it renders
- no change to the design canvas, its token or the push plugin
- no change to the artboard format
- no change to the release workflow
- no change to the skills the feature does not touch: tests, drive, slices, commits, triage, design and ui-craft

## Decisions taken
- the PR-size command lives in `pull-request`, which owns the 400-line budget; the doc command lives in `the-doc`, which owns the document
- the review path is a checklist in `the-doc`, not a new section in `writing`: the body stays the router
- the symptom trigger is one clause in `## When`, because a document that feels dense is the case the skill exists for
- the comment's command is not duplicated: `the-comment` sends the reader to `pull-request`
