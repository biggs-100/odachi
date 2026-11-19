# el-catalogo-deja
Base: 7bdc32a1e0be2e8909ef208dc9d24c6bb36a1758

## Requirements (verbatim, from the user)
R1 · "el catalogo deja de afirmar lo que no puede probar: el traspaso dice lo que la medicion mostro, el guard de citas mide cada pagina, la cita del socket tiene un solo dueno, y the-failures.md no cierra con una promesa que sus filas desmienten"

## Done when
- una pagina que pierde una linea Source falla el suite nombrando la pagina y la linea, y el baseline solo baja con una edicion deliberada
- la cita del socket aparece una sola vez, atribuida al documento que la dice, con su fecha, y la otra pagina apunta a ella
- the-failures.md no afirma mas de lo que sus filas sostienen
- docs/next-session.md dice que el movimiento ya se detecta y que una reescritura paga, con el numero medido
- go test ./... is green and odachi check --run "go test ./..." is green and odachi skills --check is green

## Steps
- [x] T1 · el traspaso dice que el movimiento se detecta y que una reescritura paga, con los numeros medidos   R1 · tree:39df0e8289581bd54656a33d3952935826228297
- [ ] T2 · una pagina que pierde una linea Source falla el suite, y el baseline solo baja a mano   R1
- [ ] T3 · la cita del socket tiene un solo dueno, comprobado en la fuente, y la otra pagina apunta a ella   R1
- [ ] T4 · the-failures.md cierra con lo que sus filas sostienen, no con una promesa mas grande   R1

## History
L1 · "el catalogo deja de afirmar lo que no puede probar: el traspaso dice lo que la medicion mostro, el guard de citas mide cada pagina, la cita del socket tiene un solo dueno, y the-failures.md no cierra con una promesa que sus filas desmienten"
