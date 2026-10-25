# 0026 - A stranger can contribute

Status: accepted · Date: 2026-10 · Session: after v0.11.9

## Context

The repository is public, `main` only takes a pull request whose gate passed, and
the README explains the tool at length. What it did not have was the path for
someone who is not the maintainer: no CONTRIBUTING, no issue forms, no pull
request template, no way to report a vulnerability privately, and no statement
about the AI that writes most of the code. A person arriving from a search
result could read for an hour and still not know whether their change would be
accepted, or how to prove it.

Two facts shape the answer. The maintainer is one person, so anything that
assumes a team (a CLA, a review rotation, a governance document) would be a
promise nobody keeps. And the brief and the fence are the maintainer's
artifacts: a contributor who could rewrite them could rewrite the rules of the
change they are asking to pass.

## Decision

- **Three short files, one path.** `CONTRIBUTING.md` (the loop, the gate, the
  commit shape, the size), `AI_POLICY.md` (disclosure and ownership),
  `SECURITY.md` (a private channel). The README links them in one place, and the
  link is tested.
- **Features start as an issue; fixes can start as a pull request.** The
  maintainer writes `brief.md` and `.odachi/fence.toml` for a feature. A
  contributor never edits those two. CI reads the fence from the base ref, so a
  branch's copy does not change the policy it is judged by; the file itself is
  protected by the CODEOWNERS request for an owner review and the maintainer's
  merge decision, not by the gate. That limit is recorded in ADR 0024.
- **The gate is the contract.** A pull request is a candidate: the workflow
  builds the gate from the base ref, reads the fence from it and runs the suite.
  The local proof is generated and ignored by git, so it does not travel; the CI
  run is the evidence a reviewer clicks. The workflow is the known limit: it
  travels with the pull request, and ADR 0024 records why that is accepted.
- **The template asks for the three things a reviewer cannot guess:** the proof,
  the risk and the rollback. It also asks for the AI disclosure, because the
  project is built that way and hiding it would be the only dishonest answer.
- **Security goes through GitHub advisories.** No email, no public issue, and a
  sentence for the person who cannot use advisories.
- **The surface is guarded by a test.** `internal/archtest/contributor_test.go`
  fails if a file disappears, if the pull request template stops asking for the
  proof, the risk or the rollback, if `SECURITY.md` loses the advisories link, or
  if the README stops linking the three files. The assertions are few on
  purpose: one per promise.
- **No CLA and no code of conduct.** MIT is the license; a CLA would add friction
  to a project with one maintainer, and a code of conduct is a community
  decision, not a repository decoration. Both can arrive later, as their own
  decision.

## Alternatives rejected

- **A wiki or a website.** Not versioned with the code, not reviewed in the same
  pull request, and it drifts on the first change that forgets it.
- **One big CONTRIBUTING with the whole odachi workflow.** Nobody reads a
  hundred lines to fix a typo; the short loop plus links to the skills is what
  gets followed.
- **Requiring a task doc for every contribution.** That is the maintainer's plan,
  not the contributor's; a bug fix with a test and a proof does not need a
  feature brief.
- **A CLA or a DCO bot.** Friction without a legal need at this size.
- **An issue bot that labels and triages.** There is no volume to triage yet;
  the forms do the work that matters.

## Consequences

- The maintainer owes an answer on every issue and a review on every pull
  request; the surface makes that promise visible.
- A docs-only change still goes through the gate and the ruleset, which is the
  price of one path for every change.
- The guard test is text matching, and text matching is brittle by design: it
  fails when a promise changes, and the fix is to update the test in the same
  pull request that changes the promise.
- AI disclosure is on the honor system. The proof binds to a tree but not to a
  provenance: the tool refuses a proof that names another tree, and CI runs the
  command again, so the evidence a merge rests on is the run. The policy says
  that instead of claiming the tool can tell a hand-written output from a real
  one.
- The fence file and the workflow are not enforced by the gate: they are
  protected by a CODEOWNERS request and a zero-approval ruleset, which asks
  rather than requires. A second reviewer or a review app is the next step, not
  a promise of this change.
- A code of conduct, a roadmap and funding stay out until someone asks for them.

## Sources

- The ruleset and the gate: ADR 0024 and ADR 0023.
- gentle-ai's contributor surface (`CONTRIBUTING.md`, `AI_POLICY.md`,
  `SECURITY.md`, `CONTRIBUTORS.md`) as the model of what a stranger needs first.
