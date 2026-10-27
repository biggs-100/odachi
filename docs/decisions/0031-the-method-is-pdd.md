# 0031 · PDD: the method has a name

Status: accepted · Date: 2026-10 · Session: v0.11.31

## Context

ADR 0008 named the discipline "ODD-lite" and ADR 0010 described the review as
"RDD's result without RDD's machinery". Both names describe odachi by what it
takes from somebody else, and the README never said a name at all. The human who
owns the vision chose one of its own.

## Decision

- **The method is Proof-Driven Development (PDD).**
- **The proof is the atom**: a green run of this tree, in `proof/`, named by the
  tree that produced it. Nothing is done without it.
- **The verdict is the judgment**: the close asks for a review of the diff the
  feature merges, and the verdict is named by the tree it judged.
- **The fence is the human's line**: what is not included is written before the
  work, and the gate refuses what falls outside it. An exception is signed and
  recorded, never silent.
- **The name adds no grammar, no verb and no file.** The decisions of ADR 0008
  stand; only the label changes.
- **PDD here means proof, not product.** Product-Driven Development uses the same
  letters for another claim.

## Consequences

- The README says the name once, in "What it is". The ADRs keep the comparison
  with gentle-ai as history, not as the identity.
- ADR 0008 keeps its text and gains a pointer: rewriting an accepted decision
  would erase what was decided then.
- "The odachi loop" stays available for the mechanical cycle. PDD is the method,
  not a second product.
