# Plan: equipo tyres (US19, US20)

## Decisiones de dominio

- **Qué es una "prueba".** La prueba ya existe como evento `tyre_test` del calendario (US7: fecha,
  campeonato, circuito). US19 le agrega el **resultado oficial**; no se crea un segundo concepto de
  prueba. Una prueba puede evaluar varios compuestos, así que se registra **un resultado por
  compuesto** (`UNIQUE (event_id, compound_id)`).
- **Tipo = compuesto ensayado.** Catálogo `tyre_compounds` (datos, no `switch`): C1 a C6 (slick),
  `intermediate`, `wet`, cada uno con su `kind` (`slick` / `intermediate` / `wet`). Se carga en
  `seeds/reference.sql`. El filtro "por tipo" de US20 filtra por compuesto.
- **Resultado = veredicto + informe.** `passed boolean` (apto / no apto) y `notes text` (hasta 2000
  caracteres: degradación, vueltas, observaciones), obligatorio cuando la prueba no aprueba y
  opcional cuando aprueba: es la regla única compartida `domain.ValidateResultNotes`. Es el mismo par que
  `technical_control_results` (aprobado + observaciones), así el bloque de formulario se comparte.
- **Fecha** = `conducted_on date`, el día en que se corrió la prueba. **Fecha inválida** significa:
  formato que no es ISO, fuera de los días `[starts_at, ends_at]` del evento, o **en el futuro**
  (no se registra el resultado de una prueba que todavía no ocurrió). Además el evento tiene que
  estar activo (no cancelado). Es la misma regla que controles (helper compartido).
- **Autor y fecha de carga:** `recorded_by` (FK `users`, el admin de la sesión) y `recorded_at`.
- **Constancia oficial = append-only.** Trigger `reject_mutation()` sobre la tabla: nadie (ni la
  FIA) modifica ni borra un resultado. No hay endpoints PATCH/DELETE.
- **Un evento con resultados es final:** su fecha y su baja quedan congeladas (igual que las
  carreras con puntajes, RC001), para que el resultado no apunte a una prueba movida o cancelada.
  Usa la guarda genérica de controles (RC020 → 409). Si necesito códigos propios, uso el rango
  RC030 a RC039 (acordado con sanciones).

## Modelo (migraciones 000032, 000033)

```sql
-- 000032_tyre_compounds
CREATE TABLE tyre_compounds (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  code text NOT NULL UNIQUE CHECK (code ~ '^[a-z0-9_]{2,32}$'),
  name text NOT NULL,
  kind text NOT NULL CHECK (kind IN ('slick', 'intermediate', 'wet')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);  -- + trigger set_updated_at

-- 000033_tyre_test_results
CREATE TABLE tyre_test_results (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  event_id bigint NOT NULL REFERENCES tyre_tests (event_id) ON DELETE RESTRICT,
  compound_id bigint NOT NULL REFERENCES tyre_compounds (id) ON DELETE RESTRICT,
  conducted_on date NOT NULL,
  passed boolean NOT NULL,
  notes text NOT NULL DEFAULT '' CHECK (char_length(notes) <= 2000),
  recorded_by bigint NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
  recorded_at timestamptz NOT NULL DEFAULT now(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (event_id, compound_id)
);
-- trigger append-only (reject_mutation) + índice sobre conducted_on
-- trigger de finalidad sobre events (función genérica acordada con controles)
```

## Endpoints

| Método y ruta | Rol | Respuestas |
|---|---|---|
| `GET /tyre-tests/types` | público | 200 `{types:[{code,name}]}` desde el catálogo de compuestos |
| `GET /tyre-tests/{eventId}/results` | `fia_admin` | 200 cadena completa con `recordedBy` y `supersedesId`; 401; 403; 404 |
| `POST /tyre-tests/results/{id}/corrections` | `fia_admin` | 201; 400 (incluye `correctionNote` vacío); 401; 403; 404; 409 si ya fue corregido |
| `GET /tyre-tests/results?from=YYYY-MM-DD&to=YYYY-MM-DD&type=<code>` | público | 200 `{results:[…]}`; 400 fecha mal formada, `from > to` o tipo desconocido |
| `POST /tyre-tests/{eventId}/results` | `fia_admin` | 201; 401 sin sesión; 403 otro rol; 404 evento inexistente o que no es `tyre_test`; 409 compuesto ya registrado; 400 validación |

