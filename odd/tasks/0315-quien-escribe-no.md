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
- [ ] T2 · el verbo: odachi verify --exec compone el prompt del verificador, corre el comando y guarda la respuesta tal cual   R1
- [ ] T3 · la puerta: check pide el registro del árbol que nombra cuando los commits declaran un riesgo, y bloquea   R1
- [ ] T4 · el catálogo: la fila del verificador deja de decir none y doctor resuelve la razón nueva   R1
- [ ] T5 · el changelog y la nota de sesión nombran la feature y lo que deja   R1

## History
L1 · "quién escribe no se verifica a sí mismo: un riesgo que pide verifier deja un registro de una verificación independiente, y la puerta no cierra sin él"
