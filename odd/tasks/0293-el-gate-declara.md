# el-gate-declara
Base: 5532e6cbbf59e7ab5586bada454f07b9b5abb0b3

## Requirements (verbatim, from the user)
R1 · "el gate declara sus razones una vez y la prueba las recorre todas: una razon nueva no entra sin un estado que la produzca ni un next que la resuelva"

## Done when
- la prueba de continuacion recorre las filas de la unica tabla, y ninguna razon queda fuera del bucle
- una razon nueva en el fuente del gate sin fila en la tabla hace fallar el suite
- cada razon bloqueada tiene un next que empieza con odachi, y ready no tiene ninguno
- go test ./... is green and odachi check --run "go test ./..." is green
- ## Unreleased names it, patch (0293)

## Steps
- [x] T1 · la tabla de ejemplos es la unica lista: la continuacion se prueba recorriendo sus filas   R1 · route:inline · tree:27171dd49ad1ccaf6f5fef0ac9eaa00747ce9207
- [x] T2 · una razon nueva en el fuente del gate sin fila en la tabla hace fallar el suite   R1 · route:inline · tree:bc59cced2e515487c3dfb8a6282dabed48670acf
- [x] T3 · el changelog nombra 0293 bajo Unreleased   R1 · route:inline · tree:3b1eb36023b31cf5a7576da9766da224a8bb3ba5

## History
L1 · "el gate declara sus razones una vez y la prueba las recorre todas: una razon nueva no entra sin un estado que la produzca ni un next que la resuelva"
