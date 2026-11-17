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
- [ ] T2 · el fixture gana una base y un defecto que solo se lee en un llamador de esa base   R1
- [ ] T3 · los dos brazos se corren y el numero va a docs/reviewer-context-measurement.md   R1

## History
L1 · "la sonda sostiene los dos brazos del contexto del revisor y la comparacion esta medida: cada mano recibe su snapshot congelado y el revisor corre fuera de las dos, el fixture gana una base con un defecto que solo se ve leyendo un llamador, y el numero de antes y despues queda escrito con su seccion de limites"
L2 · OFICIO · "La sonda sostiene dos brazos -el snapshot congelado y el ciego- y el fixture gana una base con un defecto de contexto" (porque Sin base el candidato entero va en el prompt y las dos manos no pueden diferir; el humano aprobo aflojar la regla del fixture para que el examen pueda responder la pregunta de 0320)
