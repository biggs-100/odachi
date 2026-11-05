# Next session

Written at the close of 0201 and v0.15.1, after a real new project exposed the
onboarding gap. `main` is clean, the design parity is complete, and the installed
binary and skills are the release.

## What shipped

- **0201, PR #332**: the last design parity slice. `odachi design list` reports the
  design sources and the ones newer than `design/system.md`; the skill syncs before the
  first design and names a stale source. `odachi design handoff` on a design linked to
  `odd/tasks/<feature>.md` appends one log entry naming the chosen artboard, refuses a
  link outside `odd/tasks/`, and writes no document for untracked work. The skill loads
  `ui-craft`. Review 4 of 4, no findings.
- **v0.15.1, PR #334**: 14 signed assets, installed; `odachi --version` answers
  `odachi 0.15.1`, and `odachi init` refreshed `~/.agents/skills` (the design skill
  carries the new rules). The OpenCode push plugin is installed at
  `~/.config/opencode/plugins/odachi-design/index.ts` and was verified live.

## What the new project exposed

An agent creating a new project with odachi did not know what PDD is, did not follow
the full workflow, and never created `proof/`. A read-only audit of the installed
door, `bootstrap`, `drive`, `doctor` and the block found nine candidates, not one
discipline failure:

1. `bootstrap` says `init` leaves the skills in `.agents/skills`; in a home-scope
   install (this machine) they are in `~/.agents/skills`. The agent looks in a folder
   that does not exist.
2. `bootstrap` is a setup checklist (`init`, `start`, `check`), not the first cycle; it
   does not say to load `drive`, plan a step, work, `check --run`, commit.
3. PDD (Proof-Driven Development, ADR 0031) is named only in the repo docs; the block,
   `bootstrap` and `drive` never name it or define the cycle.
4. `proof/` is created by `check --run`, not by `init`; the bootstrap does not say the
   setup is not done until the first proof, so a project can look ready with no folder.
5. `odachi start` from an agent shell is non-interactive: without `--outcome`,
   `--audience`, `--done`... it exits 2 naming the flags. The bootstrap says run start
   and ask the human, not how to pass the answers from a tool shell.
6. `odachi init` refuses a folder without `.git`; the bootstrap omits `git init`.
7. `doctor`'s preflight names `init` and `start` only; not `drive`, `check --run` or
   "no proof yet".
8. No `doctor` check compares the installed home copies (door, skills, plugin) with the
   binary; a stale copy teaches old paths (candidate 1 is the example).
9. Two product defects surfaced in the same run: `check --run` accepts an empty proof
   and says `ready`; on Windows a `.ps1` command goes through `cmd /c` and is not run
   by PowerShell.

## The plan

1. **Measure first**: a fresh folder with the installed global agent, asking for a small
   app. Record where it stops: bootstrap loaded? `git init`? `start` with flags? `drive`?
   `check --run`? `proof/`? This separates discipline from missing instructions.
2. **Fix the content**: PDD in the always-read block (`rules.go`) with its pinned tests;
   `bootstrap` as the first-cycle checklist with the home paths, the non-interactive
   `start` and `git init`; `doctor` naming the next step and auditing the home copies.
3. **Fix the two defects**: a warning (not a block) when a proof has no output, and a
   `.ps1` command run with PowerShell (or refused naming the invocation).
4. **Pin the path**: a `bench` journey for the new-project flow, so the next regression
   is caught by the tool instead of by a person.

Nothing else is queued. The feature order is the human's call; the measurement comes
first, because it tells whether the fix is text, loading or discipline.
