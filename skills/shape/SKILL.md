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
- Ask the human the question the exploration opened, not the seven questions of
  the interview. One question per real fork; what cannot be answered now is
  recorded with `odachi ask`.
- Do not write the brief for them. Draft it and hand it over: the human owns
  `brief.md` and the agent proposes.
- End in `odachi start --brief <path>`. Exploration is not a verb and writes no
  artifact; the first write records the brief, the fence and the task doc
  together, and that is the moment the feature exists.
- The fence is not shaped by convenience. What exploration found useful but out
  of scope stays out and goes to the "Not included" list.
- If exploring does not end, stop and say what you learned: an exploration
  without an end is a feature nobody planned.

## Never

- Write code, a test or a plan step before `start` recorded the brief.
- Answer your own question in the brief and call it the human's decision.
- Turn exploration into a phase, a verb or a document: the artifact is the brief.
- Declare a feature whose outcome you still cannot observe.
