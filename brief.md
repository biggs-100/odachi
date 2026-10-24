# one-workflow-tells

## What must exist
- one workflow tells the truth about a tree: the gate runs the suite on every OS

## Done when
- the gate checks the candidate on ubuntu, macos and windows, each reading the fence from the base ref and keeping its proof
- gofmt, vet, race and the goreleaser check run once in that workflow; no second workflow runs the suite
- a PR that breaks the suite on one OS fails there, and a release tag still runs the suite before publishing

## Not included (first version)
- required check and ruleset: the next feature, after main carries the merged workflow
- Go module and build caching
- self-hosted runners or job time limits

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: one definition of green.
