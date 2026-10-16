# 02 — Modelo de datos (Sprint 2)

Decisiones de modelo tomadas durante el Sprint 2. El documento equivalente del Sprint 1
(`docs/sprint1/02-modelo-de-datos.md`) queda congelado como registro histórico; lo que cambió
después vive acá.

## 1. Notificaciones = hechos puros (000010)

`notifications(id, user_id, event_id, read_at)`, con `event_id NOT NULL → events`.

- Una notificación es de un evento, **de cualquier tipo**: la base lo garantiza sin un `CHECK` de
  kinds.
- `title`, `body` y `kind` se eliminan: el texto y el tipo se **derivan al leer** del evento
  (nombre, fecha, tipo) y de la historia de publicación en `score_change_log`.
- La mensajería FIA↔escudería es otra cosa y tendrá su propia tabla; no comparte este `kind`.

## 2. Contexto del piloto (000011)

`drivers` guarda solo lo de la persona: `first_name`, `last_name`, `license_number`,
`country_code`, `deactivated_at` e `is_active` generada. Lo que depende del contexto se movió a
donde varía:

- **Rol titular/suplente — por período de pertenencia.** `driver_team_assignments` gana
  `is_starter`. Un cambio de rol o de equipo a mitad de temporada es una fila nueva
  (`effective_from`, append-only), nunca un edit.
- **Número de coche — por (piloto, temporada, categoría).** Tabla `season_drivers` (ex
  `driver_season_numbers`): `driver_id`, `season_id`, `category_id`, `number`. El número lo asigna
  el campeonato y sobrevive un cambio de equipo, por eso no tiene historial.
  - La categoría la **elige el admin** y un trigger la valida contra el historial de asignaciones
    (§7, `000019`). En `000011` se derivaba del equipo al 1 de enero, lo que dejaba sin número a
    quien se sumaba a mitad de temporada.
  - `UNIQUE (season_id, category_id, number)`: no se usa `UNIQUE (season_id, number)`, porque F1 y
    F2 numeran sus coches en forma independiente y el mismo número puede repetirse entre
    categorías de la misma temporada.

## 3. Jerarquía de eventos (000012–000015)

`events` es el padre **genérico y uniforme**: solo lo que comparten todos los tipos (tipo, timing,
baja lógica). Lo demás, incluido el campeonato, vive en una tabla hija cuya clave primaria es el
`event_id` (§5):

```mermaid
erDiagram
    events ||--o| races : "es"
    events ||--o| tyre_tests : "es"
    events ||--o| technical_controls : "es"
    rounds ||--o| races : "se corre en"
    circuits ||--o{ rounds : "aloja"
```

- `races` apunta a su `round` (`rounds`: temporada, categoría, número, circuito y nombre del GP) y
  toma el campeonato de ahí.
- `technical_controls` aporta `control_type` (texto que carga el admin) y su temporada y categoría.
- `tyre_tests` aporta su temporada, su categoría y el circuito donde se corre: la consigna pide que
  escuderías y público puedan consultarla.
- **El timing vive solo en `events`** (`starts_at`/`ends_at`): `races` y `tyre_tests` dejaron de
  duplicarlo (000015) y `technical_controls` nunca lo tuvo. Una sola fuente de horario para todo
  evento.
- `tyre_tests.event_id` es `NOT NULL` como el de las otras hijas: una prueba **es** un evento.

## 4. Integridad al cruzar tablas

La unicidad de un dato derivado que abarca dos tablas no se resuelve con un `UNIQUE` de una sola.
Dos patrones, según el caso:

- **Mover el dato a donde vive la unicidad.** La ronda es única por campeonato, así que `rounds`
  guarda temporada, categoría y número juntos y el `UNIQUE` es de una sola tabla (§5). Antes
  (000014) `races` copiaba temporada y categoría del evento con una FK compuesta; el rediseño de
  000016 eliminó esa copia.
- **Dato derivado por historial → trigger que lo materializa.** El equipo del resultado en `000008`.
  El número de piloto salió de este patrón en `000019` (§7): la categoría la elige el admin y el
  trigger solo la valida contra el historial.

Reglas transversales mantenidas: `bigint GENERATED ALWAYS AS IDENTITY`, `CHECK` en la base, **baja
lógica** derivada (`deactivated_at` + `is_active` generada) en `users`, `drivers`, `categories`,
`teams`, `events`, `rounds` y `circuits`, y ningún `DELETE` en cascada sobre datos históricos.

## 5. Integridad del calendario (000016)

Decisiones de la revisión del modelo, aplicadas editando `000016` (el esquema del Sprint 2 no está en
producción; requiere `make db-reset`).

