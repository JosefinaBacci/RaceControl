# Sprint 2 — Lista de tareas

Consigna: `docs/Sprint2_GrupoRojoFerrari.pdf`. Los pendientes que no pertenecen a ninguna US están en
`docs/sprint2/TODO.md`; las decisiones de modelo, en `docs/sprint2/02-modelo-de-datos.md`.

Cada US se cierra con sus criterios de éxito verificados (prueba independiente incluida), `make check`
en verde y un commit por cambio lógico.

## Estado

| US | Nombre | SP | Valor | Est. | Tiempo real registrado | Estado |
|----|--------|----|-------|------|------------------------|--------|
| US7 | Gestión de calendario | 5 | 8 | 5h | 2.3h | ☑ |
| US8 | Visualización y descarga de calendario | 3 | 8 | 3.5h | | ◐ |
| US11 | Gestión de pilotos | 5 | 13 | 5h | | ☑ |
| US12 | Visualización de pilotos | 3 | 13 | 4h | | ☑ |
| US9 | Gestión de puntajes | 3 | 8 | 3.5h | | ☑ |
| US23 | Visualización de puntajes | 5 | 13 | 5h | | ☑ |
| US24 | Visualización y confirmación de recibo de puntajes | 5 | 13 | 6h | | ☑ |
| US17 | Mensajería interna segura | 8 | 21 | 10h | | ☐ |

**Total estimado:** 42h

☑ cerrada · ◐ en curso · ☐ sin empezar

El orden de la tabla es el de implementación: el calendario es la base de los puntajes, y los
pilotos tienen que existir antes de cargar un resultado.

---

## Modelo de datos (transversal) ☑

Esquema `000002`–`000026` ya aplicado. Puntos de la revisión del modelo:

- [x] §A.6: catálogo `circuits` y eliminación de `events.name` (se edita `000016`, requiere `make db-reset`)
- [x] F6 / B.1: puntaje solo sobre carreras (`score_entries` y `score_change_log` → `races`)
- [x] F6 / B.2: acuse solo de un admin de esa escudería (`000023`, con US24; `02-modelo-de-datos.md` §9)
- [x] `000018`: categoría del equipo = categoría de la carrera, un ítem por piloto en cada cambio,
      evento con duración positiva, nombres de `season_drivers` (F8)
- [x] F8: renombrar los triggers de `season_drivers` (hecho por `000018`/`000019`; quedan
      `season_drivers_assert_category` y `season_drivers_set_updated_at`)
- [x] §C: poda de triggers (`000020`, ver `02-modelo-de-datos.md` §8). Se elimina el guardia del lado
      hijo; la ronda cancelada queda en el servicio y `score_entries.team_id` (F4) no cambia
- [x] F9: volcar las decisiones en `02-modelo-de-datos.md` y registrar en `prompts-ia.md`
- [x] Modelo: tests del invariante de una hija por evento y de los códigos `RC001`/`RC002`
      (`backend/tests/db/event_integrity_test.go`, `race_finality_test.go`, `score_guard_test.go`)

## US7 — Gestión de calendario

*Est. 5h · SP 5 · Valor 8*

- [x] Etiqueta del evento compuesta en `internal/calendar` (por tipo, con unit tests)
- [x] Queries `sqlc` de eventos y circuitos
- [x] Servicio: alta, modificación y baja de eventos (evento + hija en una transacción). La baja de
      una carrera da de baja también su ronda, para no dejar una carrera activa sobre una ronda
      cancelada (`02-modelo-de-datos.md` §8)
- [x] Validaciones: campos obligatorios, `ends_at > starts_at` (`000018`), ronda única por campeonato,
      fecha congelada si la carrera ya tiene resultados (`409`)
- [x] Endpoints restringidos a `fia_admin`; lectura pública
- [x] Pantalla de gestión del calendario (crear, editar, eliminar con confirmación)
- [x] Tests de integración (incluido acceso denegado a otros roles)

**Criterios de éxito**

- [x] El administrador puede crear carreras con todos los datos correspondientes
- [x] El administrador puede modificar y eliminar fechas existentes
- [x] Los cambios se reflejan de inmediato en el calendario oficial — la vista pública lee la API
      desde US8
- [x] Los demás roles no pueden crear, modificar ni eliminar carreras
- [x] El sistema valida los campos obligatorios y las fechas inválidas

## US8 — Visualización y descarga de calendario

*Est. 3.5h · SP 3 · Valor 8*

- [x] Vista del calendario sobre la API (reemplaza el mock)
- [x] Filtro por categoría
- [x] Descarga en formato estándar, incluido PDF
- [ ] Verificar que el archivo descargado se abra sin conexión

**Criterios de éxito**

- [x] Cualquier usuario puede ver el calendario actualizado
- [x] El usuario puede filtrar los eventos por categoría
- [x] El usuario puede descargar el calendario en un formato estándar
- [ ] El archivo descargado se puede consultar sin conexión y contiene los datos correctos
      (pendiente de la verificación manual)

## US11 — Gestión de pilotos

*Est. 5h · SP 5 · Valor 13*

- [x] Queries y servicio de pilotos (alta, edición, número por temporada, rol titular/suplente)
- [x] Validaciones de datos
- [x] Restringir la gestión a los pilotos de la propia escudería
- [x] Modelo: `season_drivers` deriva la categoría del equipo al 1 de enero, así que un piloto que se
      suma a mitad de temporada no puede tener número y uno que cambia de categoría queda con la
      vieja. Una fila por (piloto, temporada, categoría), con la categoría elegida por el admin y un
      chequeo de que el piloto tenga una asignación en esa categoría durante la temporada (000019)
