# odachi-init-skills

## Requisitos (textual, del usuario)
R1 · "odachi init --skills home deja todas las skills en ~/.agents/skills y el proyecto no recibe copia"

## Listo cuando
- el home tiene todas las skills y el proyecto no tiene .agents/skills
- una segunda corrida no escribe nada
- una copia previa del proyecto se avisa y no se borra
- doctor cuenta las skills del home y avisa si driftan
- go test ./... queda en verde
- odachi check queda listo

## Pasos
- [x] T1 · agregar --skills home al init   R1 · dc49caf

## Historial
L1 · "odachi init --skills home deja todas las skills en ~/.agents/skills y el proyecto no recibe copia"
