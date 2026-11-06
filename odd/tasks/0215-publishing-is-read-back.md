# publishing-is-read-back
Base: 96ca86f8dc4e0ffb1fdcd3905a95695dfaeac3c5

## Requirements (verbatim, from the user)
R1 · "a published PR or issue is read back and compared with what was sent: the title, the body, the base and the head are the ones that were posted, because a text that was sent is not a text that arrived"

## Done when
- `pull-request` says a create command is not done at the URL: `gh pr view <url> --json title,body,baseRefName,headRefName` is compared with the file that was posted, and an empty body, a wrong base or a title that lost its point is fixed before the review is requested
- `writing` says a published text is read back and compared with the file that was written, not with memory, and names `pull-request` as the owner of the command
- a test in `skills/embed_test.go` pins the readback in both skills, observed red before the text lands
- both bodies stay within the sixty-line cap and `odachi skills --check` is green
- `odachi init` leaves the installed copy equal to the tree
- the changelog's Unreleased section names the rule
- `go test ./...` is green

## Steps
- [x] T1 · pull-request and writing read the published text back   R1 · route:inline · tree:61c5b3ab41f9052620741cbadf0ebb71b00aaaf2
- [x] T2 · the changelog names the rule   R1 · route:inline · tree:fef9a0ee3b8ad2fb91eb9e4bd23c1b803af82d39

## History
L1 · "a published PR or issue is read back and compared with what was sent: the title, the body, the base and the head are the ones that were posted, because a text that was sent is not a text that arrived"
