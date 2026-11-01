# the-evidence-is-immutable
Base: 02b18764eaec92eb8109d34c95ccdfd7e5362e58

## Requirements (verbatim, from the user)
R1 · "the fence guards the bytes of a stored proof, not the code that writes them: the format may gain a versioned line, and no stored proof is ever rewritten"

## Done when
- the fence protects the stored proofs and stops naming the package that writes them
- the brief carries the same rule as a new non-goal, not as a comment
- check with a base ref does not call it a loosening, and the ADR says why

## Steps
- [ ] T1 · the fence and the brief carry one rule: the evidence is immutable   R1
- [ ] T2 · the ADR says why the evidence and the writer differ   R1
- [ ] T3 · the changelog says the policy changed   R1

## History
L1 · "the fence guards the bytes of a stored proof, not the code that writes them: the format may gain a versioned line, and no stored proof is ever rewritten"
