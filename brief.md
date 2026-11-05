# the-design-skill

## What must exist
- the design skill keeps the system fresh, links the design to its ODD feature both ways, and pairs with ui-craft

## Done when
- odachi design list reports the design sources and, when design/system.md exists, the sources newer than it; the skill syncs before the first design and offers a sync when the system is stale
- odachi design handoff appends one Log entry to the linked odd/tasks feature document naming the design, the chosen artboard and the alternatives; a feature path outside odd/tasks is refused and a design with no feature writes no document
- the design skill appends the creation Log entry for the design and its artboards, and loads ui-craft before choosing directions
- go test ./... is green; the stale report and the handoff Log are pinned

## Not included (first version)
- no change to the canvas, its token or the push plugin
- no change to the artboard format
- no automatic sync without the agent running it: list reports, the skill acts
- no two-way link for a design with no feature
- no change to the gate, the review or the proof

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: this week.
- Wants to learn: how a skill detects stale state without a daemon.
- Wants to learn: how a link stays bidirectional without a second source of truth.
