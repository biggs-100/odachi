# design-lets-the-human-tune
Base: 1f1709fad33c6cef3f67f31f1d630f87e475517d

## Requirements (verbatim, from the user)
R1 · "when the human wants to tune values, the artboard carries the controls: sliders bound to its CSS custom properties for spacing, colour and type scale, and a `Copy as prompt` button that copies the chosen values as text, so the numbers go back to the agent without a screenshot"

## Done when
- `design` says the tuning controls live in the artboard, not the canvas: sliders bound to CSS custom properties, and a `Copy as prompt` button that copies the chosen values as a prompt
- the rule keeps the artboard self-contained: the button copies text and never calls a network
- a test in `skills/embed_test.go` pins the rule, observed red before the text lands
- `odachi skills --check` is green and the design body stays under the sixty-line cap
- `odachi init` leaves the installed copy equal to the tree
- the changelog's Unreleased section names the rule
- `go test ./...` is green

## Steps
- [x] T1 · design puts the tuning controls in the artboard   R1 · route:inline · tree:862bf5b78fb71b4d5a218f45fad2fc865d433d90
- [x] T2 · the changelog names the rule   R1 · route:inline · tree:c0d2275cfd9c249b4d1e9671b7d1f0917c2a2b10

## History
L1 · "when the human wants to tune values, the artboard carries the controls: sliders bound to its CSS custom properties for spacing, colour and type scale, and a `Copy as prompt` button that copies the chosen values as text, so the numbers go back to the agent without a screenshot"
