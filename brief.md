# publishing-is-read-back

## What must exist
- a published PR or issue is read back and compared with what was sent: the title, the body, the base and the head are the ones that were posted, because a text that was sent is not a text that arrived

## Done when
- `pull-request` says a create command is not done at the URL: `gh pr view <url> --json title,body,baseRefName,headRefName` is compared with the file that was posted, and an empty body, a wrong base or a title that lost its point is fixed before the review is requested
- `writing` says a published text is read back and compared with the file that was written, not with memory, and names `pull-request` as the owner of the command
- a test in `skills/embed_test.go` pins the readback in both skills, observed red before the text lands
- both bodies stay within the sixty-line cap and `odachi skills --check` is green
- `odachi init` leaves the installed copy equal to the tree
- the changelog's Unreleased section names the rule
- `go test ./...` is green

## Not included (first version)
- no new verb and no new flag: the readback is a command and a pair of eyes, not a gate
- no change to `odachi check` or to `odachi review` behavior
- no change to the skill contract or its format: only the text of two skills and one pinned test
- no change to the six risks in `commits` or to their table
- no change to the voice or the interview's questions
- no change to the stored proofs: never touched
- no change to `odachi role` or to the skill it renders
- no change to the design canvas, its token or the push plugin
- no change to the artboard format
- no change to the release workflow
- no `assets/` in the skill contract: another feature

## Decisions taken
- the comparison is against the file that was posted, not against the command's output: `gh pr create` prints a URL, and a URL says nothing about the body
- one owner per part: `pull-request` owns the command and the PR fields, `writing` owns the comparison of any published text and sends the reader to that command
- a mismatch is fixed before the review is requested, because a reviewer reading an empty body is reading the wrong thing
- it does not become a gate: `check` cannot read GitHub, and a network call would make the tree depend on a remote
