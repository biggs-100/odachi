# brazo-limpio

## What must exist
- el brazo limpio del fixture deja de tener defectos: los hallazgos que un lente reporta sobre él se pagan en el código, o se registran como gusto y se dejan
- el número de precisión se puede leer: una corrida del probe sobre el fixture devuelve limpio 0
- los cinco defectos sembrados siguen ahí y se siguen encontrando: limpiar un brazo no puede tocar el examen

## Done when
- el código de `clean/` responde a los diez hallazgos de la primera medida: el status del remoto, el error de lectura y el de escritura, el timeout del cliente, el 404 que tapa un timeout de base, el límite que no acota negativos, y los tests que faltan del handler y del cache
- una corrida del probe con revisor devuelve `clean 0` y los cinco sembrados encontrados
- el manifiesto no se mueve: los cinco defectos, sus paths, sus líneas y sus pisos son los mismos
- `go test ./...` es verde y `odachi check --run "go test ./..."` es verde
- la medida nueva queda en `docs/review-research.md` con fecha y modelo, y reemplaza a la anterior como lectura vigente
- Unreleased nombra el fixture limpio, patch (0264)

## Not included (first version)
- no se toca el contenido de los lentes: medir editando el examen no mide nada - back when a lens is changed for a reason the probe records
- no corre en CI ni en la gate: cuesta llamadas a un modelo y no es determinista - back when a probe is cheap enough to run per change
- no se agregan defectos nuevos al brazo sembrado: la llave no crece en la misma feature que la limpia - back when the fixture outgrows five defects
- no se cambia el puntaje ni el verbo: solo el código que se mide - back when the score itself is the question

## Decisions taken
- El brazo limpio se gana el nombre: la medida es la que dice si quedó limpio, no la intención.
- Lo que un lente marque y sea real se paga en el código; lo que sea preferencia se registra en el cierre y no se deja como ruido pendiente.
- La medida anterior no se borra: se reemplaza como lectura vigente, con fecha y modelo, y la vieja queda como historia.
