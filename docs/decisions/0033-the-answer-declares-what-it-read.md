# 0033 - The answer declares what its reviewer inspected

Status: accepted (implemented in feature 0085; commits in `odd/tasks/0085-the-reviewer-declares.md`) - Date: 2026-10 - Session: v0.11.42

## Context

ADR 0003 made the verdict a receipt named after the tree; ADR 0016 gave every
row a content receipt, so a verdict says which (file, blob, lens) it answered.
Both stand. What they left open is who answered: a row is a claim that a
question was answered, and a claim is exactly what a caller can write.

The hole cost two closes in one session. A reviewer answered 49 of 52 rows; the
caller wrote the three missing ones and `review --attach` sealed the verdict, so
the store held a review nobody performed. Later the same day the reviewer left
out the row for the one file its only finding was about, and the caller wrote
that row too. Feature 0084 taught the refusal to send the gap back to the
reviewer instead of saying "fix the rows", which is the teaching half; a rule
the caller is asked to respect is not a gate.

The model to borrow was already solved elsewhere. gentle-ai's review admission
(`internal/reviewtransaction/artifact_admission.go`) requires an `inspection`
object in every reviewer answer - `status` plus the complete changed-path list,
`unavailable` with a reason when the candidate could not be read - and refuses an
answer whose declaration is short a path (`incomplete`) or names a path outside
the candidate (`out_of_scope`). Its schema makes the field `required`, so
omission is not a way to skip the check.

## Decision

- **The answer carries `inspection`.** `completed` with every changed path it
  read, or `unavailable` with a non-empty reason and no path. The vocabulary is
  gentle-ai's on purpose: one word for the same act in two projects, and
  `unavailable` is the honest exit a reviewer needs instead of inventing a path.
- **The gate compares it with the candidate in both directions.** A changed path
  the declaration leaves out is a hole; a declared path the candidate does not
  change is an invention. Both refusals name the paths.
- **It is required whenever the answer has lenses.** Not only when a blob is
  measurable: a candidate that only deletes files has no blob to stamp, and the
  exception would be a rule a reader has to hold in their head.
- **It is judged on the answer, never on the merged verdict.** The store lends a
  verdict its rows; the moment the lenses merge, a second group inherits the
  first group's declaration, and checking the union seals an answer that declared
  nothing. The review of this feature found that ordering bug; the answer is
  judged after the facts that outrank it (the lens already answered, the verdict
  for another tree) and before the coverage rows.
- **The refusal keeps the answer and asks the reviewer.** Same rule as 0084: the
  bytes were paid for, the file is named, and a verdict is not edited to pass the
  gate.
- **The prompt asks for the field and the printed schema shows it.** A gate that
  refuses an answer for a missing field is a riddle unless the question asked.
- **An older verdict still parses.** The requirement binds the answer being
  sealed, not a file already in the store.

## Alternatives rejected

- **Rows only.** The status quo, and the two incidents above: the tool cannot
  tell a row the reviewer wrote from a row the caller filled in.
- **A signature or nonce per prompt.** It proves the command ran, not that the
  reviewer read the tree; a nonce cannot be compared with a changed path, and a
  caller holding the answer holds the nonce.
- **Require it only when a row exists to answer.** A deletion-only candidate
  escapes the gate entirely, and the exception is harder to explain than the
  rule.
- **Judge the merged verdict.** It was the first implementation, and it leaves
  the door open on every second run of the same tree.
- **Leave it optional when the verdict names lenses.** Omission becomes the
  cheapest bypass, which is how the field would die.

## Consequences

- **The store format changed.** A verdict carrying `inspection` is unreadable to
  a binary before 0.11.42 (`unknown field`, the same strictness that refuses a
  typo), so the upgrade travels with the format - the pattern 0081 set with
  `project_agents`.
- **Every answer to a lens-bearing candidate is now incomplete until it
  declares.** Forty-two fixtures in this repository were writing answers that the
  new contract refuses; the CLI helpers now ask git for the changed paths instead
  of listing them by hand.
- **The gate proves coverage, not reading.** A forged declaration is still
  possible; it is a different lie than padding a row, and it must be written as a
  claim about paths rather than implied by silence.
- **One hole stays open and is recorded in `docs/review-debt.md`:** a verdict
  that names no model passes the identity gate, so a review by the model that
  built the tree can be sealed unattributed.

## Sources

- The two padded verdicts of 2026-10-28 (features 0082 and 0083).
- Feature 0084: the refusal asks the reviewer for the rows it left out.
- gentle-ai: `internal/reviewtransaction/artifact_admission.go`,
  `validateCompleteInspectionCoverage`, `reviewerprovider.LensResultSchema`.
- ADR 0016: the receipt per blob.
