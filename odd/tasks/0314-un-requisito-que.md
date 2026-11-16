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
- [ ] T1 · la comparacion: el dominio puro lee el bloque del brief y sus R, y devuelve las que no son la misma frase   R1 · route:inline
- [ ] T2 · la puerta: check bloquea nombrando la linea del brief y la del task doc, con su rojo observado antes   R1 · route:inline
- [ ] T3 · la clase gate: doctor la resuelve contra las razones que la puerta declara, y la fila deja de decir none   R1 · route:inline
- [ ] T4 · el changelog y la nota de sesion   R1 · route:inline

## History
L1 · "un requisito que no es la cita del humano no pasa la puerta: el check compara el R contra el bloque del brief que escribió start, y un requisito reescrito por el agente bloquea nombrando las dos líneas"
