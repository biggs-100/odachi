# skills-resolve-step

## What must exist
- skills --resolve <step>: ver qué skills recibiría un paso y de qué scope

## Done when
- odachi skills --resolve T1 prints every skill the step names, with its exact path and whether it came from the project or the home
- a step that names no skills says so in one line, and a name with no file is reported with the scope it was looked for
- the run writes nothing: the tree is byte-identical after it
- the tests pin the project-first order, the home fallback, the missing name and the empty step, each observed red first
- go test ./... is green and odachi check --run "go test ./..." is green
- Unreleased names the change: a new read-only surface on skills, patch (0258)

## Not included (first version)
- no change to delegate: it keeps printing the paths this verb reads
- no new package: the verb reuses the craft resolver and the task-doc parser
- no JSON output in the first version
- no change to the installer, the workflow, the canvas or the proofs

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: whether a read-only surface stops a parent from composing a hand-off with a skill that does not resolve.
