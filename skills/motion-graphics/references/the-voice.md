# The voice is the human's

## The rule

Never synthesize the voiceover, and never fill a gap with a TTS voice. The human records it because they
choose the voice, the accent and the pace, and that choice is part of the piece. An unrecorded piece
stops at the script: the animation is never built on guessed times.

## Two turns

- **Turn one: the script, then a stop.** The beat sheet becomes a reading script — one line per beat, the
  emphasis in bold, `[pause]` where the visual needs air, and each line's length in seconds. Save it where
  the human will read it, name the exact path where the audio goes, and say what waits. Then stop.
- **Turn two: the recording.** Normalize a copy, take the duration, extract the word timestamps, and write
  the real times into the beat sheet **before a frame is built**. The guessed times are gone; the
  recording is the clock.

## The script a human reads aloud

- One beat per line, no technical labels, 2-4 words of emphasis, short lines that fit one breath.
- Mark the pauses, not the beats: a pause is where the visual lands, and it is the only direction a
  reader follows without acting.
- Give each line's time and the total. A person reading cold overshoots by 20-30%, so the take is the
  length, and the piece's length is `formats.md`'s to decide.
- Ask for one take per section, not one perfect take. The sections are the edit points, and re-recording
  a single line is cheap.
- Count the words before recording: at about 3 words a second, 150 words is 50 s. Pace is the piece's
  pace.

## After the recording arrives

- **Cut the animation to the take; never stretch the human to fit the animation.** A few percent is the
  ceiling, and only with their yes, before the voice starts sounding wrong.
- **Normalize a copy**, never the original, with the loudness `formats.md` asks for, and keep the take
  beside it.
- **Extract word timestamps** with a local ASR (faster-whisper, or the engine's own whisper helper). They
  are an instrument, not the truth: a boundary is ±50 ms.
- **Put the cue on the stressed word**, not on the first frame of the sentence. The visual may start
  before the line ends, and that overlap is what separates an explainer from a metronome.
- **The beats become uneven, and that unevenness is the piece.** A person's rhythm has a long pause before
  the number and none before the list; the animation follows it.
- **The tail belongs to the piece**: hold the last frame until the last word has landed, plus 0.5 s.
- **Check the take against the script** before building: a dropped line or a changed number is a
  re-record, not a caption fix.

## Both engines, one shape

Alignment is the same shape on either engine: one paused master timeline, seeked by frame, with every cue
at a time in seconds taken from the recording.

- **HyperFrames**: register the paused GSAP timeline on `window.__timelines` and place each cue at its
  second; `the-render.md` owns the contract.
- **Remotion**: `Math.round(word.start * fps)` into `<Sequence from>`, read with `useCurrentFrame()`, the
  audio a local `<Audio>`.
- Neither reads the wall clock, both fix the fps, and both keep the audio in the project rather than at a
  URL.

## What the agent cannot decide

- Whether the take performed the line. Ask, naming the timestamp of the line in doubt.
- Whether a line that runs long should be cut or re-recorded. Report the words over and the seconds
  needed; the human chooses.
- What a misheard word was. Ask; a caption that guesses a number is worse than no caption.

## The instruments

- `ffmpeg` and `ffprobe`: the copy, the loudness, the duration.
- A local ASR for the word timestamps. A paid transcription service is a second dependency for a job that
  runs on the machine.
- No TTS, no voice changer, no synthetic filler. A second voice is a second recording.
