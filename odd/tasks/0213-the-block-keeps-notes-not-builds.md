# the-block-keeps-notes-not-builds
Base: 30ef920dc4e3e406698fcafcba0e7ee30ada0b61

## Requirements (verbatim, from the user)
R1 · "a build artifact and a one-use transport are not notes: the block sends them to the temp folder, where they die with the run, and keeps `.git/odachi/scratch/` for a note a later turn has to read"

## Done when
- the block's scratch paragraph names the two classes and the temp folder, and the sentences it has today stay (the tree stays clean, the folder is the one `clean` empties)
- `rules_test.go` pins both classes: a note a later turn reads stays in the scratch, a binary and a transport go to the temp
- `odachi init` writes the new block in this project, a second run writes nothing, and the installed agents carry it
- the changelog's Unreleased section names the rule
- `go test ./...` is green

## Steps
- [x] T1 · the block names the two classes: a note kept, a transport that dies in the run   R1 · route:inline · tree:b5acbee5afaf2c30ac244d5cdd0a36e30856d0d3
- [x] T2 · this project's block and the changelog carry the rule   R1 · route:inline · tree:db891fda807ca5a40c2fedd0b9ac615f662baef3

## History
L1 · "a build artifact and a one-use transport are not notes: the block sends them to the temp folder, where they die with the run, and keeps `.git/odachi/scratch/` for a note a later turn has to read"
