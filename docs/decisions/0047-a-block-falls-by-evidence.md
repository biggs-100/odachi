# 0047 - A block falls by evidence, not by confidence

Status: accepted - Date: 2026-11 - Session: after v0.11.92

## Context

A review verdict can carry a causal `BLOCKER`. Until this decision, one refuter could lift it:
the refuter reads the finding, cites the diff, and the block is gone. That is right when the
finding quotes the line that shows the defect -a counter-citation is checkable at a glance-
and it is thin when the finding argues. "This will break under load" has no line to look at, so
the refuter's citation is one opinion against another, and the tool recorded it as a verdict.

The comparison with gentle-ai made the gap visible: it sends BLOCKER and CRITICAL candidates to
three refuters and lifts a finding only when a majority refutes it, and it lets a finding whose
evidence is deterministic skip the refuter entirely. Odachi's reviewer runs nothing -it answers
from the prompt and executes no command- so it cannot claim "deterministic" in that sense. What
it can say is what it did: quoted a line, or argued.

## Decision

- **The finding declares its evidence.** `evidence` is `cited` when the reviewer quoted the
  line of the diff that shows the defect, and `argued` when it reasoned without one. The field
  is optional and **absence reads as argued**: a verdict written before the field keeps the
  strict meaning, and a claim nobody can point at does not get the benefit of a citation. An
  invented class is refused by name, the way an invented severity is.
- **The class decides the quorum.** A `cited` finding falls to one cited refutation. An
  `argued` finding needs two. A `stands` is never a vote to remove, and a `refuted` without a
  citation is not a vote at all: the default is always that the finding stands.
- **Three is the ceiling.** A tree takes at most three refuters; the fourth answer has no file
  to live in. A block decided by a crowd is a block nobody can read, and the ceiling keeps the
  quorum from becoming a campaign.
- **Each answer is its own file.** The first keeps the name a refutation always had
  (`<tree>.refuted.json`) and the next two are numbered (`<tree>.refuted.2.json`,
  `.3.json`), so an answer is written once and never rewritten -the promise 0044 made about
  evidence applies to refutations too- and a tree refuted before this decision is still read.
- **The gate says how many it has.** When a block survives, the report names the finding's id,
  the cited refutations it has and the ones it needs. A human then knows whether to ask for
  another witness or to fix the code.

## Consequences

- **A reviewer that quotes earns the short road.** The cheap exit is open to the evidence that
  can be checked, which is the behaviour the review wants to reward.
- **An argued block costs two refuter calls.** In this repository a blocking finding is rare -
  the last three features produced none - and the moment it appears is the moment being wrong
  is expensive. The ceiling bounds the cost at three.
- **What the tool still cannot prove, named.** The class is the reviewer's own account of what
  it did; the tool cannot verify that a `cited` finding really quotes the line it names, and it
  cannot verify that a counter-citation is about the same claim. What changed is not certainty:
  it is the number of independent accounts needed to remove a block, and the fact that the
  removal is now proportional to the evidence the claim carries.
- **Old evidence keeps its meaning.** A verdict with no class reads as argued and needs two
  witnesses from now on; a refutation stored before the numbering is round 1. No file is
  migrated, and no stored answer is rewritten.
- **The schema counter moves to 4.** An older binary meeting a verdict that declares the class
  says "upgrade" instead of calling a well-formed file broken (0094).
