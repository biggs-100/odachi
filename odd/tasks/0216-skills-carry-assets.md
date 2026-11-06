# skills-carry-assets
Base: 71e370e9797d80b78fc8a96ceca2d54c4fcda726

## Requirements (verbatim, from the user)
R1 · "a skill folder carries `assets/`: the files it copies or uses —a template, a schema— installed with the skill, declared in the body, and refused when a file is orphaned or a declaration names no file; `writing` ships the copyable document template as its first consumer"

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

## Steps
- [ ] T1 · the contract accepts assets: declared, loaded, embedded and installed   R1 · route:inline
- [ ] T2 · writing ships the document template and routes it   R1 · route:inline
- [ ] T3 · the changelog names the capability   R1 · route:inline

## History
L1 · "a skill folder carries `assets/`: the files it copies or uses —a template, a schema— installed with the skill, declared in the body, and refused when a file is orphaned or a declaration names no file; `writing` ships the copyable document template as its first consumer"
