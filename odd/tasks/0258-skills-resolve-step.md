# skills-resolve-step
Base: 53c80f25b534ef610d1b48e88b8c51027a1d1ff8

## Requirements (verbatim, from the user)
R1 · "skills --resolve <step>: ver qué skills recibiría un paso y de qué scope"

## Done when
- odachi skills --resolve T1 prints every skill the step names, with its exact path and whether it came from the project or the home
- a step that names no skills says so in one line, and a name with no file is reported with the scope it was looked for
- the run writes nothing: the tree is byte-identical after it
- the tests pin the project-first order, the home fallback, the missing name and the empty step, each observed red first
- go test ./... is green and odachi check --run "go test ./..." is green
- Unreleased names the change: a new read-only surface on skills, minor (0258)

## Steps
- [x] T1 · the resolve surface: the flag, the project and home paths, the missing name and the empty step, tests red first   R1 · route:inline · tree:3ff8edca93684d429a9618e114bbecc9aa106448

## History
L1 · "skills --resolve <step>: ver qué skills recibiría un paso y de qué scope"
L2 · OFICIO · "The resolve surface lives in cmd/odachi and reuses craftPath; it adds no app package" (porque the resolver already exists there and a second copy of the project-first order would be the way the two verbs disagree; the brief's non-goal says reuse it)
L3 · OFICIO · "A name that resolves nowhere blocks the run; an empty step is an answer" (porque the surface exists to inspect a hand-off, and a missing skill means the hand-off it describes would fail, while a step that owes none is a valid answer and exits 0)
