---
name: tests
description: Trigger: cuando hay que escribir, arreglar o decidir qué prueba un cambio. Primero la prueba que falla, después el código.
---

## Cuándo

Cuando vas a escribir código nuevo, cuando un test falla y no sabés por qué, o
cuando alguien pregunta "¿esto está probado?".

## Reglas

- Primero el test, y miralo fallar por la razón correcta. Un test que nunca falló
  no prueba nada: prueba que el código ya estaba escrito.
- Un test por comportamiento. El nombre dice qué se rompe si falla, no qué
  función llama.
- Probá el borde que duele: vacío, uno, muchos, el error, el archivo a medias, el
  CRLF, el permiso denegado. Ahí viven los defectos.
- Cuando el mensaje que lee el humano es el contrato (un exit code, una línea en
  español), el test fija ese mensaje.
- Los tests de dominio no tocan disco, red ni reloj. Si los tocás, el test tarda
  y empieza a fallar por razones ajenas al cambio.
- Un test que falla se arregla en la causa. Si el test estaba mal escrito, se
  dice en el reporte y se explica por qué.

## Nunca

- Ajustar el test para que pase sin cambiar la causa de la falla.
- Testear detalles privados que impiden refactorizar mañana.
- Tests que dependen del orden de ejecución, de la hora o de la red.
- Un `t.Skip` sin decir por qué y cuándo se destraba.
- Decir "los tests pasan" cuando corriste solo el paquete que tocaste.
