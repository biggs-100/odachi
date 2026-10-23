---
name: writing
description: Trigger: cuando hay que escribir un mensaje, la descripción de un PR, un comentario en un issue, o documentación que alguien va a leer para hacer algo.
---

## Cuándo

Cuando el texto tiene que hacer que otra persona **actúe**: entender qué cambió,
revisar un PR, reproducir un bug, o correr un comando que escribiste.

## Reglas

- Escribí para quien tiene que actuar, no para demostrar que entendiste. Primero qué
  pasó o qué falta; el detalle después.
- Lo que no se recorre en treinta segundos no se lee: título que dice el punto,
  primera línea con la conclusión, y recién después el desarrollo.
- Los ejemplos valen más que los adjetivos: "tarda 40 s" en vez de "es lento";
  `odachi check --run "npm test"` en vez de "corré las pruebas".
- Si nombrás un comando, tiene que funcionar **tal como está escrito**. Un comando
  mal citado cuesta una tarde a quien lo sigue.
- En un PR: qué problema resuelve, cómo se prueba, y qué quedó afuera. Tres
  párrafos, no tres páginas.
- Decí lo que **no** cubriste y lo que no verificaste. Un resultado vale lo que vale
  su alcance declarado.

## Nunca

- "Como se puede ver en el código…": si hay que verlo en el código, el texto sobra.
- Una lista de cambios sin decir qué problema resuelven.
- Prometer en un changelog algo que no se probó.
- Adjetivos sin número al lado.
