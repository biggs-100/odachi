---
name: drive
description: Trigger: when you work inside an odachi project: which verb goes at each moment, what each block means and how a feature closes. To start from zero, use bootstrap.
---

## When

You are inside an odachi project: there is `brief.md`, `odd/tasks/` and the odachi block in `AGENTS.md`. Before the first action, read `brief.md`: the vision and what is NOT included. If the folder is empty, the right skill is `bootstrap`, not this one.

## Rules

The four pieces and who owns each one:

- `brief.md` — the vision and the fence. The human writes it; do not rewrite it.
- `odd/tasks/` — the plan: requirements, steps and history. You move it with `odachi plan`.
- `proof/` — the evidence: real command output, tied to the tree. `check --run` writes it.
- `.odachi/` — gates and ceremony. The binary reads the profile, not you.

The feature cycle, in order:

1. No brief, no build: `odachi start`, answering; without a terminal, with `--outcome`, `--audience`, `--done`, `--exclude`, `--sensitive`, `--deadline` and `--learn`.
2. `odachi plan "<step>" --req R1`: one step per deliverable behavior.
3. Work the step. If a vision decision appears: `odachi ask "<question>"` and wait.
4. `odachi check --run "<command>"`: green is the only definition of done.
5. One commit per step, and `odachi plan --done T1 --commit <sha>`.
6. At the end, `odachi summary` and tell in three lines what the human decided.

If the request does not fit in one read, load `slices` before splitting. Its delegation policy wins over other global skills: a writer only for real parallelism or a long session, never for size or file count.

## Specs, test-first and risk

- Requirements quoted **verbatim** and "Done when" observable; if missing, `ask`.
- Before a step, load `tests` and observe RED; `check --run` only with GREEN.
- If you delegate a step, record it with `ask --craft "delegated step: ..." --because "..."`.
- Close the commit with the `Risk:` line from `commits`; a risk asks for `review`.

The verbs: `start` interviews · `ask` asks or records a craft decision (`--craft`) · `plan` opens and closes steps · `check` gates and proof · `review` verdict per tree · `doctor` health without writing · `level` profile · `summary` state · `undo` undoes the last write · `init`, `skills`, `skill new` and `pack` are installation; `--version` touches nothing.

## What a block means

- `0` ready: continue.
- `1` you can act: the `next:` line names the exact command; run it.
- `2` only the human can unblock: do not retry, do not go around, ask with `ask`.

## Review, when it applies

`odachi review` prints the prompt; `--exec "<command>"` hands it to the reviewer through stdin and `{prompt}`. A causal `BLOCKER` blocks: fix the code or use `--refute` citing the diff. The verdict is valid for THAT tree: any change invalidates it.

## Never

- Write the proof by hand or reuse an old one: the proof is the real output of this tree.
- Say "done" with `check` red, or without running it.
- Change `brief.md` or the fence without the human asking.
- Retry an exit 2, or invent the missing file or flag.
- Add work outside the scope because "it was easy": ask, do not decide.
