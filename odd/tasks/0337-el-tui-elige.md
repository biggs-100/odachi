# el-tui-elige
Base: 62aeb0624c5548c5bb8fae77da1c0454dbc229e6

## Requirements (verbatim, from the user)
R1 · "el tui elige el modelo del verificador y el de cada cliente, y lo escribe en la configuracion que ese cliente lee; la ADR 0051 se reemplaza"

## Done when
- el tui muestra y cambia el modelo del verificador y el de cada cliente instalado, y aplicar escribe cada uno en su archivo
- un modelo que no se toco deja su archivo byte a byte identico, y el archivo ajeno conserva cada clave que ya tenia
- odachi verify usa el modelo declarado y lo sustituye en {model}
- la ADR 0053 reemplaza la 0051 y dice por que el selector por cliente vive en el tui

## Steps
- [ ] T1 · odachi verify sustituye {model} con el modelo declarado, el --model lo pisa y el de la sesion es el ultimo recurso   R1 · route:delegated/odachi-worker
- [ ] T2 · init y setup aceptan --model cliente=modelo, para que el verbo haga lo que el tui hace   R1 · route:delegated/odachi-worker
- [ ] T3 · el tui gana una seccion de modelos con el verificador y uno por cliente, y aplicar escribe cada uno   R1 · route:delegated/odachi-worker
- [ ] T4 · la ADR 0053 reemplaza la 0051 y dice por que el selector por cliente vive en el tui   R1 · route:inline

## History
L1 · "el tui elige el modelo del verificador y el de cada cliente, y lo escribe en la configuracion que ese cliente lee; la ADR 0051 se reemplaza"
