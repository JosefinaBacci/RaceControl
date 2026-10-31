# Plan — equipo controls (US21, US22)

## Qué es el "resultado" de un control técnico

El enunciado dice: *"Personal Administrativo de la FIA podrá ingresar fechas y tipos de distintos
controles técnicos de los autos de todas las escuderías, registrando en cada caso si pasaron o no
cada control"*. El resultado es entonces **por escudería**: aprobado / no aprobado, con
observaciones. La fecha y el tipo del control ya existen como evento `technical_control` del
calendario (US7: ventana `starts_at`/`ends_at`, temporada, categoría y `control_type`). Lo que se
registra en US21 es, para un control ya programado, el resultado de una escudería: **fecha en que se
realizó** (dentro de la ventana del evento), aprobado, observaciones, autor y momento de registro.

Un resultado registrado es **constancia oficial**: append-only (`reject_mutation()`, como `000017`)
y uno por escudería y control (`UNIQUE`). No hay corrección en este sprint (ver preguntas).

## Modelo (migraciones 000030 y 000031)

`000030_technical_control_results`:

```sql
ALTER TABLE teams ADD CONSTRAINT teams_id_category_unique UNIQUE (id, category_id);
ALTER TABLE technical_controls ADD CONSTRAINT technical_controls_event_category_unique UNIQUE (event_id, category_id);

CREATE TABLE technical_control_results (
    id           bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_id     bigint      NOT NULL,
    category_id  bigint      NOT NULL,
    team_id      bigint      NOT NULL,
    conducted_on date        NOT NULL,
    passed       boolean     NOT NULL,
    notes        text        NOT NULL DEFAULT '',
    recorded_by  bigint      NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    recorded_at  timestamptz NOT NULL DEFAULT now(),
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    -- the team must belong to the control's category: enforced by the schema, not only the service
    CONSTRAINT tcr_control_fkey FOREIGN KEY (event_id, category_id)
        REFERENCES technical_controls (event_id, category_id) ON DELETE RESTRICT,
    CONSTRAINT tcr_team_fkey FOREIGN KEY (team_id, category_id)
        REFERENCES teams (id, category_id) ON DELETE RESTRICT,
    CONSTRAINT tcr_one_per_team UNIQUE (event_id, team_id),
    CONSTRAINT tcr_notes_length CHECK (char_length(notes) <= 1000)
);
CREATE INDEX technical_control_results_conducted_on_idx ON technical_control_results (conducted_on);
CREATE TRIGGER technical_control_results_append_only BEFORE UPDATE OR DELETE
    ON technical_control_results FOR EACH ROW EXECUTE FUNCTION reject_mutation();
```

`000031_event_results_freeze`: función genérica compartida con tyres
`assert_event_without_results_is_final()` (TG_ARGV[0] = tabla de resultados, TG_ARGV[1] = columna
clave de la fila guardada), `RAISE ... USING ERRCODE = 'RC020'`. Triggers míos:

- `events` `BEFORE UPDATE OF starts_at, ends_at, deactivated_at WHEN (OLD.kind = 'technical_control')`
  → con resultados no se mueve ni se cancela (si no, la fecha registrada quedaría fuera de la ventana).
- `technical_controls` `BEFORE UPDATE OF control_type, season_id, category_id` → no cambia el tipo
  de un control ya registrado.

Tyres agrega sus triggers sobre la misma función en `000032`. El servicio de calendario traduce
`RC020` a 409. Códigos RC acordados: sanciones RC010-019, controles RC020-029, neumáticos RC030-039.

## Validaciones ("rechaza fechas inválidas")

- `conductedOn` obligatorio, formato `YYYY-MM-DD` válido (no `2026-02-30`).
- Dentro de la ventana del evento en días UTC (`starts_at::date ≤ conductedOn ≤ ends_at::date`).
- No futuro: `conductedOn ≤ hoy` y el control ya empezó; no se registra un control que no ocurrió.
- Control cancelado → 400. Control inexistente o evento de otro tipo → 404.
- `teamId` obligatorio, escudería activa y de la categoría del control → 400 field `teamId`.
- `passed` obligatorio (booleano explícito: `null`/ausente → 400). `notes` opcional, ≤ 1000 chars,
  obligatorio cuando `passed = false` (un "no aprobado" sin motivo no sirve como constancia).
- Ya registrado para esa escudería → 409 (también bajo doble envío concurrente: lo resuelve el UNIQUE).
- Autor y momento del registro los pone el servidor (sesión y `now()`), nunca el cliente.
- Filtros: `from`/`to` mal formados o `from > to` → 400; `type` desconocido → lista vacía (es un filtro, no un alta).

## Endpoints

| Método y ruta | Rol | Respuestas |
|---|---|---|
| `POST /technical-controls/{eventId}/results` `{teamId, conductedOn, passed, notes}` | `fia_admin` | 201 resultado · 400 · 401 · 403 · 404 · 409 |
| `GET /technical-controls/results?from&to&type&category&team` | público | 200 `{"results":[...]}` · 400 |
| `GET /technical-controls/types` | público | 200 `{"types":[{code,name}]}` (tipos distintos de controles con resultados) |
| `GET /technical-controls/{eventId}/results` | `fia_admin` | 200 resultados + escuderías pendientes de la categoría (para el formulario) · 404 |

