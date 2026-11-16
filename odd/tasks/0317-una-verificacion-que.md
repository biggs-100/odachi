# una-verificacion-que
Base: 90bb26efd0d09d2680168a27df8a48c3eba1e7f9
Light: yes

## Requirements (verbatim, from the user)
R1 · "una verificación que falta se nombra como nota y solo se rechaza con --require-review"

## Done when

## Steps
- [x] T1 · una verificación que falta se nombra como nota y solo se rechaza con --require-review   R1 · tree:91b1d1d78538a440dd79cc06c9570c2668698915

## History
L1 · "una verificación que falta se nombra como nota y solo se rechaza con --require-review"
L2 · OFICIO · "la base del task doc apunta al commit que existe en main, 90bb26e, porque 345ae7d quedó fuera de la historia al squash del PR" (porque la base es el commit del que la feature partió, y tras el squash ese contenido vive en otro commit; un objeto que el remoto no tiene hace que CI no pueda leer qué cambió, que fue la falla del PR de la release)
