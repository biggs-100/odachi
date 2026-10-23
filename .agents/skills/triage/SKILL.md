---
name: triage
description: Trigger: cuando hay varios bugs, un backlog o un reporte que se repite. Clasifica por raíz y arregla una vez, no una por issue.
---

## Cuándo

Cuando llegan varios bugs, un reporte se repite, el backlog crece más rápido que
los fixes, o no sabés si un issue es nuevo, duplicado o síntoma de otra cosa.

## Reglas

- Antes de tocar código, clasificá cada reporte por **raíz**, no por superficie:
  (A) ya lo resuelve un cambio en curso, (B) duplicado de una clase conocida,
  (C) bug nuevo con una raíz identificable, (D) feature, (E) no se entiende y va
  a `odachi ask`.
- Dos o más reportes con la misma raíz son **un solo fix**: un paso de
  `odachi plan`, todos los reportes nombrados y una prueba que los cierre juntos.
- Reproducí el síntoma antes de aceptar el diagnóstico. El mecanismo que cuenta
  el reporte es una hipótesis; el síntoma es la evidencia. Cargá `tests` y escribí
  la prueba que falla.
- El fix va donde está la causa, no donde duele. Si el arreglo agrega un estado,
  un flag o un verbo, preguntate qué podés borrar o relajar primero.
- Un reporte que no se reproduce se convierte en un mensaje que se autodiagnostica
  o en una pregunta con el input exacto; no en un parche a ciegas.
- Cerrá con la prueba que reproduce el reporte y queda verde. Sin esa prueba, el
  reporte sigue abierto, aunque el código "ya debería andar".
- Si el reporte está fuera del cerco de `brief.md`, no lo arregles: `odachi ask`.
- Si el fix toca varias raíces, un paso por raíz; no un paso por reporte.

## Nunca

- Un parche por reporte cuando comparten raíz.
- Cerrar por "debería estar arreglado" o por el resumen del agente.
- Arreglar el síntoma sin reproducirlo.
- Agregar una abstracción nueva para un caso que se resuelve con un mensaje o un
  test.
- Inventar datos o pasos para que el reporte cierre.
