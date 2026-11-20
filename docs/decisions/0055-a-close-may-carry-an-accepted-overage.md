# 0055 - A close may carry an overage the human accepts, and never one it cannot read

Status: accepted - Date: 2026-11 - Session: after v0.57.0

## Context

0339 closed at 438 authored lines -59 of them the decision record, 221 the tests the
method demands- and the note it left behind asked the wrong question: it blamed
documentation and skills for a limit that had become unsatisfiable. The measurement
says otherwise. Docs and skills were 63 lines; the tests were half the slice.

Asked the same question, gentle-shell chooses differently and had already written it
down. Its budget is a review-load budget, its counting excludes only lock files and
minified assets, and its skills forbid the one escape this project was drifting
toward: "never shrink a diff by deleting comments, blank lines, docs, or tests, or by
compressing code, to fit the review budget". Its two answers to an overage are split
the work or record the overage with a rationale.

odachi's limit is narrower than gentle-shell's number: it exists only for a close that
owes an independent verifier (`gate.SliceLines`, one verifier cannot finish the tree),
and it already refuses in exactly one place. What it lacked is the second answer
gentle-shell has: an overage the human deliberately accepts, in writing, with a reason.

## Decision

- **The count does not change.** Docs, skills and tests are authored lines, because the
  verifier reads them and because an exclusion is the code-golf the method forbids.
- **The acceptance lives in the task document**, in the preamble where `Base:` and the
  waivers already live: `Slice: <lines> - <reason>`. Both halves are required -a number
  without a reason is a decision nobody can weigh- and a line inside a section is prose,
  not a decision.
- **The accepted number must reach the measured one.** Accepting 420 while the tree
  authors 438 accepts nothing, and the refusal names both numbers so a reader sees
  which one moved.
- **An accepted overage settles the size, not the record.** The close answers with
  `verification_missing` and names the accepted number and its reason; the verifier is
  still owed, because who writes does not verify itself.
- **The refusal teaches the door**: it prints the exact shape to write in the task doc,
  so nobody shrinks a diff to get past it.

## What it costs

The line is human-written, so the tool cannot prove the reason is good -it can only
prove that a reason exists and that the number covers the tree. A feature whose
human accepts everything turns the limit into a formality; what keeps that from being
silent is that the reason is printed on every close that carries it, next to the number
the tool measured.

## Alternatives

Raising the limit was rejected because the number is not the problem: the same
discussion would return with the next large feature and a new number chosen by feel.
Excluding `docs/` and `skills/` was rejected because it would have moved 0339 from 438
to 375 while hiding the half of the slice that is tests, and because gentle-shell,
asked first, counts them on purpose.
