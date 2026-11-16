# la-puerta-pregunta
Base: 647067db9c4ba86f765bbac85692ac65a8369e2a

## Requirements (verbatim, from the user)
R1 · "la puerta pregunta por el registro del verificador contra el diff completo de la feature y no contra el candidato que el cierre trae: un cierre que solo mueve documentos también lo nombra"

## Done when
- un cierre que solo mueve documentos, de una feature con un riesgo declarado, nombra el registro del verificador
- una feature sin riesgo declarado, o con un diff pasivo, no pide registro
- la línea sigue siendo un aviso por defecto y un bloqueo con --require-review
- go test ./... is green and odachi check --run "go test ./..." is green

## Steps
- [ ] T1 · la pregunta: la puerta pide el registro contra el diff de la feature, no contra el candidato del cierre   R1
- [ ] T2 · el test: un cierre de documentos con riesgo declarado lo nombra, y una feature pasiva no   R1
- [ ] T3 · el changelog y la nota   R1

## History
L1 · "la puerta pregunta por el registro del verificador contra el diff completo de la feature y no contra el candidato que el cierre trae: un cierre que solo mueve documentos también lo nombra"
