# the-design-pair-covers-product-ux

## What must exist
- the design pair closes the UX gaps the comparison found: `ui-craft` carries browse-and-compare (search, filters, pagination, navigation context), and `design` passes every artboard through the UX read, adds an artboard for another approach on the same brief, and looks at a comment's screenshot first

## Done when
- `ui-craft` ships `references/browse-and-compare.md`, routes it in `## Decision gates` and `## References`, and the body stays under the sixty-line cap
- `design` says every finished artboard passes `ui-craft`'s `the-read` (and `adaptive-interfaces` when it has states), that another approach on the same brief is a new artboard with the next `NN`, and that a comment's `screenshot` is looked at first
- tests in `skills/embed_test.go` pin both skills, observed red before the text lands
- `odachi skills --check` is green
- `odachi init` leaves the installed copy equal to the tree
- the changelog's Unreleased section names the rule
- `go test ./...` is green

## Not included (first version)
- no change to the canvas, its sliders or a copy-as-prompt control: that is the next feature
- no change to the design folder, the manifest or `comments.json`
- no change to the design verbs, the handoff or the ODD link
- no change to the other references of `ui-craft` or `design`
- no new verb and no new flag
- no change to the voice or the interview's questions
- no change to the stored proofs: never touched
- no change to `odachi role` or to the skill it renders
- no change to the design canvas, its token or the push plugin
- no change to the artboard format
- no change to the release workflow
- no change to the skills the feature does not touch: tests, drive, writing, pull-request, commits and triage

## Decisions taken
- the browse reference lives in `ui-craft`, not `design`, because design owns delivery and ui-craft owns the UX quality -the same split gentle-shell uses
- the UX pass is a rule in `design`, not a new verb: the artboard is checked before it is reported, and the Output names the read it survived
- another approach on the same brief is another artboard with the next `NN`; a new slug is a new brief, never a new direction
- a comment's screenshot is read before the element: it shows the transient state the static artboard cannot
- the canvas tuning is a separate feature: it is code, not curriculum, and it needs the fence to let the canvas change
