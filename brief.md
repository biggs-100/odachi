# review-answers-from-stdout

## What must exist
- review --exec reads the reviewer's answer from stdout, and a failing reviewer shows both streams; the proof still keeps both

## Done when
- a reviewer that writes a banner and ANSI to stderr still stores a valid verdict from its stdout
- a reviewer that fails shows stdout and stderr in the failure message
- check --run keeps both streams in the proof
- go test ./... and odachi check are green

## Not included (first version)
- changing the reviewer contract: the prompt still arrives on stdin and the answer still comes on stdout
- stripping ANSI or prose from stdout: the answer must be JSON, and decoration is the reviewer's problem
- stamping the verdict's instant from the tool (recorded as its own decision, not this feature)

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: no rush.
- Wants to learn: why a stream boundary is a contract, not a formatting detail.
