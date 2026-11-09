# The design context of a terminal product

The file is `design/system.md`, the one this project already keeps. A terminal project adds a
`## Terminal` section to it instead of inventing a second file: two design documents contradict each
other, and nobody notices until the interface disagrees with both.

Describe first, then give the exact value in parentheses, and say where it is used. One to three named
rules per section, written as a sentence somebody can break.

## The tokens

```markdown
## Terminal

**The floor is the cell.** Everything below is measured in cells, never in pixels or points.

colors:      # roles mapped to the theme's keys, not to hex literals
  canvas: <key>
  rail: <key>
  text: primary <key>, secondary <key>, muted <key>
  accent: <key>
  state: success <key>, warning <key>, error <key>, info <key>
  border: <key>
roles:       # the type roles, as attribute combinations
  title: bold + accent
  value: normal
  label: dim
  meta: dim italic
glyphs:      # one family, by meaning
  done: ✓   running: ●   pending: ○   failed: ✗
  expanded: ▾   collapsed: ▸
spacing:     # in cells and rows
  inset: 1 column
  group-gap: 1 row
  section-gap: 1 row
widths:
  rail: 50   breakpoint: 140   minimum: 80
```

## The sections, in this order

1. **Overview** — the product, its users, and a one-line metaphor ("an IDE inside a terminal").
2. **Colours** — the roles and where each one appears.
3. **Type roles** — the hierarchy, and where each role applies.
4. **Layout** — panes, rail, status bar, the widths and the breakpoint.
5. **Surfaces** — canvas, rail, elevated; how two regions separate.
6. **Glyphs** — the family and the meaning table.
7. **Components** — each panel, bar and card, with its states.
8. **Rules and refusals** — the named rules that this product keeps, and what it will not do.

## When there is no file

Propose the section, with the values the code already uses, and ask before writing it. A design document
written from invention becomes a second set of rules that the interface ignores.
