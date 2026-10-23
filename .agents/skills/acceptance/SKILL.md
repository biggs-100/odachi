---
name: acceptance
description: Trigger: when deciding whether something is done. Turn a request into observable criteria before writing code.
---

## When

When someone says "it's done", "almost there" or "can we call it done?" and
there is no list of criteria anyone can verify without asking you.

## Rules

- Before writing code, turn the request into observable criteria. If it cannot be
  observed from outside, it is not a criterion: it is an intention.
- One line per criterion, in the human's words, under "Done when". If it does not
  fit in one line, the criterion is doing two things.
- One criterion per behavior, not per file or layer. "I can book an appointment
  and it shows in the agenda" works; "the appointment module works" does not.
- Every criterion is verified with a command or real output, and that output is
  saved. Testing by hand once is not verification.
- If a criterion cannot be verified yet, say so and propose how. A hidden
  criterion is worse than a pending one.
- Criteria are written before implementation and are not changed afterwards to
  make the implementation pass.

## Never

- Accept "it works" or "I think it works" as done.
- Add criteria the human did not ask for to justify more work.
- Mark a step done without the output that proves it.
- Change the criterion to match what was already built.
