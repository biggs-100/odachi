---
name: slices
description: Trigger: cuando hay que partir un pedido grande en pasos chicos que se puedan revisar y commitear uno por uno.
---

## Cuándo

Cuando el pedido no entra en un commit, cuando hay que decir "por dónde empiezo",
o cuando el diff ya pasó las 400 líneas y nadie puede revisarlo en serio.

## Reglas

- Cada paso deja el repositorio verde y se commitea solo. Si el paso necesita
  "y", son dos pasos.
- Ordená por riesgo, no por comodidad: primero lo que puede tirar abajo el
  diseño (el contrato, los datos), después lo que se construye encima.
- Cada paso cita el requisito que satisface (R1, R2). Un paso que no cita
  ninguno sobra o revela un requisito que falta escribir.
- El primer paso es el que produce evidencia más rápido: el que permite ver algo
  funcionando antes de invertir en el resto.
- El commit dice qué cambió para quien usa el proyecto, en una línea. Si hay que
  explicar el "cómo", va en el cuerpo, no en el título.
- Un paso incluye su prueba y su documentación. No hay un paso posterior llamado
  "tests" o "docs".

## Si delegás

Un subagente cuesta una sesión nueva, lecturas sin caché y un handoff. Delegá por
razón, nunca por tamaño: entender algo que no entra en una lectura, dos unidades
independientes que no se tocan, o el contexto ya lleno.

- Al subagente le pasás **el path del task doc y el paso**, no un resumen. El brief y
  los requisitos se leen; parafrasearlos es cómo se pierden.
- Pedile que anote lo que hace con `odachi plan` y que no diga "listo" sin
  `odachi check --run`. Si no puede, que lo diga, en vez de reportar éxito.
- Cuando vuelve, verificá su salida contra `brief.md` y contra el diff, **no contra su
  resumen**: el subagente puede mentir igual que el principal, y con menos testigos.
- Si tocó algo fuera del cerco, no lo arregles en silencio: es una decisión del humano
  y va con `odachi ask`.
- **Nunca delegues la decisión de visión.** El subagente ejecuta; quién decide qué se
  construye es la persona.

## Nunca

- Commitear con tests rojos "para no perder el trabajo".
- Mezclar un refactor con un cambio de comportamiento en el mismo paso.
- Partir por archivos o por capas en vez de por comportamiento entregable.
- Dejar pasos de limpieza sin criterio de terminado: crecen y nunca se cierran.
- Commitear el trabajo de otra persona que ya estaba en el árbol.
