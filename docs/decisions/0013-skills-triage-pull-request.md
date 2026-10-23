# 0013 — Skills de triage y pull-request

Status: accepted · Date: 2026-10 · Session: v0.11.2

## Context

La comparación con gentle-ai mostró dos huecos reales en el catálogo de odachi:
atacar un backlog de bugs y cortar una entrega en PRs revisables. Gentle-ai los
resuelve con `systemic-issue-triage`, `issue-root-resolution`, `branch-pr`,
`chained-pr` y `work-unit-commits`. Odachi tenía `commits` para el mensaje y
`tests` para la reproducción, pero ninguna skill que decidiera **qué es un solo
fix** ni **cómo se parte un PR**.

Copiar esas skills enteras no entra: el formato de odachi exige `Cuándo / Reglas
/ Nunca`, un `Trigger:` en la descripción, 60 líneas de cuerpo y un nombre de tres
caracteres o más. Tampoco corresponde copiar la gramática de ODD, las etiquetas
de GitHub ni la maquinaria de RDD: odachi decidió no tenerlas (ADR 0008).

## Decision

- **`triage`**: clasifica cada reporte por raíz (A–E), junta los que comparten
  causa en un solo paso, reproduce el síntoma antes de aceptar el diagnóstico y
  cierra con la prueba que falla. Un reporte fuera del cerco va a `odachi ask`;
  varias raíces son varios pasos de `odachi plan`.
- **`pull-request`**: un work unit por PR, presupuesto de 400 líneas cambiadas,
  una sola estrategia de cadena (stacked o feature branch chain), issue aprobado
  o `Refs`, y `odachi check --run` antes del review. El slug es `pull-request`
  porque el formato exige tres caracteres; el trigger sigue diciendo “PR”.
- **Se adapta el método, no la gramática.** No hay `Specs`/`Tasks`/`Log`, ni
  etiquetas `status:approved`, ni receipts de RDD. Las dos skills usan los verbos
  que ya existen: `ask`, `plan`, `tests`, `check --run`, `commits`.
- **Se instalan como el resto del currículum.** `init` las copia a
  `./.agents/skills`; `bootstrap` sigue en `~/.agents/skills`; `doctor` cuenta las
  copias y `skills --check` valida el formato.
- **Cada una tiene una prueba de contenido** en `skills/embed_test.go`: si alguien
  borra la regla de raíz o el presupuesto de 400 líneas, el test falla.

## Not taken

- La gramática de ODD y su espejo en Engram.
- Las etiquetas y el issue-first de GitHub como obligación: la skill dice qué
  hacer si el repo usa issues, no inventa una aprobación.
- RDD, judgment-day y los receipts: odachi ya tiene `review` con lentes y
  refutador.
- `collab-perfect`, `cognitive-doc-design`, `comment-writer`, `go-testing` y
  `gentle-ai-bench`: son específicas de gentle-ai o cubren un flujo que odachi
  todavía no tiene. Si aparecen como necesidad, se agregan con su propia feature.

## Consequences

- El catálogo pasa de diez a doce skills; el proyecto instala once y el home una.
- Un backlog y una entrega tienen un método explícito sin agregar verbos ni flags.
- El costo de contexto sube poco: las dos skills suman 78 líneas y están bajo el
  cap de 60 cada una.
- La comparación con gentle-ai queda documentada: se tomó el método, no la
  maquinaria. La próxima skill que no entre en el cap va a un documento, no a un
  cuerpo más largo.
