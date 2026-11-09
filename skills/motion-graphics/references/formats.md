# Formats

The interface moves. Every number here was checked in November 2026 and every one of them can change
without a migration guide; treat them as a design margin, verify in the app before a high-stakes upload,
and never put a line of meaning closer to an edge than the strictest figure below.

## The two frames that matter

| destination | frame | ratio | practical length |
| --- | --- | --- | --- |
| YouTube | 1920x1080 | 16:9 | the idea, plus one hold; no platform limit |
| YouTube Shorts | 1080x1920 | 9:16 | up to 3 min; the first 2 s decide |
| Reels | 1080x1920 | 9:16 | 60-90 s practical; 3 min reliable |
| TikTok | 1080x1920 | 9:16 | 15-60 s for a motion piece |
| a feed square | 1080x1080 | 1:1 | the same piece when it must work everywhere |

A piece made for one frame is not the same piece cropped to another. Up to 60 s, produce the 9:16 first
and a 16:9 cut second, or produce one and say which destination was not checked.

## Safe zones, 9:16

| placement | top | bottom | sides | source |
| --- | --- | --- | --- | --- |
| Reels and Stories, ads | 14% · 269 px | 35% · 672 px | 6% · 65 px | Meta Ads Guide, checked 2026-11 |
| Reels and Stories, ads with a disclaimer | 14% | 40% | 6% | Meta Ads Guide, 2026-11 |
| Reels, organic, measured on an iPhone | 87 px | 486 px | 98 px left / 202 px right | measured Oct 2026; other tools report 220-250 top and 320-520 bottom |
| TikTok, measured from its own template | 12.5% · 135 px | 34.4% · 660 px | 11.1% · 120 px | measured from the downloadable template |
| YouTube Shorts | about 13% | about 25% · 480 px | about 6% | platform figures, 2026-11 |

Two rules survive every redesign:

- **The bottom is the expensive edge** — caption, account, audio, CTA. Nothing that matters lives below
  65% of the height on a vertical frame.
- **The right edge has a column** on Reels and TikTok. A logo or a number in the last 200 px on the
  right is covered by buttons, not cropped by the device.

A Reels cover also appears cropped to about 4:5 in the profile grid. Design the first frame so the
middle survives that crop.

## The first frame is a thumbnail

- It is read at about 120 px wide, in a grid, next to other pieces. A single word, a number or a face.
- It is also the loop's last frame when the piece loops, so it must work as both.
- It is not a title slate. A slate costs the hook and buys nothing.

## Sound

- **Assume muted.** Every caption, every number and every conclusion must survive without audio.
- **The mix for social**: about -14 LUFS integrated, true peak -1 dB; a voiceover sits 6-10 dB above the
  music, and music that fights the voice is removed rather than lowered.
- **A piece that depends on a sound effect for meaning** is a piece that most viewers do not understand.

## The file

| setting | value | why |
| --- | --- | --- |
| container and codec | MP4, H.264, AAC | every one of these platforms accepts it without a re-encode |
| frame rate | 30, or the rate the destination prefers; never mixed | 60 reads as a game capture, 24 as film |
| resolution | the native frame of the destination | upscaling a 720p render is visible on every phone |
| bitrate | 10-20 Mbps at 1080p | below 8 Mbps, gradients band and type loses its edge |
| colour | sRGB, tagged | an untagged render shifts on some players |
| length | the piece's last beat, not the platform's maximum | the platform's limit is not a target |

## The loop

- A loop is designed, not discovered: the first and last frame are the same frame, or the cut is the
  point.
- A 3-6 s loop that repeats cleanly is worth more than a 12 s piece that has to be watched once.
- If the piece does not loop, the last frame still holds: a viewer who pauses on the end sees a poster.

## What to verify before upload

- The frame size and the length in the exported file, not in the settings.
- The safe zone with an overlay at the destination's real size; a screenshot from the app is the only
  source that counts.
- The first frame as a thumbnail, at 120 px.
- The captions on a phone, at arm's length, in daylight.
- The file plays from the start to the end in a plain player, with sound and without.

If one of these was not checked, say so in the report. An unchecked upload is a guess with a deadline.
