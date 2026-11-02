# odachi-check-run

## What must exist
- odachi check --run streams the command's output while it runs, so a lock is not read as a slow suite

## Done when
- a command run with odachi check --run writes what it prints to stderr while it runs, before the command exits
- the proof file carries the same output it carries today: one copy, the terminator, and the same tree and model headers
- a command that prints nothing still changes nothing: no stream, the same report
- with --json the stream goes to stderr and the JSON on stdout stays what a caller parses
- the timeout report and its next line are unchanged

## Not included (first version)
- no change to how the reviewer runs: the stream is for the proof, and review --exec keeps its buffered call
- the proof format does not change: the stream is not a second copy of the output

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: why a stream and a file can carry the same evidence.
