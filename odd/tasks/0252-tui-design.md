# tui-design
Base: 4f721921b915486784c6370ce26f01a1bbd5506f

## Requirements (verbatim, from the user)
R1 · "a skill that designs, critiques and refines a terminal UI from a captured frame, with the floor and the eyes"

## Done when
- `skills/tui-design/SKILL.md` loads on a terminal UI request, routes with gates and stays under 60 lines
- five references: terminal-craft, critique, commands, visual-loop and design-doc
- `NOTICE.md` names the source: gentle-shell's tui-design, Apache-2.0, and impeccable behind it
- `skills/embed_test.go` names it in `bundled`, `SKILLS.md` regenerates, the impact map counts it, doctor says 21
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- Unreleased names the change: a new capability, patch (0252)

## Steps
- [x] T1 · the body and the floor: terminal-craft.md   R1 · route:inline · tree:66a82c8a643232910e2b4e2f594c66aea18add6c
- [ ] T2 · the eyes: visual-loop.md and critique.md   R1 · route:inline
- [ ] T3 · the plays: commands.md and design-doc.md   R1 · route:inline
- [ ] T4 · NOTICE, the bundled list, the registry, the map and Unreleased   R1 · route:inline

## History
L1 · "a skill that designs, critiques and refines a terminal UI from a captured frame, with the floor and the eyes"
