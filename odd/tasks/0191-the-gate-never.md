# the-gate-never
Base: 35092d9ffbe63a8c317c8a898d04c3c4d05817bd

## Requirements (verbatim, from the user)
R1 · "the gate never waits on a stream nobody reads, and a timeout leaves no process behind"

## Done when
- a check --run whose stderr is a pipe nobody drains finishes and writes the proof
- a command that times out leaves no child or grandchild alive
- a terminal still sees the output while the command runs
- the proof still keeps both streams
- go test ./... is green

## Steps
- [x] T1 · the mirror only streams to a terminal, so a pipe nobody drains cannot block the run   R1 · route:inline · tree:16084f5035bced328de979ec032d330b7cc4ea95
- [x] T2 · a timeout kills the whole process tree, so no child outlives the run   R1 · route:inline · tree:519db77f257e20355979e80a764b1498bba23961

## History
L1 · "the gate never waits on a stream nobody reads, and a timeout leaves no process behind"
