# un-cierre-que
Base: 101490ae54d7614fb0611a4a714d569f3136260b

## Requirements (verbatim, from the user)
R1 · "un cierre que trae un veredicto que pide cambios lo dice en la linea del gate y solo pasa con una dispensa humana escrita"

## Done when
- un cierre con un hallazgo bloqueante es NeedsDecision y la linea nombra el hallazgo y la dispensa que lo contesta
- odachi review waive --because registra la huella del hallazgo y quien dispensa, en el task doc
- con esa huella dispensada el cierre pasa, y la linea sigue diciendo que el review pidio cambios y quien lo dispenso
- un hallazgo distinto de un arbol posterior vuelve a frenar: la dispensa no se hereda
- go test ./... is green and odachi check --run "go test ./..." is green
- ## Unreleased names it, patch (0280)

## Steps
- [x] T1 · el gate dice que el review pidio cambios y nombra el hallazgo, la huella y la dispensa que lo contesta   R1 · route:inline · tree:63a21e4f4a358b604c5b70819ce351470bd57a92
- [ ] T2 · review waive registra la huella del hallazgo con la razon y quien, y se lee de vuelta   R1 · route:inline
- [ ] T3 · una dispensa que cubre la huella deja pasar el cierre, y una que no la cubre no   R1 · route:inline
- [ ] T4 · la cara: --show y el README dicen como se dispensa un hallazgo bloqueante   R1 · route:inline

## History
L1 · "un cierre que trae un veredicto que pide cambios lo dice en la linea del gate y solo pasa con una dispensa humana escrita"
L2 · OFICIO · "La dispensa se ata a la huella del hallazgo, no a la feature" (porque una dispensa que cubre la feature entera deja pasar cualquier fila que un review posterior encuentre, y eso es la compuerta abriendose por omision con otro nombre; con la huella, el humano contesta exactamente lo que leyo y el arbol que trae otra cosa vuelve a frenar)
L3 · OFICIO · "La dispensa vive en el task doc, donde ya vive la que existe, con quien y cuando en la misma linea" (porque el task doc viaja con el repositorio y es el registro que un lector posterior lee; el veredicto vive en el store del clon y no puede cargar la decision de una persona que no lo tiene)
