# medida-de-los-lentes

## What must exist
- una medida del revisor: cuántos defectos sembrados encuentra cada lente, y cuánto ruido levanta sobre el mismo código sin defectos
- `odachi review probe` recorre un fixture con defectos plantados y su gemelo limpio, ejecuta el mismo `--exec` que `review`, y compara lo que el revisor dice contra un manifiesto escrito a mano
- el manifiesto nombra cada defecto: path, línea, lente que debería verlo, severidad mínima y una frase de qué es
- la salida es una tabla por lente -encontrados de total, severidades, y hallazgos sobre el brazo limpio- y termina en una línea con el número
- el fixture y el manifiesto son la verdad de la medida: el defecto que no está en el manifiesto no cuenta, y el que está y nadie ve se reporta como perdido

## Done when
- `odachi review probe --fixture <dir>` sin `--exec` imprime el prompt que pediría, no llama a nadie y sale 0
- con `--exec`, corre un revisor por lente, guarda cada veredicto en el store y puntúa contra el manifiesto
- un defecto cuenta como encontrado si hay un hallazgo del mismo path, con línea dentro de la ventana del manifiesto y severidad igual o mayor al piso
- un hallazgo que no corresponde a ningún defecto cuenta como ruido, y el brazo limpio no tiene ninguno por definición
- un fixture sin manifiesto, un manifiesto con un lente que no existe, o un defecto sin path son errores que nombran lo que falta
- el fixture no entra a la suite: los tests puntúan veredictos escritos a mano, sin red y sin modelo
- `go test ./...` es verde y `odachi check --run "go test ./..."` es verde
- la primera medida queda escrita en `docs/review-research.md` con fecha, modelo y número
- Unreleased nombra el verbo y el fixture, patch (0263)

## Not included (first version)
- no corre en CI ni en la gate: cuesta llamadas a un modelo y no es determinista - back when a probe is cheap enough to run per change
- no elige un modelo distinto por lente: mide lo que `odd/review.toml` diga hoy - back when heterogeneity is what is being measured
- no mide el verifier ni el floor, solo los cuatro lentes - back when the verifier owes its own probe
- no genera defectos: el manifiesto es a mano y se revisa como código - back when the fixture grows past what a person can hold
- no toca el contenido de los lentes: medir editando el examen no mide nada - back when a lens is changed for a reason the probe records

## Decisions taken
- El instrumento es un subcomando de `review` (`odachi review probe`), igual que `review mode`: la superficie que ya existe y el mismo camino de `--exec`.
- El fixture vive versionado en `internal/probe/`: dos carpetas, `clean/` y `seeded/`, y el manifiesto.
- La medida es recall y ruido, no cobertura: un defecto cuenta una vez, y el ruido se cuenta por hallazgo.
- El fixture es Go y pequeño, con defectos visibles en el diff y sin necesidad de compilar.
- El probe escribe sus veredictos en el store como cualquier review, al lado de los demás.
