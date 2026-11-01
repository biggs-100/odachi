# check-ci-runs
Base: a58105fc4d22f20266f1a991701ccb20ba60868c

## Requirements (verbatim, from the user)
R1 · "check --ci runs the gate exactly as CI runs it: the fence read from the remote base and the base resolved from the remote, so the policy lands before the feature is verifiable before a push instead of after it"

## Done when
- a candidate the base fence forbids passes --fence-from HEAD and is blocked by --ci, and both were observed
- check --ci beside --base or --fence-from is refused, naming that --ci is what they are for
- a clone with no remote base blocks and names the command that creates it

## Steps
- [ ] T1 · the clone can name its remote base: origin/HEAD, then origin/main, or the error names the fetch that would create it   R1
- [ ] T2 · check --ci reads the fence and the base from that ref and refuses --base or --fence-from beside it   R1
- [ ] T3 · the run and the proof name the base they obeyed, and --ci blocks what --fence-from HEAD lets through   R1

## History
L1 · "check --ci runs the gate exactly as CI runs it: the fence read from the remote base and the base resolved from the remote, so the policy lands before the feature is verifiable before a push instead of after it"
L2 · OFICIO · "the fence keeps one rule for gate.yml and drops the no-network non-goal" (porque a fence rule protects a file the candidate must not touch; no glob measures a promise about behaviour, so the promise stays in the brief and the gate keeps it in its own documentation)
