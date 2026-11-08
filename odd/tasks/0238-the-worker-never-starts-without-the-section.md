# the-worker-never-starts-without-the-section
Base: 9833067c79980d3ccb9647daf1323828092b02ed

## Requirements (verbatim, from the user)
R1 · "the worker never starts a delegated step without `## Skills to load before work`: it returns `status: blocked`, names the missing section, and proceeds when the section says there are none"

## Done when
- the worker prompt returns `status: blocked` and names the missing section when `## Skills to load before work` is absent
- a test pins the sentence in the generated prompt, and the suite is green
- the voice text, the verifier, the reviewer and the door are unchanged
- go test ./... is green and odachi check --run "go test ./..." is green
- Unreleased names the worker refusal: the installed text changes

## Steps
- [x] T1 · the worker refuses a delegation without the skills section, and a test pins the sentence   R1 · route:inline · tree:1b43216ecb7b102b195d876db9e116368fa6561c

## History
L1 · "the worker never starts a delegated step without `## Skills to load before work`: it returns `status: blocked`, names the missing section, and proceeds when the section says there are none"
