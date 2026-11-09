# ui-craft-adopts-three
Base: ee0efd97b3ef19301a2719f633fad16ac2b0d1b3

## Requirements (verbatim, from the user)
R1 · "ui-craft carries the user's job and the copy of every state, and the accessibility figures cite WCAG 2.2"

## Done when
- `references/task-and-content.md` names the job table, truthful labels and the ban on invented evidence
- `references/states-and-copy.md` names the six states with what to say and what never to say
- `the-numbers.md` cites WCAG 2.2 for the target size and says the ratios alone are not conformance
- `the-read.md` checks the copy of each state, focus on an overlay and text at 200% zoom
- the body stays under 60 lines and `odachi skills --check` is green; `skills --write` regenerates SKILLS.md
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- Unreleased names the change: a new capability for the skill, patch (0251)

## Steps
- [x] T1 · the job and the copy: task-and-content.md and its gate   R1 · route:inline · tree:f3f13d7784ebe4274a49b25f8c476777c4a0042f
- [x] T2 · the six states with their copy: states-and-copy.md and its gate   R1 · route:inline · tree:f42adf81ac839484c45da5e5703001cfaf6a20b7
- [ ] T3 · WCAG 2.2 in the-numbers and the three checks the-read adds   R1 · route:inline
- [ ] T4 · the registry, the changelog bullet and the impact map   R1 · route:inline

## History
L1 · "ui-craft carries the user's job and the copy of every state, and the accessibility figures cite WCAG 2.2"
