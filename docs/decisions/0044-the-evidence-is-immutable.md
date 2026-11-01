# 0044 - The evidence is immutable, the writer is not

Status: accepted - Date: 2026-10 - Session: after v0.11.87

## Context

The fence carried one rule since 0011: *"no change to the proof format, and the stored proof
is never rewritten"*, with `internal/domain/proof/**` and `proof/**` as its paths. Two
different promises travelled inside one sentence: the bytes a tree was certified with must
never move, and the code that writes them must never change.

The first promise is the one that protects evidence. A proof is the output of a command tied
to a tree hash; a verdict and a debt row are read against it, and rewriting one would leave a
certificate about bytes nobody can inspect. The second promise was a side effect of writing
both in one line, and it became a wall: 0148 made the gate read the requirements, and the
natural next step -recording in the proof which requirement it covers- needed a new line in
the header. The fence refused it, correctly, because the rule said the format does not move.

The correction is not a patch on the rule; it is the decision the rule was standing in for,
said in its own words.

## Decision

- **The fence protects the evidence.** The rule is `proof/**` with one promise: a stored
  proof is never rewritten, whatever the writer learns. `internal/domain/proof/**` leaves the
  fence, because a writer that cannot change cannot grow, and the writer has no evidence to
  destroy -it produces files, and the existing ones stay.
- **A new line is a version, not a rewrite.** A proof that gains a header field stays valid for
  the tree it names; a proof without the field stays valid too, because the field adds a claim
  and never subtracts one. Nothing reads an absent field as a false one.
- **The licence is the brief.** `check --fence-from` allows the change only because the brief
  carries a new non-goal. The fence and the vision move together or not at all; that rule is
  older than this ADR and stays.

## Consequences

- The next feature may add `# covers` to the proof header without touching the fence.
- A stored proof is still untouchable, by hand or by a migration. If a format ever needs to
  change in a way old proofs cannot express, that is a new ADR, not an edit.
- The fence is now two rules and neither restates a package: what it protects is bytes people
  read, not files people edit.
