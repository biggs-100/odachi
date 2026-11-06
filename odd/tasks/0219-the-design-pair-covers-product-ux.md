# the-design-pair-covers-product-ux
Base: ab82e8b66afb51a1221cf49cdf77d3fe71f79a1d

## Requirements (verbatim, from the user)
R1 · "the design pair closes the UX gaps the comparison found: `ui-craft` carries browse-and-compare (search, filters, pagination, navigation context), and `design` passes every artboard through the UX read, adds an artboard for another approach on the same brief, and looks at a comment's screenshot first"

## Done when
- `ui-craft` ships `references/browse-and-compare.md`, routes it in `## Decision gates` and `## References`, and the body stays under the sixty-line cap
- `design` says every finished artboard passes `ui-craft`'s `the-read` (and `adaptive-interfaces` when it has states), that another approach on the same brief is a new artboard with the next `NN`, and that a comment's `screenshot` is looked at first
- tests in `skills/embed_test.go` pin both skills, observed red before the text lands
- `odachi skills --check` is green
- `odachi init` leaves the installed copy equal to the tree
- the changelog's Unreleased section names the rule
- `go test ./...` is green

## Steps
- [x] T1 · ui-craft ships browse-and-compare and routes it   R1 · route:inline · tree:3937561f48c449b6778c42dae9057bd7a242662c
- [x] T2 · design checks every artboard, adds a direction, reads the comment screenshot   R1 · route:inline · tree:63a0341608a3e6f4d15d1896778fe3badbd16b2a
- [x] T3 · the changelog names the rules   R1 · route:inline · tree:f542816010d5c62683d5d328498b76d59c0bc076

## History
L1 · "the design pair closes the UX gaps the comparison found: `ui-craft` carries browse-and-compare (search, filters, pagination, navigation context), and `design` passes every artboard through the UX read, adds an artboard for another approach on the same brief, and looks at a comment's screenshot first"
