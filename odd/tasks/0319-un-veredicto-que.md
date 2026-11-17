# un-veredicto-que
Base: 99c612ef82d0d949e152678aec5b21f8f4bf78ea

## Requirements (verbatim, from the user)
R1 · "un veredicto que pide cambios frena el cierre aunque sus filas no bloqueen, y la dispensa escrita por fila es la salida"

## Done when
- un veredicto changes_requested para el arbol frena el cierre aunque ninguna fila bloquee
- el cierre nombra el veredicto y la fila que lo sostiene
- la dispensa por fila deja pasar el cierre con la razon escrita
- un veredicto aprobado no frena nada
- go test ./... is green and odachi check --run "go test ./..." is green

## Steps
- [x] T1 · la puerta: un veredicto que pide cambios frena el cierre aunque ninguna fila bloquee   R1 · tree:f794b3a98f8034fb5f39d8a532b0fcd85f73ca0d
- [x] T2 · la dispensa: la fila que sostiene el veredicto se deja pasar con la razon escrita   R1 · tree:f794b3a98f8034fb5f39d8a532b0fcd85f73ca0d
- [x] T3 · el changelog y la nota   R1 · tree:ea1e865080f43894cdb07dcc2af77349880a1aa0

## History
L1 · "un veredicto que pide cambios frena el cierre aunque sus filas no bloqueen, y la dispensa escrita por fila es la salida"
L2 · PREGUNTA P1 · "Cómo deja el revisor de estar ciego (deuda 4 de review-research, Part 7): (1) candidate congelado en el prompt con read: deny, o revisor que lee el árbol; (2) con max_prompt_kb ya en 120, crece el prompt o se mantiene/baja; (3) qué medición justifica creerle más: fixture de defectos sembrados (con el brazo limpio arreglado) o sonda de antes/después sobre un candidato real"
L3 · PREGUNTA P2 · "La cerca (odd/fence.toml regla de 0298) protege docs/review-research.md: no se reescribe ninguna medicion ya publicada, y check la lee del base. ¿Donde escribo la pagina de decision P1: (a) un feature nuevo que levante la regla para ese archivo, (b) un archivo nuevo docs/reviewer-context-research.md sin tocar la cerca, o (c) solo el scratch del clon, que no viaja y se borra con clean"
L4 · VOS P1 · "la que tu me recomiendes"
L5 · VOS P2 · "la que tu me recomiendes"
L6 · OFICIO · "Elijo el arbol congelado con lectura acotada, sin bash ni worktree vivo, prompt sin crecer y medicion con defectos sembrados, canario y prompt grabado; y escribo la pagina en docs/reviewer-context-research.md" (porque El humano delego la eleccion. El arbol congelado con lecturas acotadas es la unica forma con evidencia en dos fuentes independientes que quita los falsos positivos sin romper la garantia del tree; no subir max_prompt_kb evita cambiar truncamiento por dilucion; y la medicion con brazo limpio, canario y prompt grabado es la que puede decir si algo mejoro. El archivo nuevo no toca la cerca que protege docs/review-research.md)
