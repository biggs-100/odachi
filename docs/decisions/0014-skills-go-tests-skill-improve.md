# 0014 — Skills de go-tests y skill-improve

Status: accepted · Date: 2026-10 · Session: v0.11.2

## Context

Después de `triage` y `pull-request`, quedaban dos huecos que no eran de flujo
sino de oficio: el patrón concreto de los tests de Go y la auditoría de una skill
que creció o dejó de usarse. Odachi ya tenía `tests` (la disciplina TDD) y
`skill-craft` (cómo nace una skill), pero ninguna elegía el patrón de Go ni decía
cómo recortar una skill existente sin romper el formato.

Copiar `go-testing` y `skill-improver` de gentle-ai tal cual no entra: el primero
está atado a Bubbletea, golden files y `teatest`, que son útiles pero no son toda
la disciplina; el segundo audita contra un style guide de 1.000 tokens que odachi
no tiene. Las dos skills tienen que complementar lo que ya existe, no repetirlo.

## Decision

- **`go-tests`**: se carga después de `tests` y elige el patrón de Go. Tabla con
  `t.Run`, `t.TempDir`, `testing.Short`, `teatest`, golden files, mocks en el
  borde y paquete focal antes de la suite. No repite “un test por comportamiento”
  ni “miralo fallar”: eso es de `tests`.
- **`skill-improve`**: audita una skill que no se usa o creció. Empieza por el
  trigger, exige reglas observables, corta en 60 líneas, deja un dueño por regla
  y manda a `skills --check`/`--write`. El formato lo sigue validando el binario;
  la skill no repite las reglas de `internal/domain/skill`.
- **Nada nuevo en el binario.** Son dos skills, no dos verbos: se instalan con
  `init`, se validan con `skills --check` y se cuentan con `doctor`.
- **Cada una tiene una prueba de contenido** en `skills/embed_test.go`: si alguien
  borra `t.TempDir` o la regla de 60 líneas, el test falla.

## Not taken

- `go-testing` entera: se adaptó el patrón, no se copió el inventario de Bubbletea.
- `skill-improver` entera: su style guide de 1.000 tokens no es el contrato de
  odachi; el contrato es el validador y el cap de 60 líneas.
- `judgment-day`, `rdd-*`, `cognitive-doc-design` y `comment-writer`: siguen fuera
  porque odachi ya tiene `review`, `writing` o decidió no tener RDD.

## Consequences

- El catálogo pasa de doce a catorce skills; el proyecto instala trece y el home
  una.
- Un proyecto Go tiene un patrón concreto sin cargar una skill de 1.000 tokens.
- Una skill que creció tiene un camino de auditoría que termina en el validador,
  no en una reescritura a mano.
- El costo de contexto sube poco: 78 líneas entre las dos, cada una bajo el cap.
- La próxima skill que no entre en el cap va a un documento, no a un cuerpo más
  largo.
