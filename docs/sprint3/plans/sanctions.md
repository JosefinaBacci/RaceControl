# Plan — equipo sanciones (US15, US25, US26)

> **Decisiones finales (aprobadas por el manager).** Esto prevalece sobre lo que dice el resto del plan:
> 1. Cada versión se notifica a **todas** las escuderías activas del campeonato, no solo a la
>    responsable. Los destinatarios quedan fijos en `sanction_recipients` al escribir la versión.
>    La sanción a la propia escudería o a su piloto se destaca en la bandeja.
> 2. Las revocadas siguen visibles en lo público, marcadas "Revocada" y con su nota. El filtro de
>    estado (`status=all|active|revoked`) toma `all` por defecto, y el detalle muestra el historial
>    de versiones.
> 3. No hay descuento de puntos de campeonato: el motivo está en `docs/sprint2/02-modelo-de-datos.md` §11.
> 4. La modificación es `PUT /sanctions/{id}`, con los términos completos y una nota. El seguimiento
>    de la FIA y la autoría solo aparecen en endpoints de la FIA.

## Decisiones de modelo

| Pregunta | Decisión | Por qué |
|----------|----------|---------|
| ¿Qué es la "penalización"? | Tabla de referencia `penalty_types` (seed) + `penalty_value` entero opcional según la unidad del tipo | El enunciado nombra incidentes de carrera, cambios no autorizados y arreglos fuera de horario; el reglamento FIA responde con un catálogo cerrado de penas, algunas con magnitud. Un tipo nuevo es un `INSERT`, no un `CHECK` nuevo (regla §2 del modelo) |
| Tipos | `warning` (advertencia), `reprimand` (reprimenda), `time_penalty` (s), `drive_through`, `stop_and_go` (s), `grid_penalty` (puestos), `license_points` (puntos de superlicencia), `fine` (€), `disqualification`, `suspension` (carreras) | Son las penas del Código Deportivo Internacional / reglamento deportivo de F1. Deduccion de puntos de campeonato **no** se incluye: la sanción es un registro y no recalcula la clasificación; si hace falta, la FIA corrige el resultado con US23 |
| ¿Cómo se valida el valor? | `penalty_types.unit` (NULL = sin valor) y `max_value`. Con unidad: valor obligatorio entre 1 y `max_value`; sin unidad: valor prohibido. Lo valida el servicio (400 con `field`) y un trigger lo garantiza (RC010) | La regla depende de otra tabla, así que no entra en un `CHECK` |
| Destinatario | `driver_id` **o** `team_id`, `CHECK (num_nonnulls(driver_id, team_id) = 1)` (diseño de Sprint 1 §2) | Integridad referencial en las dos ramas |
| ¿Qué escudería se notifica? | Solo la **escudería responsable**: la sancionada, o la del piloto **en la fecha de la sanción** (`assigned_team(driver_id, issued_at)`). Se guarda en `sanctions.accountable_team_id`, completada por trigger y nunca escrita a mano | Igual que US24: se notifica a quien la decisión toca. Se guarda (desnormalización deliberada, como `score_entries.team_id`) porque una asignación nueva puede empezar antes de una sanción ya registrada y derivarla después cambiaría el histórico. Un piloto sin escudería en esa fecha no se puede sancionar (400) |
| ¿Se puede modificar? | Sí. Cada alta, modificación o revocación es una **versión** append-only en `sanction_versions`; la versión vigente es la última. La escudería acusa recibo **por versión**, como `000009` en US24 | US26 pide notificar al registrar **o modificar**: una corrección deja viejo el acuse anterior y vuelve a quedar pendiente. Nada se sobrescribe: el historial es evidencia |
| ¿Qué se puede editar? | Motivo, tipo y valor de la penalización, evento. **No** el destinatario: cambiarlo cambiaría a quién se notificó; si está mal, se revoca y se registra otra | Mantiene el acuse coherente con la escudería notificada |
| ¿Se puede revocar? | Sí: versión con `action = 'revoked'` y nota obligatoria. Es la baja lógica de la sanción (nunca `DELETE`). Se notifica a la escudería como cualquier cambio | La FIA anula sanciones en apelación; la revocación también es una decisión que la escudería debe acusar |
| Fecha | `issued_at` (alta) la pone el servidor (`now()`); cada versión tiene su `changed_at` y su autor | "Fecha y autor" de US15 sin confiar en el cliente |
| Notificación | Derivada, sin filas en `notifications`: vista `sanction_notices` (una fila por versión × escudería responsable, con su acuse) | Mismo patrón que `score_notices` (000025): cualquier versión escrita se vuelve aviso sola |

