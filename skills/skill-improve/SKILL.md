---
name: skill-improve
description: Trigger: cuando una skill no se usa, creció o el agente no la carga. Audita el trigger, recorta el cuerpo y vuelve a validar el formato.
---

## Cuándo

Cuando una skill existe pero el agente no la carga, cuando su cuerpo pasó de 60
líneas, cuando se repite con otra, o cuando una regla dejó de ser observable.

## Reglas

- Empezá por el trigger: si no dice la situación del usuario, el agente no la va a
  cargar. Reescribilo como “cuando…” y no como un tema.
- Una regla por línea. Si una regla necesita dos, son dos reglas; si no se puede
  violar de forma observable, no es una regla: es una intención.
- Lo que el modelo ya sabe no va. Lo que el proyecto decidió, sí. Lo que se puede
  verificar con un test o un glob del cerco va al test o al cerco antes que a la
  skill.
- Si el cuerpo pasó de 60 líneas, mové el detalle a un documento y dejá en la
  skill el disparador y las reglas duras. No la partas en dos skills sin una razón
  de carga distinta.
- Si se repite con otra skill, elegí un dueño y borrá la copia. Una regla con dos
  dueños se desincroniza.
- Después de editar: `odachi skills --check` y `odachi skills --write` si cambió
  el trigger o el nombre. El registro es generado, no se edita a mano.
- Si la skill no se usa, probá el trigger antes de agregar más contenido: el
  problema suele ser cuándo se carga, no qué dice.
- Medí el antes y el después: líneas, reglas, y si el agente la cargó en una sesión
  real.
- Si la regla es de una sola vez, no es skill: va al `brief.md`, al cerco o a un
  test.

## Nunca

- Agregar reglas para “por si acaso”.
- Copiar una skill entera de otro proyecto sin adaptar el trigger.
- Dejar una skill con dos dueños o con una sección vacía.
- Pasar de 60 líneas y llamarlo “documentación”.
- Tocar `SKILLS.md` a mano.
