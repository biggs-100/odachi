# The numbers

| what | value |
|---|---|
| scale | 1.2: 12, 14, 16, 20, 24, 32, 40, 56 |
| body | 16-18px, leading 1.5, measure 45-90 characters (65 is ideal) |
| display | leading 1.1-1.2, tracking -0.02em at 32px and up |
| space | 4/8: 4, 8, 12, 16, 24, 32, 48, 64, 96 |
| radius | one per level; a nested one is the outer minus the padding |
| contrast | 4.5:1 for text, 3:1 for large text and UI edges (WCAG 2.2 AA) |
| pointer | 24x24 minimum (WCAG 2.2 target size); 44x44 when space allows |
| zoom | text reflows at 200% without clipping; a button grows, never truncates |
| layout | 12 columns, 24 gutter, content 1200-1280, text 60-75ch |
| colour space | OKLCH for the ramp, one accent tinted through the neutrals; a P3 accent with an sRGB fallback |
| fluid scale | `clamp()` on the display sizes, 1.2 ratio; body stays 16-18px at every width |

## How to use them

They are defaults, not a brand: start from the project's tokens when it has
them, and keep these relationships when it does not. The scale is a ratio, not a
list; the spacing is a rhythm of two steps, not a menu of every value in between.
Measure contrast with a tool, never by eye: a colour that looks fine on a laptop
fails on a phone in the sun. Breakpoints come from where the content breaks, not
from the names of devices.

A ratio is not conformance: WCAG 2.2 asks for a 24x24 CSS px target size with its
spacing exceptions, wants status not carried by colour alone, and wants text that
survives 200% zoom. These numbers are the floor a screen starts from, not a
certificate it earns.

## The colour space, and the one that muddies

Build the ramp in OKLCH: lightness is a number a person can reason about, and the
neutrals tint evenly instead of turning grey-green halfway down. A wide-gamut accent
(`color(display-ps3 ...)` or the `oklch()` value inside an `@supports`) is what makes
an accent look current on a good screen; it needs a declared sRGB fallback, because the
person with the ordinary monitor is not an edge case. Measure the contrast after the
fallback, not before: the fallback is what most people see.

A fluid display size is a `clamp()` on one or two sizes, never on the body text: text
that shrinks with the viewport is text somebody cannot read on the phone they are
holding.

