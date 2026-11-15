# portar-ai-project
Base: 2505e966162a02b7bbe69ea8376668fa738e828c

## Requirements (verbatim, from the user)
R1 · "portar ai project foundation completo a odachi, con una skill nueva arch-project para cualquier stack"

## Done when
- odachi skills --check pasa y el indice lista arch-project con su trigger
- las cuatro piezas que hoy no tienen dueno entran: el layout hexagonal, la plantilla de ADR, la matriz de comandos por stack y la guia de Go + HTMX
- bootstrap queda como la primera puerta y arch-project como la segunda, y ninguna repite el material de la otra
- una prueba fija el trigger de arch-project y su tabla de decision, y la prueba que fija bootstrap sigue verde
- docs/impact-map.md y SKILLS.md nombran la carpeta nueva, y el pin de bytes sube con el numero que mide el test
- go test ./... is green and odachi check --run "go test ./..." is green
- ## Unreleased names it, patch (0307)

## Steps
- [x] T1 · arch-project nace con sus cinco references movidos y bootstrap rutea a ella   R1 · tree:73f61a9cb34738ad1614917f6a592cbd51e02854
- [x] T2 · las cuatro piezas sin dueno entran: layout, ADR, matriz por stack y Go + HTMX   R1 · tree:b79753b5780274507621914dce23f75388f51715
- [x] T3 · el indice, el impact map, la prueba y el changelog nombran la puerta nueva   R1 · tree:af67f9be313d4d9adc68d461135a3e725c45da6a

## History
L1 · "portar ai project foundation completo a odachi, con una skill nueva arch-project para cualquier stack"
L2 · OFICIO · "los cinco references se mueven a arch-project y no se copian" (porque dos carpetas con el mismo material son dos duenos, y el catalogo ya pago ese defecto en 0302 y 0304)
L3 · OFICIO · "la skill se llama arch-project y no ai-project-foundation" (porque el nombre viejo es el que colisiona con el trigger de bootstrap; arch-project forma familia con arch-tauri y deja sitio a arch-<stack>)
