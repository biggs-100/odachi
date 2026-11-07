---
name: shape
description: Trigger: when a request does not fit in one reading, or before the brief of a feature whose edges are not decided. Explore first, then declare.
---

## When

A request arrives and you cannot say, in one sentence, what has to exist, what it
must not touch and how the human will know it worked. Or the human asks for
something whose edges they have not decided yet. Both are the same moment: the
brief would be written with the least information the feature will ever have.

Not for a request already shaped by a brief, a one-line fix, or work whose fence
is obvious: there `drive` is enough.

## Rules

- Explore before you declare. Read the code the request names, run the command
  that fails, look at the data. Three reads that end in "now I know what to
  build" are cheaper than a brief that has to be rewritten.
- **The smallest change that satisfies the requirement.** Before you declare,
  name what the change reuses, what it adds (files, dependencies, abstractions)
  and what it deletes. A file or an abstraction needs a reason; an abstraction
  needs a second caller, not a guess about the future.
- **Wait for three.** Copy until the same pattern appears three times for the
  same reason -one decision would change every copy at once- and extract then. A
  wrong abstraction costs more than the duplication it replaced; if it is already
  wrong, inline it back first.
- **The props test.** If the props describe the situation (mode, variant,
  isSpecial), you merged similar things; if they describe what is rendered, the
  abstraction is probably sound.
- Ask the human the question the exploration opened, not the seven questions of
  the interview. One question per real fork; what cannot be answered now is
  recorded with `odachi ask`.
- Do not write the brief for them. Draft it and hand it over: the human owns
  `brief.md` and the agent proposes.
- End in `odachi start --brief <path>`. The draft may live anywhere; the run
  records it as `brief.md`, together with the fence and the task doc, and that is
  the moment the feature exists. Exploration is not a verb and writes no artifact.
- The fence is not shaped by convenience. What exploration found useful but out
  of scope stays out and goes to the "Not included" list.
- If exploring does not end, stop and say what you learned: an exploration
  without an end is a feature nobody planned.

## Output

The draft it hands over: the outcome in one sentence, the observable done-when, the non-goals with the path each one protects, and the decisions taken with their reason. The questions exploration opened that nobody answered, recorded with `odachi ask`, and the `start` command that records the brief.

## Never

- Write code, a test or a plan step before `start` recorded the brief.
- Answer your own question in the brief and call it the human's decision.
- Turn exploration into a phase, a verb or a document: the artifact is the brief.
- Declare a feature whose outcome you still cannot observe.
