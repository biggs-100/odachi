# un-guard-se
Base: 2b5c90546d4003e91eb26701065d394cf6cf3382

## Requirements (verbatim, from the user)
R1 · "un guard se prueba rompiendolo: la corrida nombra el archivo y la cadena, el suite tiene que fallar diciendo esa cadena, y el archivo se restaura"

## Done when
- check --run acepta el archivo y la cadena, corre el suite en una copia sin la cadena, y exige que la salida contenga la cadena esperada
- si la copia queda verde, check bloquea y dice que el guard no noto nada
- si falla sin nombrar la cadena, check bloquea y muestra la salida: un rojo por otra razon no es evidencia
- si la cadena no esta en el archivo, se niega antes de correr: una mutacion que no muta no prueba nada
- la evidencia queda escrita en proof con el arbol, el archivo, la cadena, la esperada y la salida
- go test ./... is green and odachi check --run "go test ./..." is green
- ## Unreleased names it, patch (0297)

## Steps
- [x] T1 · el veredicto de la mutacion: lectura, render y las cuatro respuestas, con sus tests   R1 · route:inline · tree:2e120fac7b0cb98e90eac9469e054432b007e468
- [ ] T2 · la corrida: copia el arbol, quita la cadena, corre el suite y exige el mensaje   R1 · route:inline
- [ ] T3 · el changelog nombra 0297 bajo Unreleased, patch   R1 · route:inline

## History
L1 · "un guard se prueba rompiendolo: la corrida nombra el archivo y la cadena, el suite tiene que fallar diciendo esa cadena, y el archivo se restaura"
L2 · OFICIO · "La mutacion necesita una razon propia en el gate, y no se escribe al final de una sesion larga" (porque un guard que no dispara tiene que bloquear el cierre, y bloquear con una razon prestada seria mentir sobre la causa; la razon nueva entra en la tabla que 0293 dejo como unica lista, con su continuacion, y eso es un cambio de contrato que merece contexto fresco)
