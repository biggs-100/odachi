# lo-que-la-compuerta-ve
Base: 0a84a94125ae69b93d7a131fad67a78543e0ce9a

## Requirements (verbatim, from the user)
R1 · "un artefacto que no decodifica como UTF-8, o que trae las secuencias de un UTF-8 leído como CP850, se rechaza en la compuerta nombrando el archivo, para que la doble codificación se vea sin modelo y sin review"

## Done when
- a `brief.md` whose bytes are not UTF-8 blocks the gate, naming the file and how to save it
- a `brief.md` with a BOM blocks, naming the BOM
- a brief block whose requirement no longer matches its task document's R1 blocks, naming both files
- a clean tree stays ready, so the guards do not tax every run
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- `## Unreleased` names it, patch (0274)

## Steps
- [x] T1 · los artefactos decodifican: UTF-8 valido y sin BOM, y la compuerta nombra el archivo   R1 · route:inline · tree:01528f1285cebfe1f084ff66c0b6c51641cf80b9
- [ ] T2 · el brief coincide con su documento: la oracion de cada bloque contra el R1 del task doc   R1 · route:inline

## History
L1 · "un artefacto que no decodifica como UTF-8, o que trae las secuencias de un UTF-8 leído como CP850, se rechaza en la compuerta nombrando el archivo, para que la doble codificación se vea sin modelo y sin review"
L2 · OFICIO · "T2 shipped the double-encoding guard and not the brief-to-document cross-check the step names" (porque the cross-check blocked valid states: a project brief before its first document, and a block whose document was removed -the plan is the record and the record accumulates. The guard catches the case R1 names, and the step's other half is dropped with this note)
