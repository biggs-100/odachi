# 0029 - A base is a commit, not a text

Status: accepted · Date: 2026-10 · Session: v0.11.27

## Context

Closing feature 0055 hit a dead end the tool built itself. The feature was
reviewed with `odachi review --base 6aeb877`: the seven-character spelling of the
commit the feature started from, which git resolves to the same commit as the
forty characters the task doc records. The reviewer judged the right diff, and the
receipt recorded the base it was given.

`odachi check` then refused that receipt: *"the review of this tree judged another
diff (base "6aeb877"): the feature starts at 6aeb87793d…"*. It compared the base as
text. `odachi review` refused to judge again -"this tree already has a verdict"- so
the only ways out were editing a receipt by hand or paying for the same review
twice. A paid review, filed with a true base, could not answer the question it was
paid to answer.

## Decision

- **The gate resolves the base.** `check` asks git which commit each spelling
  names (`rev-parse --verify <ref>^{commit}`) and accepts the receipt when both
  name the same one. A ref that cannot be resolved is still refused: what cannot
  be measured is not evidence.
- **The receipt keeps recording the invocation.** `review` stores the base it was
  given, not a normalized one: a receipt says what ran. The gate, which asks
  whether this verdict judged this diff, asks git what that base named.
- **Two spellings of one commit are one diff.** A prefix, a tag or a branch that
  points at the feature's base answers the same question as the full hash, and it
  costs nothing to accept it.

## Alternatives rejected

- **Normalize only in `review`.** It would fix the receipts of tomorrow and leave
  the ones already paid for dead, which is the incident that motivated this; and a
  receipt written by a script or by CI would still be compared as text.
- **Rewrite the receipt to the full hash.** Making a receipt match the gate by
  hand is the failure receipts exist to prevent; the gate should read what is
  true, not demand a spelling.
- **Compare the trees of the two bases.** Two commits can carry the same tree, and
  equal content is not equal scope: the diff a reviewer judged starts at a commit,
  not at a tree.
- **A verb that re-files a verdict under another spelling.** That charges the human
  for a spelling and turns a typo into a bill.
- **Compare by prefix.** `6aeb877` is a prefix of the full hash, but so is any
  shorter or ambiguous text, and a tag is not a prefix at all.

## Consequences

- A review paid once answers the diff it judged, whatever spelling its receipt
  carries.
- The gate asks the repository one more question per closed feature; the cost is
  one `rev-parse`.
- A base spelled as a moving ref -a branch, a tag- is resolved when the gate reads
  it. The task doc's `Base:` line is written from HEAD as a commit, so the
  feature's own base cannot drift under its receipt.
- The check port gained `Resolve`, and gitx answers it; three tests pin it, one of
  them end to end through the CLI, and the old refusal still has a test.

## Sources

- Feature 0055: the receipt with the short base and the refusal that had no way
  out.
- Decisions 0016 (the receipt) and 0018 (a receipt records what ran).
