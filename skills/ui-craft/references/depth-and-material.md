# Depth has a job, or it is a filter

A translucent layer, a shadow, a blur: each one says something is nearer, and the
moment nothing is nearer, the screen is texture. The current look -the glass of a
toolbar, the light of a floating panel- works because the depth answers a question the
layout asked, and it fails the same way every time: blur over text, contrast lost, a
page that cannot be read in the sun.

## The four jobs depth is allowed to do

1. **Hierarchy**: what floats over what. A command palette, a menu, a toast. One
   elevation level per layer, and the levels are tokens, not per-component guesses.
2. **Continuity**: a surface that stays while the content behind it moves -a toolbar
   over a long list, a header that survives the scroll.
3. **Grouping**: a boundary that a border cannot draw without noise, such as a region
   that overlaps two others by design.
4. **Focus**: a modal that removes the page behind it. It is the one place where a
   blur earns its cost, because the point is that the page is no longer available.

Anything else is a filter. A card over a gradient with a 12px blur says "nearer" about
a thing that does not move.

## Translucency, when it costs more than it gives

- The blur is only as good as its worst backdrop. Test it over the busiest image on the
  page, not over the neutral one in the mockup.
- Text on a translucent surface gets its own contrast measurement: 4.5:1 against what
  is actually behind it at the scroll position where it is read (`the-numbers.md`).
- Never blur under body text. A subtitle over glass is a readability debt paid by every
  person with a cheaper screen.
- The material is a token: one blur radius, one tint, one border-on-glass. Three
  different glass values is three different materials on one page.
- Honour reduced transparency and increased contrast: the same surface becomes opaque,
  and the information survives. It is the depth version of the reduced-motion pass.

## What it costs

- Compositing: a backdrop blur is the most expensive effect on the page and it competes
  with the signature move (`motion-identity.md`) for the same frame budget.
- Legibility: every translucent layer is a place where the contrast of the text depends
  on the content behind it, which is the content nobody designed.
- Weight: shadows that stack -a border, a shadow and a gradient for "polish"- are the
  tell in `the-tells.md`. One elevation system means one thing floats.

## The current version of the look

The 2026 glass is thin and functional: a toolbar that stays legible over content, a
sheet that keeps the page visible behind it, a pressed state that refracts for a frame.
It is thin because the effect is a hint, not a material to be admired. Copy the
restraint, not the blur radius: the products that ship it use it on one surface per
screen and keep every other surface opaque.
