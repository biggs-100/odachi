# odachi-check-run
Base: 9337c6534bc27703390d4f6070da63f05f869be1

## Requirements (verbatim, from the user)
R1 · "odachi check --run streams the command's output while it runs, so a lock is not read as a slow suite"

## Done when
- a command run with odachi check --run writes what it prints to stderr while it runs, before the command exits
- the proof file carries the same output it carries today: one copy, the terminator, and the same tree and model headers
- a command that prints nothing still changes nothing: no stream, the same report
- with --json the stream goes to stderr and the JSON on stdout stays what a caller parses
- the timeout report and its next line are unchanged

## Steps
- [ ] T1 · the shell edge mirrors the output to a writer while it captures it   R1 · route:inline
- [ ] T2 · check gives the stream a writer and the proof keeps one copy   R1 · route:inline
- [ ] T3 · the CLI sends the stream to stderr, and --json keeps stdout parseable   R1 · route:inline

## History
L1 · "odachi check --run streams the command's output while it runs, so a lock is not read as a slow suite"
