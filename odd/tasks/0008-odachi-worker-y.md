# odachi-worker-y

## Requisitos (textual, del usuario)
R1 · "odachi-worker y odachi-verify como subagentes nativos, con handoff acotado, para que el agente principal no cargue el trabajo pesado"

## Listo cuando
- init escribe odachi-worker y odachi-verify para OpenCode y Claude, y una segunda corrida no escribe nada
- doctor compara los subagentes con lo generado y avisa si faltan o driftan
- slices manda a pasar task doc, paso, superficies, comandos y paths de skills; el retorno es <=2k
- opencode debug agents lista los dos como subagent
- go test ./... queda en verde
- odachi check queda listo

## Pasos
- [x] T1 · escribir los subagentes y que init los instale   R1 · e36e275
- [x] T2 · que doctor compare los subagentes   R1 · e36e275
- [x] T3 · handoff en slices y docs   R1 · e36e275

## Historial
L1 · "odachi-worker y odachi-verify como subagentes nativos, con handoff acotado, para que el agente principal no cargue el trabajo pesado"
