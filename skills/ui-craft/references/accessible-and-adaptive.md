# Accessible, and still the same interface when it narrows

Two questions in one page, because a screen that survives a phone and a keyboard is
one screen or it is none. The states and their sentences are `states-and-copy.md`'s;
the measurements in `the-numbers.md`; this page is what the markup has to be.

## Say what it is, not what it looks like

- A link navigates; a button acts. A `div` with a click handler is neither, and it
  cannot be reached by a keyboard or announced by a reader.
- Every input has a label that stays visible, not a placeholder that disappears with
  the first keystroke. A placeholder is a hint, never the name of the field.
- An icon-only control carries an accessible name. The painted size of an icon says
  nothing about the area a finger needs.
- A heading level describes the structure of the document, not the size of the text.
  Choosing `h3` for its look and reordering with CSS leaves a reader with a page that
  has no outline.

## The keyboard is a first-class path

- Every essential action is reachable without a pointer, and nothing essential
  appears only on hover: a hover-revealed control does not exist on a phone or for a
  keyboard.
- Focus is visible, and it is not the same thing as selection: one says where the
  keyboard is, the other says what was chosen.
- Selected, disabled, hover and focus are four distinct states: focus is not
  selection, hover is neither, and a disabled control says why it is unavailable.
- An overlay moves focus to a control inside it, and returns it to the trigger when
  it closes. Escape closes it, and never discards work without warning.
- When the layout rearranges, the order of the DOM follows reading and keyboard
  order. Reordering with `flex-direction` or `order` leaves the tab sequence behind.

## It narrows without losing the decision

- What fits a rail on a desktop is a navigation problem on a phone: at the narrow
  breakpoint the rail becomes a labelled control, the current location stays visible
  and the primary action stays reachable.
- A wide comparison table keeps the fields that identify a row and the comparison
  that matters, or scrolls horizontally with that signalled. Hiding the column that
  decides is not adapting.
- Long names, translated labels and text at 200% zoom wrap or reflow. A fixed-height
  button that truncates its label is a bug with a rounded corner: let it grow.
- A compact control keeps separation and an accessible name. `the-numbers.md` carries
  the 24×24 floor and the 44×44 goal; this page says why the area, not the glyph, is
  what gets measured.
- Motion is not the only signal of a change: with `prefers-reduced-motion` the same
  information survives without travel. `motion.md` owns the two-property rule.

## Announce the change once

A status change reaches assistive technology in context, once. Repeating an unchanged
message is noise that gets muted, and a muted region is worse than no region: the
sentence for each state is `states-and-copy.md`'s, and this page only says that it
must arrive.
