---
name: motion-graphics
description: Trigger: a title card, a stat, a chart, a logo sting, kinetic type, or a short video for YouTube, Shorts or Reels where the movement carries the message.
---

## When

The human asks for a motion graphic —a title card, a stat, a chart, a logo sting, kinetic type— or for a
short video for YouTube, Shorts or Reels whose message the movement carries. Not for a narrated explainer
with b-roll, footage editing or an avatar: `hyperframes` and `media-use` own that.

## Rules

- **One idea per piece.** If you cannot say it in one sentence a viewer repeats, it is two pieces.
- **Plan before you animate**: `assets/scene-plan.md`, a row per beat. A fix in the plan is free.
- **The first 1.5 s carry the piece**: muted feeds show the tension or the result, never the logo.
- **Motion that explains nothing does not ship**: remove it and the message should survive.
- **Vary the defaults**: not one ease, one duration, or `y: 30 + opacity` for everything.
- **Exits are faster than entrances**: 0.4 s in, 0.25 s out.
- **Hierarchy is choreography**: the important element moves first, the whole entrance under 500 ms.
- **Every scene builds, breathes, resolves**; one ambient move at most, and stillness is a tool.
- **Two focal points and three layers per frame**: ground, content, accent.
- **Type is read at arm's length**: 3-6 words, 1.5 s minimum, measured contrast, burned-in captions.
- **The clock belongs to the renderer**: no wall clock, no unseeded random, no screen recording.
- **Preview a still, then 3 s, then the render**; watch the MP4, not the preview.
- **Fit the destination before rendering**: frame, safe zone, and a first frame that reads as a thumbnail.
- **Say what you did not render or verify.** A frame you did not see is a guess.

## Decision gates

| Situation | Load |
|---|---|
| a piece to plan: the one idea, the beats, the hook, the length | `references/the-piece.md` |
| easing, durations, stagger, transitions, weight | `references/motion-principles.md` |
| the piece looks machine-made and you need the tell and its fix | `references/the-ai-look.md` |
| kinetic type, captions, numbers, reading speed | `references/type-in-motion.md` |
| YouTube, Shorts or Reels: size, duration, safe zone, loop | `references/formats.md` |
| building and rendering: determinism, HyperFrames, Remotion, the last pass | `references/the-render.md` |

## Output

The plan, the renderer, the format, the file that rendered, and what was not seen or verified.

## References

- `references/the-piece.md` — the one idea, the beat sheet, the hook, the length and the ending.
- `references/motion-principles.md` — easing, speed, stagger, choreography, transitions, weight.
- `references/the-ai-look.md` — the tells a generated piece repeats, and the fix for each.
- `references/type-in-motion.md` — kinetic type, reading speed, captions, numbers.
- `references/formats.md` — 16:9 and 9:16, durations, safe zones, loop, first frame.
- `references/the-render.md` — deterministic frames, HyperFrames, Remotion, the checklist.

## Assets

- `assets/scene-plan.md` — the template a piece is written on before a frame is animated.

## Never

- Screen-record a browser in real time and call it a render.
- Ship a piece you only watched in the preview.
- Animate every element because it is there.
