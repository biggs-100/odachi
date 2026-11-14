---
name: measure
description: Trigger: when a number is about to be taken (a probe, a benchmark, a cost, an A/B run), or when a document claims one. Write the criterion before the run.
---

## When

When you are about to take a number the project will act on, when a document
claims one, or when someone asks "did that help?".

## Rules

- Write the criterion before the run: the sentence that decides pass or fail has
  to be readable before the result arrives.
- Say what the instrument measures: bytes are not tokens, and
  instruction accounting is not a cost measurement.
- Name the exact command and the tree it ran on: a number with no command cannot
  be re-run, and a number nobody can re-run cannot be disputed.
- One row per run in the verdict table: the run, the number, the criterion, and
  the verdict it earned.
- Name every excluded run with its reason: the one you dropped is the one that
  would have changed the answer.
- Write what the data does not show: the arm nobody measured, the sample, and the
  question the number cannot answer.
- A document that claims a number quotes the tool that counts it, or it is a
  sentence: `docs/next-session.md` carries the totals line `odachi summary`
  prints, and `odachi doctor` compares it.
- A number measured once is not a trend: say the n, and say which draw it was.
- Write what would have changed your mind. A measurement with no decision
  attached is a story with a table.

## The document

The record lives in `docs/`, named by the question, and carries the five
sections in this order:

1. **The question and the criterion** - what is measured, and the pass/fail line
   decided before the run.
2. **The instrument** - the exact command, its version, and what it actually
   measures.
3. **The verdict table** - one row per run, with the number and the verdict it
   earned against the criterion.
4. **The excluded runs** - each run that did not count, and why.
5. **What the data does not show** - the limits, the sample, and the question
   left open.

## Output

The measurement document with its five sections, the verdict table that decides
it, and the commands a person re-runs to disagree. A number nobody can re-run is
returned as a sentence with no record.

## Never

- Take the number first and write the criterion after: that is a result looking
  for a rule.
- Report a byte count, a transcript or an instruction count as a cost.
- Drop a run without saying which one and why.
- Claim a total in a document the tool can compute.
