# Security policy

## Reporting a vulnerability

Report it privately through GitHub Security Advisories:

<https://github.com/biggs-100/odachi/security/advisories/new>

Do not open a public issue, and do not paste a proof of concept in a pull
request. If GitHub advisories are not available to you, open an issue that says
only "I need a private channel for a security report" and nothing else; the
maintainer will answer with one.

## What is in scope

- **The binary and its artifacts:** `brief.md`, the task doc, `proof/`, the
  fence, `.odachi/reviews/`, and the ruleset script.
- **The gate's trust model:** the workflow reads the fence from the base ref and
  builds the gate from it, `main` only takes a pull request whose four checks
  passed, and a change has to meet the policy of the base branch. The workflow
  itself is the known limit below: a pull request can edit the file that asks
  for all of that.
- **The workflows** in `.github/workflows` and the ruleset in
  `.github/main-ruleset.json`.

The interesting class is a change that gets a green gate it did not earn: a
proof that names another tree, a fence that is obeyed from the worktree, a check
that reports without running, a rule that looks enforced and is not. That is the
defect this project exists to make impossible, with one known exception below.

## What a release proves

The tag workflow signs `checksums.txt` without a key: cosign asks Fulcio for a
short-lived certificate that names this workflow and the tag, and the release
publishes the bundle that holds the signature and that certificate. The README
has the command to verify it.

- The signature proves the artifact came from this repository's release
  workflow and from that tag, not from someone who can edit a release.
- The provenance attestation names the repository, the commit and the workflow
  that produced each archive, and `gh attestation verify` reads it. Together
  with the signature it answers "which source built this file".
- Neither proves the source was good: a signed and attested build of a
  compromised tree is still signed and attested. What it rules out is a binary
  from somewhere else wearing this release's name.
- Each archive ships an SPDX SBOM beside it, so what is inside can be read
  without asking the maintainer.
- The installer verifies `sha256` only. Requiring cosign to install would turn
  verification into a barrier, and a barrier is skipped: the signature is there
  for whoever needs it, not as a condition to use the tool.

## The known limit

The workflow file travels with the pull request, and the ruleset requires zero
approvals, because a solo maintainer cannot approve their own pull request
(ADR 0024). GitHub's default extra approval for changes from people without
write access is turned off for the same reason: without that, a stranger's pull
request could not merge in a one-person repository. So a pull request that edits
`.github/workflows/` can weaken the jobs that judge it. `.github/CODEOWNERS` asks
for an owner review, and it is a request, not an enforcement. That is the known
state of a one-person project,
not a vulnerability; a report about it is welcome as a design discussion. What is
in scope is a path that reports green *without* editing the workflow.

## What is not in scope

- The third-party agents and models that run odachi. Report those to their
  vendor; a model answering badly is not a vulnerability of this tool.
- A repository that removed the ruleset, edited `CODEOWNERS` or no longer runs
  the gate: the protection is the configuration, and a fork can drop it.
- The content of a project's own brief, fence, task doc or proofs.
- Denial of service by running an expensive command in `--run`: the command is
  the user's.

## What to expect

An acknowledgement as soon as the maintainer reads it, and a target of seven
days for a first answer. A fix ships with a test that fails without it, and the
advisory names the versions affected. Credit is given unless you ask otherwise.