Cuerpo del POST: `{conductedOn, compoundCode, passed, notes}`. Cada ítem del historial:
`{id, eventId, conductedOn, type:{code,name}, passed, notes, circuitName, seasonYear, categoryCode,
recordedBy:{username}, recordedAt}`. Orden: `conducted_on DESC, id DESC`.
No existen PUT/PATCH/DELETE: cualquier intento responde 405.

Capas: `transport/tyre_tests_handlers.go` → `internal/tyretests` (servicio) → `sql/tyre_tests.sql`.
El parseo de `from/to/type` lo provee controles (compartido).

## Validaciones (servicio)

`compoundCode` existente; `passed` presente (puntero, nulo → 400); `notes` según
`domain.ValidateResultNotes` (recortado, hasta 2000, obligatorio si `passed=false`); `conductedOn` válida según la regla de fecha; evento activo y de tipo
`tyre_test`; duplicado → 409 (lo detecta la `UNIQUE` y se traduce en una sola capa).

## App

- **Público y escudería** (solo lectura): historial de pruebas de neumáticos con filtros de rango de
  fechas y compuesto, sobre la lista de historial compartida de controles. Navegación propuesta:
  una pestaña pública compartida "Técnica" (Neumáticos | Controles); la escudería la ve desde
  "Sitio público".
- **FIA:** formulario de registro (bloque de resultado compartido + selector de compuesto) al que se
  llega desde el detalle de un evento `tyre_test` en el calendario del panel.
- `src/api/tyreTests.ts` y hooks en `src/data/tyreTests.ts`; nada de lógica en las pantallas.

## Compartido con otros equipos

- **controles (dueño), acordado:** `internal/domain/history.go` (`EventWindow`,
  `ValidateConductedOn`, `ParseDay`, `DateRange` con `Until` exclusivo, `HistoryFilter`),
  `internal/transport/history_filter.go` (`historyFilterOf`), `pgconv.OptionalDate`; función
  genérica `assert_event_without_results_is_final()` (RC020 → 409) en su migración 000031, a la que
  yo le agrego mi trigger `WHEN (OLD.kind = 'tyre_test')` con `('tyre_test_results', 'id')`. En la
  app, `src/features/history/*` (`HistoryFilters`, `HistoryList`, `OutcomeBadge`, `ResultFields`,
  `resultForm.ts`), `src/api/history.ts` (`HistoryEntry`), la pestaña pública
  `(public)/technical.tsx` con el registro `technicalSections.ts` (agrego una línea `tyres`) y
  `features/calendar/recordResultRoutes.ts` (agrego `tyre_test → /fia/calendar/[id]/tyre-results`).
  Validación → 400 `{code:"validation", field}` como el resto del código. Consumo sus commits con
  `git merge`.
- **sanciones:** solo el reparto de códigos SQLSTATE (tyres usa RC030 a RC039). Sin notificaciones.

## Plan de tests

- `tests/db`: append-only (UPDATE/DELETE rechazados), `UNIQUE`, FK a un evento que no es prueba,
  un evento con resultados no se puede mover ni cancelar.
- `tests/tyretests`: reglas de validación a través de la API pública del servicio.
- `tests/transport` (integración): la FIA registra y el resultado aparece en el historial público
  con los mismos datos; sin sesión 401; `team_admin` 403; evento inexistente o carrera 404;
  duplicado 409; fecha futura, fuera del rango o mal formada, notas vacías, compuesto desconocido,
  `passed` ausente → 400; JSON malformado → 400; evento cancelado → 400; filtros por fecha (bordes
  inclusivos), por tipo y combinados; `from > to` → 400; PATCH/DELETE → 405; dos POST concurrentes
  del mismo compuesto → uno 201 y el otro 409.

## Decisiones del manager (aprobado)

1. Tipo = compuesto. 2. Correcciones append-only con `supersedes_id` y motivo obligatorio; el
público ve el vigente marcado "Corregido". 3. Pestaña pública "Técnica". 4. Sin usuarios de la FIA
en las respuestas públicas. 5. Días en UTC con `formatDay`. 6. Atajo "Nueva prueba de neumáticos".
7. Tipo desconocido: 400 en neumáticos (catálogo), lista vacía en controles (texto libre).

## Preguntas abiertas (resueltas arriba)

1. ¿"Tipo" es el compuesto (propuesta) o el tipo de ensayo (desarrollo, durabilidad, lluvia)?
2. Al ser append-only, un error de carga no se puede corregir. ¿Alcanza, o se quiere una
   "rectificación" que agregue un resultado nuevo que reemplace al anterior?
3. ¿Pestaña pública compartida "Técnica" o pantallas sueltas?
