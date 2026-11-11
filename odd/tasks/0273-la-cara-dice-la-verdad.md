# la-cara-dice-la-verdad
Base: 94645f09ea8bb951660d6e664c038fa130c4388c

## Requirements (verbatim, from the user)
R1 · "atacá todo eso: que el help liste todos los verbos, que el README no afirme lo que el repo desmiente (provenance, surfaces, packages, install.ps1), y los topics y la descripción del repo"

## Done when
- every verb the dispatcher handles appears in `odachi --help`, and a test fails naming the one that is missing (observed red with `design` today)
- `review probe`, `review mode` and `review waive` are visible where a person looking for review finds them
- the README no longer claims a package count, a surface count, that the release has no provenance, or that `install.ps1` has no automated test; each place names the command or keeps no number
- the README opens with a short table of contents that links the sections it has
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- `## Unreleased` names it, patch (0273)

## Steps
- [x] T1 · el help no pierde un verbo: el uso los lista y un test lee los casos del despachador   R1 · route:inline · tree:aa04a25a14bba917d2ea69f5d0f6549c03f690f5
- [ ] T2 · el README dice lo que el repo hace: sin numeros, provenance con su comando, el test de install.ps1 y un indice   R1 · route:inline

## History
L1 · "atacá todo eso: que el help liste todos los verbos, que el README no afirme lo que el repo desmiente (provenance, surfaces, packages, install.ps1), y los topics y la descripción del repo"
