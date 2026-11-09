# Motion principles

## The five decisions

Every tween answers five questions. A piece that answers them differently each time reads as a
compilation; one that answers them the same way everywhere reads as a template.

1. **What changes** — position, scale, opacity, colour, a mask, a value. One property that carries the
   meaning, at most one more supporting it.
2. **How far** — the distance, in the frame's own units, not in pixels of a design tool.
3. **How long** — the band below, chosen by the weight the element has.
4. **With what curve** — the adverb. This is the part most generated work gets wrong.
5. **When it starts** — relative to what, and with how much overlap.

## Easing is emotion

The transition is the verb; the easing is the adverb. The same slide can be confident, dreamy or
childish, and the curve alone decides which.

| direction | use it for | why |
| --- | --- | --- |
| `out` (ease-out) | anything entering or landing | starts fast, decelerates: responsive, decisive |
| `in` (ease-in) | anything leaving | starts slow, accelerates away: it is gone, not stuck |
| `inOut` | an element moving between two positions | both ends at rest: a deliberate journey |
| linear | progress bars, a loop, a constant pan | no meaning to give; any curve would lie |
| spring / elastic | a gesture, a playful brand | overshoots on purpose; costs the seriousness of the piece |

Getting the direction backwards is the most common defect in generated motion: ease-in on an entrance
feels sluggish, ease-out on an exit feels reluctant.

Vary the curve. No more than two independent tweens in a scene share an ease, and a piece that uses
`power2.out` on everything has one adverb for a whole vocabulary.

## Speed is weight

| band | duration | reads as |
| --- | --- | --- |
| fast | 0.15-0.3 s | energy, urgency, a UI response |
| medium | 0.3-0.5 s | professional, most content |
| slow | 0.5-0.8 s | gravity, luxury, a card turning over |
| very slow | 0.8-2.0 s | cinematic, atmospheric, a camera move |

The slowest scene should be about 3x the fastest in the piece. Uniform speed is the second tell of a
generated piece: every element moves like it weighs the same, and nothing lands.

The biggest element is not always the heaviest: a 12 px label that snaps in 0.2 s and a 400 px panel
that arrives in 0.7 s say different things about what matters.

## Choreography

- **Hierarchy is the order of movement.** The element that moves first is read as the most important,
  whatever its size.
- **Overlap, don't queue.** The second entrance starts before the first ends; waiting for completion
  makes a list, not a scene.
- **Total entrance under 500 ms** in a scene, whatever the count: stagger 20-50 ms between supporting
  elements and overlap the rest.
- **Exits are faster than entrances.** A card that takes 0.4 s to appear takes 0.25 s to disappear.
- **Anticipation and settle.** A small move opposite the direction before a big one, or a 10-20%
  overshoot that settles, sells weight. More than that is cartoon; none of it is a slideshow.
- **One lead per beat.** Everything else is supporting and later or smaller.

## Stillness, secondary motion and the frame

- **Not every scene needs ambient motion.** A hold after a landing is the strongest thing a 6 s piece
  can do; ambient zoom on every scene flattens the whole piece.
- **One ambient move per scene at most**: a slow push, a pan, a drift, a breathing glow. It runs under
  the content and never competes with it.
- **Move the frame or the object, not both.** A camera push and an element sliding at the same time
  reads as a wobble unless the two are deliberately locked.
- **Continuity beats re-entry.** An element that survives a scene change moves across it; re-entering it
  as if new resets the viewer's model of the piece.

## Transitions carry meaning

| transition | says | cost |
| --- | --- | --- |
| crossfade | this continues | looks like two videos layered unless the layouts match |
| hard cut | wake up, a register change | unforgiving: a mismatch of position reads as a mistake |
| slow dissolve | drift with me | slow; it spends 0.6-1.0 s of the piece |
| wipe or push with a direction | the story moves that way | implies space, so the next scene must agree |
| match cut on a shape | these two things are the same | the strongest and the most work |

A transition chosen for variety rather than for meaning is the third tell: every scene ends the same
way, and the piece reads as slides with dissolves.

## Frames and blur

- **Match the destination's frame rate**, 24, 25, 30 or 60. Never mix inside one piece.
- **Motion blur is a shutter, not a filter.** It belongs where the speed would smear in a real camera;
  on a 0.4 s fade it is noise, and on text it destroys legibility.
- **The last frame holds.** A piece that ends on the frame the motion stopped has no ending.

## The test before shipping motion

- Remove the animation in your head: if the piece still says the same thing, the animation was
  decoration.
- Watch it at 0.25x: any tween that never settles, any element that flickers or teleports, shows there.
- Watch it muted and at 1x: if the meaning depends on a sound, the piece is not finished.
- Name the adverb of each scene out loud. If the name is the same for all of them, the piece has one
  emotion and probably not the one you intended.
