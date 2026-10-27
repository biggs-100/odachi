# 0032 · Organic PDD: the work is shaped before it is declared

Status: accepted · Date: 2026-10 · Session: v0.11.33

## Context

ADR 0031 named the method Proof-Driven Development. ADR 0008 took ODD's
discipline without its grammar, ADR 0010 took RDD's result without its machinery,
and ADR 0007 refused the orchestrator: `slices` owns delegation, the parent owns
the work. What that left is a cycle that is honest at the back and rigid at the
front:

- `start` interviews before the code is explored, so the brief is written with
  the least information the feature will ever have.
- The plan grows (`plan` adds steps), but a requirement the human changes mid-
  feature has no home: the brief is the vision and rewriting it is a new feature.
- Ceremony does not scale with what the change can break. Every feature pays the
  fence, the proof and, at the close, a verdict; the waiver exists, but it is the
  human's exception, not a measured decision.
- An independent verifier exists (`odachi-verify`) but the policy names it for
  delegated work, not for a change that touches something dangerous.

The reference implementation is gentle-ai's organic RDD: ODD's seven steps
(authorize, explore, resolve, classify, track, implement, close) and RDD's atomic
review over the frozen candidate. Its story is the argument for enforcement:
good human decisions do not survive contact with an agent unless something makes
them binding, and the first version that ignored that shipped three monthly
resets of infrastructure nobody asked for. The lesson is the enforcement, not the
machinery: phases with contracts, one writer per lane, verify before asserting.
Odachi enforces with a binary and four files. That is its own way of being
binding, and it does not need a lineage, a consent envelope or a lease.

## Decision

**Organic PDD is the PDD spine with an organic front**: the work is shaped by
exploration and the human's corrections, and the ceremony is sized by what the
change can break. The proof is never optional. It is one method with an
adjective, not a second cycle.

1. **Shape before declare.** Exploration and the human's real questions come
   before the brief; the agent may draft it, `start` records it, and the human
   owns it. Exploration is not a new verb: it is a skill `drive` loads when the
   request does not fit in one reading, and it ends in `start --brief`.
2. **The plan grows, the vision does not.** Steps are added when understanding
   grows. A requirement the human changes appends their words to the History and
   reopens the affected steps; `brief.md` changes only when the human changes the
   vision, and that is a new feature or a deliberate edit, never the agent's.
3. **Ceremony by blast radius, not by size.** The six items - data, security,
   contract, concurrency, delivery, no-test - decide the lenses, the verifier and
   whether a waiver is on the table. The fence and the proof stay for every
   change: that is what PDD means, and size is never a reason to skip them.
4. **The verifier is named by a risk item, not by delegation.** When a change
   touches one of the six, an independent read-only verification happens even if
   the parent wrote inline. It reads the whole task doc and returns a verdict per
   requirement, with probes, never against the writer's summary.
5. **One cycle, no second grammar.** No `S#`/`Tasks`/`Log` document, no Engram
   mirror, no orchestrator, no lineage/consent envelopes/leases. The four
   artifacts are the memory and the binary is the gate; what does not fit is a
   new skill or an ADR, not a heavier core.
6. **Close with evidence, not with a summary.** The verdict of this tree plus the
   proof of this tree. Missing proof becomes one **Needs your decision** result
   naming the blocker, never a "partial" that reads as done.

## Not taken

- The ODD feature grammar and its mirror: ADR 0008 already decided against them.
- The size classification that erases artifacts for small work. A one-line fix in
  PDD still declares its fence and leaves its proof; the waiver is the only
  exception, and it is signed and named.
- The orchestrator and RDD's transaction: ADR 0007 and ADR 0010 decided against
  the machinery, with reasons.
- The ~400-line slice budget: odachi's unit is the step, with its test and its
  proof, not a line count.
- An `assess` verb: the floor already measures the candidate from its paths and
  content, and `review --show` prints the plan. A second measurement would be a
  second source of truth to drift.

## Consequences

- **This ADR records the design, not the implementation.** Until the three
  follow-ups land, the current cycle stands and nothing about what `check`
  accepts changes.
- The follow-ups, each with its own evidence: a `shape` skill that ends in
  `start --brief`; the six-item table in `slices`/`drive` with the verifier
  trigger; the requirement-change rule in `drive` and `ask`.
- If the follow-ups are never worth building, this record still says why the
  front was the missing half. A design that can wait is cheaper than machinery
  nobody uses, which is the first lesson of the story.
