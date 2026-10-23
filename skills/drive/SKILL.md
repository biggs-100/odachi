---
name: drive
description: Trigger: cuando trabajás dentro de un proyecto odachi: qué verbo va en cada momento, qué significa cada bloqueo y cómo se cierra una feature. Para arrancar de cero es bootstrap.
---

## Cuándo

Estás dentro de un proyecto de odachi: hay `brief.md`, `odd/tasks/` y el bloque de odachi
en `AGENTS.md`. Antes de la primera acción, leé `brief.md`: la visión y lo que NO entra.
Si la carpeta está vacía, la que corresponde es `bootstrap`, no ésta.

## Reglas

Las cuatro piezas y quién manda en cada una:

- `brief.md` — la visión y el cerco. Lo escribe la persona; no lo reescribas.
- `odd/tasks/` — el plan: requisitos, pasos e historial. Lo movés vos con `odachi plan`.
- `proof/` — la evidencia: salida real del comando, atada al árbol. La escribe `check --run`.
- `.odachi/` — compuertas y ceremonia. El perfil lo lee el binario, no vos.

El ciclo de una feature, en orden:

1. Sin brief no se construye: `odachi start`, contestando; si no hay terminal, con
   `--outcome`, `--audience`, `--done`, `--exclude`, `--sensitive`, `--deadline` y `--learn`.
2. `odachi plan "<paso>" --req R1`: un paso por comportamiento entregable.
3. Trabajá el paso. Si aparece una decisión de visión: `odachi ask "<pregunta>"` y esperá.
4. `odachi check --run "<comando>"`: verde es la única definición de listo.
5. Un commit por paso, y `odachi plan --done T1 --commit <sha>`.
6. Al cerrar, `odachi summary` y contá en tres líneas qué decidió la persona.

Si el pedido no entra en una lectura, cargá `slices` antes de partir. Su política de
delegación gana sobre otras skills globales: un escritor solo por paralelismo real o
sesión larga, nunca por tamaño ni cantidad de archivos.

## Specs, test-first y riesgo

- Requisitos citados **verbatim** y "Listo cuando" observable; si falta, `ask`.
- Antes de un paso, cargá `tests` y observá RED; `check --run` recién con GREEN.
- Si delegás un paso, anotalo con `ask --craft "paso delegado: ..." --because "..."`.
- Cerrá el commit con la línea `Risk:` de `commits`; un riesgo pide `review`.

Los verbos: `start` entrevista · `ask` pregunta o decisión de oficio (`--craft`) · `plan`
abre y cierra pasos · `check` compuertas y prueba · `review` veredicto por árbol · `doctor`
salud sin escribir · `level` perfil · `summary` estado · `undo` deshace lo último · `init`,
`skills`, `skill new` y `pack` son de instalación; `--version` no toca nada.

## Qué significa un bloqueo

- `0` listo: seguí.
- `1` podés actuar: la línea `siguiente:` nombra el comando exacto; corrélo.
- `2` solo la persona puede destrabar: no lo reintentes, no lo rodees, preguntá con `ask`.

## El review, cuando toca

`odachi review` imprime el prompt; `--exec "<comando>"` se lo entrega al revisor por stdin
y por `{prompt}`. Un `BLOCKER` causal bloquea: se destraba arreglando el código o con
`--refute` citando el diff. El veredicto vale para ESE árbol: cualquier cambio lo invalida.

## Nunca

- Escribir la prueba a mano o reusar una vieja: la prueba es la salida real de este árbol.
- Decir "listo" con `check` en rojo, o sin haberlo corrido.
- Cambiar `brief.md` o el cerco sin que la persona lo pida.
- Reintentar un exit 2, o inventar el archivo o el flag que falta.
- Agregar trabajo fuera del alcance porque "era fácil": se pregunta, no se decide.
