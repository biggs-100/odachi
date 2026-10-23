# review-proporcional-al

## Requisitos (textual, del usuario)
R1 · "review proporcional al riesgo y con modo de proyecto: pasivo no pide juicio, el humano prende o apaga, y el riesgo mira contratos y concurrencia"

## Listo cuando
- check --require-review no pide veredicto para un cambio solo de docs
- review mode on|off|status persiste en .odachi/review.toml y check lo respeta; los flags lo override
- Classify marca full para contratos, concurrencia y entrega, no solo auth y pagos
- go test ./... queda en verde
- odachi check queda listo

## Pasos
- [x] T1 · saltar el review en candidatos pasivos   R1 · 9b76b93
- [x] T2 · review mode on/off/status y check   R1 · 1a0160e
- [x] T3 · ampliar las senales de riesgo   R1 · db33de9

## Historial
L1 · "review proporcional al riesgo y con modo de proyecto: pasivo no pide juicio, el humano prende o apaga, y el riesgo mira contratos y concurrencia"
L2 · OFICIO · "El modo de review vive en .odachi/review.toml y check lo lee; los flags lo override por una corrida" (porque el switch es del humano y el hook ya corre check, asi que no hay que editar el hook ni agregar un flag en cada comando)
L3 · OFICIO · "Un candidato pasivo no pide veredicto aunque el modo este on" (porque no hay lente que responder; pedir un review de docs es ceremonia, no evidencia)
L4 · OFICIO · "Las senales de riesgo siguen siendo por path, sin assessor LLM" (porque el arbol ya es el candidato y el path es suficiente para decidir cuantas lentes pedir sin agregar una llamada que pueda fallar)
