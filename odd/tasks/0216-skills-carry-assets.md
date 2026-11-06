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
- [x] T1 · the contract accepts assets: declared, loaded, embedded and installed   R1 · route:inline · tree:f7a56e8e877507168e8c2fc9ebac3c0107f50c6d
- [x] T2 · writing ships the document template and routes it   R1 · route:inline · tree:c781cab3e809900cece57184ee5f89978d3e8754
- [x] T3 · the changelog names the capability   R1 · route:inline · tree:763b8f81700f212f0e74295beb7ff62a0ac26762
- [x] T4 · the impact map names the asset folder   R1 · route:inline · tree:fcd7fc336a7c4089039de85bd785daa8af6a4e5d
- [x] T5 · the payload count knows assets   R1 · route:inline · tree:56dc7766b18b7b3b2ddff24e4dc7fdd14e4b3064
- [x] T6 · the project catalog and the embed carry every asset   R1 · route:inline · tree:0ecebb5a46d8aa0c67be944eee65c22774e1f9f6

## History
L1 · "a skill folder carries `assets/`: the files it copies or uses —a template, a schema— installed with the skill, declared in the body, and refused when a file is orphaned or a declaration names no file; `writing` ships the copyable document template as its first consumer"
L2 · OFICIO · "the pinned test was red before the asset landed: writing carried no assets/doc.md" (porque the RED is the proof the test guards the asset; test and asset landed in one commit, so the record cannot show it)
L3 · OFICIO · "the embed uses all:* and a build probe listed data/.hidden, data/_under and data/doc.md" (porque the loader lists dot and underscore names, and the old * would have let a check-green tree install unequal)
