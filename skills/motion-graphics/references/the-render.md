# The render

## The clock belongs to the renderer

A browser plays animation in real time, and real time is a promise nobody keeps: a busy CPU drops
frames, a late font shifts the first second, a slow disk stalls a decode. A renderer takes the clock
away from the page, asks for frame 0, frame 1, frame 2, and writes each one to disk. That is the whole
difference between a video and a recording of a browser.

Consequences that decide the code:

- **No `Date.now()`, `performance.now()`, timers or unseeded `Math.random()`** anywhere in the piece. The
  same composition renders the same frames twice; a wall clock makes that false.
- **No infinite animation** (`repeat: -1` with no bound): the render has no end to discover. A loop is a
  finite timeline whose first and last frame match.
- **No measuring at tween time**: `getBoundingClientRect()` during a tween returns whatever the renderer
  happened to have on screen. Compute positions once, at setup, and reuse them.
- **No `from()` tweens that write their start state at build time** inside a composition that is seeked:
  the state is applied before the scene starts, and the element flashes or never enters. Use `fromTo`
  and state both ends.
- **No two transform tweens on one element at the same time.** A `y` entrance and a `scale` push
  concurrently overwrite each other and the element can stay offscreen with no error. Combine them, or
  put one on a parent and one on a child.
- **Ambient motion lives on the timeline**, never in a bare `gsap.to()`: a standalone tween runs on wall
  clock and is simply absent from the rendered file while looking right in the preview.

## HyperFrames (default)

HTML, CSS and GSAP rendered to MP4 by a deterministic engine. The composition is a folder whose
`index.html` is the video: clips carry their time (`data-start`, `data-duration`), one paused GSAP
timeline is registered on `window.__timelines`, and the renderer seeks it frame by frame.

- The framework ships its own agent skills: `npx hyperframes skills update` once, then read the
  `hyperframes-*` skill that matches the work. This file does not restate its contract; those skills own
  it.
- The loop: `npx hyperframes lint` → `npx hyperframes check` → `npx hyperframes snapshot` (stills) →
  `npx hyperframes preview` (real time, for the human) → `npx hyperframes render`.
- Media is local: `assets/` in the project, no remote URL at render time.
- Use it when the composition is HTML-shaped, when the piece is mostly type and shapes, and when a
  human should be able to edit the same file afterwards.

## Remotion (the alternative)

React and TypeScript: a component receives the frame number and returns the tree for that frame. Use it
when the piece is already inside a React codebase, when the data is React-shaped, or when the team
already renders with it. `npx remotion render`.

The two produce the same kind of file; the choice is about what the project can maintain, not about
quality. Do not port a composition between them without being asked: the timing and the easing do not
translate one to one.

## The order that costs least

1. **Plan.** The scene plan, the times, the one idea. `assets/scene-plan.md`.
2. **Stills first.** Render three key frames —the hook, the middle, the end— and look at them. A layout
   error found here costs a minute; the same error found after the render costs the render.
3. **Three seconds.** Render the first beat and watch it. This is where easing, weight and the hook are
   judged, because they only exist over time.
4. **The full render.** Only when the three seconds are right.
5. **Watch the MP4**, from the start, with the sound on and then muted. The preview plays in real time
   and lies about timing; the file is what the platform will show.

## The last pass

- **At 1x, muted, on a phone**: the hook, the reading, the landing.
- **At 0.25x**: tweens that never settle, elements that flicker, text that reflows.
- **Frame by frame around a landing**: the value stops where it should, the name is fully formed, the
  still holds.
- **The safe zone overlay** at the destination's size, and the first frame at thumbnail size.
- **The sound**: peaks, the voice against the music, the silence before the logo.
- **The file**: size, length, codec, the first and last frame as a loop.

## When the renderer is not there

Say it and stop: "HyperFrames is not installed; `npx hyperframes --version` fails, so there is no
render." Do not produce a screen recording as a substitute, do not write an MP4 by hand, and do not
report a video that does not exist. Offer the install command and the plan, so the work continues from
the plan when the tool is available.

## What a render does not prove

A green render proves the file exists and the frames were written. It does not prove the piece is good,
that the safe zone was respected, or that the sound is mixed. Those are in the checklist above, and
each one is a thing a person looked at, not a thing a command reported.
