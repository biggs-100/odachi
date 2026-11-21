# la-tui-edita
Base: 80fffac08daa4cc6272072101ec2f0bff53f00d0
Light: yes

## Requirements (verbatim, from the user)
R1 · "la TUI edita nombre, comando y remedio de un check declarado sin tocar su severidad"

## Done when

## Steps
- [x] T1 · la TUI edita nombre, comando y remedio de un check declarado sin tocar su severidad   R1 · tree:902e1f6739404f773c56e851f21ebd76184deaf5

## History
L1 · "la TUI edita nombre, comando y remedio de un check declarado sin tocar su severidad"
L2 · OFICIO · "la severidad queda fuera del editor porque fija el veredicto; nombre y comando se pueden editar pero ambos entran en la huella de aprobacion, asi que el apply debe decir que hace falta review checks approve de nuevo y el remedio no la toca" (porque check.go:318-321 incluye Name, Command y Severity en ChecksFingerprint y check_approval_test.go:54 lo prueba para el comando; la decision de politica era solo sobre bajar la puerta, no sobre corregir un comando roto)
