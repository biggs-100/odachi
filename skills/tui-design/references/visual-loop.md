# The eyes

## The frame is the evidence

A TUI is what the terminal draws, not what the source says. Capture the frame before the change and
after it, at the sizes that ship.

| width | rows | what it proves |
| --- | --- | --- |
| 180 | 50 | the full layout: rail, panes, status bar |
| 140 | 40 | exactly at the breakpoint, where the rail usually leaves |
| 80 | 24 | the narrow case: what survives when everything competes |

Pick the project's real breakpoint if it has one; the three widths above are the default when nobody
decided.

## Capturing it

The frame is text. A text frame is deterministic, diffable and runs anywhere; a PNG is better when the
tool exists, and it is named and never required.

- **Go and Bubble Tea.** The model's `View()` is already a string: render it with a fixed width and
  height in a test and write `testdata/<target>-<cols>x<rows>.txt`. `teatest` adds the interactive half
  —send keys, read the output, compare a golden— which is how a focus state or a loading state is
  captured without a person at the keyboard. `go-tests` carries the harness.
- **Any stack with a snapshot flag.** Run the binary at `--cols 80 --rows 24` (or the equivalent) and
  keep its output. This is the cheapest and the one to ask for first.
- **Any program at all.** `tmux new -d -x 80 -y 24 './app'` then `tmux capture-pane -p` gives the
  drawn text; `script` on macOS and Linux does the same for a one-shot run.
- **A PNG** when `vhs`, a headless browser or the project's own screenshot tool exists. Note the tool
  and its version in the report: a harness that stopped working is worse than none. In this project's
  own experience, `vhs` wrote no output at version 0.12 on one machine, so never make it the only path.

Never type a prompt into a captured shell that inherits the user's permissions. Feed it a scripted
input, or capture a view that needs none.

## The loop

1. **Capture before** at the three widths, named `<target>-<step>-<cols>x<rows>`.
2. **Write the finding before editing**, from the frame. A finding written after the fix is a
   justification.
3. **Change one named target**, then capture again at the same widths.
4. **Compare point by point** against the brief and the previous frame: what moved, what was supposed
   to move, what moved that nobody asked.
5. **Stop at three rounds**, or earlier when no P0-P2 is left. Report what remains.

The loop is a reading, not a ritual: one frame read carefully beats four captured in a hurry.

## Pair it with a rendered test

A visual check is a person looking. The project's tests keep it from coming back: a golden frame, a
render count that proves an unchanged frame costs nothing, an overflow test at the narrow width, a
breakpoint test at the width where the rail leaves. A change that fixes a screenshot and breaks a golden
is not fixed.

## When there is no way to capture

Say it and stop before editing: "there is no snapshot path in this project, so I cannot see the frame".
Offer the smallest one (a `View()` dump at a fixed size is a dozen lines) and, until it exists, review
the code and label the review as reading, not seeing. Do not describe a frame nobody looked at.
