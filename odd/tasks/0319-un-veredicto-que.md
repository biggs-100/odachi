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
- [ ] T1 · la puerta: un veredicto que pide cambios frena el cierre aunque ninguna fila bloquee   R1
- [ ] T2 · la dispensa: la fila que sostiene el veredicto se deja pasar con la razon escrita   R1
- [ ] T3 · el changelog y la nota   R1

## History
L1 · "un veredicto que pide cambios frena el cierre aunque sus filas no bloqueen, y la dispensa escrita por fila es la salida"
