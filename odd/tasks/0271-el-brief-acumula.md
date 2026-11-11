# el-brief-acumula
Base: c3c700425ab017d13f71b50ca646eba720c0c55a

## Requirements (verbatim, from the user)
R1 · "el brief deja de reemplazarse: cada start agrega la seccion de su feature, y lo que estaba escrito antes queda"

## Done when
- after a second `start` with a new requirement, `brief.md` holds the first block and the second, each under its own `## NNNN · slug` header
- `brief.Parse` returns both outcomes and both non-goals, and the gate still calls the brief complete
- the fence the second run writes still carries the first feature's rule with its sentence
- running the same requirement twice does not append a second block
- the first run of a project writes the file as it does today: a `# name` and one block
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- `## Unreleased` names it, patch (0271)

## Steps
- [x] T1 · el brief acumula: cada start agrega su bloque y el parser lee la union de todos   R1 · route:inline · tree:83edc00899b9d84d3e1955d9b799259477ba6248
- [x] T2 · una corrida no duplica: el mismo requisito no agrega un bloque dos veces   R1 · route:inline · tree:f06834b5d23c0e0c4762f493f581b75197b67783

## History
L1 · "el brief deja de reemplazarse: cada start agrega la seccion de su feature, y lo que estaba escrito antes queda"
L2 · OFICIO · "opening 0271 with v0.38.11 replaced brief.md again, and this run merged the 0270 block back by hand" (porque start still rewrites the brief until T1 lands; the hand merge is the shape the feature will write, and it keeps the plan readable with the block that justified the previous feature)
