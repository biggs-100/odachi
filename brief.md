# go-tests-y

## Qué tiene que existir
- go-tests y skill-improve: dos skills nuevas para tests de Go y auditoria de skills

## Listo cuando
- odachi skills --check queda en verde con go-tests y skill-improve, cada una con Trigger, Cuando, Reglas y Nunca en 60 lineas o menos
- odachi init instala las dos en .agents/skills y doctor cuenta 14 skills y 13 instaladas
- go test ./... queda en verde
- odachi check queda listo

## NO incluye (primera versión)
- no copiar la gramatica de ODD ni las etiquetas de GitHub de gentle-ai
- no duplicar tests ni skill-craft: go-tests elige el patron de Go y skill-improve audita el formato y el trigger

## Decisiones tomadas
- Se usa desde la compu.
- No hay datos sensibles en la primera versión.
- Primera versión usable: esta semana.
- Quiere aprender: como se escribe una skill que complementa otra sin repetirla.
