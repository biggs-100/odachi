# a-failed-run

## Requirements (verbatim, from the user)
R1 · "a failed run keeps its whole output, so a red gate says why"

## Done when
- a failing --run writes proof/<feature>.failed.txt with the tree, the command, the exit code and the whole output
- the gate prints the last line and the path of that file, and its exit code does not change
- a green run writes only <feature>.txt, byte-identical to today, and leaves no failed file
- a red pull request uploads the failed output without touching the workflow
- the README and an ADR say what a failed file is, and what it is not

## Steps
- [x] T1 · a failed run keeps its whole output in proof/<feature>.failed.txt   R1 · d6b3a5b
- [x] T2 · a red pull request uploads that file without touching the workflow   R1
- [x] T3 · README and ADR 0025: a failed run is evidence, not a proof   R1 · a38dc90

## History
L1 · "a failed run keeps its whole output, so a red gate says why"
L2 · OFICIO · "The failed file starts with # failed <tree>, not # tree" (porque the proof parser rejects it by construction instead of trusting a suffix or a third line)
L3 · OFICIO · "A green run removes the stale failed file" (porque proof/ must not show a failure beside the proof of the tree that passed)
