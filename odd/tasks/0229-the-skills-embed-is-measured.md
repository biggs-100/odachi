# the-skills-embed-is-measured
Base: 05d62e6c1a429ff23e83d3da7466bf46f219659c

## Requirements (verbatim, from the user)
R1 · "a regression from `all:*` to `*` in the skills embed fails a repo test, not a release: the suite reads a hidden fixture through the bundled payload, so the build that carries every skill is measured where it is compiled"

## Done when
- `skills/.embed-fixture/` carries a dot file and an underscore file, and the embed keeps them
- a test in the `skills` package reads both through the bundled filesystem and fails when the directive goes back to `*`; observed red by flipping it and back
- the fixture is not a skill: `All()` keeps its count, `odachi skills --check` stays green, and `init` and `pack` never write it
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- the changelog's Unreleased section names the pin

## Steps
- [x] T1 · the embed carries a hidden fixture   R1 · route:inline · tree:1f2d3894d495dd9596ac1a17f90d7700d6f3a1fe
- [x] T2 · the changelog names the pin   R1 · route:inline · tree:b6d08915c611468b1ea641d48922f2cb919e0b55

## History
L1 · "a regression from `all:*` to `*` in the skills embed fails a repo test, not a release: the suite reads a hidden fixture through the bundled payload, so the build that carries every skill is measured where it is compiled"
