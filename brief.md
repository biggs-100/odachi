# design-lets-the-human-tune

## What must exist
- when the human wants to tune values, the artboard carries the controls: sliders bound to its CSS custom properties for spacing, colour and type scale, and a `Copy as prompt` button that copies the chosen values as text, so the numbers go back to the agent without a screenshot

## Done when
- `design` says the tuning controls live in the artboard, not the canvas: sliders bound to CSS custom properties, and a `Copy as prompt` button that copies the chosen values as a prompt
- the rule keeps the artboard self-contained: the button copies text and never calls a network
- a test in `skills/embed_test.go` pins the rule, observed red before the text lands
- `odachi skills --check` is green and the design body stays under the sixty-line cap
- `odachi init` leaves the installed copy equal to the tree
- the changelog's Unreleased section names the rule
- `go test ./...` is green

## Not included (first version)
- no change to the canvas: a focused board already hides the shield and lets the artboard receive the pointer, and the button copies text in the page
- no change to the design folder, the manifest, `comments.json` or the handoff
- no change to the design verbs or the ODD link
- no change to the other references of `ui-craft` or `design`
- no new verb and no new flag
- no change to the voice or the interview's questions
- no change to the stored proofs: never touched
- no change to `odachi role` or to the skill it renders
- no change to the artboard format
- no change to the release workflow
- no change to the skills the feature does not touch: tests, drive, writing, pull-request, commits and triage

## Decisions taken
- the controls live in the artboard, not the canvas: the artboard is the self-contained artifact, and gentle-shell does the same
- the button copies a prompt, not a JSON blob: the agent reads the values it will apply
- sliders bind to CSS custom properties so the same values are read back and pasted
- the canvas stays untouched: focusing a board hides the shield and the artboard receives the pointer
- the comparison said this was a canvas feature; exploration showed the canvas already allows it and the rule is the whole gap
