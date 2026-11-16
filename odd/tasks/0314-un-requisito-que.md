# un-requisito-que
Base: d46e621e3603f458e53d532de47e4afc6be608bc

## Requirements (verbatim, from the user)
R1 · "un requisito que no es la cita del humano no pasa la puerta: el check compara el R contra el bloque del brief que escribió start, y un requisito reescrito por el agente bloquea nombrando las dos líneas"

## Done when
- el check compara cada R del task doc contra el bullet de What must exist del bloque del brief de esa feature, y los dos son la misma frase
- un requisito reescrito por el agente bloquea nombrando la línea del brief y la del task doc
- una feature light o un task doc sin bloque de brief no se juzga, y la línea dice por qué
- odd/rules.toml nombra la regla con gate: y doctor resuelve esa clase contra las razones que la puerta declara
- go test ./... is green and odachi check --run "go test ./..." is green

## Steps
- [x] T1 · la comparacion: el dominio puro lee el bloque del brief y sus R, y devuelve las que no son la misma frase   R1 · route:inline · tree:1464cd83ae132e780802c946fb9d4e81d81130a6
- [x] T2 · la puerta: check bloquea nombrando la linea del brief y la del task doc, con su rojo observado antes   R1 · route:inline · tree:6d9f61a326e9b4860453332fa974a4c3dad833b4
- [x] T3 · la clase gate: doctor la resuelve contra las razones que la puerta declara, y la fila deja de decir none   R1 · route:inline · tree:358b686eaa308ef5aee17623281f1a87a7bb9ef8
- [x] T4 · el changelog y la nota de sesion   R1 · route:inline · tree:c4ab32083b2e582c89c1f1e263556096196cf9d8

## History
L1 · "un requisito que no es la cita del humano no pasa la puerta: el check compara el R contra el bloque del brief que escribió start, y un requisito reescrito por el agente bloquea nombrando las dos líneas"
