# go-tests-y

## Requisitos (textual, del usuario)
R1 · "go-tests y skill-improve: dos skills nuevas para tests de Go y auditoria de skills"

## Listo cuando
- odachi skills --check queda en verde con go-tests y skill-improve, cada una con Trigger, Cuando, Reglas y Nunca en 60 lineas o menos
- odachi init instala las dos en .agents/skills y doctor cuenta 14 skills y 13 instaladas
- go test ./... queda en verde
- odachi check queda listo

## Pasos
- [x] T1 · la skill go-tests elige el patron de Go sin repetir tests   R1 · 8c28cf7
- [x] T2 · la skill skill-improve audita trigger, tamano y dueno de una skill   R1 · 864f87a
- [x] T3 · dogfood: init instala las dos, el registro y los docs las cuentan, y check queda verde   R1 · 3e3a47b

## Historial
L1 · "go-tests y skill-improve: dos skills nuevas para tests de Go y auditoria de skills"
L2 · OFICIO · "skill-improve audita trigger y contenido; el formato lo sigue validando el binario" (porque si la skill repite las reglas de skills --check, se desincroniza del validador)
L3 · OFICIO · "go-tests no repite tests: el trigger manda a cargar tests primero" (porque dos skills que dicen lo mismo se desincronizan; una elige el patron y la otra la disciplina)
