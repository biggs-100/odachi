# Next session

Written at the close of 0203, `the-first-cycle`, on `feat/first-cycle`. The
feature is stacked on #338, the light installer-test fix; both wait for their
pull requests.

## What shipped

- **0203, the first cycle.** The `AGENTS.md` block names Proof-Driven Development
  and the first proof; the `bootstrap` skill is the first-cycle checklist: `init`
  creates the repository, `start` asks the human and passes their words, `drive`
  plans one step and `check --run` closes with the first proof. The doctor names
  that cycle in an empty folder, compares the installed skills with the binary
  when the project has no `skills/`, and audits the design plugin beside the
  door. `check --run` warns when a proof has no output instead of calling it
  evidence; a `.ps1` command is refused with the PowerShell invocation that runs
  it; `review` reads a repository with no commits as an empty history instead of
  a fatal git error. The bench has a journey from an empty folder to the first
  proof and uses `echo ok`, which both shells run, instead of `true`.
- **#338, the installer test.** The PowerShell installer test snapshots the user
  PATH and restores it, so a suite run no longer leaves a dead temp entry behind.
- **The measurement** is in `.git/odachi/scratch/probe-new-project.jsonl` and
  `probe-new-project-2.jsonl`: a fresh folder, the installed 0.15.1 agent, and a
  small expense app.

## What the measurement changed

The fresh agent completed the full cycle and created `proof/`; `odachi init`
already creates the repository, so the audit's `git init` fix was unnecessary.
The real gap was the invented interview answers in a non-interactive run: the
harness cancelled the question and the agent assumed. The bootstrap now says
never invent the vision; the rest of the fixes close the path around it.

## The decision that waits

- The two pull requests (#338 first, then the stacked #339) and the review of
  0203. The release renames `## Unreleased`.
- Nothing else is queued: the next feature is the human's call.
