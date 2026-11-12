# govulncheck-corre-en
Base: ae4e2492515a69fb160a5578f854ec9716e177c1

## Requirements (verbatim, from the user)
R1 · "govulncheck corre en cada review como una fila del piso, y la deuda que encuentra se paga: no queda una vulnerabilidad que el codigo llame"

## Done when
- el toolchain sube a la version que arregla las 25: govulncheck ./... sale 0
- odd/review.toml declara govulncheck a CRITICAL, este clon la aprobo, y un review de este arbol dice govulncheck: passed
- go test ./... is green and odachi check --run "go test ./..." is green
- ## Unreleased names it, patch (0283)

## Steps

## History
L1 · "govulncheck corre en cada review como una fila del piso, y la deuda que encuentra se paga: no queda una vulnerabilidad que el codigo llame"
