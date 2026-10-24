# 0022 — The gate runs from the base ref

Status: accepted · Date: 2026-10 · Session: v0.11.9

## Context

Odachi's gate was local and honest, and that is also its limit in a team: the
change can edit the fence it must obey, and the hook can be skipped. A public
repository or a company needs the gate to run where the candidate cannot rewrite
it, and evidence that survives the runner.

gentle-ai, a mature public repo, solves one half of this with a trusted policy
source: `pr-size-policy.yml` checks out the default branch, reads the live PR
facts from the API, and never executes candidate code. Its own audit also shows
the other half is easy to get wrong: issue and approval checks read scripts from
the PR head, so a motivated PR can weaken its gate.

## Decision

- **`check --fence-from <ref>` reads the fence from a git ref, not the worktree.**
  CI passes the base branch; the candidate's copy is not consulted. A missing
  file at a valid ref is "this revision has no fence"; a ref that does not exist
  blocks, because a typo must never turn the gate off.
- **The proof names the ref.** `# fence origin/main` sits between the command and
  the output. Old proofs without it still parse: they used the worktree's fence.
- **The CI gate is built from the base ref.** The workflow builds odachi from
  `origin/main` and runs it against the candidate worktree. Gate code shipped by
  the candidate is advisory; code the candidate cannot edit is the gate.
- **The proof is uploaded even when the gate fails.** A failed run is evidence
  too, and a green check with no artifact is a claim without a receipt.
- **`ODACHI_NO_SETUP=1` means binary only.** Both installers skip `setup`;
  `setup` and `init --global` refuse with the variable named. A CI job must not
  write a home door or a default agent on a shared runner.
- **Local behavior without the flags is unchanged.** `check` still reads the
  worktree fence by default; `init` still leaves the home alone without
  `--global`.

The workflow in this repository runs on `main` and by hand for now. The
`pull_request` trigger becomes a required check after `main` carries
`--fence-from`: on the change that adds the flag, the base ref would build a
binary without it and the job would fail for the wrong reason.

## Alternatives rejected

- **Overwrite the worktree fence with the base fence before running.** It
  changes the tree hash the proof certifies, so the evidence would describe a
  tree that is not the candidate.
- **Run the candidate's `odachi`.** That is the hole gentle-ai's audit found: the
  gate's code and its policy move with the change.
- **Pin a released binary in CI.** It adds a release dependency and an ordering
  trap (the workflow cannot exist before the release that supports it). Building
  from the base ref is the same trust with no release coupling.
- **Make the PR check required immediately.** The first PR would fail for a
  missing flag, not for a defect. The rollout is explicit in the workflow.

## Consequences

- A PR cannot loosen the fence in the same change that asks to pass it; the
  proof says which ref it obeyed.
- CI gets a proof artifact tied to the exact tree, with the command that produced
  it, even on failure.
- A shared runner is not mutated: no home door, no default agent, no config.
- The local hook stays advisory; the CI job is the authority once it is required.
- Identity (author vs reviewer), org policy beyond the fence, monorepo scopes and
  signing/provenance remain open, in that order.

## Sources

- gentle-ai `pr-size-policy.yml` (trusted policy from the default branch) and
  `docs/audits/2026-07-27-post-ratchet-audit.md` (the gate code read from the PR
  head).
- Feature 0023 (installation is the setup) for the installer path and the
  `ODACHI_NO_SETUP` switch.
