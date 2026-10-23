# odachi-tiene-las

## Requisitos (textual, del usuario)
R1 · "odachi tiene las skills triage y pr, adaptadas a su gramatica, para clasificar bugs por raiz y entregar PRs chicos"

## Listo cuando
- odachi skills --check queda en verde con triage y pr, cada una con Trigger, Cuando, Reglas y Nunca en 60 lineas o menos
- odachi init instala las dos skills en .agents/skills y doctor las cuenta
- go test ./... queda en verde
- odachi check queda listo

## Pasos
- [x] T1 · la skill triage clasifica reportes por raiz y arregla una vez   R1 · eb9a2bb
- [x] T2 · la skill pr parte la entrega en work units y elige una estrategia de cadena   R1 · 2b6d9f5
- [ ] T3 · dogfood: init instala las dos, el registro y los docs las cuentan, y check queda verde   R1

## Historial
L1 · "odachi tiene las skills triage y pr, adaptadas a su gramatica, para clasificar bugs por raiz y entregar PRs chicos"
L2 · OFICIO · "La skill se llama pull-request, no pr, porque el formato exige nombres de 3 caracteres" (porque el slug es parte del contrato que valida skills --check; relajar la regla por una skill romperia el formato de todas)
L3 · OFICIO · "Se porta el metodo de gentle-ai, no su gramatica: triage y pull-request usan ask, plan, tests y check --run" (porque odachi ya decidio no tener Specs/Tasks/Log ni RDD; copiar la gramatica habria duplicado el task doc)
L4 · OFICIO · "triage no depende de etiquetas ni de un issue tracker" (porque la skill tiene que servir en un repo sin GitHub; el reporte y la raiz son el contrato, no la etiqueta)
L5 · OFICIO · "pull-request toma el presupuesto de 400 lineas y las dos estrategias de cadena, pero no vuelve obligatorio el issue-first" (porque el corte de PR sirve sin copiar la burocracia de un repo OSS; si el repo usa issues, la skill dice como referenciarlos)
