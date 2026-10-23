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

## If you delegate

Inside an odachi project, this is the delegation policy: if another global skill
says otherwise, this one wins.

A subagent starts without cache: **it does not save total tokens, it protects
your context.**

- Budget: if understanding took more than ~3 reads or ~10k tokens, or the session
  is already long, the next unit goes to a subagent. Delegate for a reason, never
  for size.
- What is **read a lot** gets delegated; what is **written little** does not.
- A **writer** is delegated for one reason only: two or more independent units
  with disjoint surfaces, each heavier than starting a subagent; or ~150k tokens
  of session. Never for size, files or price; without a reason, write it yourself.
- One writer per unit; in parallel they do not share files and you own git and the
  final check. If you delegate the writing, delegate the reading that prepares it.
- The request carries the **task doc path and the step**, not a summary; also the
  **edit surfaces**, the **verification commands** and the **exact skill paths**
  (`## Skills to load before work`).
- To write, use **`odachi-worker`**; to verify, **`odachi-verify`** (read-only).
  The worker does not commit or delegate; the parent keeps git, the gate and the
  vision.
- The return is bounded: **=2k tokens**, `path:line` evidence and the list of files
  read; the rest stays inside.
- Ask it to record with `odachi plan` and not to say "done" without
  `odachi check --run`.
- When it returns, verify its output against `brief.md` and the diff, **not
  against its summary**: the subagent can lie like the main agent.
- **Never delegate the vision decision.** The subagent executes; the human decides.

## Never

- Commit with red tests "so the work is not lost".
- Mix a refactor with a behavior change in the same step.
- Split by files or layers instead of by deliverable behavior.
- Leave cleanup steps without a done criterion: they grow and never close.
- Delegate for size or file count, or to save tokens on something you do in one
  turn.
