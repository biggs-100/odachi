# el-brazo-limpio
Base: ad20f5dde70cd28be67d6cbf007ee18da5b7088d

## Requirements (verbatim, from the user)
R1 · "el brazo limpio del fixture paga las dos mejoras que el brazo ciego reclamo: la rama del 4xx no reintentable tiene su prueba, y el backoff espera cancelable y no duerme despues del ultimo intento"

## Done when
- un test del fixture prueba que un 404 devuelve el error sin reintentar, y el fixture compila y pasa sus pruebas en un modulo aparte
- el backoff no duerme despues del ultimo intento y una cancelacion durante la espera corta el fetch antes de agotar los intentos, con su test
- la pagina de la medicion dice que el examen se movio despues del numero y nombra las dos mejoras pagadas
- go test ./... is green and odachi check --run go test ./... is green

## Steps
- [ ] T1 · la prueba del 4xx en el brazo limpio, y el fixture compila y pasa sus pruebas en un modulo aparte   R1 · route:inline
- [ ] T2 · el backoff espera cancelable y no duerme despues del ultimo intento   R1 · route:inline
- [ ] T3 · la pagina dice que el examen se movio despues del numero, y el changelog   R1 · route:inline

## History
L1 · "el brazo limpio del fixture paga las dos mejoras que el brazo ciego reclamo: la rama del 4xx no reintentable tiene su prueba, y el backoff espera cancelable y no duerme despues del ultimo intento"
