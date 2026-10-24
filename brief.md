# a-failed-run

## What must exist
- a failed run keeps its whole output, so a red gate says why

## Done when
- a failing --run writes proof/<feature>.failed.txt with the tree, the command, the exit code and the whole output
- the gate prints the last line and the path of that file, and its exit code does not change
- a green run writes only <feature>.txt, byte-identical to today, and leaves no failed file
- a red pull request uploads the failed output without touching the workflow
- the README and an ADR say what a failed file is, and what it is not

## Not included (first version)
- a diff of the output, or truncating it
- renaming or parsing the green proof format
- changing the gate's exit codes

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: why evidence for a failure is not a proof.
