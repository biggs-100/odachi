# el-prefijo-que-resuelve
Base: bf190d2ae146c02699d6b2a8abb17c4070885c69

## Requirements (verbatim, from the user)
R1 · "`--since` acepta el árbol que un humano copia: un prefijo del veredicto guardado, y resuelve al árbol completo antes de medir el delta"

## Done when
- `--since <prefix>` of a single stored verdict measures the delta from its full tree, and the run says which tree it completed
- a prefix that matches two verdicts is refused, naming both and asking for more characters
- a prefix that matches none is refused with a message that no stored verdict starts with it, and names `odachi review --all`
- a full tree keeps working without a resolver in the way
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- `## Unreleased` names it, patch (0272)

## Steps
- [x] T1 · el prefijo resuelve: --since con un pedazo del arbol mide desde el veredicto guardado   R1 · route:inline · tree:495212d0f467326bd8738048b6af441a45be7ba9
- [x] T2 · el prefijo ambiguo o ausente se explica: dos se nombran, ninguno dice como listarlos   R1 · route:inline · tree:495212d0f467326bd8738048b6af441a45be7ba9

## History
L1 · "`--since` acepta el árbol que un humano copia: un prefijo del veredicto guardado, y resuelve al árbol completo antes de medir el delta"
