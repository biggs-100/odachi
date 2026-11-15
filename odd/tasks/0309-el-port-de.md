# el-port-de
Base: 20fb16ae79f0aea4ff23d44a5162c55a67bb1ff3

## Requirements (verbatim, from the user)
R1 · "el port de gentle-shell cierra sus huecos de texto: el aviso que perdio tui-design y los dos references de UI que ui-craft no tiene"

## Done when
- odachi skills --check pasa y tui-design lleva el aviso de atribucion
- ui-craft lleva adaptive-interfaces y composition-and-feedback, ruteados desde ## Decision gates y listados en ## References
- el cuerpo de ui-craft sigue con 60 lineas o menos y su tabla tiene una fila por reference
- una prueba fija el ruteo de los dos references nuevos y la presencia del aviso
- docs/impact-map.md cuenta los references nuevos y el changelog lo nombra
- go test ./... is green and odachi check --run "go test ./..." is green
- ## Unreleased names it, patch (0309)

## Steps
- [x] T1 · el aviso de atribucion vuelve a tui-design como asset   R1 · tree:3ebe691cd8da489d76371375cd58eeee9a2c40f3
- [x] T2 · los dos references de UI entran en ui-craft con su fila y su bullet   R1 · tree:af835c05726598442c4656909a3ccdfa3c5c7cdd
- [ ] T3 · la prueba, el impact map y el changelog nombran los huecos cerrados   R1

## History
L1 · "el port de gentle-shell cierra sus huecos de texto: el aviso que perdio tui-design y los dos references de UI que ui-craft no tiene"
L2 · PREGUNTA P1 · "odachi exige commits convencionales (regex type(scope): description con labels type:*) como gentle-shell, o conserva el mensaje libre con la linea Risk: que el gate ya lee"
L3 · VOS P1 · "odachi conserva el mensaje libre con la linea Risk: y no exige commits convencionales; la politica de PR de branch-pr no se porta"
L4 · OFICIO · "el aviso de atribucion no se perdio: vive en el cuerpo de tui-design, y T1 fija su presencia con una prueba en vez de mover un archivo" (porque el port lo movio al cuerpo a proposito porque el instalador solo copia SKILL.md, references y assets; el diff de nombres de archivo lo leyo como hueco, y esa es la leccion de la auditoria)
