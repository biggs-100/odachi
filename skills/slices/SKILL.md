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

## Nunca

- Commitear con tests rojos "para no perder el trabajo".
- Mezclar un refactor con un cambio de comportamiento en el mismo paso.
- Partir por archivos o por capas en vez de por comportamiento entregable.
- Dejar pasos de limpieza sin criterio de terminado: crecen y nunca se cierran.
- Commitear el trabajo de otra persona que ya estaba en el árbol.
