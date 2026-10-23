# cuando-start-no

## Requisitos (textual, del usuario)
R1 · "cuando start no puede leer una respuesta, el mensaje nombra los flags que faltan en vez de morir con 'no pude leer la respuesta'"

## Listo cuando
- un start con EOF en la primera pregunta sale 2 y nombra --outcome, --audience, --done, --sensitive y --deadline
- go test ./... queda en verde
- odachi check queda listo

## Pasos
- [x] T1 · devolver needsAnswer cuando el interview falla   R1 · 46dfa09

## Historial
L1 · "cuando start no puede leer una respuesta, el mensaje nombra los flags que faltan en vez de morir con 'no pude leer la respuesta'"