- **Tipo con discriminador.** `events.kind` (`race`, `tyre_test`, `technical_control`) y cada hija
  fija el suyo con una columna constante (`event_kind`, `DEFAULT` + `CHECK`) y una FK compuesta a
  `events (id, kind)`. Es el patrón estándar de subtipo exclusivo: "un evento de un solo tipo" queda
  **declarativo** y es seguro bajo concurrencia, porque depende de un único valor en una única fila.
  Se evaluó eliminar `kind` y contar hijas con un trigger: funciona, pero con dos transacciones
  concurrentes se probó que un evento puede quedar con dos hijas si el trigger no bloquea la fila del
  padre. Se descartó: lo que se puede expresar con una FK no va a un trigger.
- **Una hija por evento, verificada al insertar.** "Como mucho una" la dan la PK y la FK de tipo.
  "Al menos una" no es expresable con FK ni `CHECK`, así que un `CONSTRAINT TRIGGER` diferido lo
  verifica al commit sobre el alta del evento y el cambio de `kind` (se inserta el evento y su hija en
  la misma transacción). Solo un bug del backend lo puede disparar, por eso no tiene código `RC`
  propio: un 500 es la respuesta correcta. Un tipo que se agregue al `CHECK` sin agregarse al trigger
  hace fallar al evento en lugar de dejarlo pasar sin hija. El trigger **no** vigila el borrado de la
  hija: esa mitad se retiró en `000020` (§8), así que un borrado por SQL crudo deja el evento sin hija.
- **PK compartida.** La hija no tiene id propio: `event_id` es su clave primaria. Nada referenciaba el
  id de la hija, y así una carrera tiene un solo id.
- **El campeonato es de la hija.** `events` guarda solo tipo y timing; `tyre_tests` y
  `technical_controls` tienen su `season_id`/`category_id` propios, y `races` los toma de su ronda.
  Ninguna columna es copia de otra, así que no hay nada que sincronizar.
- **Tabla `rounds`.** Temporada, categoría, número, circuito y nombre del GP. El número es único entre
  las rondas **activas** del campeonato (índice parcial sobre la propia tabla): si la FIA cancela la
  ronda 5 y arma otra ronda 5, la vieja queda dada de baja y libera el número, sin copiar el estado a
  ninguna otra tabla. Se evaluó antes copiar el estado activo del evento en `races` con una FK
  compuesta; `rounds` lo resuelve sin copia.
- **Una carrera por ronda.** La consigna no tiene sprints, así que `races_round_id_unique` deja una
  sola carrera por ronda. Si una entrega futura pidiera carreras sprint, alcanza con quitar esa
  restricción y distinguir la sesión.
- **Puntajes solo sobre carreras (B.1).** `score_entries` y `score_change_log` apuntan a
  `races (event_id)` en lugar de `events`: una prueba de neumáticos o un control técnico no pueden
  recibir puntos. Los acuses cuelgan del log, así que también quedan cubiertos. `notifications` sigue
  apuntando a `events`: cualquier tipo de evento se puede notificar. B.2 (quién puede acusar) se
  resolvió en `000023` (§9).
- **Ronda cancelada — la cierra el servicio (§8).** Nada en la base impide una carrera activa **sin
  resultados** sobre una ronda dada de baja; el servicio cancela la ronda y da de baja su carrera en
  la misma transacción, y no se agrega un trigger para esto (ver §8). Con resultados, la base lo
  impide (punto siguiente).
- **Sin nombre almacenado.** Se eliminó `events.name`. La etiqueta la compone el backend por tipo: la
  carrera con el nombre del GP de su ronda y el número, el control con su `control_type`, la prueba de
  neumáticos con un texto fijo y la fecha. Un solo lugar, testeable, y el texto en español no entra a
  las migraciones.
- **Catálogo `circuits`.** `name` (único), `locality`, `country_code` y baja lógica. Es dato de
  referencia (`seeds/reference.sql`, los 24 circuitos de 2026): en producción hace falta para crear
  rondas. El nombre del GP no está acá sino en la ronda, porque un mismo circuito puede albergar GP
  con distinto nombre según la categoría o el año.
- **Carrera disputada y equipo derivado.** Una carrera con resultados ya se corrió: lo que la ubica
  queda fijo. No cambia su fecha ni se da de baja, no cambia de ronda, y su ronda no cambia de
  temporada, categoría ni circuito ni se cancela (`RC001`, el servicio responde `409`); si no, los
  resultados pasarían a otra fecha, otra tabla de posiciones u otro circuito, o colgarían de algo que
  nadie ve. El número de ronda y el nombre del GP siguen editables: las rondas se renumeran tras una
  cancelación y un error de tipeo se tiene que poder corregir. El equipo del resultado se re-deriva en
  cada escritura.
