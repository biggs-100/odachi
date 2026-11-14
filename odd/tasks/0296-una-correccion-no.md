# una-correccion-no
Base: eac2eae2c58459a3959f4173636ddc8b463b0050

## Requirements (verbatim, from the user)
R1 · "una correccion no se repite sin una razon: el veredicto recuerda a cual corrige, y la tercera vuelta se niega salvo que el humano diga por que"

## Done when
- un veredicto de correccion guarda el arbol del veredicto que corrige, y el esquema sube a 9
- una tercera correccion en la misma cadena se niega con un mensaje que nombra la salida: --because con el motivo
- con --because la tercera correccion se guarda y el motivo queda en la linea que el humano lee
- una cadena de dos correcciones y el veredicto original siguen funcionando sin motivo
- go test ./... is green and odachi check --run "go test ./..." is green
- ## Unreleased names it, minor (0296)

## Steps
- [x] T1 · el veredicto guarda el arbol que corrige, y el esquema sube a 9   R1 · route:inline · tree:2140952936273ef44871f4f6aba9ff76931672cc
- [x] T2 · la cadena se camina: la tercera correccion se niega sin motivo, y con --because se guarda   R1 · route:inline · tree:99ce095a3d485625892e523f44e58c07062efb1c
- [x] T3 · el changelog nombra 0296 bajo Unreleased, minor   R1 · route:inline · tree:b69bc60eedb0519700bfa2cb79f5ccba906fb7f1

## History
L1 · "una correccion no se repite sin una razon: el veredicto recuerda a cual corrige, y la tercera vuelta se niega salvo que el humano diga por que"
