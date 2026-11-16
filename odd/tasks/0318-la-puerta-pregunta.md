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
- [x] T1 · la pregunta: la puerta pide el registro contra el diff de la feature, no contra el candidato del cierre   R1 · tree:4b42c14474b4979943e1f5c70cc5638d50601e33
- [x] T2 · el test: un cierre de documentos con riesgo declarado lo nombra, y una feature pasiva no   R1 · tree:4b42c14474b4979943e1f5c70cc5638d50601e33
- [x] T3 · el changelog y la nota   R1 · tree:358da851e001a3ed8d814f512b3096c0e6f6256b

## History
L1 · "la puerta pregunta por el registro del verificador contra el diff completo de la feature y no contra el candidato que el cierre trae: un cierre que solo mueve documentos también lo nombra"
