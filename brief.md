# skills-carry-assets

## What must exist
- a skill folder carries `assets/`: the files it copies or uses —a template, a schema— installed with the skill, declared in the body, and refused when a file is orphaned or a declaration names no file; `writing` ships the copyable document template as its first consumer

## Done when
- `internal/domain/skill` accepts `assets/<file>` and refuses an undeclared asset, a declaration that names no file, and a tree under `assets/`, each with its own line
- the loader reads `assets/`, the embed carries it, and `odachi pack` installs it beside `SKILL.md` and `references/`
- `skills/writing/assets/doc.md` exists, `writing` declares it under `## Assets`, and its rules send the writer to copy it
- a test in `skills/embed_test.go` pins the template and the route, observed red before the asset lands
- `odachi skills --check` is green with the asset, and a fixture skill with an orphaned asset is refused
- `odachi init` leaves the installed copy equal to the tree, asset included
- `docs/impact-map.md` names `assets/` in the skill format
- the changelog's Unreleased section names the capability
- `go test ./...` is green

## Not included (first version)
- no nested trees under `assets/`: one file, flat, because a template is a file and a tree needs a walk nobody asked for
- no new verb and no new flag
- no change to `odachi check` or to `odachi review` behavior
- no change to the six risks in `commits` or to their table
- no change to the voice or the interview's questions
- no change to the stored proofs: never touched
- no change to `odachi role` or to the skill it renders
- no change to the design canvas, its token or the push plugin
- no change to the artboard format
- no change to the release workflow
- no change to the skills the feature does not touch: drive, tests, slices, pull-request, commits and triage

## Decisions taken
- an asset is a file, not a tree: a template or a schema is one file, and a tree needs a walk and path rules nobody asked for
- every asset is declared in `## Assets` the same way a reference is declared in `## References`: a file nobody routes is a file nobody reads
- any extension is allowed: the contract does not guess what a template is made of
- `writing` is the consumer because its document shape is prose today and a copyable skeleton is the artifact it lacks
- it is a contract change: `Risk: contract`, and the release is minor
