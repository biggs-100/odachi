# The piece

## Before a frame

A motion graphic is one sentence made visible. Write the sentence before anything else; if it needs a
comma to breathe, it is two pieces.

| question | the answer that decides the work |
| --- | --- |
| Who watches it, and where? | A phone, muted, at arm's length, half-attention. That is the default. |
| What is the one idea? | The sentence the viewer repeats to someone else. |
| How long? | The shortest time the idea needs, plus one hold. Never the maximum the platform allows. |
| What carries it? | The movement, the type, or one image. Two carriers compete; one wins. |
| What is the last thing they see? | A name, a result, a CTA, or a frame that reads as a poster. Decide it before the first beat. |

Length by kind, as a starting point and not a rule: a logo sting 2-4 s; a title card 3-6 s; a stat or
a chart 6-12 s; a kinetic-type piece 8-20 s; a reel that carries an argument 20-60 s.

## The beat sheet

A beat is a change the viewer can notice: something enters, leaves, moves, lands or stops. A scene
without a beat is a still with extra steps.

Write the beats on `assets/scene-plan.md` before animating. One row per beat, times to the tenth, and
the motion column says what changes, not how it eases:

| # | in-out | beat | on screen | motion | sound |
| --- | --- | --- | --- | --- | --- |
| 1 | 0.0-1.2 | hook | the result, already true | the number lands | dry hit |
| 2 | 1.2-2.8 | it came from | a bar fills from zero | scaleX from the base | low pulse |
| 3 | 2.8-5.0 | hold | the sentence | one slow push, nothing else | silence |

The table is the contract with the renderer: every row has a start and an end, and the piece's length
is the last end, not the sum of the entrances.

## The hook

The first 1.5 s decide whether the rest is watched, and the feed shows them muted. Three hooks that
work, and what each costs:

- **The result first.** Show the number, the outcome, the finished thing, then rewind to how. Cheapest
  to understand; it spends the surprise.
- **The tension.** A question the viewer already has, in five words. It buys attention with a promise
  the piece must keep.
- **The motion itself.** A shape, a cut, a change of state that does not look like the others in the
  feed. It earns the second look but says nothing yet, so the second beat must land fast.

Never open with the logo, a slow fade, or a title that repeats the human's request.

## The shape: build, breathe, resolve

Every scene, and the piece as a whole, has three phases:

- **Build, the first 30%.** The elements enter, in order of importance, overlapping. Do not dump
  everything at once; the eye can track one thing.
- **Breathe, 30-70%.** The content is fully readable and one ambient motion keeps it alive: a slow
  push, a pan, a glow that drifts, a line that draws. One. Stillness is also allowed and often better.
- **Resolve, the last 30%.** An exit, a landing or a decisive stop. Exits are faster than entrances and
  the last frame holds long enough to be read as a still — at least 1 s, and longer if it carries the
  name.

A piece that dumps everything in the build and cuts has no resolve, and reads as an ad for itself.

## Keeping the eye

- One lead per beat. Everything else is a supporting entrance, smaller or later.
- Stagger in order of importance, never in DOM order: the viewer reads the motion as a hierarchy.
- The total entrance sequence of a scene stays under 500 ms, whatever the number of elements. Ten items
  at 60 ms apart is already 600 ms; overlap them instead of shortening them to nothing.
- A persisting element that crosses scenes does not re-enter: it moves, and the continuity is what makes
  the piece feel authored rather than assembled.

## The ending

- The last frame is a poster: a person who pauses there understands the piece without the preceding 3 s.
- If the piece loops, the last frame and the first must be the same frame, or a cut the viewer is meant
  to notice. A loop that jumps reads as a mistake.
- The name and the CTA arrive last and hold. A logo that appears for 0.4 s in a 6 s piece was not shown.
- Do not fade to black between scenes inside a piece: black reads as "over" and the viewer leaves.

## What this file does not decide

The platform numbers, the safe zones and the durations are in `formats.md`; easing and weight are in
`motion-principles.md`; the type sizes are in `type-in-motion.md`. Read those when the plan is written,
not while you annotate.
