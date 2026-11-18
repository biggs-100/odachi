# skills-evals
Base: 47706a2687a4d3941b71e97eaeaba90427495212

## Requirements (verbatim, from the user)
R1 · "un examen de skill se corre con un comando y deja su puntaje guardado: dos manos con el mismo prompt -una con la skill y otra sin ella-, el puntaje por comportamiento leido del transcript, y el registro en el store del clone"

## Done when
- `odachi skills eval tests --exec "<command>" --runs 3` corre las dos manos, imprime una fila por corrida con los comportamientos que el transcript decide, y guarda el registro en `.git/odachi/evals/`
- la mano sin skill no carga ninguna: su transcript no muestra una carga y el canario lo responde, sin tocar `~/.agents/skills`
- un puntaje nombra el evento que lo decidio: el test escrito antes del codigo, y `go test` corrido entre los dos
- el mismo examen sobre los mismos bytes otra vez se niega o reusa el registro: lo pagado no se paga dos veces
- `go test ./...` is green and `odachi check --run "go test ./..."` is green

## Steps
- [ ] T1 · el instrumento prueba como correr la mano sin skill sin tocar el home, y el verbo se niega si no hay forma limpia   R1 · route:inline
- [ ] T2 · el examen de tests corre las dos manos con la exec del humano y puntua el orden de los eventos   R1 · route:inline
- [ ] T3 · el puntaje se guarda en el store del clone, clean lo retiene y una corrida repetida se niega   R1 · route:inline
- [ ] T4 · la pagina del bucle de evals y el changelog   R1 · route:inline

## History
L1 · "un examen de skill se corre con un comando y deja su puntaje guardado: dos manos con el mismo prompt -una con la skill y otra sin ella-, el puntaje por comportamiento leido del transcript, y el registro en el store del clone"
