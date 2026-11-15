# el-port-de
Base: 1dbd9fbbf0c93a4b90bb920b907d2416f951a2e9

## Requirements (verbatim, from the user)
R1 · "el port de Tauri queda completo: los nueve pasos del scaffold, las tres condiciones de no cargar, y arch-tauri mirando a arch-project"

## Done when
- odachi skills --check pasa y el cuerpo de arch-tauri sigue con 60 lineas o menos
- los nueve pasos viven en una reference ruteada desde ## Decision gates y listada en ## References
- las tres condiciones de no cargar estan en el cuerpo, y la que habla de un proyecto sin Rust nombra arch-project
- tauri-bridge.md manda el layout y los ADRs a arch-project, no a bootstrap
- una prueba fija el ruteo de la reference nueva y el nivel de la puerta
- docs/impact-map.md cuenta la reference nueva y el changelog la nombra
- go test ./... is green and odachi check --run "go test ./..." is green
- ## Unreleased names it, patch (0308)

## Steps
- [x] T1 · la reference del scaffold entra con el orden de los nueve pasos y su fila   R1 · tree:07a4a8487f07d9832d0f68dfb6d3b34116e88042
- [ ] T2 · la puerta deja de mirar a bootstrap: las tres condiciones y el nivel   R1
- [ ] T3 · la prueba, el impact map y el changelog nombran el port completo   R1

## History
L1 · "el port de Tauri queda completo: los nueve pasos del scaffold, las tres condiciones de no cargar, y arch-tauri mirando a arch-project"
