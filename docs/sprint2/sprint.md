# Sprint 2 — Lista de tareas

Consigna: `docs/Sprint2_GrupoRojoFerrari.pdf`. Los pendientes que no pertenecen a ninguna US están en
`docs/sprint2/TODO.md`; las decisiones de modelo, en `docs/sprint2/02-modelo-de-datos.md`.

Cada US se cierra con sus criterios de éxito verificados (prueba independiente incluida), `make check`
en verde y un commit por cambio lógico.

## Estado

| US | Nombre | SP | Valor | Est. | Tiempo real registrado | Estado |
|----|--------|----|-------|------|------------------------|--------|
| US7 | Gestión de calendario | 5 | 8 | 5h | | ◐ |
| US8 | Visualización y descarga de calendario | 3 | 8 | 3.5h | | ☐ |
| US11 | Gestión de pilotos | 5 | 13 | 5h | | ☐ |
| US12 | Visualización de pilotos | 3 | 13 | 4h | | ☐ |
| US9 | Gestión de puntajes | 3 | 8 | 3.5h | | ☐ |
| US23 | Visualización de puntajes | 5 | 13 | 5h | | ☐ |
| US24 | Visualización y confirmación de recibo de puntajes | 5 | 13 | 6h | | ☐ |
| US17 | Mensajería interna segura | 8 | 21 | 10h | | ☐ |

**Total estimado:** 42h

☑ cerrada · ◐ en curso · ☐ sin empezar

El orden de la tabla es el de implementación: el calendario es la base de los puntajes, y los
pilotos tienen que existir antes de cargar un resultado.

---

## Modelo de datos (transversal) ◐

Esquema `000002`–`000017` ya aplicado. Pendiente de la revisión del modelo:

- [x] §A.6: catálogo `circuits` y eliminación de `events.name` (se edita `000016`, requiere `make db-reset`)
- [ ] F6: puntaje solo sobre carreras (B.1) y acuse solo de un admin de esa escudería (B.2)
- [ ] F8: renombrar los triggers de `season_drivers` (cosmético)
- [ ] F9: volcar las decisiones en `02-modelo-de-datos.md` y registrar en `prompts-ia.md`

## US7 — Gestión de calendario

*Est. 5h · SP 5 · Valor 8*

- [ ] Etiqueta del evento compuesta en `internal/calendar` (por tipo, con unit tests)
- [ ] Queries `sqlc` de eventos y circuitos
- [ ] Servicio: alta, modificación y baja de eventos (evento + hija en una transacción)
- [ ] Validaciones: campos obligatorios, `ends_at >= starts_at`, ronda única por campeonato,
      fecha congelada si la carrera ya tiene resultados (`409`)
- [ ] Endpoints restringidos a `fia_admin`; lectura pública
- [ ] Pantalla de gestión del calendario (crear, editar, eliminar con confirmación)
- [ ] Tests de integración (incluido acceso denegado a otros roles)

**Criterios de éxito**

- [ ] El administrador puede crear carreras con todos los datos correspondientes
- [ ] El administrador puede modificar y eliminar fechas existentes
- [ ] Los cambios se reflejan de inmediato en el calendario oficial
- [ ] Los demás roles no pueden crear, modificar ni eliminar carreras
- [ ] El sistema valida los campos obligatorios y las fechas inválidas

## US8 — Visualización y descarga de calendario

*Est. 3.5h · SP 3 · Valor 8*

- [ ] Vista del calendario sobre la API (reemplaza el mock)
- [ ] Filtro por categoría
- [ ] Descarga en formato estándar, incluido PDF
- [ ] Verificar que el archivo descargado se abra sin conexión

**Criterios de éxito**

- [ ] Cualquier usuario puede ver el calendario actualizado
- [ ] El usuario puede filtrar los eventos por categoría
- [ ] El usuario puede descargar el calendario en un formato estándar
- [ ] El archivo descargado se puede consultar sin conexión y contiene los datos correctos

## US11 — Gestión de pilotos

*Est. 5h · SP 5 · Valor 13*

- [ ] Queries y servicio de pilotos (alta, edición, número por temporada, rol titular/suplente)
- [ ] Validaciones de datos
- [ ] Restringir la gestión a los pilotos de la propia escudería
- [ ] Formulario de alta y edición en la app del responsable de escudería
- [ ] Tests de integración (incluido intento sobre un piloto de otra escudería)

