# the-block-keeps-notes-not-builds

## What must exist
- a build artifact and a one-use transport are not notes: the block sends them to the temp folder, where they die with the run, and keeps `.git/odachi/scratch/` for a note a later turn has to read

## Done when
- the block's scratch paragraph names the two classes and the temp folder, and the sentences it has today stay (the tree stays clean, the folder is the one `clean` empties)
- `rules_test.go` pins both classes: a note a later turn reads stays in the scratch, a binary and a transport go to the temp
- `odachi init` writes the new block in this project, a second run writes nothing, and the installed agents carry it
- the changelog's Unreleased section names the rule
- `go test ./...` is green

## Not included (first version)
- no change to `clean`: it still empties the scratch, and the temp is outside its reach
- no new verb and no new flag
- no change to the skill contract or the curriculum
- no change to the voice or the interview's questions
- no change to the stored proofs: never touched
- no change to `odachi role` or to the skill it renders
- no change to the design canvas, its token or the push plugin
- no change to the artboard format
- no change to the release workflow

## Decisions taken
- two classes, not one rule: a note a later turn reads is kept and sweepable; a transport is single use and dies in the run
- the temp is the operating system's folder, not the clone: `clean` reclaims what it owns, and a build is not evidence
- the paragraph says why (the clone keeps notes, not builds) so an agent does not have to guess
