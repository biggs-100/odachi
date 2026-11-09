# tui-design

## What must exist
- a skill that designs, critiques and refines a terminal UI from a captured frame, with the floor and the eyes

## Done when
- `skills/tui-design/SKILL.md` loads on a terminal UI request, routes with gates and stays under 60 lines
- five references: terminal-craft, critique, commands, visual-loop and design-doc
- `NOTICE.md` names the source: gentle-shell's tui-design, Apache-2.0, and impeccable behind it
- `skills/embed_test.go` names it in `bundled`, `SKILLS.md` regenerates, the impact map counts it, doctor says 21
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- Unreleased names the change: a new capability, patch (0252)

## Not included (first version)
- no harness of its own: the loop uses the project's test or snapshot flag - back when a project cannot capture
- no change to `ui-craft` or `design`: a terminal is not a screen - back when one rule must hold for both
- no runtime, installer, workflow, canvas or proof change - back when one of those paths changes
- no verbatim port of their text: the method is adopted, the words are ours - back when both must stay in step

## Decisions taken
- The eyes are a text frame at 80x24, 140 and 180x50; a PNG is optional and named when it exists.
- The design context lives in `design/system.md` with a terminal section, not in a second file.
- The method is Apache-2.0 and this project is MIT: original text and a NOTICE, the way they credited impeccable.
- The floor is implementation-agnostic; the loop is concrete for Go and gives a way out for other stacks.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: whether a scored critique read from a text frame finds the same defects a screenshot does.
