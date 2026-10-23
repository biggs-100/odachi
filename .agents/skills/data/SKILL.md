---
name: data
description: Trigger: cuando hay que guardar, migrar, leer o borrar datos. Decide dueño, vida y respaldo antes de escribir la primera fila.
---

## Cuándo

Cuando aparece un archivo que hay que persistir, un esquema que cambia, una
migración, un borrado o la pregunta "¿dónde guardamos esto?".

## Reglas

- Antes de escribir, contestá tres cosas: quién es dueño del dato, cuánto vive y
  quién lo lee. Si no hay respuesta, todavía no es momento de guardarlo.
- Escribí atómico: archivo temporal en el mismo directorio y rename al final. Un
  corte a la mitad deja el archivo viejo entero, no uno truncado.
- Toda migración es reversible o tiene un backup que ya probaste restaurar. Un
  backup que nunca se restauró no es un backup.
- Leé verificando: si el archivo lleva checksum o longitud, comprobalo al abrir.
  Asumir que está entero es cómo se pierden los datos.
- Lo sensible (nombres, salud, plata) no entra al modelo ni al log si no hace
  falta. Lo que no se guarda no se filtra.
- Los datos de prueba viven en fixtures chicos y legibles; los reales, nunca en
  el repositorio.

## Nunca

- Correr una migración destructiva sin backup verificado ni forma de volver.
- Guardar secretos, tokens o datos personales en el repo, en fixtures o en logs.
- Cambiar el esquema y el código que lo lee en un deploy que no se puede
  revertir.
- Borrar en el momento: marcá, esperá, y borrá con fecha y responsable.
- Inventar datos para que un test pase.
