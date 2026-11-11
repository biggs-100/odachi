# la-medida-se-repite

## What must exist
- el probe deja de reportar una tirada: corre el fixture N veces y reporta, por lente y por brazo, una mediana y el rango en que se movió
- `--runs N`: sin el flag la medida es una corrida, como hoy; con `--runs 3` son tres por brazo y por lente
- cada corrida guarda su respuesta numerada en el store del probe, para que el número se pueda auditar después

## Done when
- sin `--runs`, el comando se comporta igual que hoy y lo dice en el resumen
- con `--runs N`, se ejecutan N corridas por cada brazo y cada lente, y el informe imprime mediana y rango de encontrados, ruido y limpio
- un `--runs 0` o negativo se rechaza nombrando el mínimo, y un `--runs` mayor que 10 se rechaza nombrando el costo
- los tests puntúan tres corridas distintas con respuestas escritas a mano y pinchan mediana y rango, sin red y sin modelo
- `go test ./...` es verde y `odachi check --run "go test ./..."` es verde
- la nueva lectura de la precisión queda en `docs/review-research.md` con fecha, modelo y la dispersión observada
- Unreleased nombra la repetición, patch (0265)

## Not included (first version)
- no se toca el fixture ni los lentes: la repetición mide lo que ya existe - back when the fixture or a lens changes for a reason this feature records
- no corre en CI ni en la gate: N veces una llamada a un modelo multiplica el costo - back when a cheap reviewer makes a probe per change affordable
- no hay intervalos de confianza ni tests estadísticos: mediana y rango, que es lo que una persona lee - back when the number has to support a claim
- no se cambia el puntaje: la regla de encontrado, bajo y perdido queda como está - back when the scoring rule itself is the question

## Decisions taken
- Mediana y no media: una tirada rara no mueve el centro.
- El rango es el hallazgo de 0264, así que viaja en el informe al lado de la mediana y no en una nota.
- El tope de corridas es 10: más que eso es un experimento, y un experimento se diseña aparte.
- N corridas por brazo y por lente se guardan numeradas, porque el número sin las respuestas no se puede auditar.
