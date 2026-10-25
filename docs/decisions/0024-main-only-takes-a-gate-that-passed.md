# 0024 - main only takes a gate that passed

Status: accepted · Date: 2026-10 · Session: after v0.11.9

## Context

The gate ran in CI, but it was advice. A change could push to `main` directly
and never meet the fence, the suite or the proof, and a red pull request could be
merged by the person who opened it. Feature 0024 in `odd/tasks` gave the gate a
pull request trigger; nothing required the result.

The repository has one maintainer. That single fact rules out the obvious
configuration: GitHub does not let you approve your own pull request, so
"require one approval" would deadlock every change.

## Decision

- **A ruleset on the default branch.** Pull request required, the four check runs
  required (`gate (ubuntu-latest)`, `gate (macos-latest)`,
  `gate (windows-latest)`, `checks (linux)`), no force push, no deletion.
- **Zero approving reviews.** A solo maintainer cannot approve their own pull
  request; requiring one would block the only person who can merge. The gate, not
  a second person, is what protects `main` for now.
- **The policy is a file, not a click.** `.github/main-ruleset.json` holds the
  rules; `scripts/protect-main.ps1` applies it by name, so a second run updates
  the same ruleset instead of adding another. A settings change that is not in
  the repository is a settings change nobody can review.
- **The check names are the job names.** Renaming a job in `gate.yml` without
  renaming it in the ruleset leaves a required check that never reports, and
  every pull request stays blocked. Both files say so.
- **No bypass list.** An admin bypass would turn the protection back into advice.

## Alternatives rejected

- **Legacy branch protection.** The same rules through an older API, with no
  single document to review and no dry name for the policy.
- **One approving review.** Deadlocks a solo repository; it becomes real the day
  a second maintainer or a review app exists.
- **Require branches to be up to date (`strict`).** More rebases for one writer,
  and it does not close the hole below.
- **A GitHub App that approves.** The right answer for enforced review without a
  human; out of scope until there is a published app to point at.

## What still does not hold

The workflow file travels with the pull request: GitHub runs the version from the
head, so a change can weaken the jobs that judge it. `.github/CODEOWNERS` asks
for an owner review of `.github/` and `.odachi/`, but with zero required
approvals that request is not enforced. Closing it needs one of: a second
reviewer, a review app, or a ruleset with a bypass list used deliberately by the
owner. Until then, the protection is against mistakes and haste, not against the
owner or a hostile contributor with push access.

## Consequences

- Every change lands through a pull request, including the ones that change the
  workflow: the feature branch is the candidate, and the gate runs on it.
- A red job on one OS blocks the merge. That is the point: the three-OS promise
  is now a condition, not a report.
- A direct push to `main` is rejected by the remote; the local `main` has to be
  synced after a merge instead of being pushed.
- GitHub's default extra approval for changes from people without write access is
  turned off. Without that, an outside contribution could never merge in a
  one-person repository; the gate, not an approval, is the requirement.
- Renaming a job in `gate.yml` breaks every pull request until
  `.github/main-ruleset.json` follows. The script prints the required checks
  after applying them, so the drift is one run away from being seen.
- The owner can still edit the ruleset. The protection is a habit with a
  receipt, and the receipt is the JSON file.

## Sources

- Feature 0026 and ADR 0023: the single workflow, the base ref and the four jobs.
- GitHub rulesets API (`POST/PUT /repos/{owner}/{repo}/rulesets`), and the check
  run names the API reports for a merged commit.
