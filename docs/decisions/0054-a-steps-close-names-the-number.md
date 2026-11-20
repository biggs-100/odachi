# 0054 - A step's close names the number, and only a feature's close refuses

Status: accepted - Date: 2026-11 - Session: after v0.56.0

## Context

The four-hundred line limit has two moments, and neither caught the feature that
needed catching. `plan` weighs the candidate from the feature's base before a step is
written, and `check` refuses to owe a verifier past it at the close. 0337 closed at 3678
authored lines and 0338 at 1801, and both plans were written whole before the code
existed: at each `plan` call the diff was empty, so the refusal had nothing to weigh, and
the number only appeared at the close, where it can no longer stop anything.

gentle-shell, asked the same question, answers it in a different place. Its number is
`SLICE_LINES = 400` in `lib/commit-slice-signal.ts`, counted from the branch's merge base,
and it is **reported after every commit** as a line appended to the tool result -never a
block. Its only hard refusal is at delivery: `lib/review-delivery-preflight.ts` denies
`git push`, `gh pr create` and `gh pr merge` when the native assessment reports a review
due, and `slice_budget_reached` is one of the two reasons. Overages are recorded with a
rationale and one bounded slicing pass; no code is ever shrunk to fit.

## Decision

- **The step's close names the number.** `odachi plan --done T1` reports the authored lines
  from the feature's base, the base it counted from, and the limit the close applies, in
  the one line a person reads. Same measure as the plan and the gate, from one function.
- **Past the limit, the sentence says what follows in the gate's own words**: the close
  owes no verifier and the rest of the work belongs to a feature of its own. It is a note.
- **It never refuses.** The exit code and the result are those of a done step.
- **The only refusal stays at the feature's close**, and it stays narrow: the verifier a
  close owes past the limit is a verifier the tree cannot finish.
- **The delivery decision stays the human's.** Nothing here cuts a branch or opens a
  chain; the number is what a person uses when they decide, at publish time.

## Alternatives rejected

- **Refuse the step's close past the limit.** The code is already written. Blocking the
  record would hide the work instead of removing it, and the document is the memory of
  what happened, not a gate on what may exist.
- **A per-step forecast at plan time.** A plan written whole has no diff to weigh, so the
  number would be the model's guess -exactly what gentle-shell moved away from when it
  replaced the estimate with a count read from Git.
- **Move the limit to delivery only, as gentle-shell does.** odachi does not run push or
  pull requests; they are the human's commands, outside the tool. A gate it cannot see
  is not a gate.
- **Lower the limit, or count differently.** The number is not the defect: 0337 and 0338
  were large because the work was large, and a smaller limit would have refused them at
  the same late moment.

## Consequences

- The number becomes visible at every step from the one that crosses, which is while
  splitting is still cheap: the next step can be a feature of its own.
- A feature that passes the limit keeps saying so, once per step, until it closes -the
  same fact the close will then apply.
- The close still refuses the one thing it can: a verifier for a tree it cannot finish.
  Nothing about the refusals that existed moves.
- Nothing shrinks code, and no step is re-planned to fit a number: the note is a
  measurement, and the split is a decision.
