---
name: slices
description: Trigger: cuando hay que partir un pedido grande en pasos chicos, o cuando entender el código que hay que tocar ya no entra en una lectura.
---

## Cuándo

Cuando el pedido no entra en un commit, cuando hay que decir "por dónde empiezo",
cuando el diff ya pasó las 400 líneas y nadie puede revisarlo en serio, o cuando
entender lo que hay que tocar te llevó más de tres lecturas.

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

Un subagente arranca sin caché: re-lee el prompt y los archivos. **No ahorra tokens
totales, protege tu contexto.** Por eso la regla no es el tamaño de la tarea.

- Presupuesto: si entender esto pasó de ~3 lecturas o de ~10k tokens, o si la sesión
  ya viene larga, la próxima unidad va a un subagente. Delegá por razón —explorar
  mucho, dos unidades que no se tocan, contexto lleno—, nunca por tamaño.
- Lo que **se lee mucho** se delega; lo que **se escribe poco** no. Explorar un
  módulo desconocido: delegá. Cambiar una línea que ya encontraste: hacelo vos.
- El pedido lleva el **path del task doc y el paso**, no un resumen. Parafrasear los
  requisitos es cómo se pierden.
- El retorno está acotado: **≤2k tokens**, evidencia `path:line`, y la lista de
  archivos que leyó. Todo lo demás se queda adentro del subagente: eso es lo que no
  contamina.
- Pedile que anote con `odachi plan` y que no diga "listo" sin `odachi check --run`.
- Cuando vuelve, verificá su salida contra `brief.md` y contra el diff, **no contra su
  resumen**: el subagente puede mentir igual que el principal, y con menos testigos.
- **Nunca delegues la decisión de visión.** El subagente ejecuta; quién decide qué se
  construye es la persona.

## Nunca

- Commitear con tests rojos "para no perder el trabajo".
- Mezclar un refactor con un cambio de comportamiento en el mismo paso.
- Partir por archivos o por capas en vez de por comportamiento entregable.
- Dejar pasos de limpieza sin criterio de terminado: crecen y nunca se cierran.
- Delegar para ahorrar tokens en algo que hacés en un turno.
