# writing-carries-the-doc-and-the-comment
Base: 7d9645d86371f5869fa8aaa91685b35b3479e3bf

## Requirements (verbatim, from the user)
R1 · "`writing` teaches the long document and the collaboration comment without growing its body: the shapes and the tone live in `references/`, and the body routes them"

## Done when
- `writing` declares `## Output`, a `## Decision gates` table with one row per reference, and `## References` naming both files
- `references/the-doc.md` carries the document's shape (answer first, quick path, details, checklist, next step) and how a reviewer reads it
- `references/the-comment.md` carries the comment's tone and formula, including the language of the thread and no em dashes
- the body stays within sixty lines and every rule it has today survives
- `odachi skills --check` is green, and `odachi init` leaves the installed copy equal to the tree
- the changelog's Unreleased section names the two texts
- `go test ./...` is green

## Steps
- [ ] T1 · writing routes the long document and the comment, with the depth beside it   R1 · route:inline
- [ ] T2 · the changelog names the two texts   R1 · route:inline

## History
L1 · "`writing` teaches the long document and the collaboration comment without growing its body: the shapes and the tone live in `references/`, and the body routes them"
