# a-block-falls-by-evidence
Base: f6e60e15eb2ff873f749ad6b2f7c3d39a3bc35fb

## Requirements (verbatim, from the user)
R1 · "a finding says whether its evidence is cited from the diff or argued, and a blocking finding is lifted only by the evidence its class requires: one cited counter-argument for a cited claim, two for an argued one"

## Done when
- a reviewer answer that does not say cited or argued is refused naming the field, and a verdict stored before the class existed reads as argued and keeps today's meaning
- two cited refutations lift an argued BLOCKER and one leaves it blocking, naming the vote it still needs; a cited BLOCKER falls with one -both observed in a test
- each refuter's answer is stored once, the legacy `<tree>.refuted.json` keeps the meaning it has today, and a test reads one
- an ADR says what a citation buys, why an argued claim needs two votes, and what the tool still cannot prove

## Steps
- [ ] T1 · a finding declares cited or argued evidence, and an old verdict reads as argued   R1
- [ ] T2 · an argued blocker needs two cited refutations to fall, and a cited one needs one   R1
- [ ] T3 · the ADR says what a citation buys and what the tool still cannot prove   R1

## History
L1 · "a finding says whether its evidence is cited from the diff or argued, and a blocking finding is lifted only by the evidence its class requires: one cited counter-argument for a cited claim, two for an argued one"