Fila: `{id, eventId, conductedOn, type:{code,name}, passed, notes, seasonYear, categoryCode,
team:{id,name}, recordedBy:{username}, recordedAt}` — mismas claves comunes que tyres. Ningún
`PUT`/`PATCH`/`DELETE` existe (405), y la tabla rechaza UPDATE/DELETE aunque se llegue por SQL.

Capas: `internal/technicalcontrols` (servicio + reglas) → `sql/technical_controls.sql` (sqlc) →
`internal/transport/technical_controls_handlers.go`. Router: un bloque propio.

## Piezas compartidas (las comiteo primero, en commits propios)

Backend:
- `internal/domain/history.go`: `DateRange{From, Until *time.Time}` (Until exclusivo, días UTC),
  `ParseDateRange`, `HistoryFilter{Range, Type}`, `ParseDay`, `EventWindow`, `ValidateConductedOn`.
- `internal/transport/history_filter.go`: `dateRangeOf(r)`, `historyFilterOf(r)`.
- `internal/db/pgconv.go`: `OptionalDate`, `OptionalTimestamptz`.
- Migración `000031` (función de congelamiento) + mapeo `RC020` en `calendar.translateWriteError`.

App (`app/src/features/history/`):
- `historyFilter.ts` (`DateRangeValue`, `HistoryFilterValue`, `validateDateRange`, `dateRangeQuery`),
  `DateRangeFilter.tsx`, `HistoryFilters.tsx` (rango + chips de tipo).
- `HistoryList.tsx` genérico (fecha, tipo, `OutcomeBadge`, notas, autor; `renderSubject` por feature).
- `ResultFields.tsx` + `resultForm.ts` (sujeto, fecha, aprobado/no aprobado, observaciones).
- `app/src/api/history.ts` (`HistoryEntry`, `HistoryQuery`).
- Pestaña pública "Técnica" `app/(public)/technical.tsx` con registro de secciones
  `technicalSections.ts` (tyres agrega una línea), y `features/calendar/recordResultRoutes.ts`
  (kind → ruta de registro) usado por `app/fia/calendar/[id].tsx`.

Sanciones consume `DateRange`/`dateRangeOf` y `DateRangeFilter`. Tyres consume todo lo anterior.

## Pantallas por rol

- **FIA:** en el detalle de un control del calendario, botón "Registrar resultados" →
  `app/fia/calendar/[id]/control-results.tsx`: fecha y tipo del control (solo lectura), lista de
  escuderías de la categoría con su resultado o "pendiente", y el formulario `ResultFields`
  (escudería, fecha realizada, aprobado, observaciones). Al guardar, la fila pasa a registrada.
- **Público y escudería:** pestaña "Técnica" → sección Controles: historial con filtros por rango
  de fechas y tipo (y categoría), cada fila con fecha, tipo, escudería, resultado y observaciones.
  Solo lectura: no hay acciones. La escudería llega por su pestaña "Sitio público".

## Tests (`backend/tests/technicalcontrols`, `tests/transport`, `tests/db`)

- Alta feliz → 201, y aparece igual en `GET /technical-controls/results` (datos coinciden).
- 401 sin sesión, 403 `team_admin`, 403 FIA con sesión revocada (401).
- Fechas: mal formada, inexistente (`2026-02-30`), antes/después de la ventana, futura, control
  futuro, control cancelado → 400. Filtros `from`/`to` inválidos y `from > to` → 400.
- Escudería de otra categoría, inactiva, inexistente → 400; evento de tipo carrera → 404.
- `passed` ausente, `notes` > 1000, "no aprobado" sin observaciones, campos desconocidos → 400.
- Duplicado → 409; 10 POST concurrentes de la misma escudería → exactamente un 201.
- Inmutabilidad: `PUT/PATCH/DELETE` → 405; UPDATE/DELETE por SQL rechazados por el trigger;
  mover/cancelar/cambiar tipo de un control con resultados → 409 (RC020).
- Filtros: rango inclusivo en los bordes, `type`, combinación, sin filtros.
- Dominio: `ParseDateRange` y `ValidateConductedOn` con tabla de casos.

## Decisiones del manager (aplicadas en la implementación)

1. **Correcciones:** sí, append-only. Una corrección es una fila nueva con `supersedes_id` y motivo
   obligatorio; el público ve el resultado vigente marcado "Corregido" y la FIA ve la cadena
   (`POST /technical-controls/results/{id}/corrections`).
2. **Fecha y tipo** quedan en el calendario (US7); la pantalla de resultados de la FIA tiene el atajo
   "Nuevo control técnico", que abre el formulario del calendario con el tipo preseleccionado.
3. **UTC** para los días, también en la app (`formatDay` compartido).
4. **Sin nombres de usuario en las respuestas públicas**; `recordedBy` solo en los endpoints de la FIA.
5. Tipo desconocido: lista vacía (texto libre); en neumáticos, 400 (catálogo).

Cambios respecto del borrador: la función de congelamiento quedó en `000030` (commit compartido) y la
tabla en `000031`; las observaciones aceptan hasta 2000 caracteres (regla común con neumáticos); un
evento que todavía no empezó tampoco admite resultados; las rutas de la FIA son
`/fia/results/controls/[id]` y la pestaña "Resultados" de la FIA muestra una tarjeta por tipo de evento.