- **Regla en el servicio, no en la base.** "La fecha del evento cae dentro de su temporada" abarca
  `events`, `races`, `rounds`, `tyre_tests` y `technical_controls`; hacerla cumplir en la base
  exigiría cinco triggers diferidos. La valida el servicio en cada escritura.

## 6. Puntajes y nombres (000018)

Migración aditiva sobre `000017` (no requiere `make db-reset`).

- **Categoría del puntaje.** El trigger que deriva el equipo del resultado también exige que ese
  equipo corra en la categoría de la ronda de la carrera: un piloto de F2 no puntúa en una carrera de
  F1. Usa `RC002`, porque es un error de carga del admin y el servicio lo responde con un 4xx.
- **Un ítem por piloto en cada cambio.** `UNIQUE (change_log_id, driver_id)` en
  `score_change_log_items`: dos líneas del mismo piloto dejarían ambiguos los puntos viejos y nuevos.
- **Duración positiva.** `ends_at > starts_at`: un evento que termina cuando empieza no significa
  nada en un calendario.
- **Nombres de `season_drivers`.** `000013` había renombrado la tabla pero no sus restricciones,
  índice, secuencia, triggers ni función.

## 7. Categoría del número de piloto (000019)

Migración aditiva sobre `000018` (no requiere `make db-reset`).

- **La categoría la elige el admin; el trigger la valida.** `000011` la derivaba del equipo del
  piloto al 1 de enero de la temporada. Eso castigaba a quien se suma a mitad de temporada —sin
  asignación en esa fecha, no podía tener número— y dejaba con la categoría vieja al piloto que
  cambia de categoría. Ahora la carga el admin y el trigger exige que el piloto tenga una asignación
  a un equipo de esa categoría cuya vigencia solape la temporada (año calendario, la misma convención
  que `000011`). Si no, `RC003`: es un error de carga y el servicio lo responde con un 4xx.
- **Una fila por (piloto, temporada, categoría).** `UNIQUE (driver_id, season_id, category_id)`
  reemplaza a `UNIQUE (driver_id, season_id)`, así un piloto promovido de F2 a F1 mantiene un número
  en cada campeonato durante la misma temporada.

## 8. Poda de triggers (000020)

La revisión del modelo fijó llevar los invariantes a la base en lugar de validarlos en la app. Al
revisar qué triggers valía la pena mantener, la regla se precisó: un trigger gana su lugar cuando el
estado no debe existir **por ningún escritor** (API, seeds, SQL) y su forma es barata; una operación
del servicio o una validación de entrada no lo justifican.

- **Se elimina el guardia del lado hijo.** `assert_event_keeps_child` (tres `CONSTRAINT TRIGGER`
  diferidos más una función con SQL dinámico) impedía que un evento quedara sin hija al borrar o
  reapuntar una hija. Reapuntar `event_id` no es una operación —la hija *es* su evento— y el proyecto
  nunca borra filas referenciadas: las da de baja con `deactivated_at`. El guardia cubría solo la
  mitad "se borra la hija" (el caso "se borra el padre también" estaba ignorado a propósito), así que
  vigilaba algo que el código no hace. No queda nada en su lugar: un borrado a mano por SQL es un
  problema de SQL crudo, y `reject_mutation` es la red más simple si alguna vez se quiere. La mitad
  "al menos una hija" (`assert_event_has_child`) se conserva porque no es expresable con FK ni `CHECK`,
  pero solo dispara desde el evento (alta y cambio de `kind`, §5): no cubre el borrado de la hija, que
  es justamente el caso que deja de vigilarse. La brecha queda documentada y cubierta por tests, no
  silenciada.
- **Ronda cancelada — en el servicio, no en la base (§5).** El caso sin resultados lo cierra el
  servicio en una transacción. Cubrir las tres vías (ronda, evento, vínculo) exigiría `DEFERRABLE` y
  tres triggers más una función: el mismo precio que se acaba de rechazar. Con resultados lo sigue
  impidiendo la base (`RC001`).
- **`score_entries.team_id` no cambia (F4).** El `drift` que preocupaba lo cierra la finalidad de la
  carrera: una vez que hay resultados, `starts_at` es inmutable (`RC001`), así que el equipo
  materializado no puede quedar desfasado por un cambio de fecha posterior.
- **Se conservan** `set_updated_at` (×14), `reject_mutation` (×5), `assert_event_has_child`, el trío
  de finalidad y `assert_season_driver_category` (§7): son baratos, de escritura rara, y ninguno
  protege una operación inexistente.
- **Retroactividad de la asignación (US11, `000022`).** Un trigger nuevo y barato: una asignación
  solo puede extender el historial, porque el dato que protege (el equipo derivado de una carrera ya
  puntuada) ya existe. Ver §2.