## Migraciones (000034–000037, solo DDL)

- **000034_penalty_types**: `penalty_types (id identity, code UNIQUE, name, unit CHECK IN ('seconds','grid_places','points','euros','races') NULL, max_value int NULL, CHECK (unit IS NULL) = (max_value IS NULL), created_at, updated_at)`. Filas en `seeds/reference.sql`.
- **000035_sanctions**:
  - `sanctions (id identity, driver_id → drivers RESTRICT NULL, team_id → teams RESTRICT NULL, accountable_team_id → teams RESTRICT NOT NULL, issued_by → users RESTRICT, issued_at default now(), created_at, updated_at, CHECK num_nonnulls(driver_id, team_id) = 1, CHECK team_id IS NULL OR accountable_team_id = team_id)`; trigger BEFORE INSERT completa `accountable_team_id` (RC011 si el piloto no tiene escudería) y trigger que impide cambiar destinatario/fecha/autor (RC012).
  - `sanction_versions (id identity, sanction_id → sanctions RESTRICT, action CHECK IN ('issued','amended','revoked'), penalty_type_id → penalty_types RESTRICT, penalty_value int NULL, reason text CHECK btrim <> '' y largo ≤ 2000, event_id → events RESTRICT NULL, note text NULL (obligatoria en amended/revoked por CHECK), actor_user_id → users RESTRICT, changed_at default now())`; trigger del valor de penalización (RC010); `reject_mutation()` append-only; índices (`sanction_id, changed_at DESC`), (`accountable_team_id`), (`driver_id`), (`issued_at`).
  - Vista `current_sanctions`: última versión por sanción (`DISTINCT ON`) + `is_revoked`.
- **000036_sanction_acknowledgements**: `sanction_acknowledgements (id identity, sanction_version_id → sanction_versions RESTRICT, team_id → teams RESTRICT, acknowledged_by → users RESTRICT, acknowledged_at default now(), UNIQUE (sanction_version_id, team_id))`; trigger que toma `team_id` del usuario y rechaza (RC013) si no es la escudería responsable (igual que 000023); append-only.
- **000037_sanction_notices**: vista `sanction_notices` (versión × `accountable_team_id` + acuse + `is_latest`).

Down de cada una en orden inverso. Demo: 3–4 sanciones en `seeds/demo.sql` (ids fijos + `setval`), una con dos versiones y una con acuse.

## Endpoints

| Método y ruta | Rol | Respuestas |
|---------------|-----|-----------|
| `GET /penalty-types` | público | 200 |
| `GET /sanctions?driverId=&teamId=&category=&from=&to=` | público | 200 (excluye revocadas); 400 filtros mal formados. `teamId` = escudería responsable (incluye sanciones a sus pilotos) |
| `GET /sanctions/{id}` | público | 200 detalle (destinatario, motivo, penalización, evento, fecha de alta, última modificación); 404 inexistente o revocada |
| `POST /sanctions` | fia_admin | 201; 400 JSON inválido / campos (`reason`, `penaltyType`, `penaltyValue`, `target`, `eventId`); 401; 403 |
| `PATCH /sanctions/{id}` | fia_admin | 200 nueva versión `amended` (nota obligatoria); 200 sin versión si no cambió nada (no re-notifica); 400; 404; 409 revocada |
| `POST /sanctions/{id}/revocation` `{note}` | fia_admin | 200; 400 sin nota; 404; 409 ya revocada |
| `GET /sanction-acknowledgements` | fia_admin | 200: por sanción, versión vigente, escudería responsable, estado pendiente/confirmada, quién y cuándo, e historial de versiones |
| `GET /team/sanction-notices` | team_admin | 200 `{pendingCount, notices[]}` (incluye revocadas y versiones anteriores) |
| `POST /team/sanction-notices/{versionId}/acknowledgement` | team_admin | 200 idempotente (conserva el primer acuse); 404 versión ajena o inexistente (RC013 → 404, como US24) |

Validación: cuerpo con campos desconocidos rechazado (como el resto de la API), `reason` recortado 3..2000, exactamente uno de `driverId`/`teamId`, existencia y actividad de piloto/escudería, evento existente y del mismo campeonato que la escudería responsable, tipo existente, valor según unidad. Concurrencia: PATCH/revocación toman `SELECT … FOR UPDATE` sobre la sanción, así dos ediciones simultáneas no se pisan y una revocación doble da un 200 y un 409.

