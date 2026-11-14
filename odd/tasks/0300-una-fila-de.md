# una-fila-de
Base: 6aee47bc834a4d8254165f3ef7a7128793cb3cca

## Requirements (verbatim, from the user)
R1 · "una fila de deshacer que no compila deja de ser un no juzgado: si el error de compilacion nombra un simbolo que el cambio agrego, la copia noto el revert y eso es evidencia"

## Done when
- undone extrae los nombres declarados en las lineas agregadas del diff -func, metodo, type, const, var- con sus tests
- un fallo de compilacion que nombra uno de esos nombres es noticed, con una nota que dice cual
- un fallo de compilacion que no nombra ninguno sigue unknown, con la nota de hoy
- la fila pasa el diff y escribe veredicto y nota en el archivo undo de la feature
- go test ./... is green and odachi check --run "go test ./..." is green
- ## Unreleased names it, patch (0300)

## Steps
- [x] T1 · los nombres que el diff agrega: declaraciones de lineas agregadas, con sus tests   R1 · route:inline · tree:c0ad853abb0d46cccaf0bdc4062afcd0753e1eca
- [ ] T2 · el rojo de compilacion que nombra uno de esos nombres es noticed, y el que no sigue unknown   R1 · route:inline
- [ ] T3 · la fila pasa el diff y escribe la nota con el nombre   R1 · route:inline
- [ ] T4 · el changelog nombra 0300 bajo Unreleased, patch   R1 · route:inline

## History
L1 · "una fila de deshacer que no compila deja de ser un no juzgado: si el error de compilacion nombra un simbolo que el cambio agrego, la copia noto el revert y eso es evidencia"
