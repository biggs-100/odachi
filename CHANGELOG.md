# Changelog

Lo que cambió en cada versión publicada, en el orden en que lo lee una persona.
Las versiones anteriores a la 0.10.0 se reconstruyeron desde sus tags: el repo
todavía no tenía este archivo, y un changelog que empieza hoy borra la historia
de quien llega mañana.

## v0.10.0 · 2026-10-22

Listo en un comando, y el agente sabe manejarlo.

- **`odachi init` deja el proyecto listo de una vez**: historial de git si faltaba,
  bloque de reglas en `AGENTS.md`, bloque de `.gitignore`, hook de pre-commit, las
  skills del proyecto y las dos personas. Ya no son dos comandos.
- **Skills en la ruta estándar**: `.agents/skills` en el proyecto y `~/.agents/skills`
  en el home, la convención cross-client que recomienda el estándar Agent Skills.
  `pack` gana `--scope project|home`, `--dest <carpeta>` y el pseudo-agente `standard`.
- **Personas generadas**: `novato` y `senior` en `.opencode/agents/` (se eligen con
  Tab) y en `.claude/output-styles/` (desde `/config`). La ceremonia la sigue
  decidiendo `.odachi/profile.toml`; la persona decide solo cómo habla el agente.
- **Skill `drive`**: el manual que el agente carga para manejar la herramienta —los
  verbos, el ciclo de una feature, qué significa cada exit— y la línea del bloque
  que lo obliga a cargarla.
- **Skill `bootstrap`** en el home: enseña a arrancar un proyecto donde todavía no
  hay nada, que es justo cuando no hay otro archivo que leer.
- **`doctor` con dos formas**: en una carpeta que no es un proyecto responde una
  situación y un paso (`init`, después `start`); en el resto, los diez chequeos.
- **El bloque de `AGENTS.md` dice qué es odachi**, quién decide qué y dónde está el
  contexto; y quedó fijado byte a byte por un test.
- `pack --json` cambia de forma: las filas son por destino e incluyen `agents: []`.
- `doctor --json` gana `lines`.

## v0.9.1 · 2026-10-21

- La línea de cierre de `init` deja de asumir que el humano commitea: el que guarda
  el avance es el agente.

## v0.9.0 · 2026-10-21

- **`init` crea el historial de git** cuando falta, y lo explica en castellano llano.
  Antes se negaba con «corré git init» a alguien que no sabe qué es git.
- Sin git en la máquina sigue bloqueando, con una frase que dice qué instalar.

## v0.8.0 · 2026-10-21

- Tres skills nuevas para cerrar el currículum del que construye: `commits`,
  `writing` y `skill-craft`. Ocho skills, todas bajo el cap de 60 líneas.
- El registro de skills pasa a preguntarse al payload y al directorio: cambiar el
  contenido ya no obliga a editar un test.

## v0.7.1 · 2026-10-21

- La regla de delegación queda con presupuesto medible (~3 lecturas o ~10k tokens)
  y retorno acotado (≤2k tokens, evidencia `path:line`), y el trigger se amplía.

## v0.7.0 · 2026-10-21

- La mitad de delegación de un orquestador vive en la skill que ya parte pedidos:
  delegar por razón, pasar el task doc por path, verificar contra el brief y no
  contra el resumen, y no delegar nunca la decisión de visión.

## v0.6.1 · 2026-10-21

- El doctor deja de mentir en dos casos medidos: un worktree (`.git` es un archivo)
  ya no se reporta como «no hay repositorio», y un perfil inventado sin `extends`
  ya no valida en verde.

## v0.6.0 · 2026-10-21

- **`doctor`**: diez chequeos de salud, de solo lectura, cada uno con el comando que
  lo arregla.
- **`bench`**: ocho recorridos que clasifican cada bloqueo (in-band, out-of-band,
  templated, interactive, dead end) con ratchet: un dead end hace fallar la corrida.

## v0.5.0 · 2026-10-21

- **Las cuatro lentes 4R** con selección por evidencia: sin lentes para documentos,
  `reliability` para lógica, las cuatro cuando el cambio roza algo que puede lastimar.
- **Refutador**: un `BLOCKER` causal solo se destraba citando la línea del diff que
  lo desmiente. Lo que no se puede citar, no se refuta.

## v0.4.0 · 2026-10-21

- **`review`**: veredicto de diseño por árbol, content-addressed en
  `.odachi/reviews/<hash>.json`. La aprobación no puede flotar porque el nombre del
  archivo *es* el estado del código.
- `check --require-review` opcional: el review como compuerta.

## v0.3.1 · 2026-10-21

- `init` avisa cuando `odachi` no resuelve desde un shell: una regla que el agente no
  puede ejecutar es peor que ninguna.

## v0.3.0 · 2026-10-21

- **OpenCode** como destino de las skills, con destinos por alcance (proyecto vs home)
  y la lista de agentes soportados.
- `undo` respalda lo que reemplaza y se niega a elegir por vos cuando hay más de un
  candidato; `plan` agrega y cierra pasos sin editar Markdown a mano.

## v0.2.0 · 2026-10-21

- **`pack`**: las skills llegan al directorio nativo de cada agente, sin borrar nada,
  con respaldo de lo que difiere.
- Registro generado de skills, validador y creador.

## v0.1.0 · 2026-10-21

- Primera versión que se instala: `start`, `ask`, `level`, `summary`, `check`, `init`,
  `--version`, hook de pre-commit, instaladores `sh` y `ps1`, y seis binarios.
