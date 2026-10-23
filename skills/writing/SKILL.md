---
name: writing
description: Trigger: when writing a message, a PR description, an issue comment, or documentation someone will read to act.
---

## When

When the text has to make another person **act**: understand what changed, review
a PR, reproduce a bug, or run a command you wrote.

## Rules

- Write for the person who has to act, not to prove you understood. What happened
  or what is missing first; details after.
- What cannot be scanned in thirty seconds is not read: a title that says the
  point, a first line with the conclusion, and only then the development.
- Examples beat adjectives: "takes 40 s" instead of "is slow";
  `odachi check --run "npm test"` instead of "run the tests".
- If you name a command, it has to work **exactly as written**. A wrong command
  costs the reader an afternoon.
- In a PR: what problem it solves, how it is tested, and what was left out. Three
  paragraphs, not three pages.
- For a long document: answer first, then the happy path, then details and edge
  cases. Use small sections, tables and checklists so the reader can scan.
- Say what you did **not** cover and did not verify. A result is worth what its
  stated scope is worth.
- For a review, say what to look at first, what is out of scope, and which
  decision each section carries.

## Never

- "As you can see in the code...": if it has to be seen in the code, the text is
  unnecessary.
- A list of changes without saying what problem they solve.
- Promise in a changelog something that was not tested.
- Adjectives without a number next to them.
