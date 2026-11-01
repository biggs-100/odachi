# 0048 - A second reading is not a second opinion

Status: accepted - Date: 2026-11 - Session: after v0.11.93

## Context

The gate refused a verdict written by the model that produced the proof: a tree approving
itself is not a review. That rule assumed two things that turned out not to be true for
everybody. The first is that a human can afford a second model. The second is that comparing the
whole model string answers the question -it does not, because a model value carries an optional
`#variant`, which is an effort or a setting of the same model, so `provider/name#high` and
`provider/name` read as two judges when they are one.

Two questions were tangled in one refusal. **Who judged** -the model, and it must be named- and
**whether a second reading buys a second opinion**, which is a judgement about evidence and
belongs to the human, the way the review's authority always has (0043).

## Decision

- **The identity is the model, not its setting.** The gate compares `provider/name`, ignoring
  case and the `#variant`. The variant is an effort, and the sentence a human reads names the
  effort each side ran with, because that is the part that can be chosen and paid for.
- **A self-review is named by default.** With the project's mode on, a verdict by the builder's
  own model leaves the close ready and says what it is: *a second reading, not a second
  opinion*. This was already the behaviour -every review refusal but a causal finding is
  softened, 0043- and it had no test; it has one now.
- **The refusal lives where somebody asks for it.** `--require-review` still refuses a verdict
  from the builder's model, and a verdict that names no model keeps the same pair of
  treatments: named by default, refused when the run asks for the guarantee.
- **A waiver is a decision on the document.** `odachi review waive --because "<why>"` writes
  `Review: waived` into the current feature's task doc and the reason into its history, through
  the same write guard every other verb uses. It removes the verdict and nothing else: the
  proof and the fence apply, and a finding the review already made still stops the ship. A
  waiver without a reason is refused.
- **What a second reading buys, and what it does not.** The reviewer runs in a fresh session
  with a different role prompt, no tools and no memory of having written the code: that removes
  the author's bias. It does not remove correlated blind spots -a model that does not think of a
  class of bug will not think of it at another effort either. Raising the effort buys a more
  careful reading of the same mind; it does not buy a different one.

## Consequences

- A project that owns one model works without friction, and its record says exactly what it
  got: a review by the builder's model, or a waiver with a reason. Nothing pretends the first is
  the second.
- A project that wants the strong guarantee turns it on, and pays for a second model. The
  default no longer decides that for it.
- The variant hole is closed in both directions: a reviewer at another effort is the same judge
  for this rule, and the message says which effort each side ran with so a human can tell a
  cheap reading from an expensive one.
- What the tool still cannot prove, named: that the reviewer read the diff at all (only that it
  declared it), and that a second model is better than the builder's own on any particular
  candidate. The gate compares declarations; the human weighs evidence.
