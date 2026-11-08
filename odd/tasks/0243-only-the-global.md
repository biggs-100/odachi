# only-the-global
Base: 9dab0ed3ba17f8ed34eac99f0a03603b100df186

## Requirements (verbatim, from the user)
R1 · "the agent lives only in the home: no command writes or refreshes a project agent file; `start` and `level` rewrite only the voice of the block, and a file an older release left is named and left alone"

## Done when
- `odachi start` writes no agent file in the project and does not touch one an older release left: the file keeps its bytes and no backup is written
- the home door is the only agent `start` reads or refreshes; a missing one is `init`'s to install
- `odachi doctor` still names a project agent leftover as a warning
- go test ./... is green and odachi check --run "go test ./..." is green
- Unreleased names the change: a contract change, minor

## Steps
- [ ] T1 · start rewrites the voice of the block and leaves a project agent leftover alone   R1 · route:inline

## History
L1 · "the agent lives only in the home: no command writes or refreshes a project agent file; `start` and `level` rewrite only the voice of the block, and a file an older release left is named and left alone"
L2 · CAMBIO R1 · "se va a eliminar de los repos y solo se usara el global"
