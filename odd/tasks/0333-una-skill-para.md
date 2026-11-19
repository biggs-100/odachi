# una-skill-para
Base: 7def21b70e08792b9421731e6560fcbeb2e5a913

## Requirements (verbatim, from the user)
R1 · "una skill para trabajar como un profesional senior en todo lo que tenga que ver con contenedores, docker, podman, devops, etc, buenas practicas y errores que no se deben cometer"

## Done when
- odachi skills --check pasa y el indice lista container-craft con su trigger
- el cuerpo tiene 60 lineas o menos y cada regla se puede violar de forma observable en el diff
- las siete references estan ruteadas desde ## Decision gates y listadas en ## References, y toda regla sobre un sistema externo lleva version y fecha
- the-failures.md es un catalogo: el error, el sintoma, el arreglo y que lo habria atrapado antes
- una prueba fija el trigger y el ruteo de cada reference
- SKILLS.md se regenera, el pin de bytes sube con el numero que mide el test, y docs/impact-map.md lo nombra
- go test ./... is green and odachi check --run "go test ./..." is green

## Steps
- [ ] T1 · el cuerpo nace con sus decisiones y la tabla que rutea las siete paginas   R1 · route:delegated/odachi-worker · skills:tests
- [ ] T2 · las siete paginas llevan su version y su fecha, y el catalogo de fallas cierra la puerta   R1 · route:delegated/odachi-worker · skills:tests
- [ ] T3 · la prueba fija el trigger y el ruteo, y el indice con su pin bajan el numero que el test mide   R1 · route:delegated/odachi-worker · skills:tests

## History
L1 · "una skill para trabajar como un profesional senior en todo lo que tenga que ver con contenedores, docker, podman, devops, etc, buenas practicas y errores que no se deben cometer"
