# 0023 - One workflow, one truth about a tree

Status: accepted · Date: 2026-10 · Session: after v0.11.9

## Context

Two workflows could judge the same tree. `ci.yml` ran the suite on three
operating systems, plus `vet`, `gofmt`, the race detector, `build` and the
goreleaser check. `gate.yml` ran the suite again on ubuntu and called it the
gate. On a pull request the ubuntu suite ran twice, and a person reading the
checks had to decide which of the two greens was the real one. When they
disagreed, the pull request had two answers and no rule for choosing.

The gate also read the fence from the base ref but compared the changed paths to
its own `HEAD`. A checkout is clean, so the changed-path set was empty and the
fence was applied to nothing: the policy came from the base, and the change never
met it. On a push to `main`, `origin/main` after checkout is the pushed commit
itself, so the "base" was the candidate.

## Decision

- **One workflow judges a tree.** `gate.yml` holds the gate and the static
  checks. `ci.yml` is deleted. A tag still runs `go test ./...` in `release.yml`
  before publishing, so a release is never built from an unproven tree.
- **The gate runs on ubuntu, macos and windows**, one job each, and each keeps
  its own proof artifact (`proof-<os>`). A green ubuntu says nothing about
  Windows; a proof is evidence for the tree on that runner, not for the project
  in general.
- **The base is the ref the change cannot edit.** On a pull request it is
  `github.event.pull_request.base.sha`; on a push it is `github.event.before`,
  the previous tip; a zero or empty SHA falls back to `origin/main`.
- **`--base` goes with `--fence-from`.** `--fence-from` says where the policy is
  read; `--base` says what the policy is applied to. The gate passes both, so the
  fence meets the paths that differ from the base.
- **The proof is kept for a green run.** odachi writes the proof only when the
  command exits 0, so a red job has no artifact and is read from the log. The
  workflow says that instead of claiming an upload that cannot happen.
- **The static checks are jobs of the same workflow.** `gofmt`, `go vet`,
  `go test -race ./internal/...` and `goreleaser check` run once on Linux. The
  race detector runs the same tests with a different instrument: it is a second
  measurement, not a second opinion.

## Alternatives rejected

- **Keep both workflows and require both checks.** Two definitions of green is
  the defect, not the redundancy: the first one people learn to ignore wins.
- **Let `gate.yml` do the fence on ubuntu and `ci.yml` the matrix.** The fence
  would be applied on one OS while the suite breaks on another, and a required
  check would still be two workflows.
- **Use `origin/main` as the base on a push.** After checkout it is the pushed
  commit, so the diff is empty and the gate approves whatever was just pushed.
  `github.event.before` is the previous tip, which the change cannot edit.
- **Keep the `build` job.** `go vet` and `go test` already compile every package;
  a third compile asks no new question.
- **Fabricate a proof on red.** A proof is the real output of a green run; an
  artifact that says "failed" is a log with extra ceremony.

## Consequences

- One file defines green, and a pull request runs the suite once per OS. Removing
  the duplicate ubuntu run pays for the two extra gate builds.
- A direct push to `main` is now compared against the previous tip, so the push
  path meets the fence before the ruleset exists.
- A red run leaves no proof artifact. If an artifact on red is wanted later,
  odachi must write the proof before exiting nonzero; that is a product change,
  not a workflow change.
- The gate is still advisory until the repository ruleset marks it required, and
  a pull request can edit the workflow file it runs under. `CODEOWNERS` asks for
  an owner review of `.github/` and `.odachi/`; the ruleset is the next feature.
- The matrix triples the gate minutes on a push to `main`, which is the price of
  the three-OS promise; caching was left out of this feature on purpose.

## Sources

- Feature 0024 and ADR 0022: the fence read from the base ref, the trusted
  binary, and the rollout of the `pull_request` trigger.
- The review of feature 0026: the first tree read the fence from the base and
  compared to `HEAD`; the probe showed `ready` where it had to say
  `out of scope`.
