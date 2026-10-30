---
name: drive
description: Trigger: when you work inside an odachi project: which verb goes at each moment, what each block means and how a feature closes. To start from zero, use bootstrap.
---

## When

You are inside an odachi project: there is `brief.md`, `odd/tasks/` and the odachi block in `AGENTS.md`. Before the first action, read `brief.md`: the vision and what is NOT included. If the folder is empty, the right skill is `bootstrap`, not this one.

## Rules

The four pieces and who owns each one:

- `brief.md` — the vision and the fence. The human writes it; do not rewrite it.
- `odd/` — the plan and the rules: `odd/tasks/` (moved with `odachi plan`) plus `fence.toml`, `profile.toml`, `review.toml` and `roles/`, which travel with the repository.
- `proof/` — the evidence: real command output, tied to the tree. `check --run` writes it.
- `.git/odachi/` — the clone's store: verdicts, versions, the lock and the client choice. It never travels. An older project carrying `.odachi/`: `store move` adopts it and prints the git commands.

The feature cycle, in order:

1. No brief, no build: `odachi start`, answering; without a terminal, with `--outcome`, `--audience`, `--done`, `--exclude`, `--sensitive`, `--deadline` and `--learn`, or with `--brief` when `shape` drafted it.
2. `odachi plan "<step>" --req R1`: one step per deliverable behavior.
3. Commit what start wrote -brief, fence, profile and task doc, steps included- before you branch. If they do not land together, the next change is judged by the last feature's fence, or a reset loses them.
4. Work the step. A vision decision: `odachi ask "<question>"` and wait. A
   requirement the human changes: `odachi ask --change "<words>" --req R1`: its steps reopen and `brief.md` does not move.
5. `odachi check`: the gate. The proof is owed by the closing tree, not by every step.
6. One commit per step, and `odachi plan --done T1`: it records the tree the step produced, which a rebase cannot rewrite.
7. At the end, `odachi summary` and tell in three lines what the human decided.

If the request does not fit in one reading -you cannot say what must exist or what
it must not touch- load `shape` before `start`: it ends in `odachi start --brief`.
Once the work is understood, load `slices` before splitting when it does not fit in
one step. Its delegation policy wins over other global skills: a writer only for
real parallelism or a long session, never for size or file count.

## Specs, test-first and risk

- Requirements quoted **verbatim** and "Done when" observable; if missing, `ask`.
- Before a step, load `tests` and observe RED; `check --run` only with GREEN.
- If you delegate a step, record it with `ask --craft "delegated step: ..." --because "..."`.
- Close the commit with the `Risk:` line from `commits`: its table names the lens and the verifier, and a risk allows no waiver.

The verbs: `start` interviews · `ask` asks or records a craft decision (`--craft`) · `plan` opens and closes steps · `check` gates and proof · `review` verdict per tree · `doctor` health without writing · `level` profile · `summary` state · `undo` undoes the last write · `init`, `skills`, `skill new` and `pack` are installation; `--version` touches nothing.

## What a block means

- `0` ready: continue. `1` you can act: the `next:` line names the exact command; run it.
- `2` only the human can unblock: do not retry, do not go around, ask with `ask`.

## Review, when it applies

`odachi review --show` says whether this tree asks for one and why: ask when it does or the
human asks. `--exec` hands the prompt to the reviewer; a causal `BLOCKER` blocks: fix it or
`--refute` citing the diff. The verdict is valid for THAT tree: one correction and one
re-review; a `WARNING` or `SUGGESTION` goes to the task doc and earns no round.

## Never

- Write the proof by hand or reuse an old one: the proof is the real output of this tree.
- Say "done" with `check` red, or without running it.
- Change `brief.md` or the fence without the human asking.
- Retry an exit 2, or invent the missing file or flag.
- Add work outside the scope because "it was easy": ask, do not decide.
- Read, run odachi in, or touch another folder or repository: another clone is another project, with its own brief, fence, store and memory. The human asks for that; never decide it.
