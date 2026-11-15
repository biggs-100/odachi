# The board and the hand

## What this look is for

A structure the viewer has to build in their head: boxes, arrows, states, a flow. The metaphor is that
somebody explained it to you at a board, and each frame is a moment of that explanation. It wins where the
argument is spatial, and loses where the product is the star or the message is a feeling.

The cost is planning: every station has to exist on one canvas before the camera moves.

## The ground

| surface | value | why |
| --- | --- | --- |
| paper | `#FAF9F5` | warm cream; white reads as a screen, cream as a sheet |
| dot grid | `#DCD9D0`, 28 px apart, 1.2 px | texture at a glance, invisible at reading distance |
| ruled sheet | `#E5E7EB` lines, `#FCA5A5` margin | opening and closing only: a ruled page inside a diagram reads as a school notebook |
| ink | `#1A1A1A`, 2-3 px | one ink, two widths at most; a diagram in five weights is a poster |
| card | `#FFFFFF`, shadow `0 10px 25px rgba(0,0,0,.06)` | a card is paper on paper: the shadow is the only depth there is |
| tape | `#E5DEC9` at 0.85 | the mark of something placed by a person, not laid out by a grid |

Type: a mono for commands, paths and labels (`JetBrains Mono`, `Fira Code`), a geometric bold for the
claim (`Inter`, `Plus Jakarta Sans`). Mono says "this is a command", and a board is where commands are
written. The sizes and the reading speed are `type-in-motion.md`'s question, not this one's.

## One accent per meaning

| colour | means | appears as |
| --- | --- | --- |
| `#1D4ED8` | the way the argument goes | underline, oval, arrow, check |
| `#DC2626` | a way that does not work | cross, tally mark, the wrong line |
| `#9CA3AF` | not now | the item the voice has not reached |

Gray is "not now", never "disabled": a board has no disabled buttons, and a viewer reads a greyed command
as broken. A word never carries two accents, and a scene never invents a third meaning. That grammar is
what makes the piece read as authored — the viewer learns it in the first 5 s and reads the rest faster.

## The imperfection is seeded

- `rough.js` with `roughness: 1.1-1.3`, `bowing: 1-1.5`, and a **seed per element derived from its name**,
  written in the source. Same box, same wobble, in every scene and every render.
- Generate the paths **once** — at module level or in a build step — never inside the frame function. Per
  frame generation is CPU, and a floating seed is a vibrato.
- Normalize the stroke with `pathLength="1"`, `strokeDasharray="1"` and `strokeDashoffset = 1 - p`: no
  measurement, so nothing depends on layout at render time. `getTotalLength()` is a DOM measure — once,
  never per frame.
- Imperfection is not sloppiness: a right angle stays within 1-2 px of square, a circle closes, and the
  same arrow has the same curve in every scene.
- Tape and rotation are for what was placed **later**: one or two per scene. A grid of taped cards is the
  machine-made tell in a paper costume (`the-ai-look.md`, "a uniform grid of equal cards").

## Draw-on is a verb, and it has weight

- A shape enters by being drawn, not by fading in: the eye follows the tip and the shape arrives as a
  thought.
- Duration follows distance and weight: a 40 px underline 0.15-0.25 s, a 400 px arrow 0.35-0.5 s, a box
  that fills the frame 0.6-0.9 s. One duration for every stroke is the "same duration on everything" tell.
- Reading order is drawing order: left to right, top to bottom, the container before its contents, the
  cause before its arrow.
- A stroke stops where a hand stops: at a corner, at the end of a word, at the arrowhead. It does not fade
  out mid-line.
- The tip leads. Animate the dash, never a wipe mask: a wipe says "a layer was revealed", a dash says
  "somebody drew it".

## The board and the camera

- **One canvas, not slides.** Place the beats as stations on one plane before anything moves; a pan that
  discovers empty paper is a plan that was never made.
- **The camera is the transition inside one argument.** It says "same board, one step later"
  (`motion-principles.md` has the row). A cut inside a flow breaks the metaphor; a cut between two
  arguments is free.
- **One move at a time**: the camera or a stroke, not both. Lock the camera while a line draws, and move
  it while nothing draws.
- **The payoff shot is a zoom-out**: the whole structure connected, everything the piece built, in one
  frame. It is the reason to plan the map.
- **The camera has a floor and a ceiling.** Below 0.85 the ink thins against the grain and the dots
  moiré; above 1.6 the frame reads as a close-up and the negative space the look depends on disappears.
  0.85-1.6 is the working range.
- **Paper may be empty.** The ground carries the frame, so a hold is not a void: this look can hold longer
  than a dark UI. 40-50% negative space is the look, not a waste.

## Correction is the drama

A board explains by changing its mind, and that is the one thing a slide cannot do:

- a word crossed out and rewritten beside it;
- tally marks counting a cost, one per beat;
- an arrow crossed in red, with the working arrow drawn in blue underneath;
- a box erased and redrawn smaller.

Build at least one correction into a piece that compares before and after. Marks accumulate, and erasing
one is itself a beat, not a cleanup.

## What the look costs, and the traps it sets

- **A generator or a fixed set of paths.** Under about ten shapes, authored paths with a hand-made wobble
  are cheaper than a dependency; above it, `rough.js` earns itself. An icon is a path, not a package.
- **The grid under a moving camera is the first thing a low bitrate smears.** Check the dots at the
  destination's bitrate; if they crawl, keep the grid on the ground layer and off the moving one.
- **A font that arrives late reflows the board.** Load it before frame 0 (`document.fonts.ready`, the
  engine's font loader, or `delayRender`) so the first frames do not render with the fallback.
- **A handwriting face for long text** is decoration: a sentence in a marker font is unreadable at arm's
  length. Mono or sans for anything longer than a label.
- **A 6 s piece cannot pay for this look.** The metaphor needs two stations and one camera move before it
  reads as a board, so a sting or a title card is the wrong shape for it.
