# 0012 — Elegir los clientes del agente nativo

Status: accepted · Date: 2026-10 · Session: v0.11.1

## Context

`odachi init` escribía el agente nativo para los dos clientes que tienen un
selector: OpenCode (`.opencode/agents/`) y Claude Code (`.claude/output-styles/`
y `.claude/agents/`). Un proyecto que usa uno solo recibía las carpetas del
otro, y `doctor` las reclamaba como faltantes si alguien las borraba. En este
repo el humano borró `.claude/` porque no usa Claude Code; `doctor` siguió
diciendo «corré odachi init», e `init` volvía a crear lo que acababa de borrar.
Elegir un cliente era una decisión del humano, pero no había forma de decirla.

## Decision

- **`odachi init --agents opencode|claude|both`**. `both` es el default y el
  comportamiento que ya existía; los otros dos filtran el espejo del proyecto y
  la puerta del home. Un valor desconocido se rechaza antes de escribir nada.
- **La elección vive en `.odachi/agents.toml`**, con una sola clave `agents`.
  El archivo se escribe sólo cuando el humano elige: sin flag y sin archivo, el
  default sigue siendo `both` y no aparece un artefacto nuevo. Un `init`
  posterior sin flag conserva la elección guardada; un flag explícito la cambia.
- **Nunca se borra un cliente que ya existe.** `--agents opencode` deja de
  escribir `.claude/`, pero no toca los archivos que ya están en disco. Borrar
  trabajo ajeno no es parte de elegir.
- **La elección gobierna todo lo que escribe o compara el agente nativo.**
  `init` filtra las dos superficies; `start` y `level` refrescan sólo los
  clientes elegidos; `doctor` compara sólo esos y lo dice en la misma línea.
- **Un store roto no se adivina.** `init`, `start` y `level` se frenan con el
  path y el comando que lo arregla; `doctor` lo nombra como nota y se salta los
  agentes, porque el bloque sigue funcionando sin selector.
- **El bloque de `AGENTS.md` sigue siendo el dueño de las reglas.** El store
  decide qué espejos se escriben, no qué reglas se obedecen.

## Consequences

- Un proyecto de OpenCode puede dejar de generar `.claude/` sin perder lo que
  ya tenía, y `doctor` deja de pedirle archivos que nunca eligió.
- El default no cambia para nadie: sin archivo, `both`, y el segundo `init`
  sigue sin escribir nada.
- La elección es un artefacto versionable más. Si se edita a mano, se valida
  estricto: una clave o un valor que el binario no entiende se nombra en vez de
  interpretarse.
- Un cliente nuevo se agrega en la tabla de `internal/domain/persona` y en el
  vocabulario de la selección; `both` sigue significando «todos los de la
  tabla», no «los dos de hoy».
- Tests: el vocabulario y el filtro en `persona`, el formato en `agentstoml`,
  `init` con y sin flag, `start`/`level` refrescando sólo lo elegido, `doctor`
  con la elección y con el store roto, y la CLI de punta a punta.
