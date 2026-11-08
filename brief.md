# only-the-global

## What must exist
- the agent lives only in the home: no command writes or refreshes a project agent file; `start` and `level` rewrite only the voice of the block, and a file an older release left is named and left alone

## Done when
- `odachi start` writes no agent file in the project and does not touch one an older release left: the file keeps its bytes and no backup is written
- the home door is the only agent `start` reads or refreshes; a missing one is `init`'s to install
- `odachi doctor` still names a project agent leftover as a warning
- go test ./... is green and odachi check --run "go test ./..." is green
- Unreleased names the change: a contract change, minor

## Not included (first version)
- no change to the bundled skill texts: the 19 keep their words
- no change to the review or the classifier
- no change to odachi bench and no new verb
- no change to the stored proofs: never touched
- no change to the design canvas, its token or the installers
- no change to the release workflow

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: whether removing the last writer keeps a project clean.
