# un-revisor-que
Base: 55a88161a4a30b87eee217a1d244d674ea5b9928

## Requirements (verbatim, from the user)
R1 · "un revisor que sigue transmitiendo no se mata por el reloj total: la espera mide el silencio, con un plazo para la primera salida y otro para cada pausa, mientras la prueba sigue usando solo su plazo total"

## Done when
- un comando mudo se detiene en el plazo de primera salida y devuelve timedOut antes del plazo total
- un comando que escribe con pausas cortas termina sin timedOut aunque supere el plazo de primera salida
- un comando que calla despues de escribir se detiene en el plazo de inactividad
- check --run sigue midiendo solo el plazo total y no corta una prueba silenciosa
- go test ./... is green and odachi check --run go test ./... is green

## Steps
- [x] T1 · el reloj: RunInput mide primera salida e inactividad, Run y RunStream quedan igual   R1 · route:inline · tree:104da85867cce3985c3a17b5dacf9cb6ba0ff849
- [x] T2 · la prueba conserva su plazo total y una prueba lo fija   R1 · route:inline · tree:939533ac07002e9f0842d3647534bfda2ef7b3e0
- [ ] T3 · el changelog y la nota de cierre   R1 · route:inline

## History
L1 · "un revisor que sigue transmitiendo no se mata por el reloj total: la espera mide el silencio, con un plazo para la primera salida y otro para cada pausa, mientras la prueba sigue usando solo su plazo total"
L2 · OFICIO · "Abro el reloj de silencio (recomendacion 1 de la comparacion con gentle-shell) con esta frase como requisito" (porque El humano dijo ok, comencemos despues de que propuse el orden 1, 2, 3; acepto empezar por el primero y dejo la redaccion del requisito a mi recommendation. El reloj va solo en RunInput, que usan review y verify; Run y RunStream siguen con su plazo total para no matar un go test silencioso)
