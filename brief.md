# a-block-falls-by-evidence

## What must exist
- a finding says whether its evidence is cited from the diff or argued, and a blocking finding is lifted only by the evidence its class requires: one cited counter-argument for a cited claim, two for an argued one

## Done when
- a reviewer answer that says neither cited nor argued reads as argued, an unknown value is refused naming the two, and a verdict stored before the class existed keeps its meaning
- two cited refutations lift an argued BLOCKER and one leaves it blocking, naming the vote it still needs; a cited BLOCKER falls with one -both observed in a test
- each refuter's answer is stored once, the legacy `<tree>.refuted.json` keeps the meaning it has today, and a test reads one
- an ADR says what a citation buys, why an argued claim needs two votes, and what the tool still cannot prove

## Not included (first version)
- no change to the four lens prompts: the class is asked in the reviewer's answer, not in what a lens looks for
- no network, no second reviewer run and no automatic retry: the extra refuters are asked by a human, the same way the first one is
- no change to the severity vocabulary: cited or argued, BLOCKER means what it means today

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: why a claim is worth what can be cited from it, and not how convincing it sounds.
