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
- [ ] T1 · la cerca deja de cerrar internal/domain/gate/** para este motivo, y esa politica aterriza sola antes que el codigo   R1 · route:inline
- [ ] T2 · la puerta aprende el estado: un arbol cuyos tests no fallan al deshacer el cambio no cierra   R1 · route:inline
- [ ] T3 · check corre el comando sobre una copia con el cambio deshecho y guarda la corrida como evidencia   R1 · route:inline

## History
L1 · "check mide si el test puede fallar, no solo si el comando paso: si el suite sigue verde al deshacer el cambio, el arbol no cierra"
