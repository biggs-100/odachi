# the-probe-s
Base: 2c79b254253593451c8add90d79d2c9474750182

## Requirements (verbatim, from the user)
R1 · "the probe's reading separates a suggestion from an invented row: noise counts only what claims a defect, both arms report both numbers, and the 0281 answers re-read under the split are in review-research"

## Done when
- a suggestion on either arm is a note, and a warning or above where nothing was planted is noise
- the table and the summary line print noise, notes, clean noise and clean notes across the draws
- the answers 0281 stored are re-read by severity and review-research Part 5c carries the corrected numbers
- go test ./... is green and odachi check --run "go test ./..." is green
- ## Unreleased names it, patch (0282)

## Steps
- [x] T1 · el puntaje separa lo que reclama un defecto de lo que solo sugiere uno, y las dos cuentas se leen de la misma lista   R1 · route:inline · tree:a7a59b712a62a2465361216ee2da482240d2d7cb
- [x] T2 · la tabla y la linea final nombran ruido y notas en los dos brazos   R1 · route:inline · tree:9303ddd3c68a38e7e83675947aca63c2300aa1ac
- [x] T3 · las respuestas que 0281 guardo se releen por severidad y review-research Parte 5c lleva los numeros corregidos   R1 · route:inline · tree:18d92372e27be67cfb2bea79f929b32893de5e88
- [x] T4 · cada version publicada tiene su seccion: v0.40.2 y v0.40.1 salen de Unreleased, y 0282 escribe la suya   R1 · route:inline · tree:d2eb0d8ef4a402df761ac7ebcd8e0f58e3e9e3c6

## History
L1 · "the probe's reading separates a suggestion from an invented row: noise counts only what claims a defect, both arms report both numbers, and the 0281 answers re-read under the split are in review-research"
L2 · OFICIO · "La regla de la sonda deja de proteger el puntaje y conserva el fixture y la llave" (porque la llave es lo que se compara y el puntaje es como se lee la respuesta, y esta feature cambia justamente esa lectura: una regla que protege el archivo que la feature existe para mover se abre sola o bloquea lo que el humano decidio; el brief movio sus non-goals en el commit de apertura, asi que la compuerta no se abre por omision)
L3 · OFICIO · "La tabla y la linea final nombran ruido y notas en los dos brazos, en vez de una columna noise y otra clean" (porque el ruido y la nota son lo que un hallazgo reclama, no lo que un brazo es: la columna clean mezclaba dos preguntas (cuanto inventa el brazo con defectos y cuanto sugiere el brazo limpio), y un lector que compara dos modelos no puede restar dos numeros que no son el mismo)
L4 · OFICIO · "Las respuestas de 0281 se releen reproduciendo el archivo guardado a traves de la sonda nueva, no recontandolas a mano" (porque un recontador propio es una segunda lectura de la misma regla y las dos pueden discrepar; la reproduccion corre el codigo que mide y ademas reproduce los numeros viejos exactos (5 · 6 y 1 · 0), asi que lo que cambia es el significado y no la medicion)
