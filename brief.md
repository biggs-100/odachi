# the-gate-never

## What must exist
- the gate never waits on a stream nobody reads, and a timeout leaves no process behind

## Done when
- a check --run whose stderr is a pipe nobody drains finishes and writes the proof
- a command that times out leaves no child or grandchild alive
- a terminal still sees the output while the command runs
- the proof still keeps both streams
- go test ./... is green

## Not included (first version)
- no change to what proof/ and odd/tasks are: never touched
- no change to the retention numbers or which trees survive
- no new verb and no new flag

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: why a stream nobody reads can deadlock a command, and why killing the parent is not killing the work.
