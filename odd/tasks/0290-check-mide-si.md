# check-mide-si
Base: 9a1f6486de88adc47d04f239034b4c6873f7976d

## Requirements (verbatim, from the user)
R1 · "check mide si el test puede fallar, no solo si el comando paso: si el suite sigue verde al deshacer el cambio, el arbol no cierra"

## Done when
- un feature que cierra con tests que no fallan al deshacer el cambio no cierra: check lo nombra con el comando, la salida y el next
- la fila corre sobre el arbol que cierra y su evidencia queda escrita como una prueba: arbol, comando y salida
- un test que si falla al deshacer el cambio deja el arbol cerrar
- go test ./... is green and odachi check --run "go test ./..." is green

## Steps
- [x] T1 · la cerca deja de cerrar internal/domain/gate/** para este motivo, y esa politica aterriza sola antes que el codigo   R1 · route:inline · tree:113962368616b6a4b59558d8974202f05866db63
- [x] T2 · la puerta aprende el estado: un arbol cuyos tests no fallan al deshacer el cambio no cierra   R1 · route:inline · tree:1fcd2552fe839e9a7d8d24011801b9c0d7ebd68e
- [x] T3 · el veredicto del deshacer es una funcion pura: se entero, no se entero, o no se pudo juzgar   R1 · route:inline · tree:09ecec7e2892a310372e44c6cab787730d6e5411
- [x] T4 · el diff que se revierte son los .go de produccion: ni tests ni testdata   R1 · route:inline · tree:350c34ba30e451a5212a8f532df21febfa1d4c46
- [ ] T5 · check revierte en una copia y corre el mismo comando sobre ella   R1 · route:inline
- [ ] T6 · start acepta --waive-undo y lo guarda en el task doc con su razon   R1 · route:inline

## History
L1 · "check mide si el test puede fallar, no solo si el comando paso: si el suite sigue verde al deshacer el cambio, el arbol no cierra"
