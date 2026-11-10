---
name: drive
description: Trigger: when you work inside an odachi project: which verb goes at each moment, what each block means and how a feature closes. To start from zero, use bootstrap.
---

## When

You are inside an odachi project: there is `brief.md`, `odd/tasks/` and the odachi block in `AGENTS.md`. Before the first action, read `brief.md`: the vision and what is NOT included. If the folder is empty, the right skill is `bootstrap`, not this one.

## Rules

- **A question is answered, not implemented**: an answer is not a proposal, and work starts when the human asks for it or approves one.
- **An irreversible action names its exact target and waits for a yes**: push, pull request, merge, tag, release, `init`, another folder.
- **A fence rule protects a file the feature must not touch**, never a reason (ADR 0027).
- `brief.md` is the vision and the fence: the human writes it, you do not rewrite it.
- `odd/` is the plan and the rules: `odd/tasks/` plus `fence.toml`, `profile.toml`, `review.toml` and `roles/`, and it travels with the repository.
- `proof/` is the evidence: real command output tied to the tree, written by `check --run`.
- `.git/odachi/` is the clone's store: verdicts, versions, the lock and the client choice; it never travels, and `store move` adopts an older `.odachi/`.
- **No brief, no build**: `odachi start`, answering; without a terminal, with `--outcome`, `--audience`, `--done`, `--exclude`, `--sensitive`, `--deadline` and `--learn`, or `--brief` when `shape` drafted it.
- `odachi plan "<step>" --req R1 --route inline`: one step per deliverable behavior, and the step says who runs it, `--route delegated --agent <name>` included.
- **Commit what start wrote** -brief, fence, profile and task doc, steps included- before you branch; apart, the next change is judged by the last feature's fence.
- Work the step. A vision decision is `odachi ask "<question>"` and a wait; a requirement the human changes is `odachi ask --change "<words>" --req R1`, whose steps reopen and `brief.md` does not move.
- `odachi check` is the gate, and the proof is owed by the closing tree, not by every step.
- One commit per step, then `odachi plan --done T1`: it records the tree the step produced, which a rebase cannot rewrite.
- At the end, `odachi summary` and three lines on what the human decided.
- If the request does not fit in one reading, load `shape` before `start`: it ends in `odachi start --brief`.
- Once the work is understood, load `slices` before splitting: a writer only for real parallelism or a long session, never for size.
- A change that fits one line opens with `odachi start --light "<requirement>"`: one quoted requirement, one step with the same words, no brief, and the fence and voice untouched.
- A change whose diff is the whole record writes `Odachi-Change: <line>` in its commit instead: `check` and `review` read it from the merge-base, the proof is named after the line, and no document is written.
- That lane lives on a branch and reads the remote's default branch, or the `--base` the human names; a second step, a vision decision, or anything the fence protects means you left it.
- Requirements quoted **verbatim**, done-when observable, and observe RED before the code; a delegated step declares `--route delegated` at plan time, and every commit carries its `Risk:` line.

## Specs, test-first and risk

- Requirements quoted **verbatim** and "Done when" observable; if missing, `ask`.
- Before a step, load `tests` and observe RED; `check --run` only with GREEN.
- A delegated step declares it when it is written (`--route delegated --agent <name>`): the task doc keeps the route, and a chat is not the record.
- Close the commit with the `Risk:` line from `commits`: its table names the lens and the verifier, and a risk allows no waiver.

The verbs: `start` interviews · `ask` asks or records a craft decision (`--craft`) · `plan` opens and closes steps · `check` gates and proof · `review` verdict per tree · `doctor` health without writing · `level` profile · `summary` state · `undo` undoes the last write · `init`, `skills`, `skill new` and `pack` are installation; `--version` touches nothing.
`docs/releases.md`: a release is a tag for a person's change, never for a test alone.

## Output

The next verb and why: the command that moves the feature, the block it read and the file it changes. When the work is blocked, the question only the human can answer.

## References

- `references/the-gate.md` — what a block means, and when a review applies: `--show`, `--exec`, a blocker and a warning.

## Never

- Write the proof by hand or reuse an old one: the proof is the real output of this tree.
- Say "done" with `check` red, or without running it.
- Change `brief.md` or the fence without the human asking.
- Retry an exit 2, or invent the missing file or flag.
- Add work outside the scope because "it was easy": ask, do not decide.
- Read, run odachi in, or touch another folder or repository: another clone is another project, with its own brief, fence, store and memory. The human asks for that; never decide it.