- **Auditoría de puntajes (US9, `000026`) — el guardia, no el armado.** La versión
  (`score_change_log` + sus ítems) la escribe el **servicio** en la misma transacción que toca
  `score_entries`; lo que la base garantiza es que ninguna escritura pasa sin declarar un autor
  (`SET LOCAL racecontrol.actor_id`, o falla con `RC005`). El trigger **no** arma la versión: agrupar
  las filas de una edición en un solo `score_change_log` exigía llevar un `change_log_id` entre filas
  (un GUC local) o tablas de transición, y eso es más maquinaria de la que la auditoría vale para un
  escritor único (el servicio). La brecha queda documentada: un escritor por SQL crudo que declare un
  actor podría saltarse el log, el mismo caso de "SQL crudo" que §8 ya acepta en otras partes.

Los puntos pendientes de la revisión del modelo se resolvieron en su user story (detalle en
`docs/sprint2/sprint.md`): el guardia de auditoría de puntajes (US9, `000026`), el cierre y la
retroactividad de las asignaciones (US11, `000022`) y el tipo de notificación y la validación del
acuse (US24, `000023`, resuelto en §9 sin tocar `notifications`). Se descartaron: guardar posición y
estado (DNF, DSQ), porque la consigna solo pide puntajes y `000006` deriva la posición de los puntos;
índices sin una query que los use; quitar las columnas `is_active`, que son convención del proyecto;
y la zona horaria del circuito, que ninguna historia pide.

## 9. Puntajes: clasificación, avisos y acuse (000023–000025)

Migraciones aditivas sobre `000022` (no requieren `make db-reset`). Cubren US23 y US24.

- **Posición y clasificación se calculan al leer.** La vista `race_scores` (`000024`) es cada
  resultado de una carrera activa con su campeonato, el piloto, el equipo acreditado y la posición:
  `RANK()` por puntos dentro de la carrera, así que dos pilotos con los mismos puntos comparten
  posición (el apellido solo ordena la visualización). Los resultados por carrera, el campeonato de
  pilotos y el de escuderías leen esa única definición. El de escuderías suma
  `score_entries.team_id`, el equipo con el que el piloto corrió esa carrera (§4): si cambia de equipo,
  sus puntos anteriores quedan en el equipo anterior. Los puntos se suman como `numeric` y se
  convierten a número una sola vez, así el total coincide con lo que cargó el administrador.
- **El aviso de puntajes se deriva, no se guarda.** La vista `score_notices` (`000025`) tiene una fila
  por cambio de `score_change_log` y por escudería que figura en sus ítems: eso es lo que esa escudería
  tiene que confirmar. El tipo es *publicación* si es el primer cambio de la carrera (orden
  `(changed_at, id)`, así hay exactamente uno) y *corrección* si no. El estado es *confirmado* si
  existe el acuse de (cambio, escudería) y *pendiente* si no. Se descartó lo que proponía el sprint
  —agregar a `notifications` el tipo y el vínculo al cambio con unicidad (usuario, cambio)—: obligaba
  a escribir una fila por usuario en cada cambio y mantenerla sincronizada, cuando el acuse es por
  escudería y no por usuario. Derivado, US24 queda desacoplada de la carga de puntajes (US9):
  cualquier cambio que se registre en el log aparece como aviso por sí solo. `notifications` queda
  para avisos de otros eventos.
- **Quién puede acusar (B.2, `000023`).** Un trigger `BEFORE INSERT` en `score_acknowledgements`
  toma `team_id` del usuario que acusa e ignora el que mande quien escribe, así nadie confirma en
  nombre de otra escudería; un administrador FIA no tiene escudería y nunca puede acusar. Además exige
  que esa escudería figure en los ítems del cambio: una escudería solo confirma lo que la tocó. Un
  cambio inexistente, uno de otra escudería y un usuario sin escudería fallan en el mismo chequeo con
  `RC005`.
- **`RC005` se responde con 404, no con 409 ni 400.** Para quien acusa, un cambio que no lo afecta no
  existe, igual que un piloto de otra escudería en US11: distinguir "no existe" de "es de otra
  escudería" revelaría los cambios ajenos.
- **Acusar dos veces es idempotente.** `ON CONFLICT (change_log_id, team_id) DO NOTHING` conserva la
  primera confirmación con su fecha, hora y usuario, y la respuesta es la misma. Un 409 no le sirve a
  quien tocó dos veces el botón.
- **Una corrección pide una confirmación nueva.** El acuse es por versión (`000009`): la confirmación
  de la publicación queda en la publicación y la corrección aparece pendiente. La FIA ve, por cada
  cambio, qué escuderías afectó y cuál confirmó, cuándo y quién.
