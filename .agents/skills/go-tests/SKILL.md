---
name: go-tests
description: Trigger: cuando escribís o revisás tests de Go, golden files o flujos de Bubbletea. Carga `tests` para la disciplina y esta skill para el patrón concreto.
---

## Cuándo

Cuando el cambio es Go y hay que decidir qué test escribir, cómo nombrarlo,
cuándo usar `t.TempDir`, `testing.Short`, `teatest` o un golden file.

## Reglas

- Cargá `tests` primero: un test por comportamiento y miralo fallar por la razón
  correcta. Esta skill no repite esa disciplina; elige el patrón.
- Tabla para varios casos: `for _, tt := range cases { t.Run(tt.name, ...) }`. El
  nombre del caso dice el escenario, no el input.
- Probá comportamiento y estado, no detalles privados. Si el test se rompe al
  refactorizar sin cambiar la conducta, está mal ubicado.
- Archivos y permisos: `t.TempDir()`; nunca el home real. Un test que escribe en
  `$HOME` falla en otra máquina.
- Integración lenta o con comandos externos: `testing.Short()` y un `t.Skip` con
  la razón y cuándo se destraba.
- Bubbletea: `Model.Update()` directo para estado; `teatest` sólo para el flujo
  interactivo.
- Golden files: deterministas, se actualizan con el flag `-update` del repo y se
  vuelven a correr sin `-update` para probar que el diff quedó.
- Mocks en el borde del sistema (comando, reloj, red), no en el medio de la lógica.
- Corré el paquete focal primero y la suite completa después; no digas "pasan"
  con un solo paquete.
- Si el test falla, arreglá la causa. No ajustes el test para que pase.

## Nunca

- Un test que depende del orden, la hora o la red.
- `t.Skip` sin razón y sin condición de salida.
- Golden files actualizados a mano sin mirar el diff.
- Testear getters, setters o detalles privados.
- Inventar datos para que un caso pase.
