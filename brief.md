# odachi-worker-y

## Qué tiene que existir
- odachi-worker y odachi-verify como subagentes nativos, con handoff acotado, para que el agente principal no cargue el trabajo pesado

## Listo cuando
- init escribe odachi-worker y odachi-verify para OpenCode y Claude, y una segunda corrida no escribe nada
- doctor compara los subagentes con lo generado y avisa si faltan o driftan
- slices manda a pasar task doc, paso, superficies, comandos y paths de skills; el retorno es <=2k
- opencode debug agents lista los dos como subagent
- go test ./... queda en verde
- odachi check queda listo

## NO incluye (primera versión)

## Decisiones tomadas
- Se usa desde la compu.
- No hay datos sensibles en la primera versión.
- Primera versión usable: esta semana.
- Quiere aprender: contexto y subagentes.
