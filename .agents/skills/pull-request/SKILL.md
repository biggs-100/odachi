---
name: pull-request
description: Trigger: when a change goes to a PR or the diff passes 400 lines. Split the delivery into reviewable work units and pick one chain strategy.
---

## When

When the change is ready for review, when a PR passes 400 changed lines, when
several PRs have to be stacked, or when the reviewer cannot understand the diff
in one sitting.

## Rules

- A PR is **one work unit**: a behavior, a fix or a migration that can be
  reverted alone. If the title needs "and", they are two PRs.
- Budget: **400 changed lines** (additions + deletions). If you pass it, split
  before opening; do not ask for an exception without saying why.
- Pick **one** strategy and do not mix them: stacked to main if each slice can
  land alone; feature branch chain with a tracker PR if the feature has to
  integrate before main.
- Every PR says what problem it solves, how it is tested, what was left out and
  which PR it depends on. The first one in a chain depends on nobody.
- Tests and docs travel in the PR of the behavior they verify; there is no later
  "tests" or "docs" PR.
- Required checks are read from the repo, not guessed: if you do not know whether
  a check is required, the PR is not ready.
- If the repo uses issues, reference the approved issue with
  `Closes/Fixes/Resolves #N`; otherwise use `Refs #N`. Do not invent an approval.
- `commits` decides the message; this skill decides the cut. The PR reads in order.
- Close with `odachi check --run "<command>"` green and the saved output: the PR
  does not replace the tree's proof.

## Never

- Open an 800-line PR "so the work is not split".
- Mix stacked and feature branch chain in the same delivery.
- A PR that does not say what was left out.
- Ask for review without the required checks green.
- Rewrite a protected branch or force-push to main.
- Invent an issue or an approval to satisfy issue-first.
