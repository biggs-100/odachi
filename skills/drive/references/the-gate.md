# What the gate answers

## What a block means

- `0` ready: continue. `1` you can act: the `next:` line names the exact command; run it.
- `2` only the human can unblock: do not retry, do not go around, ask with `ask`.

## Review, when it applies

`odachi review --show` says whether this tree asks for one and why: ask when it does or the
human asks. `--exec` hands the prompt to the reviewer; a causal `BLOCKER` blocks: fix it or
`--refute` citing the diff. The verdict is valid for THAT tree: one correction and one
re-review; a `WARNING` or `SUGGESTION` goes to the task doc and earns no round.
