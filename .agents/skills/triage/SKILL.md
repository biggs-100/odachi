---
name: triage
description: Trigger: when there are several bugs, a backlog or a report that repeats. Classify by root and fix once, not once per issue.
---

## When

When several bugs arrive, a report repeats, the backlog grows faster than the
fixes, or you do not know whether an issue is new, a duplicate or a symptom of
something else.

## Rules

- Before touching code, classify each report by **root**, not by surface:
  (A) an in-flight change already solves it, (B) duplicate of a known class,
  (C) new bug with an identifiable root, (D) feature, (E) unclear and goes to
  `odachi ask`.
- Two or more reports with the same root are **one fix**: one `odachi plan` step,
  every report named, and one test that closes them together.
- Reproduce the symptom before accepting the diagnosis. The mechanism the report
  names is a hypothesis; the symptom is the evidence. Load `tests` and write the
  failing test.
- The fix goes where the cause is, not where it hurts. If the fix adds a state, a
  flag or a verb, ask what you can delete or relax first.
- A report that cannot be reproduced becomes a message that diagnoses itself or a
  question with the exact input; not a blind patch.
- Close with the test that reproduces the report and turns green. Without that
  test, the report stays open, even if the code "should work now".
- If the report is outside the fence in `brief.md`, do not fix it: `odachi ask`.
- If the fix touches several roots, one step per root; not one step per report.

## Never

- One patch per report when they share a root.
- Close because "it should be fixed" or because of the agent's summary.
- Fix the symptom without reproducing it.
- Add a new abstraction for a case solved by a message or a test.
- Invent data or steps to close the report.
