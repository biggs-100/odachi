# the-design-skill
Base: a379edff32c1fb0997daeba6f6e1ac3a84518b5e

## Requirements (verbatim, from the user)
R1 · "the design skill keeps the system fresh, links the design to its ODD feature both ways, and pairs with ui-craft"

## Done when
- odachi design list reports the design sources and, when design/system.md exists, the sources newer than it; the skill syncs before the first design and offers a sync when the system is stale
- odachi design handoff appends one Log entry to the linked odd/tasks feature document naming the design, the chosen artboard and the alternatives; a feature path outside odd/tasks is refused and a design with no feature writes no document
- the design skill appends the creation Log entry for the design and its artboards, and loads ui-craft before choosing directions
- go test ./... is green; the stale report and the handoff Log are pinned

## Steps
- [ ] T1 · design list reports the sources and the ones newer than design/system.md   R1 · route:inline
- [ ] T2 · design handoff appends one Log entry to the linked feature document and refuses a feature outside odd/tasks   R1 · route:inline
- [ ] T3 · the design skill syncs first, offers a stale sync, appends the creation log and loads ui-craft   R1 · route:inline

## History
L1 · "the design skill keeps the system fresh, links the design to its ODD feature both ways, and pairs with ui-craft"
