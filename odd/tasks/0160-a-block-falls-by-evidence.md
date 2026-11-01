# a-block-falls-by-evidence
Base: f6e60e15eb2ff873f749ad6b2f7c3d39a3bc35fb

## Requirements (verbatim, from the user)
R1 · "a finding says whether its evidence is cited from the diff or argued, and a blocking finding is lifted only by the evidence its class requires: one cited counter-argument for a cited claim, two for an argued one"

## Done when
- a reviewer answer that says neither cited nor argued reads as argued, an unknown value is refused naming the two, and a verdict stored before the class existed keeps its meaning
- two cited refutations lift an argued BLOCKER and one leaves it blocking, naming the vote it still needs; a cited BLOCKER falls with one -both observed in a test
- each refuter's answer is stored once, the legacy `<tree>.refuted.json` keeps the meaning it has today, and a test reads one
- an ADR says what a citation buys, why an argued claim needs two votes, and what the tool still cannot prove

## Steps
- [x] T1 · a finding declares cited or argued evidence, and an old verdict reads as argued   R1 · tree:057e6dfdd7d22791481104b64d9984f4644bf2b4
- [ ] T2 · an argued blocker needs two cited refutations to fall, and a cited one needs one   R1
- [ ] T3 · the ADR says what a citation buys and what the tool still cannot prove   R1

## History
L1 · "a finding says whether its evidence is cited from the diff or argued, and a blocking finding is lifted only by the evidence its class requires: one cited counter-argument for a cited claim, two for an argued one"
L2 · OFICIO · "the evidence class is the reviewer's own account -cited or argued- and absence reads as argued" (porque a veredicto escrito antes del campo tiene que conservar su significado, y el lado estricto es el que exige dos testigos; exigir el campo al guardar romperia las correcciones que heredan filas viejas)
L3 · OFICIO · "a cited block falls with one refutation and an argued one needs two, capped at three witnesses" (porque una cita se puede mirar y refutar en el acto; una afirmacion que solo argumenta necesita que dos lectores independientes coincidan, y tres es el techo que evita que una multitud levante un bloqueo; el archivo viejo <tree>.refuted.json se lee como el primer testigo, asi que ningun arbol revisado antes pierde su significado)
