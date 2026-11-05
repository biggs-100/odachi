# Next session

Written at the close of 0204, `the-release-publishes`. The installer the README
documents is now an asset of every release; v0.16.2 is the release that carries
the fix.

## What shipped

- **0204, the release publishes the installer.** `.goreleaser.yaml` carries
  `scripts/install.ps1` and `scripts/install.sh` as release extra files, and
  `TestTheReleasePublishesTheInstallers` pins the config. v0.16.1 and v0.15.1
  never carried them, so the documented command returned 404; v0.16.2 does.
- **0203, the first cycle.** The `AGENTS.md` block names Proof-Driven Development
  and the first proof; the `bootstrap` skill is the first-cycle checklist: `init`
  creates the repository, `start` asks the human and passes their words, `drive`
  plans one step and `check --run` closes with the first proof. The doctor names
  that cycle in an empty folder, compares the installed skills with the binary
  when the project has no `skills/`, and audits the design plugin beside the
  door. `check --run` warns when a proof has no output; a `.ps1` command is
  refused with the PowerShell invocation; `review` reads a repository with no
  commits as an empty history. The bench has the new-project journey.
- **#338, the installer test.** The PowerShell installer test snapshots the user
  PATH and restores it.
- **The measurement** is in `.git/odachi/scratch/probe-new-project.jsonl` and
  `probe-new-project-2.jsonl`: a fresh folder, the installed 0.15.1 agent, and a
  small expense app.

## What the measurement changed

The fresh agent completed the full cycle and created `proof/`; `odachi init`
already creates the repository, so the audit's `git init` fix was unnecessary.
The real gap was the invented interview answers in a non-interactive run: the
harness cancelled the question and the agent assumed. The bootstrap now says
never invent the vision.

## The decision that waits

- **0205, the skill contract.** The references system (a skill's `references/`
  installed with it), the skill grammar (`Decision gates`, `Output`,
  `References`), `ui-craft` as the first router with four references, and
  `design` with its output contract. The brief draft is the next session's first
  move.
- The two WARNING rows in `docs/review-debt.md` remain open: the `Messages` guard
  reads every `Head` failure as an empty history, and the doctor's default note
  returns before the design plugin note. The next feature that touches those
  paths pays them.
- Nothing else is queued: after 0205, the other seventeen skills and the
  routing/measurement are the follow-up.
