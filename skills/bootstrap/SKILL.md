---
name: bootstrap
description: Trigger: cuando alguien quiere arrancar un proyecto desde cero, pregunta qué es odachi, o no sabe por dónde empezar. No se carga dentro de un proyecto que ya es de odachi.
---

odachi es un binario y cuatro artefactos de texto plano —brief, task doc, prueba y
perfil—. El binario hace cumplir el gate; los artefactos son la memoria. La persona
decide la visión y lo que no se negocia; el agente ejecuta.

## Cuándo

Cuando alguien quiere arrancar un proyecto desde cero, pregunta qué es odachi, o no
sabe por dónde empezar. También cuando la carpeta está vacía y hay que decidir el
primer paso.

## Reglas

Tres comandos, en este orden. Cada uno hace una cosa y la deja hecha:

1. `odachi init` — deja el proyecto listo: el bloque de reglas que el agente lee, el
   hook que frena un commit sin prueba, el `.gitignore` que no ensucia el árbol y las
   skills en `.agents/skills`. Esta skill, la de arranque, ya está en `~/.agents/skills`.
2. `odachi start` — la entrevista que escribe `brief.md` (la visión y lo que NO
   entra), el cerco y el primer task doc. Es lo que le da rumbo al trabajo.
3. `odachi check` — dice si está listo: brief, cerco, preguntas abiertas y prueba. Con
   `--run "<comando>"` corre la prueba y la guarda.

- Un proyecto ya es de odachi si hay `brief.md`, un bloque de odachi en `AGENTS.md`, o
  `odd/tasks/`. Si lo hay, no arranques de nuevo: leé ese bloque y seguí sus reglas.
- Si no sabés en qué estado está, `odachi doctor` lo dice sin escribir nada.
- Ya con el proyecto armado, el manual del día a día es la skill `drive`.
- La visión la decide la persona. Vos ejecutás y explicás el porqué en una línea.

## Nunca

- Correr `odachi check`, `odachi review` o el hook en una carpeta que no es un
  proyecto odachi: gates sobre algo ajeno no dicen nada.
- Correr `odachi init` en un repositorio que no es de la persona: escribe en el
  proyecto y en su home.
- Arrancar de nuevo un proyecto que ya existe, ni reescribir un `brief.md` ajeno.
