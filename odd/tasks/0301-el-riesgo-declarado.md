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
- [x] T1 · la tabla riesgo a lente: cada riesgo de commits contesta su lente, con su test   R1 · route:inline · tree:c75d2a303c0e3cbaba99307417c3e3bc115e2df4
- [x] T2 · la demanda por archivo lleva las lentes declaradas, sin duplicar   R1 · route:inline · tree:de3ef7aebaf72fa7c8b55414571c4c0dd50a2053
- [x] T3 · la puerta pasa el riesgo de los commits y la negativa nombra el comando   R1 · route:inline · tree:def0f13898fd4af349165ad63f6ca2b70a1067a9
- [x] T4 · el changelog nombra 0301 bajo Unreleased, minor   R1 · route:inline · tree:ae367bdbc11a57543fd78a84b40f62eedb3bb9bb

## History
L1 · "el riesgo declarado en los commits agrega su lente a lo que la puerta exige, y un veredicto sin esa lente no cierra"
