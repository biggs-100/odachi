# lo-que-un
Base: 1d67cf707ac3734ae5f2bd7e8bb1a26085899111

## Requirements (verbatim, from the user)
R1 · "lo que un agente lee en cada sesion queda fijado por bytes: crecer cuesta un commit deliberado, y acortarlo baja el pin"

## Done when
- el bloque que init inyecta, los cuatro lentes y SKILLS.md tienen su tamano en un baseline, y el suite falla cuando alguno crece
- un texto que se acorta hace fallar la prueba hasta que el pin baja, asi el numero no miente hacia arriba
- el baseline vive en un solo archivo y la comparacion en un solo lugar, no una copia por paquete
- go test ./... is green and odachi check --run "go test ./..." is green

## Steps
- [x] T1 · la comparacion en un solo lugar: paquete budget puro, dos direcciones, con sus tests   R1 · route:inline · tree:6b218261e145536fe77a0cad6558c42ba154fbe8
- [ ] T2 · los pines: bloque, lentes y SKILLS.md en un baseline, y el suite falla cuando crecen   R1 · route:inline

## History
L1 · "lo que un agente lee en cada sesion queda fijado por bytes: crecer cuesta un commit deliberado, y acortarlo baja el pin"
