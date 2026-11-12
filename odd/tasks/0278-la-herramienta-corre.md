# la-herramienta-corre
Base: bfe51fd8e65fda2a992b91f651102e94d5755da5

## Requirements (verbatim, from the user)
R1 · "la herramienta corre antes que el modelo y su fallo es un defecto que no se puede ignorar"

## Done when
- una comprobación declarada que sale distinta de cero deja en el veredicto un finding con su nombre, su salida y la severidad que el proyecto le declaró
- una comprobación que sale cero no deja finding, y el prompt dice que pasó
- una comprobación cuyo comando no existe deja una fila que lo dice, en vez de caerse
- las comprobaciones corren en el worktree del candidato y no en el repositorio
- go test ./... is green and odachi check --run "go test ./..." is green
- ## Unreleased names it, patch (0278)

## Steps
- [x] T1 · el archivo del review declara las comprobaciones: nombre, comando y severidad, y un archivo sin ellas se lee igual   R1 · route:inline · tree:b8709618731d17669f6eecc59c84c20736bb6514
- [ ] T2 · el review corre cada comprobacion en el worktree del candidato y el prompt lleva el resultado antes de cualquier lente   R1 · route:inline
- [ ] T3 · una comprobacion que falla es un finding de la herramienta, con su salida y su severidad   R1 · route:inline
- [ ] T4 · el veredicto dice lo que corrio: --show nombra las comprobaciones y Unreleased nombra la feature   R1 · route:inline

## History
L1 · "la herramienta corre antes que el modelo y su fallo es un defecto que no se puede ignorar"
L2 · OFICIO · "Una comprobacion que falla es un finding con un campo check y sin path, porque un path inventado seria una mentira en el registro" (porque una fila que nombra un archivo que no existe manda a arreglar el archivo equivocado; el campo dice de donde viene la fila y el lector sabe que no es una linea del diff)
L3 · OFICIO · "El veredicto sube a esquema 6 y review.toml a 4: un binario viejo nombra el archivo antes de leerlo mal" (porque el campo check y la tabla de comprobaciones llegan como version nueva, y el decodificador estricto no distingue un typo de un escritor que va adelante)
L4 · OFICIO · "La severidad de una comprobacion la declara el proyecto y no hay default" (porque un default que bloquea gasta una decision que nadie tomo: un gofmt y un govulncheck no valen lo mismo, y solo el proyecto lo sabe)
L5 · OFICIO · "Una comprobacion declarada y no instalada es una fila que dice que no existe, nunca una instalacion ni un fallo de la corrida" (porque odachi no instala nada en una maquina ajena, y una comprobacion que no corre en silencio es peor que una que falla fuerte)
