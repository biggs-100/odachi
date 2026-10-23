# odachi-tiene-las

## Qué tiene que existir
- odachi tiene las skills triage y pr, adaptadas a su gramatica, para clasificar bugs por raiz y entregar PRs chicos

## Listo cuando
- odachi skills --check queda en verde con triage y pr, cada una con Trigger, Cuando, Reglas y Nunca en 60 lineas o menos
- odachi init instala las dos skills en .agents/skills y doctor las cuenta
- go test ./... queda en verde
- odachi check queda listo

## NO incluye (primera versión)
- no copiar la gramatica de ODD ni las etiquetas de GitHub de gentle-ai
- no agregar un flag ni un verbo nuevo: son skills, no binario

## Decisiones tomadas
- Se usa desde la compu.
- No hay datos sensibles en la primera versión.
- Primera versión usable: esta semana.
- Quiere aprender: como se escribe una skill que el binario valida y que el agente puede cargar.
