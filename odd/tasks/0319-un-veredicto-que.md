# un-veredicto-que
Base: 99c612ef82d0d949e152678aec5b21f8f4bf78ea

## Requirements (verbatim, from the user)
R1 · "un veredicto que pide cambios frena el cierre aunque sus filas no bloqueen, y la dispensa escrita por fila es la salida"

## Done when
- un veredicto changes_requested para el arbol frena el cierre aunque ninguna fila bloquee
- el cierre nombra el veredicto y la fila que lo sostiene
- la dispensa por fila deja pasar el cierre con la razon escrita
- un veredicto aprobado no frena nada
- go test ./... is green and odachi check --run "go test ./..." is green

## Steps
- [x] T1 · la puerta: un veredicto que pide cambios frena el cierre aunque ninguna fila bloquee   R1 · tree:f794b3a98f8034fb5f39d8a532b0fcd85f73ca0d
- [x] T2 · la dispensa: la fila que sostiene el veredicto se deja pasar con la razon escrita   R1 · tree:f794b3a98f8034fb5f39d8a532b0fcd85f73ca0d
- [ ] T3 · el changelog y la nota   R1

## History
L1 · "un veredicto que pide cambios frena el cierre aunque sus filas no bloqueen, y la dispensa escrita por fila es la salida"
