# odachi-init-skills

## Qué tiene que existir
- odachi init --skills home deja todas las skills en ~/.agents/skills y el proyecto no recibe copia

## Listo cuando
- el home tiene todas las skills y el proyecto no tiene .agents/skills
- una segunda corrida no escribe nada
- una copia previa del proyecto se avisa y no se borra
- doctor cuenta las skills del home y avisa si driftan
- go test ./... queda en verde
- odachi check queda listo

## NO incluye (primera versión)

## Decisiones tomadas
- Se usa desde la compu.
- No hay datos sensibles en la primera versión.
- Primera versión usable: esta semana.
- Quiere aprender: skills globales.