**Criterios de éxito**

- [ ] El responsable de escudería puede cargar nuevos pilotos con su información
- [ ] Puede modificar los datos de sus pilotos
- [ ] No puede ver ni editar pilotos de otras escuderías
- [ ] El sistema valida los datos obligatorios
- [ ] Los cambios se reflejan en el sistema de forma inmediata

## US12 — Visualización de pilotos

*Est. 4h · SP 3 · Valor 13*

- [ ] Listado público de pilotos agrupado por escudería (alineación vigente)
- [ ] Vista de detalle de cada piloto
- [ ] Búsqueda y filtros
- [ ] Acceso de solo lectura para todos los usuarios

**Criterios de éxito**

- [ ] Cualquier usuario puede consultar la información de los pilotos
- [ ] La alineación vigente se muestra agrupada por escudería
- [ ] Los datos mostrados coinciden con los cargados por cada responsable de escudería
- [ ] Se puede buscar o filtrar pilotos
- [ ] Los usuarios sin permisos de edición no pueden modificar la información

## US9 — Gestión de puntajes

*Est. 3.5h · SP 3 · Valor 8*

- [ ] Formulario de carga y modificación de puntajes por carrera
- [ ] Validaciones (negativos, vacíos, pilotos repetidos)
- [ ] Recalcular la clasificación general al modificar un puntaje
- [ ] Restringir a `fia_admin` y registrar fecha y autor de cada modificación (`score_change_log`)

**Criterios de éxito**

- [ ] El administrador puede cargar y modificar los puntajes de cada carrera
- [ ] El sistema rechaza valores inválidos, como negativos o vacíos
- [ ] La clasificación general se actualiza automáticamente tras cada cambio
- [ ] Cada modificación queda registrada con fecha y autor
- [ ] Solo el administrador de la FIA puede editar puntajes

## US23 — Visualización de puntajes

*Est. 5h · SP 5 · Valor 13*

- [ ] Vista pública de resultados por carrera
- [ ] Clasificación del campeonato de pilotos y de escuderías
- [ ] Filtro de carrera
- [ ] Acceso público sin iniciar sesión

**Criterios de éxito**

- [ ] Cualquier persona puede ver los puntajes de cada carrera sin iniciar sesión
- [ ] Se muestra la clasificación acumulada de pilotos y escuderías
- [ ] Los puntajes mostrados coinciden con los cargados por el administrador

## US24 — Visualización y confirmación de recibo de puntajes

*Est. 6h · SP 5 · Valor 13*

- [ ] Vista de puntajes para el responsable de escudería
- [ ] Notificación al publicar o modificar puntajes
- [ ] Botón de confirmación de recibo y su estado (pendiente / confirmada)
- [ ] Vista de seguimiento de confirmaciones para el administrador

**Criterios de éxito**

- [ ] El responsable de escudería ve los puntajes de cada carrera
- [ ] Recibe una notificación cuando se publican o modifican los resultados
- [ ] Puede confirmar la recepción, y la confirmación queda registrada con fecha y hora
- [ ] El estado de la notificación (pendiente o confirmada) es visible
- [ ] El administrador puede ver qué escuderías confirmaron

## US17 — Mensajería interna segura

*Est. 10h · SP 8 · Valor 21*

- [ ] Modelo de conversaciones y mensajes (migración nueva)
- [ ] Envío de mensajes entre responsables de escudería y administradores
- [ ] Bandeja de entrada, enviados y lectura de mensajes
- [ ] Seguridad: autorización por participante, validación de entradas, conexión cifrada
- [ ] Indicador de no leídos y marcado como leído
- [ ] Pruebas de envío, recepción y acceso no autorizado

**Criterios de éxito**

- [ ] Responsables de escudería y administradores pueden enviar y recibir mensajes
- [ ] Solo los participantes de una conversación pueden acceder a sus mensajes
- [ ] Los mensajes se transmiten de forma segura y las entradas se validan
- [ ] Los mensajes no leídos se identifican claramente
- [ ] Un usuario no autenticado no puede acceder a ninguna conversación
