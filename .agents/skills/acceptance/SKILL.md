---
name: acceptance
description: Trigger: cuando hay que decidir si algo está terminado. Convierte un pedido en criterios observables antes de escribir código.
---

## Cuándo

Cuando alguien dice "ya está", "falta poco" o "¿lo damos por hecho?" y no hay una
lista de criterios que cualquiera pueda verificar sin preguntarte nada.

## Reglas

- Antes de escribir código, convertí el pedido en criterios observables. Si no se
  puede observar desde afuera, no es un criterio: es una intención.
- Una línea por criterio, en las palabras del humano, en "Listo cuando". Si no
  entra en una línea, el criterio está haciendo dos cosas.
- Un criterio por comportamiento, no por archivo ni por capa. "Puedo reservar un
  turno y aparece en la agenda" sirve; "el módulo de turnos funciona" no.
- Cada criterio se verifica con un comando o una salida real, y esa salida queda
  guardada. Probar a mano una vez no es verificar.
- Si un criterio no se puede verificar todavía, decilo y proponé cómo. Un
  criterio escondido es peor que uno pendiente.
- Los criterios se escriben antes de implementar y no se tocan después para que
  la implementación pase.

## Nunca

- Aceptar "anda" o "creo que anda" como cierre.
- Agregar criterios que el humano no pidió para justificar más trabajo.
- Marcar un paso como hecho sin la salida que lo prueba.
- Cambiar el criterio para que coincida con lo que ya se construyó.
