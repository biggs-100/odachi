# the-run-that
Base: 0a953e92bf5473929d3f00f919691e4ad574a2db

## Requirements (verbatim, from the user)
R1 · "the run that reads a draft leaves the brief written: start --brief records it as brief.md, and the shape skill stops promising a file the run never wrote"

## Done when
- a draft outside brief.md becomes brief.md in the canonical shape, and the draft itself is untouched
- a draft that is already brief.md is not rewritten
- an existing brief.md beside a draft elsewhere is refused without --force, naming the flag
- the shape skill says the draft may live anywhere and the run records it
- tests pin the three paths and go test ./... is green

## Steps
- [x] T1 · start --brief records the draft as brief.md: the three paths, the skill, and the tests   R1 · route:inline · tree:8a384aca2ec9d3669826936bbbc949241a6f230b

## History
L1 · "the run that reads a draft leaves the brief written: start --brief records it as brief.md, and the shape skill stops promising a file the run never wrote"
L2 · OFICIO · "the run records the draft as brief.md: a draft elsewhere is written in the canonical shape, a draft that is brief.md is never rewritten, and an existing brief.md needs --force" (porque the skill promises the first write records the brief; a run that leaves no brief.md leaves the project without its vision, and the human's file is only replaced with permission)
