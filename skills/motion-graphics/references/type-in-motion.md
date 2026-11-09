# Type in motion

## Read it at arm's length

The viewer holds a phone at arm's length, muted, in a feed where everything else is also moving. Type
that would be polite on a 27-inch monitor is unreadable there.

- **Sizes as a share of the frame's height**, never as points: a hero line is 8-14% of the height on
  1080x1920; a caption is 3.5-5%; a label is 2.5-3.5%. Under 3% on a vertical frame is decoration.
- **A hero line fills 60-80% of the width.** A line that fills 40% is a web paragraph that was pasted
  into a video.
- **Contrast is measured, never estimated**: 4.5:1 for captions and small text, 3:1 for a hero. Over
  video the ground changes frame to frame, so a plate, a scrim or a shadow behind the text is part of
  the type, not a decoration.
- **One phrase at a time.** 3-6 words. A full sentence is a paragraph, and a paragraph on a phone is a
  slideshow with a narrator nobody can hear.

## Timing is reading

| the line | minimum on screen |
| --- | --- |
| a 2-3 word label | 1.0 s |
| a 4-6 word phrase | 1.5 s |
| a 7-10 word sentence | 2.5-3.0 s |
| a number that lands | 1.0 s after it lands |
| the name or the CTA | 2.0 s |

Reading speed is about 3-4 words per second at 1x. A line that needs 2.5 s to read and is on screen for
1.2 s was never read; it was flashed.

## Kinetic type

- **Animate at the phrase level.** Word-by-word is for a hook of 5 words or a lyric; letter-by-letter is
  for a logo or a lockup, never for a sentence.
- **Let the type do one thing.** A mask wipe, a tracking change, a line that reveals a baseline, a
  weight shift. Two effects on the same line and the eye reads neither.
- **A mask reveal beats an opacity fade** for a headline: the letterform stays put and the frame opens
  onto it, which reads as authored rather than as a PowerPoint.
- **Tracking and weight animate well**; letter-spacing on a wrap reflows the line mid-tween, and that is
  always a mistake. Animate one line, or split the text into elements first.
- **A moving line of text needs a fixed anchor**: the baseline holds and the words arrive into it. Text
  that moves and re-centers at the same time is unreadable.
- **Never animate a hyphen, a wrap or a justified column.** Type set for a page does not animate; type
  set for a frame does.

## Numbers

- **Tabular figures**, so a count-up does not shift its width on every digit.
- **A count-up lands on the real value and holds**; a number that spins past its target and stops on a
  guess is worse than no number.
- **Carry the unit and the date**: "$4.2M, Q3 2026" is a fact; "$4.2M" is a claim.
- **Large numbers need a thousand separator or a scale**; the eye cannot parse `4200000` in 1 s.
- **A percentage needs its base**, or it is a magic trick.

## Captions

- **Burned in**, because the feed autoplays muted. A caption file that the platform may or may not show
  is not a caption.
- **Two lines maximum**, 32-42 characters per line, one phrase per screen, 15-20 characters per second.
- **High contrast on every frame**: white with a heavy outline, or a solid plate. Thin type over footage
  survives the preview and dies on a phone in daylight.
- **Inside the safe zone**, not at the very bottom: the interface covers it there. `formats.md` has the
  numbers.
- **Verbatim, not cleaned up**: captions that drop words are read as inaccurate, and auto-captions that
  misspell a price are worse than none.

## The type system of a piece

- **One family, two weights, one accent.** A second family is a second voice, and a piece has one.
- **A scale, not a size per scene**: use three sizes (display, body, label) and stay inside them.
- **Licensed and loaded before the first frame**: a font that arrives late reflows the first second, and
  the render keeps the reflow.
- **No hyphenation and no justification** inside a frame: a hyphen at a line break in motion looks like
  a broken word.
- **All caps is a label, not a paragraph**: three words maximum. Long all-caps text loses its word
  shapes and becomes a texture.

## What this file does not cover

The palette, the layout and the composition live in `the-piece.md` and in `ui-craft`; the numbers here
are for the type. If the piece is a screen recording or a UI mockup with type over it, `ui-craft`'s
numbers still hold.
