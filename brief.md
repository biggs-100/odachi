# writing-carries-the-doc-and-the-comment

## What must exist
- `writing` teaches the long document and the collaboration comment without growing its body: the shapes and the tone live in `references/`, and the body routes them

## Done when
- `writing` declares `## Output`, a `## Decision gates` table with one row per reference, and `## References` naming both files
- `references/the-doc.md` carries the document's shape (answer first, quick path, details, checklist, next step) and how a reviewer reads it
- `references/the-comment.md` carries the comment's tone and formula, including the language of the thread and no em dashes
- the body stays within sixty lines and every rule it has today survives
- `odachi skills --check` is green, and `odachi init` leaves the installed copy equal to the tree
- the changelog's Unreleased section names the two texts
- `go test ./...` is green

## Not included (first version)
- no change to the other eighteen skills
- no change to the grammar: references stay `.md` under `references/`, and `--check` keeps the rule it has
- no new skill, verb or flag
- no change to the voice or the interview's questions: an artifact stays English, and a comment follows the thread's language -the reference says so
- no change to the stored proofs: never touched
- no change to `odachi role` or to the skill it renders
- no change to the design canvas, its token or the push plugin
- no change to the artboard format
- no change to the release workflow
- no adoption of gentle-ai's frontmatter, its section names or its license text

## Decisions taken
- the depth goes to two references, not the body: a document and a comment are two shapes, and two shapes justify a router
- the rules that live in the body today stay there; the references add the shape and the examples
- a comment follows the language of the thread; an artifact stays English
