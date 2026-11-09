---
name: tui-design
description: Trigger: a terminal interface, a panel, a rail, a status bar, a prompt; a TUI that looks busy, flat or generic; a critique before shipping one.
---

## When

The human asks to design, review or refine a terminal interface —panes, rails, status bars, cards,
prompts— or says one looks busy, flat or generic. Not for a web page (`ui-craft` owns that) and not for
the plain output of a command.

## Rules

- **Never judge a TUI from the code**: capture the frame, look at it, change, capture again.
- **The grid is the cell**: space is columns and rows; a blank row is the most expensive unit there is.
- **Hierarchy from role, weight and alignment**, never from blank rows, boxes or colour alone.
- **One primary per panel**, two or three secondary, the rest muted; each fact once per screen.
- **Glyphs are the icon system**: one family by meaning, used everywhere; no emoji in chrome.
- **Colour is roles, not swatches**: one accent for action, selection and live state; state has a glyph.
- **Density follows frequency**: chrome is one row per fact; rare detail waits behind a key.
- **IDE vocabulary, not a costume**: panes, tabs, status bar; no banners, no fake key bars, no neon.
- **Every control has every state**: default, hover, focus, selected, disabled, loading.
- **Every width is designed**: 180x50, the breakpoint, 80x24; what leaves moves to the status bar.
- **Render cheap**: no I/O in a render path; a cache keyed by style, width and theme.
- **Three rounds, then report**: fix, re-capture, compare; what is left is named, not hidden.

## Decision gates

| the situation | load |
|---|---|
| the floor before any edit: the medium, the refuses, what to verify | `references/terminal-craft.md` |
| how to capture the frame, and the loop that compares two of them | `references/visual-loop.md` |
| a review: the score, the severities and the report | `references/critique.md` |
| one named change: distill, quieter, bolder, harden, adapt, polish | `references/commands.md` |
| the project's terminal context, or writing it | `references/design-doc.md` |

## Output

The frames the review read, each finding with its severity and its fix, what was fixed, what is left
open, and the checks that ran.

## References

- `references/terminal-craft.md` — the floor: medium, hierarchy, space, roles, the refuses, the verify.
- `references/visual-loop.md` — the frame at each width, the capture, the three-round loop.
- `references/critique.md` — ten heuristics scored, the severities, the report shape.
- `references/commands.md` — the playbooks, one named target each.
- `references/design-doc.md` — `design/system.md` with its terminal section.

## Never

- Ship a TUI you did not look at, or refine a look that is being replaced.
- Add a box, a divider or a blank row to create hierarchy.
- Encode state with colour alone.

## Attribution

The method comes from `gentle-shell`'s `tui-design` (Apache-2.0); `NOTICE.md` names it. The text is ours.
