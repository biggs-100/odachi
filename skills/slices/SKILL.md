---
name: slices
description: Trigger: cuando hay que partir un pedido grande en pasos chicos, cuando el código ya no entra en una lectura, o hay que decidir si delegar a un subagente.
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

Dentro de un proyecto odachi, esta es la política de delegación: si otra skill
global dice otra cosa, gana esta.

Un subagente arranca sin caché: **no ahorra tokens totales, protege tu contexto.**

- Presupuesto: si entender pasó de ~3 lecturas o ~10k tokens, o la sesión viene
  larga, la próxima unidad va a un subagente. Delegá por razón, nunca por tamaño.
- Lo que **se lee mucho** se delega; lo que **se escribe poco** no.
- Un **escritor** se delega solo por una razón: dos o más unidades independientes
  con superficies disjuntas, cada una más pesada que arrancar un subagente; o
  ~150k tokens de sesión. Nunca por tamaño, archivos ni precio; sin razón,
  escribí vos.
- Un escritor por unidad; en paralelo no comparten archivos y vos sos dueño de
  git y del control final. Si delegás la escritura, delegá también la lectura que
  la prepara.
- El pedido lleva el **path del task doc y el paso**, no un resumen; además, las
  **superficies de edición**, los **comandos de verificación** y los **paths
  exactos de las skills** (`## Skills to load before work`).
- Para escribir, usá **`odachi-worker`**; para verificar, **`odachi-verify`**
  (read-only). El worker no commitea ni delega; el padre conserva git, el gate y
  la visión.
- El retorno está acotado: **≤2k tokens**, evidencia `path:line` y la lista de
  archivos leídos; el resto se queda adentro.
- Pedile que anote con `odachi plan` y que no diga "listo" sin `odachi check --run`.
- Cuando vuelve, verificá su salida contra `brief.md` y el diff, **no contra su
  resumen**: el subagente puede mentir igual que el principal.
- **Nunca delegues la decisión de visión.** El subagente ejecuta; la persona decide.

## Nunca

- Commitear con tests rojos "para no perder el trabajo".
- Mezclar un refactor con un cambio de comportamiento en el mismo paso.
- Partir por archivos o por capas en vez de por comportamiento entregable.
- Dejar pasos de limpieza sin criterio de terminado: crecen y nunca se cierran.
- Delegar por tamaño o cantidad de archivos, o para ahorrar tokens en algo que hacés
  en un turno.
