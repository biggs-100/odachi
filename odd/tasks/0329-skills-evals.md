# skills-evals
Base: 47706a2687a4d3941b71e97eaeaba90427495212

## Requirements (verbatim, from the user)
R1 · "un examen de skill se corre con un comando y deja su puntaje guardado: dos manos con el mismo prompt -una con la skill y otra sin ella-, el puntaje por comportamiento leido del transcript, y el registro en el store del clone"

## Done when
- `odachi skills eval tests --exec "<command>" --runs 3` corre las dos manos, imprime una fila por corrida con los comportamientos que el transcript decide, y guarda el registro en `.git/odachi/evals/`
- la mano sin skill no carga ninguna: el verbo pide permiso explicito, mueve la copia del home y la restaura siempre, y el canario lo responde
- un puntaje nombra el evento que lo decidio: el test escrito antes del codigo, y `go test` corrido entre los dos
- el mismo examen sobre los mismos bytes otra vez se niega o reusa el registro: lo pagado no se paga dos veces
- `go test ./...` is green and `odachi check --run "go test ./..."` is green

## Steps
- [x] T1 · el instrumento prueba como correr la mano sin skill sin tocar el home, y el verbo se niega si no hay forma limpia   R1 · route:inline · tree:d871b0e989814401b7d0c0b47f0dad9595c2206f
- [x] T2 · el examen de tests corre las dos manos con la exec del humano y puntua el orden de los eventos   R1 · route:inline · tree:a9569e2d90040dff7786a4944bbbdae997c1d81d
- [x] T3 · el puntaje se guarda en el store del clone, clean lo retiene y una corrida repetida se niega   R1 · route:inline · tree:5de48f61da36a45120c4ce1960dff752c4144e50
- [x] T4 · la pagina del bucle de evals y el changelog   R1 · route:inline · tree:67635a09fc4f0896ddd6ec1f4788cb6fa68ffdaf

## History
L1 · "un examen de skill se corre con un comando y deja su puntaje guardado: dos manos con el mismo prompt -una con la skill y otra sin ella-, el puntaje por comportamiento leido del transcript, y el registro en el store del clone"
L2 · OFICIO · "El aislamiento de la mano sin skill no se puede hacer con un home copiado: sin la db la corrida no autentica (401) y con la db trae el estado real, asi que el cliente sigue leyendo el home de verdad" (porque El canario en el .agents/skills del proyecto se carga y responde BANANA; el mismo canario en el home del sandbox no se carga. La unica exclusion limpia que ofrece el esquema es aditiva (skills.paths); no hay switch para no leer el home)
L3 · OFICIO · "La mano sin skill se aisla pidiendo permiso y moviendo la copia del home, con diario y restauracion siempre" (porque El humano eligio entre negarse, un stub y mover: copiar el home no autentica sin la db y con ella trae el estado real, y un stub dejaria el nombre de la skill en juego. Mover conserva el sentido de ausente, y el permiso explicito mas el diario lo hacen reversible)
