# reserva-de-turno

## Requisitos (textual, del usuario)
R1 · "quiero que el paciente reserve un turno desde el telefono y que no se pisen dos turnos del mismo medico"
R2 · "la secretaria tiene que ver todos los turnos del dia en una sola pantalla"
R3 · "el paciente tiene que poder cancelar sin llamar"

## Listo cuando
- Reservar el mismo horario con el mismo médico dos veces deja un solo turno (R1).
- La agenda del día muestra los turnos ordenados por hora (R2).
- Cancelar desde el teléfono libera el horario para otro paciente (R3).

## Pasos
- [x] T1 · horario libre u ocupado en el dominio, sin base de datos todavía   R1 · 4a1c9f
- [ ] T2 · reserva con verificación de choque de horario   R1
- [ ] T3 · agenda del día ordenada por hora   R2

## Historial
L1 · "quiero que el paciente reserve un turno desde el telefono y que no se pisen dos turnos del mismo medico"
L2 · se decide que un turno dura 30 minutos y no se configura: el usuario no lo pidió.
L3 · queda abierto si la cancelación necesita confirmación por WhatsApp; el humano lo decide antes de T3.
