# el-revisor-juzga
Base: a274e13d89458ad9d1b2aa861cff27f17d9f4516

## Requirements (verbatim, from the user)
R1 · "el revisor juzga un arbol congelado al que puede leer con herramientas acotadas, sin bash ni worktree vivo, sin crecer el prompt, y la mejora se mide con defectos sembrados, un canario de acceso y el prompt grabado"

## Done when
- odachi-reviewer puede leer, glob y grep sobre el snapshot congelado y el veredicto nombra ese arbol
- bash, edit y las herramientas de web siguen denegadas al revisor
- max_prompt_kb sigue en 120 y el prompt de un candidato full no crece
- la sonda registra defectos sembrados, brazo limpio, canario de acceso y el prompt materializado

## Steps
- [x] T1 · el snapshot congelado: el revisor lee el arbol con herramientas acotadas y el veredicto lo nombra   R1 · route:inline · tree:3dc18bf0a0dc07bea7f1f358a7e531b571cbe0bf
- [x] T2 · la vara del prompt: max_prompt_kb queda en 120 y una prueba fija que el prompt de un candidato full no crece   R1 · route:inline · tree:83c84706e781a6fd8094d58aa289abf8b5f45564
- [x] T3 · la sonda: canario de acceso, brazo limpio y prompt materializado guardado con cada respuesta   R1 · route:inline · tree:5200e38f9b54711977d7cb4d07550f8e63ba5e13
- [x] T4 · el changelog y la nota de cierre   R1 · route:inline · tree:4ebdb3375042e838ab6dc528ab97978b2074b48b

## History
L1 · "el revisor juzga un arbol congelado al que puede leer con herramientas acotadas, sin bash ni worktree vivo, sin crecer el prompt, y la mejora se mide con defectos sembrados, un canario de acceso y el prompt grabado"
