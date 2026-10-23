---
name: pull-request
description: Trigger: cuando un cambio va a PR o el diff pasa las 400 líneas. Parte la entrega en work units revisables y elige una sola estrategia de cadena.
---

## Cuándo

Cuando el cambio ya está para revisión, cuando un PR supera las 400 líneas
cambiadas, cuando hay que apilar varios PRs, o cuando el revisor no puede
entender el diff en una sentada.

## Reglas

- Un PR es **un work unit**: un comportamiento, un fix o una migración que se
  puede revertir solo. Si el título necesita "y", son dos PRs.
- Presupuesto: **400 líneas cambiadas** (adiciones + deleciones). Si lo pasás,
  partí antes de abrir; no pidas una excepción sin decir por qué.
- Elegí **una** estrategia y no la mezcles: stacked a main si cada slice puede
  aterrizar solo; feature branch chain con un PR tracker si la feature tiene que
  integrarse antes de main.
- Cada PR dice qué problema resuelve, cómo se prueba, qué quedó afuera y de qué
  PR depende. El primero de una cadena no depende de nadie.
- Tests y docs viajan en el PR del comportamiento que verifican; no hay un PR
  posterior de "tests" o "docs".
- Los checks requeridos se leen del repo, no se adivinan: si no sabés si un check
  es obligatorio, el PR no está listo.
- Si el repo usa issues, referenciá el issue aprobado con
  `Closes/Fixes/Resolves #N`; si no, usá `Refs #N`. No inventes una aprobación.
- `commits` decide el mensaje; esta skill decide el corte. El PR se lee en orden.
- Cerrá con `odachi check --run "<comando>"` en verde y la salida guardada: el PR
  no reemplaza la prueba del árbol.

## Nunca

- Abrir un PR de 800 líneas "para no partir el trabajo".
- Mezclar stacked y feature branch chain en la misma entrega.
- Un PR que no dice qué quedó afuera.
- Pedir review sin los checks requeridos verdes.
- Reescribir una rama protegida o hacer force-push a main.
- Inventar un issue o una aprobación para cumplir el issue-first.