Capas: `internal/sanctions` (service + errores), `sql/sanctions.sql`, `transport/sanctions_handlers.go`, `sanction_notices_handlers.go`, un bloque en `router.go` y `cmd/api/main.go`.

## App

- `src/api/sanctions.ts` + `src/data/sanctions.ts` (hooks sobre `useRemote`), `TeamSanctionNoticesProvider` como el de puntajes.
- **Público** `(public)/sanctions.tsx`: lista real con filtros piloto, escudería, categoría y rango de fechas (componente compartido de controles); nueva ruta `(public)/sanctions/[id].tsx` con el detalle. Portada pública: últimas 3 desde la API.
- **FIA**: `fia/sanctions/` lista + formulario de alta/edición (selector de destinatario piloto/escudería, tipo, valor si corresponde, motivo, evento opcional, nota en la edición) + botón Revocar; la pantalla `fia/notifications` suma una sección "Acuses de sanciones". KPI "Sanciones sin acuse" real.
- **Escudería** `team/inbox.tsx`: avisos reales (registrada / modificada / revocada), estado pendiente/confirmada con `AcknowledgementBadge`, botón Acusar recibo, badge del tab con `pendingCount`; KPI del dashboard real.
- Borrar `sanctions` y el tipo `Sanction` de `src/mocks/data.ts`, `useSanctions`/`useTeamInbox`/`usePendingSanctions` de `hooks.ts`, y `MockNotice` si queda sin uso.

## Con otros equipos

- **Controles** (acordado): consumo con `git merge` de su commit `domain.DateRange{From, Until}` y `domain.ParseDateRange` (`internal/domain/history.go`, días en UTC, `Until` exclusivo = `to` + 1 día), `dateRangeOf(r)` (`internal/transport/history_filter.go`), `db.OptionalTimestamptz` (`internal/db/pgconv.go`), y en la app `features/history/DateRangeFilter.tsx` + `historyFilter.ts` (`DateRangeValue`, `validateDateRange`, `dateRangeQuery`). El filtro de sanciones aplica sobre `issued_at`.
- **Códigos SQLSTATE** (acordado con controles y neumáticos): sanciones RC010–RC019, controles RC020–RC029, neumáticos RC030–RC039.

## Tests (`backend/tests/sanctions`, `tests/transport`, `tests/db`)

- Esquema: `CHECK` de un solo destinatario (ninguno / ambos), trigger de valor (falta, sobra, fuera de rango), escudería responsable derivada en la fecha (piloto que cambió de equipo después), piloto liberado rechazado, versiones y acuses append-only (UPDATE/DELETE fallan), destinatario inmutable.
- Servicio/HTTP: alta piloto y escudería con fecha y autor; sin motivo, motivo en blanco, sin destinatario, ambos destinatarios, ids inexistentes, tipo inválido, valor negativo/0/enorme/no entero, evento de otro campeonato, JSON roto, campos extra → 400.
- RBAC: anónimo 401 y team_admin 403 en POST/PATCH/revocación/seguimiento; fia_admin 403 en el acuse; público lee sin sesión; métodos de escritura sobre rutas públicas no existen.
- US26: escudería responsable ve el aviso pendiente, acusa (fecha y hora guardadas), repetir es idempotente; otra escudería → 404; modificar crea aviso nuevo pendiente; revocar notifica; FIA ve quién confirmó cada versión.
- Filtros: piloto, escudería (incluye sus pilotos), rango de fechas, combinados, formato inválido; datos de la lista = datos registrados.
- Concurrencia: dos acuses simultáneos → una fila; dos revocaciones simultáneas → 200 + 409; ediciones simultáneas → dos versiones ordenadas.

## Preguntas abiertas para el manager

1. ¿Notificar solo a la escudería responsable (propuesto) o a **todas** las escuderías del campeonato? El enunciado dice "inmediatamente disponibles a todas las escuderías" (lo cubre la vista pública) y "las escuderías deberán dejar asentado que se han notificado". Cambiar a todo el campeonato es solo la vista `sanction_notices` y el trigger del acuse.
2. ¿La lista pública oculta las revocadas (propuesto) o las muestra marcadas como "Revocada"?
3. ¿OK dejar fuera la deducción de puntos de campeonato (no recalcula la clasificación)?
