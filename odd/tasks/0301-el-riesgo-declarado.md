# el-riesgo-declarado
Base: a5ff71099235571a07ac614e6114f42fc7df42a8

## Requirements (verbatim, from the user)
R1 · "el riesgo declarado en los commits agrega su lente a lo que la puerta exige, y un veredicto sin esa lente no cierra"

## Done when
- review mapea cada riesgo de la tabla de commits a su lente, con su test
- un cambio cuyos commits declaran security y un veredicto que no contesta risk bloquea nombrando odachi review --lens risk
- el mismo cambio con la lente contestada cierra
- un riesgo none o sin linea no agrega nada, y la lente que el piso ya pide no se duplica
- go test ./... is green and odachi check --run "go test ./..." is green
- ## Unreleased names it, minor (0301)

## Steps
- [ ] T1 · la tabla riesgo a lente: cada riesgo de commits contesta su lente, con su test   R1 · route:inline
- [ ] T2 · la demanda por archivo lleva las lentes declaradas, sin duplicar   R1 · route:inline
- [ ] T3 · la puerta pasa el riesgo de los commits y la negativa nombra el comando   R1 · route:inline
- [ ] T4 · el changelog nombra 0301 bajo Unreleased, minor   R1 · route:inline

## History
L1 · "el riesgo declarado en los commits agrega su lente a lo que la puerta exige, y un veredicto sin esa lente no cierra"
