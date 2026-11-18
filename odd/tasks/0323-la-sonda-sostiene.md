# la-sonda-sostiene
Base: 970de471210c059596cf26f1aeae38407ed5a11d

## Requirements (verbatim, from the user)
R1 · "la sonda sostiene los dos brazos del contexto del revisor y la comparacion esta medida: cada mano recibe su snapshot congelado y el revisor corre fuera de las dos, el fixture gana una base con un defecto que solo se ve leyendo un llamador, y el numero de antes y despues queda escrito con su seccion de limites"

## Done when
- una corrida con --exec nombra el directorio congelado en cada prompt, y el canario responde attached and scoped
- el fixture tiene un defecto que no aparece en el diff: el brazo que lee el snapshot lo cuenta como encontrado y el brazo ciego lo cuenta como missed
- el revisor corre en un directorio que no contiene la otra mano, asi que un listado del cwd no alcanza seeded/
- una pagina nueva lleva el numero de los dos brazos con sus draws, el modelo y una seccion de lo que no muestra
- go test ./... is green and odachi check --run go test ./... is green

## Steps
- [x] T1 · cada mano se congela en su snapshot y el revisor corre dentro; --blind corre el brazo sin snapshot   R1 · tree:b30122a82edaebad83c2d557aff97b0809071c43
- [x] T2 · el fixture gana una base y un defecto que solo se lee en un llamador de esa base   R1 · tree:18f95e2f8255d6ea6dd7b2b81c9d9ebc96a6cfc4
- [x] T3 · los dos brazos se corren y el numero va a docs/reviewer-context-measurement.md   R1 · tree:24d3046ebac1c32760f92cd04a2da1408fef92dd
- [x] T4 · una respuesta que narra antes del JSON se lee, no se rechaza   R1 · tree:225c580795076553f1651798e205467504587dd0
- [x] T5 · el mundo del revisor es su propio temp: desde su cwd, .. no alcanza la otra mano   R1 · tree:225c580795076553f1651798e205467504587dd0

## History
L1 · "la sonda sostiene los dos brazos del contexto del revisor y la comparacion esta medida: cada mano recibe su snapshot congelado y el revisor corre fuera de las dos, el fixture gana una base con un defecto que solo se ve leyendo un llamador, y el numero de antes y despues queda escrito con su seccion de limites"
L2 · OFICIO · "La sonda sostiene dos brazos -el snapshot congelado y el ciego- y el fixture gana una base con un defecto de contexto" (porque Sin base el candidato entero va en el prompt y las dos manos no pueden diferir; el humano aprobo aflojar la regla del fixture para que el examen pueda responder la pregunta de 0320)
L3 · OFICIO · "Las dos manos se miden con el mismo binario y la misma fixture, y el brazo ciego hereda el prompt de hoy sin la seccion del arbol congelado" (porque Cambiar el prompt a mitad de la corrida haria que cada mano la midiera un binario distinto; el brazo ciego no es una re-edicion del prompt de antes de 0320 sino un control de la misma pregunta con el directorio vacio, y la pagina lo dice en sus limites)
L4 · OFICIO · "La corrida de los dos brazos queda anulada: el brazo ciego encontro el defecto de contexto citando auth.go sin tener snapshot, porque las manos viven en work y el revisor corre en un directorio vecino; y el revisor narra antes del JSON, asi que la sonda lo rechaza" (porque Una corrida cuyo brazo ciego lee el examen no mide el contexto: mide el sistema de archivos. Las respuestas quedan en el store como evidencia del defecto, no como numero)
L5 · OFICIO · "La medida de 0323 queda bloqueada en su T3: el revisor no esta acotado por el cliente, y el spike del 0324 muestra que external_directory deny en la configuracion del proyecto alcanza para acotarlo sin quitarle glob ni grep" (porque El brazo ciego cito auth.go:16 sin snapshot y glob busco en C:/.../Temp: la garantia de 0320 es una regla del prompt, no una frontera. Medir sin acotar mide lo que el modelo decide hacer, no lo que el snapshot compra)
