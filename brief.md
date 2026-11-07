# the-skills-embed-is-measured

## What must exist
- a regression from `all:*` to `*` in the skills embed fails a repo test, not a release: the suite reads a hidden fixture through the bundled payload, so the build that carries every skill is measured where it is compiled

## Done when
- `skills/.embed-fixture/` carries a dot file and an underscore file, and the embed keeps them
- a test in the `skills` package reads both through the bundled filesystem and fails when the directive goes back to `*`; observed red by flipping it and back
- the fixture is not a skill: `All()` keeps its count, `odachi skills --check` stays green, and `init` and `pack` never write it
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- the changelog's Unreleased section names the pin

## Not included (first version)
- no change to the loader's globs, the skill grammar or the registry
- no new skill and no change to the skills' text
- no change to the brief's format or to `start`
- no change to the review, the doctor or the classifier
- no change to the stored proofs: never touched
- no change to the voice
- no change to the design canvas, its token or the push plugin
- no change to the artboard format
- no change to the release workflow
- no change to the installers

## Decisions taken
- the fixture lives at the top of `skills/` with a dot name: `*` skips it, `all:*` keeps it, and the loader never lists it because it carries no SKILL.md
- the test reads the filesystem the binary carries, not the disk: the pin is about the build, and a disk read would pass even when the payload is short
- the fixture carries a dot file and an underscore file inside it, so the recursive exclusion is measured too and not only the top-level one
- no text assertion on the directive: a test that greps `embed.go` would pass on a comment and would not prove the payload
