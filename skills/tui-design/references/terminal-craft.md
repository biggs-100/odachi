# The floor

Read before any terminal UI edit. Every line is checkable in a captured frame.

## The medium

- The cell is the unit. Height is rows, width is columns; there is no half cell and no font size.
- **Type roles come from attributes**, not size: bold, dim, italic, case and colour. Keep three or four
  roles, and apply each one identically wherever it repeats.
- **Glyphs are the icon system.** Pick one family —geometric: `● ○ ◐ ✓ ✗ ▸ ▾`— and never mix it with
  emoji or decorative sets for the same meaning.
- **Emoji and CJK are two cells wide** and break every alignment and width budget. Keep them out of
  chrome; when a value can contain one, measure with the string's visible width, not its byte length.

## Hierarchy

- One primary element per panel, two or three secondary, everything else muted.
- Labels dim, values normal, the one that matters bright or accented.
- Align key and value in a column (`label  value`). A label stacked above its value costs a row and
  reads slower.
- Group by proximity first, add a separator when proximity fails, add a box only when the region is
  interactive or scrolls on its own.
- **The squint test**: blur the frame. The order primary > secondary > groups must still read.

## Space and density

- Density matches how often a thing is used: always-visible chrome is one row per fact; detail used
  twice a week waits behind a key or expands on demand.
- Rhythm comes from contrast —tight inside a group, one gap between groups— never from the same gap
  everywhere.
- A panel's chrome is a header row plus content. Padding above, below and between earns its place or
  goes.
- Truncate chrome with `…`; wrap prose at 60-80 columns. A fixed-width rail does not wrap.
- Finished items collapse to one summary row: a completed task does not need its three sub-rows.

## Colour roles

- Build roles, not swatches: canvas, elevated surface, rail surface (a second neutral, slightly tinted),
  text primary / secondary / muted, accent, success / warning / error / info, border.
- The accent belongs to action, selection and live state, and there is one accent hue.
- Never grey text on a coloured band; use a darker shade of that band's own hue.
- Never encode state by colour alone: pair it with a glyph or a word.
- A dark theme is designed, not an inversion. Check dim text on the tinted surfaces, which is where it
  always fails.

## Operate mode: the vocabulary an IDE taught

- **Status bar**: one row, fixed position, segments divided by a thin rule, the most volatile data on
  the right where the eye returns.
- **Side panel**: sections with a one-row header (`▾ NAME  count`), collapsible, dense lists below.
- **Panes** are divided by one thin rule or a surface tint, not by framed boxes with gaps.
- **Key hints** live in the status bar or appear on focus; not repeated in every header.
- **The command palette is the navigation model.** Do not invent a new affordance beside it.

## Refuse

- A box inside a box, or a card per tiny fact.
- A coloured bar (`▌`, `┃`) on every item as decoration.
- Blank rows as the hierarchy tool, or one gap value repeated down the whole screen.
- Text flush against a coloured band with no inset cell.
- The same fact twice on one screen: the header and the panel.
- A label repeated three times in one panel, or the word where a glyph would do.
- Letter-spaced or all-caps body text.
- Decorative animation; a spinner only while work is genuinely running.
- Retro costume: fake function-key bars, ASCII banners, neon chosen for the vibe.
- A colour that is not a theme token.
- Marketing words, aphorisms and em-dash chains in status text.

## Verify

- Contrast of dim and muted text on every surface it appears on.
- Every interactive row in all six states.
- Long values (paths, branch names), an empty list, 0 items, 1000 items, the narrow width.
- No line exceeds its column budget; an overflowing rail can be dropped by the layout engine entirely.
- Render cost: an unchanged frame re-renders nothing, and a style or width change invalidates the cache.
