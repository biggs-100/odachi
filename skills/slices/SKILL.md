---
name: slices
description: Trigger: when a large request has to be split into small steps, when the code no longer fits in one read, or when deciding whether to delegate to a subagent.
---

## When

When the request does not fit in one commit, when you have to say "where do I
start", when the diff passed 400 lines and nobody can review it seriously, or
when understanding what to touch took more than three reads.

## Rules

- Each step leaves the repository green and is committed alone. If the step needs
  "and", they are two steps.
- Order by risk, not by comfort: first what can bring the design down (the
  contract, the data), then what is built on top.
- Each step cites the requirement it satisfies (R1, R2). A step that cites none
  is extra work or reveals a requirement that is missing.
- The first step is the one that produces evidence fastest: the one that lets you
  see something working before investing in the rest.
- The commit says what changed for the person using the project, in one line. If
  the "how" needs explaining, it goes in the body, not the title.
- A step includes its test and its documentation. There is no later step called
  "tests" or "docs".

## Output

The steps in the order they run, each with the requirement it satisfies and the step that produces evidence first; the reason for the order and who runs each step. A diff that passed 400 lines returns the split that makes it reviewable.

## References

- `references/delegation.md` — when a step goes to a subagent: the budget, the request it carries, the bound it returns and what the parent keeps.

## Never

- Commit with red tests "so the work is not lost".
- Mix a refactor with a behavior change in the same step.
- Split by files or layers instead of by deliverable behavior.
- Leave cleanup steps without a done criterion: they grow and never close.
- Delegate for size or file count, or to save tokens on something you do in one
  turn.
