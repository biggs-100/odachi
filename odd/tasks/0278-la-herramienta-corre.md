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
- [x] T2 · el review corre cada comprobacion en el worktree del candidato y el prompt lleva el resultado antes de cualquier lente   R1 · route:inline · tree:62f4ca56abc94cc777c9d8d5bb959d7a894d7255
- [x] T3 · una comprobacion que falla es un finding de la herramienta, con su salida y su severidad   R1 · route:inline · tree:c0c821ab17d9d610c520e5696f5196c5b60e6613
- [ ] T4 · el veredicto dice lo que corrio: --show nombra las comprobaciones y Unreleased nombra la feature   R1 · route:inline

## History
L1 · "la herramienta corre antes que el modelo y su fallo es un defecto que no se puede ignorar"
L2 · OFICIO · "Una comprobacion que falla es un finding con un campo check y sin path, porque un path inventado seria una mentira en el registro" (porque una fila que nombra un archivo que no existe manda a arreglar el archivo equivocado; el campo dice de donde viene la fila y el lector sabe que no es una linea del diff)
L3 · OFICIO · "El veredicto sube a esquema 6 y review.toml a 4: un binario viejo nombra el archivo antes de leerlo mal" (porque el campo check y la tabla de comprobaciones llegan como version nueva, y el decodificador estricto no distingue un typo de un escritor que va adelante)
L4 · OFICIO · "La severidad de una comprobacion la declara el proyecto y no hay default" (porque un default que bloquea gasta una decision que nadie tomo: un gofmt y un govulncheck no valen lo mismo, y solo el proyecto lo sabe)
L5 · OFICIO · "Una comprobacion declarada y no instalada es una fila que dice que no existe, nunca una instalacion ni un fallo de la corrida" (porque odachi no instala nada en una maquina ajena, y una comprobacion que no corre en silencio es peor que una que falla fuerte)
L6 · OFICIO · "El contador de esquema es uno solo: la feature lo sube a 6 y los tres archivos lo declaran" (porque el campo check y la tabla de comprobaciones llegan a la vez, y un contador por archivo dejaria pasar un archivo que el binario no lee entero; la linea anterior que hablaba de review.toml en 4 queda corregida por esta)
L7 · OFICIO · "La feature es minor y no patch: cambia la forma de un registro guardado, y docs/releases.md llama contrato a eso" (porque el veredicto sube de esquema y review.toml gana una tabla; un binario viejo tiene que decir actualiza, y eso es la senal visible de que algo puede tener que adaptarse)
L8 · OFICIO · "Una fila de herramienta se desestima con razon aunque sea causal y bloquee" (porque un analizador no razona: no hay cita que refutar, y la unica refutacion posible es que la regla no aplique a este codigo; la razon queda escrita al lado del veredicto con quien y cuando, asi que la compuerta no se abre por omision)
