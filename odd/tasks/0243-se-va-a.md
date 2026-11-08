# se-va-a
Base: 9dab0ed3ba17f8ed34eac99f0a03603b100df186
Light: yes

## Requirements (verbatim, from the user)
R1 · "se va a eliminar de los repos y solo se usara el global"

## Done when
- `odachi start` writes no agent file in the project and does not touch one an older release left: the file keeps its bytes and no backup is written
- the home door is the only agent `start` reads or refreshes; a missing one is `init`'s to install
- `odachi doctor` still names a project agent leftover as a warning
- go test ./... is green and odachi check --run "go test ./..." is green
- Unreleased names the change: contract, minor

## Steps
- [ ] T1 · se va a eliminar de los repos y solo se usara el global   R1

## History
L1 · "se va a eliminar de los repos y solo se usara el global"
L2 · OFICIO · "start leaves a project agent file alone, like level" (porque the file is a leftover nobody reads; rewriting it would keep alive what doctor tells the human to remove)
