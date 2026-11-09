# motion-graphics

## What must exist
- a skill that turns a request for a motion graphic or a YouTube or Reels video into a piece that reads senior

## Done when
- `skills/motion-graphics/SKILL.md` has a human trigger, rules, gates, references and output, under 60 lines
- six references: the-piece, motion-principles, the-ai-look, type-in-motion, formats, the-render
- `assets/scene-plan.md` is the template a piece is written on before a frame is animated
- `odachi skills --check` is green, `skills --write` regenerates SKILLS.md, doctor counts 20 skills
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- Unreleased names the change: a new capability, patch (0250)

## Not included (first version)
- no renderer or dependency shipped: the skill names HyperFrames and stops when missing - back when odachi must render
- no sourcing, narration, avatars, b-roll or footage editing: media-use and hyperframes own it - back when a piece needs them
- no change to ui-craft's motion reference: screen motion is not video motion - back when one rule must hold for both
- no caption timing or transcript alignment - back when captions must be generated from audio, not written
- no bench, verb, installer, workflow, canvas or proof change - back when the bench must judge a skill

## Decisions taken
- It is used from the computer, by a person who asks for a video or a motion graphic for YouTube or Reels.
- HyperFrames is the default renderer and Remotion the named alternative; the skill claims no frame it did not render.
- Scope: motion-led pieces up to about 60 s, narrated or not, where the movement carries the message.
- Depth lives in references/ and the body stays under 60 lines (ADR 0050).
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: whether naming the machine-made tells in one place makes a piece read senior on the first render.
