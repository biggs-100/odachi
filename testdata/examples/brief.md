# turnos-consultorio

## Qué tiene que existir
- Un paciente reserva un turno con un médico en menos de 1 minuto desde el teléfono.
- La secretaria ve todos los turnos del día en una sola pantalla.
- El paciente recibe un recordatorio el día anterior.
- El paciente cancela su turno sin llamar a nadie.

## Listo cuando
- Reservo un turno desde el teléfono y aparece en la agenda de la secretaria.
- La secretaria ve los turnos del día ordenados por hora en una sola pantalla.
- Un turno reservado a las 10:00 envía un recordatorio a las 10:00 del día anterior.
- El paciente cancela desde su teléfono y el turno desaparece de la agenda.
- Dos pacientes no pueden reservar el mismo horario con el mismo médico.

## NO incluye (primera versión)
- Historia clínica y recetas.
- Cobros, facturación y obras sociales.
- Aplicación móvil nativa.
- Múltiples sedes.
- Panel para el médico.

## Decisiones tomadas
- La agenda se maneja en hora local del consultorio, sin zonas horarias por usuario.
- Un turno dura 30 minutos; la duración no se configura todavía.
- Los recordatorios salen por WhatsApp; no se envía email en la primera versión.
- No hay pagos en línea: los cobros afuera del sistema.
- Los datos de los pacientes no salen del servidor del consultorio.
