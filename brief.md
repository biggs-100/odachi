# the-worker-never-starts-without-the-section

## What must exist
- the worker never starts a delegated step without `## Skills to load before work`: it returns `status: blocked`, names the missing section, and proceeds when the section says there are none

## Done when
- the worker prompt returns `status: blocked` and names the missing section when `## Skills to load before work` is absent
- a test pins the sentence in the generated prompt, and the suite is green
- the voice text, the verifier, the reviewer and the door are unchanged
- go test ./... is green and odachi check --run "go test ./..." is green
- Unreleased names the worker refusal: the installed text changes

## Not included (first version)
- no change to the voice text: the How you speak section and the profile wording stay as they are
- no change to the verifier, the reviewer or the door: only the worker prompt moves
- no re-run of the probe here: the installed binary is the gate, and it reopens after the release installs the new worker
- no change to the skills
- no change to odachi bench and no new verb
- no change to the brief format or to start
- no change to the review, the doctor or the classifier
- no change to the stored proofs: never touched
- no change to the design canvas, its token or the installers
- no change to the release workflow

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: this week.
- Wants to learn: whether the worker refusal makes the parent pass the paths, measured after the release installs it.
