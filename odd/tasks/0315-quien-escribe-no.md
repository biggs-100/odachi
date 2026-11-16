# quien-escribe-no
Base: f78794a9f7cbaede695fa2a3cc6f34e7f820c67b

## Requirements (verbatim, from the user)
R1 · "quién escribe no se verifica a sí mismo: un riesgo que pide verifier deja un registro de una verificación independiente, y la puerta no cierra sin él"

## Done when
- un riesgo declarado en un commit pide su verificador, y la puerta bloquea nombrando el árbol y el riesgo cuando no hay registro
- el registro guarda el árbol que verificó, quién lo escribió y la hora en que lo selló, como el veredicto
- un registro de otro árbol no cubre este, y un registro que no dice quién lo escribió no cuenta
- el comando que produce el registro corre una verificación independiente con el agente verificador del perfil y deja la respuesta tal cual la dio
- doctor cuenta la fila del catálogo que deja de decir none
- go test ./... is green and odachi check --run "go test ./..." is green

## Steps
- [x] T1 · el registro: el dominio sella esquema, árbol, autor y hora, lee estricto, y una respuesta sin autor no cubre   R1 · tree:5f6f1534edfa27557fbc337512c081b69c2ec736
- [x] T2 · el verbo: odachi verify --exec compone el prompt del verificador, corre el comando y guarda la respuesta tal cual   R1 · tree:59f0fdc750234b2aee9ceb96a13a1d5d46a33c67
- [x] T3 · la puerta: check pide el registro del árbol que nombra cuando los commits declaran un riesgo, y bloquea   R1 · tree:f7294576c9962f0fe133def54e9d95f53697b899
- [x] T4 · el catálogo: la fila del verificador deja de decir none y doctor resuelve la razón nueva   R1 · tree:fdcff4ff296b85776ad91208ef995cac17b30f84
- [x] T5 · el changelog y la nota de sesión nombran la feature y lo que deja   R1 · tree:15c3306d6a7bd4e6975b5685d6997fe42d915b74

## History
L1 · "quién escribe no se verifica a sí mismo: un riesgo que pide verifier deja un registro de una verificación independiente, y la puerta no cierra sin él"
L2 · OFICIO · "El registro del verificador lleva su propio contador de esquema (verify.Schema), no el compartido" (porque el contador de schema.Current es de cinco archivos que un binario viejo lee; un sexto archivo que empieza aquí no debe mover ninguno de los cinco, y un binario que nunca oyó de él nunca lo lee)
L3 · OFICIO · "La exigencia vive en el gate y no se suaviza con el modo del review" (porque el requisito dice que la puerta no cierra sin el registro; el review se convierte en nota cuando el proyecto lo pide (ADR 0043), y una verificación que también fuera nota sería el silencio que la feature existe para romper)
L4 · OFICIO · "La independencia del verificador es el rol y el nombre escrito, no una comparación con el modelo del builder" (porque esta máquina corre un modelo para todo, y comparar los nombres cerraría la cola del pstack contra su ítem siguiente; el registro deja el nombre escrito y el día que haya un segundo vendor la comparación es una línea)
L5 · OFICIO · "El prompt del verificador pide sondas enfocadas y no el suite completo" (porque la primera corrida de campo gastó sus treinta minutos en go test ./... y no respondió nada; el suite es del gate, y el catálogo ya decía que el verificador corre comandos enfocados para no competir por la máquina)
