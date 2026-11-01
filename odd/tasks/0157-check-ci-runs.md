# check-ci-runs
Base: a58105fc4d22f20266f1a991701ccb20ba60868c

## Requirements (verbatim, from the user)
R1 · "check --ci runs the gate exactly as CI runs it: the fence read from the remote base and the base resolved from the remote, so the policy lands before the feature is verifiable before a push instead of after it"

## Done when
- a candidate the base fence forbids passes --fence-from HEAD and is blocked by --ci, and both were observed
- check --ci beside --base or --fence-from is refused, naming that --ci is what they are for
- a clone with no remote base blocks and names the command that creates it

## Steps
- [x] T1 · the clone can name its remote base: origin/HEAD, then origin/main, or the error names the fetch that would create it   R1 · tree:d76c1efa28d567b4ce277792bd3c947917424128
- [x] T2 · check --ci reads the fence and the base from that ref and refuses --base or --fence-from beside it   R1 · tree:0c32f540beed3be36b98cf7e924ecba78cfc0a02
- [x] T3 · the run and the proof name the base they obeyed, and --ci blocks what --fence-from HEAD lets through   R1 · tree:cb39d6d092443a5cc55b5a9c8ff6b0e8075c4eae

## History
L1 · "check --ci runs the gate exactly as CI runs it: the fence read from the remote base and the base resolved from the remote, so the policy lands before the feature is verifiable before a push instead of after it"
L2 · OFICIO · "the fence keeps one rule for gate.yml and drops the no-network non-goal" (porque a fence rule protects a file the candidate must not touch; no glob measures a promise about behaviour, so the promise stays in the brief and the gate keeps it in its own documentation)
L3 · OFICIO · "a step is marked done with --commit <sha>, not from a clean worktree" (porque the mark itself dirties the task doc, so the next step's worktree can never equal its commit; naming the commit records the tree without asking the doc to be clean)
L4 · OFICIO · "the fence keeps an explicit proof/** rule beside the gate.yml one" (porque a new brief licenses start to rewrite the fence, and the archtest that pins the evidence boundary is the only thing that noticed proof/** had gone; protection this repository wants permanently cannot rely on the brief of the day)
