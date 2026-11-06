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
- A text that was sent is not a text that arrived: read the published copy back and
  compare it with the file you wrote, not with your memory. For a PR, the command
  and the fields are `pull-request`'s.
- For a long document: answer first, then the happy path, then details and edge
  cases. Use small sections, tables and checklists so the reader can scan.
- Copy `assets/doc.md` for a long document: the skeleton carries the shape, and
  `references/the-doc.md` says how to fill each part.
- Say what you did **not** cover and did not verify. A result is worth what its
  stated scope is worth.
- For a review, say what to look at first, what is out of scope, and which
  decision each section carries.

## Decision gates

| Situation | Load |
|---|---|
| a guide, an RFC, a README or an onboarding page someone reads to act | `references/the-doc.md` |
| a comment, a review reply or an async update | `references/the-comment.md` |

## Output

The text the reader acts on: the point in the title, the conclusion in the first line, examples with their numbers, and what was not covered. A PR description carries the problem, the test and what was left out, in three paragraphs.

## References

- `references/the-doc.md` — the long document: answer first, quick path, details, checklist, next step.
- `references/the-comment.md` — the comment: the formula, the tone, the language of the thread.

## Assets

- `assets/doc.md` — the copyable skeleton of a long document: title, conclusion, quick path, details, checklist and next step.

## Never

- "As you can see in the code...": if it has to be seen in the code, the text is
  unnecessary.
- A list of changes without saying what problem they solve.
- Promise in a changelog something that was not tested.
- Adjectives without a number next to them.
