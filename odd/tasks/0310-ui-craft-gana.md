# ui-craft-gana
Base: de26ee3880286e9748aba78b88ad93deb0581cea

## Requirements (verbatim, from the user)
R1 · "ui-craft gana el paso que le falta: direccion de arte, superficies de IA, motion con identidad y profundidad con funcion, para que el agente disene cosas de 2026 y no de 2020"

## Done when
- ui-craft lleva cuatro references nuevas, cada una con su fila en ## Decision gates y su bullet en ## References
- el cuerpo dice que una pantalla tiene una direccion elegida y nombrada, y sigue con 60 lineas o menos
- ai-native cubre streaming, transparencia de la herramienta, aprobacion antes de lo irreversible, undo y las tres formas de generative UI
- art-direction exige tres direcciones que difieran en estrategia, cada una con su compromiso declarado
- motion-identity trata la interaccion firma y que puede cargar; depth-and-material da la version correcta del vidrio y su costo
- the-numbers gana gama amplia y escala fluida
- una prueba fija el ruteo de las cuatro y su contenido
- docs/impact-map.md y el changelog nombran los cuatro
- go test ./... is green and odachi check --run "go test ./..." is green
- ## Unreleased names it, patch (0310)

## Steps
- [x] T1 · ai-native entra: las superficies que nadie disenaba antes de 2025   R1 · tree:17d8444bf4503c1772cb95864580155bbf0b4e91
- [ ] T2 · art-direction entra: tres direcciones que difieran en estrategia   R1
- [ ] T3 · motion-identity y depth-and-material entran, con su modo de fallo   R1
- [ ] T4 · the-numbers gana gama amplia y escala fluida, y los registros lo nombran   R1

## History
L1 · "ui-craft gana el paso que le falta: direccion de arte, superficies de IA, motion con identidad y profundidad con funcion, para que el agente disene cosas de 2026 y no de 2020"
L2 · OFICIO · "la casa no tiene estilo: las referencias ensenan a elegir direccion, no imponen una" (porque un look propio del catalogo haria que todo proyecto se parezca, que es el defecto que the-tells ya castiga; el default que si entra es la postura (densa y editorial) cuando el brief no dice nada)