- [x] Modelo: una asignación no se puede terminar (un piloto liberado sigue en la alineación).
      Fila de "baja" en `driver_team_assignments` (sin escudería desde una fecha), sin romper el
      append-only (`000022`, `team_id` nullable)
- [x] Modelo: una asignación retroactiva cambia `assigned_team()` para fechas pasadas y deja viejo el
      equipo guardado en los puntajes. Guardia: una asignación nueva empieza después de la última del
      piloto y nunca antes de una carrera en la que ya tenga resultados (`000022`, RC004)
- [x] Formulario de alta y edición en la app del responsable de escudería
- [x] Tests de integración (incluido intento sobre un piloto de otra escudería)

**Criterios de éxito**

- [x] El responsable de escudería puede cargar nuevos pilotos con su información
- [x] Puede modificar los datos de sus pilotos
- [x] No puede ver ni editar pilotos de otras escuderías
- [x] El sistema valida los datos obligatorios
- [x] Los cambios se reflejan en el sistema de forma inmediata

## US12 — Visualización de pilotos

*Est. 4h · SP 3 · Valor 13*

- [x] Listado público de pilotos agrupado por escudería (alineación vigente; una asignación ya se
      puede terminar con la fila de baja de `000022`, ver US11)
- [x] Vista de detalle de cada piloto
- [x] Búsqueda y filtros
- [x] Acceso de solo lectura para todos los usuarios

**Criterios de éxito**

- [x] Cualquier usuario puede consultar la información de los pilotos
- [x] La alineación vigente se muestra agrupada por escudería
- [x] Los datos mostrados coinciden con los cargados por cada responsable de escudería
- [x] Se puede buscar o filtrar pilotos
- [x] Los usuarios sin permisos de edición no pueden modificar la información

## US9 — Gestión de puntajes

*Est. 3.5h · SP 3 · Valor 8*

- [x] Formulario de carga y modificación de puntajes por carrera
- [x] Validaciones (negativos, vacíos, pilotos repetidos)
- [x] Recalcular la clasificación general al modificar un puntaje
- [x] Restringir a `fia_admin` y registrar fecha y autor de cada modificación (`score_change_log`)
- [x] Modelo: nada impide cambiar `score_entries` sin declarar quién lo hizo. Guardia de auditoría
      (`000026`, `RC005`): ninguna escritura pasa sin `SET LOCAL racecontrol.actor_id`. La versión
      (`score_change_log` + sus ítems) la arma el **servicio** en la misma transacción, no el trigger:
      agrupar las filas de una edición exigía estado entre filas (un `change_log_id` local o tablas de
      transición), más maquinaria de la que la auditoría vale (`02-modelo-de-datos.md` §8)
- [x] Traducir `RC002` (piloto de otra categoría que la carrera) a un 4xx
- [ ] Decidir si se guarda el orden de llegada: hoy la posición se deriva de los puntos (`RANK()` en
      la vista `race_scores`, `000024`), así que todos los pilotos que terminan fuera de los puntos
      comparten la misma posición. Si se carga el orden, `race_scores` debe ordenar por él

**Criterios de éxito**

- [x] El administrador puede cargar y modificar los puntajes de cada carrera
- [x] El sistema rechaza valores inválidos, como negativos o vacíos
- [x] La clasificación general se actualiza automáticamente tras cada cambio
- [x] Cada modificación queda registrada con fecha y autor
- [x] Solo el administrador de la FIA puede editar puntajes

## US23 — Visualización de puntajes

*Est. 5h · SP 5 · Valor 13*

- [x] Vista pública de resultados por carrera (`GET /results`, `GET /results/{id}`)
- [x] Clasificación del campeonato de pilotos y de escuderías (`GET /standings`, calculada sobre la
      vista `race_scores` de `000024`; los empates comparten posición)
- [x] Filtro de carrera
- [x] Acceso público sin iniciar sesión

**Criterios de éxito**

- [x] Cualquier persona puede ver los puntajes de cada carrera sin iniciar sesión
- [x] Se muestra la clasificación acumulada de pilotos y escuderías
- [x] Los puntajes mostrados coinciden con los cargados por el administrador

## US24 — Visualización y confirmación de recibo de puntajes

*Est. 6h · SP 5 · Valor 13*

- [x] Vista de puntajes para el responsable de escudería (pestaña Puntajes)
- [x] Notificación al publicar o modificar puntajes (`GET /team/score-notices`)
- [x] ~~Modelo: una notificación solo apunta al evento, así que no distingue "publicado" de
      "corregido". Agregar el tipo y el vínculo al cambio de puntajes que anuncia (como el acuse), con
      la unicidad por (usuario, cambio).~~ Resuelto por derivación, sin columnas nuevas en
      `notifications`: la vista `score_notices` (`000025`) arma un aviso por cambio de
      `score_change_log` y por escudería afectada, *publicación* si es el primer cambio de la carrera y
      *corrección* si no (`02-modelo-de-datos.md` §9)
- [x] Modelo (F6 / B.2): el acuse toma la escudería del usuario que acusa (trigger que la deriva, un
      admin FIA no tiene) y exige que esa escudería figure en los ítems del cambio acusado (`000023`,
      `RC005`)
- [x] Botón de confirmación de recibo y su estado (pendiente / confirmada)
      (`POST /team/score-notices/{id}/acknowledgement`, idempotente)
- [x] Vista de seguimiento de confirmaciones para el administrador (`GET /score-changes`, pestaña
      Acuses)

**Criterios de éxito**

- [x] El responsable de escudería ve los puntajes de cada carrera
- [x] Recibe una notificación cuando se publican o modifican los resultados
- [x] Puede confirmar la recepción, y la confirmación queda registrada con fecha y hora
- [x] El estado de la notificación (pendiente o confirmada) es visible
- [x] El administrador puede ver qué escuderías confirmaron

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
