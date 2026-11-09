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
