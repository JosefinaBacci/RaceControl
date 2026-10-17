# Registro de prompts de IA

## Regla de trabajo

Antes de dar por cerrada cualquier tarea que haya usado IA:

1. Agregar la entrada acá con el prompt utilizado.
2. Revisar la salida: leer el código generado, ejecutar los tests, verificar que no haya
   introducido dependencias no deseadas ni secretos.

Las entradas se registran en el momento de la tarea, nunca a posteriori. Los prompts se redactan
de forma clara y autocontenida, conservando la intención, el alcance y las restricciones del
pedido original. Cuando el pedido se apoyó en el contexto previo de la conversación, ese contexto
se explicita en el campo **Contexto**.

Todas las entradas siguen la misma estructura: **Contexto**, **Prompt**, **Respuesta de la IA**
(cuando aporta algo que no queda reflejado en el resultado) y **Resultado**.

---

## Entradas

### 2026-10-02 — Definición del stack tecnológico (US1)

**Contexto:** teníamos que decidir backend, base de datos, framework de
app móvil y mecanismo de autenticación para el Enunciado 1 (plataforma FIA). Restricción del
enunciado: la misma funcionalidad debe existir como sistema web y como app móvil.

**Prompt:**

> Comparar alternativas para el stack de una plataforma web + móvil de gestión deportiva
> (calendario, escuderías, pilotos, controles técnicos, sanciones, puntajes). Requisitos: roles
> FIA / escudería / público, la misma funcionalidad en web y móvil, y foco en seguridad de
> autenticación porque hay que poder desactivar usuarios. El equipo viene de TypeScript. Evaluar
> frameworks de backend en Go, base de datos, framework de app móvil multiplataforma, y
> mecanismos de autenticación (JWT vs. sesiones). Recomendar una opción por área justificando el
> porqué de los descartes.

**Resultado:** `docs/sprint1/01-stack.md`

---

### 2026-10-02 — Estructura de repositorio y convenciones (AGENTS.md)

**Contexto:** repositorio vacío salvo los PDF del enunciado. Necesitábamos la estructura de
carpetas, el layout de capas del backend y las convenciones de código antes de implementar US5.

**Prompt:**

> Proponer la estructura de carpetas para un monorepo con backend Go (chi + pgx + sqlc +
> golang-migrate) y una app Expo con expo-router. Necesito un layout en capas
> handler → service → repo, un módulo de configuración desde entorno, y las convenciones de
> nomenclatura de tablas y columnas. Incluir qué carpetas corresponden a qué user story del
> Sprint 1.

**Resultado:** `AGENTS.md`, `backend/`, `app/`, `docs/sprint1/TODO.md`

---

### 2026-10-02 — Modelo de datos y pipeline sqlc (US2)

**Contexto:** había que modelar el dominio del Enunciado 1 (categorías, temporadas, escuderías,
usuarios, sesiones, pilotos, calendario, puntajes, controles técnicos, sanciones, acuses de
notificación) respetando tres restricciones del equipo: base normalizada, construcción incremental,
y poder cambiar decisiones sin consecuencias (los conflictos entre ramas se resuelven regenerando
datos, no fusionándolos).

**Prompt:**

> Modelar en PostgreSQL el dominio de una plataforma de gestión de la Fórmula 1 y sus categorías:
> categorías, temporadas, escuderías, usuarios con roles, sesiones, pilotos titulares y suplentes,
> calendario de carreras y pruebas de neumáticos, puntajes por carrera, acuse de notificación de
> puntajes por escudería, controles técnicos con resultado por escudería, sanciones que aplican a
> un piloto o a una escudería, y acuse de notificación de sanciones. Requisitos: tercera forma
> normal, migraciones incrementales que nunca se editan, y facilidad para cambiar decisiones de
> diseño después. Una sanción puede apuntar a piloto o escudería. Los pilotos cambian de
> escudería a mitad de temporada. El seed debe ser idéntico en todos los entornos.

**Resultado:** `backend/migrations/000001_core.*`, `backend/sqlc.yaml`, `backend/sql/users.sql`,
`docs/sprint1/02-modelo-de-datos.md`

---

### 2026-10-02 — Separación de esquema y datos de referencia

**Contexto:** al cerrar US2, los datos de referencia habían quedado como una migración más
(`000002_seed_reference`). Al revisar la salida surgió la duda de si los datos debían vivir dentro
de la cadena de migraciones.

**Prompt:**

> Justificar por qué los datos de referencia (seed) se incluyeron como una migración dentro de
> `backend/migrations/`, y evaluar si esa ubicación es correcta para el proyecto.

**Respuesta de la IA:** se señaló que el seed se había metido como migración `000002` por comodidad
(`make db-reset` dejaba la base poblada en un solo comando), y que esa comodidad era un error de
diseño. Se resumieron tres consecuencias concretas: la cadena de migraciones se ejecuta también
en producción, con lo que las cuentas de demo de US5 con contraseña conocida se desplegarían; un
archivo de seed ya aplicado no se puede editar, de modo que cada corrección del dato de demo
agregaba una migración y la cadena terminaba siendo mayoritariamente datos; y `migrate-down 1`
borraba filas de forma silenciosa porque el seed era la última migración.

Se propuso la separación: `migrations/` queda solo con DDL y `seeds/` con datos idempotentes
(`ON CONFLICT DO NOTHING`) fuera de `schema_migrations`, con `make db-reset = drop + migrate + seed`.

**Decisión del equipo:** se eligió la variante **migraciones solo DDL**, es decir, también las
cuatro categorías pasan a las seeds; y las cuentas de demo se crean con un seed exclusivo de
desarrollo, con la contraseña tomada de una variable de entorno y nunca presente en producción.

**Resultado:** `backend/seeds/*`, `backend/internal/seed/seed.go`, `backend/cmd/seed/main.go`,
`backend/internal/config/config.go`, `Makefile`, `.env.example`, `AGENTS.md`,
`docs/sprint1/02-modelo-de-datos.md`, `docs/sprint1/TODO.md`

---

### 2026-10-02 — Arquitectura, roles y seguridad (US3)

**Contexto:** US1 y US2 cerradas y el seed separado del esquema de migraciones. Lo que pide el
Sprint 1 para US3 (verificado leyendo `docs/Sprint1_GrupoRojoFerrari.pdf`): definir la
arquitectura general (capas, servicios, módulos, comunicación) con 3 h de estimación; definir el
modelo de permisos por rol para administrador de la FIA, escuderías y público general, con 2 h;
establecer los estándares de seguridad (autenticación, cifrado, políticas), con 3 h; y
documentarlo, con 1 h.

**Prompt:**

> Con US1 y US2 cerradas, desarrollar la US3 a partir del PDF del Sprint 1 y de
> `docs/sprint1/TODO.md`: definir la arquitectura general (capas, servicios, módulos y
> comunicación), el modelo de permisos por rol (FIA, escudería y público) y los estándares de
> seguridad (autenticación, cifrado y políticas), y documentarlo en
> `docs/sprint1/03-arquitectura-roles-seguridad.md`.

**Resultado:** `docs/sprint1/03-arquitectura-roles-seguridad.md`, actualización de
`docs/sprint1/TODO.md` (US3 cerrada con tiempo real y desviación).

---

### 2026-10-02 — Revisión general del repositorio

**Contexto:** al terminar US5 y antes de arrancar US6, se hizo una revisión integral del
repositorio para validar lo implementado contra lo documentado y decidir cómo seguir. El pedido
fue explícitamente de solo lectura: no modificar archivos.

**Prompt:**

> Revisar el repositorio completo (`AGENTS.md`, la documentación de diseño de `docs/sprint1/`,
> el backend y la app) para obtener una visión general del estado del proyecto. Contrastar lo
> documentado con lo implementado y señalar bugs, inconsistencias y riesgos, sin
> modificar archivos. Después, explicar en detalle cada bug encontrado: causa, impacto y forma de
> corregirlo.

**Respuesta de la IA:** un estado por user story y una lista priorizada de hallazgos. Se
identificaron tres bugs:

1. **`CHECK` de exclusividad de sanciones invertido** (documentado en `02-modelo-de-datos.md` y en
   la entrada de US2 de este registro, todavía no migrado). La tabla de verdad muestra que
   `(driver_id IS NULL) <> (team_id IS NOT NULL)` rechaza los dos casos válidos (solo piloto, solo
   escudería) y acepta los dos inválidos (ambos, ninguno). La forma correcta es
   `(driver_id IS NULL) <> (team_id IS NULL)` o `num_nonnulls(driver_id, team_id) = 1`.
2. **El migrador descartaba el error original.** En `migrate.New`, las dos ramas de error
   envolvían el resultado de `db.Close()` en lugar de `err`. Como `sql.Open` no conecta, el
   primer intento real de conexión ocurre ahí, y los errores más comunes de configuración (base
   apagada, credenciales incorrectas) terminaban como `create migration driver: %!w(<nil>)`.
3. **La cookie de sesión no funciona con el deploy en Vercel.** El diseño de §3.4 de US3 asume
   que la app web se sirve desde el mismo sitio que la API (cookie `SameSite=Strict`, sin token
   en el cliente web). El deploy estático en Vercel pone la app en otro sitio: el navegador no
   guarda ni envía la cookie, y el login web se pierde en el primer request posterior.

Además se señalaron inconsistencias entre `01-stack.md` y `03-arquitectura-roles-seguridad.md`
(rotación del token, migraciones al arranque, `ENUM` frente a `CHECK`), las casillas de US4 y
US5 sin marcar en `TODO.md` pese a estar implementadas, la falta de `04-interfaces.md`, tests de
integración que se omiten sin `DATABASE_URL` (con lo que `make check` en verde dice poco),
`go.mod` sin ordenar y código sin uso en `internal/db/queries.go`.

**Resultado:** sin cambios en el repositorio. Los hallazgos definen los próximos pasos: US6,
deploy, `04-interfaces.md` y consistencia de la documentación.

---

### 2026-10-02 — Corrección del manejo de errores del migrador

**Contexto:** de los tres bugs de la revisión anterior, el equipo decidió corregir primero el del
migrador, el único que afecta hoy al flujo de trabajo de todos (`make migrate`, `make db-reset`).
En el mismo pedido se aprovechó para normalizar este registro.

**Prompt:**

> Corregir el manejo de errores de `migrate.New` para que el error original de la conexión no se
> pierda, respetando la regla de `AGENTS.md` de envolver los errores con contexto. Aprovechar para
> normalizar el formato de `docs/prompts-ia.md`, que alterna entre "Prompt:" y
> "Prompt (literal):", e informar cualquier otra inconsistencia del registro.

**Resultado:** `backend/internal/db/migrate/migrate.go`, `docs/prompts-ia.md`

---

### 2026-10-02 — Corrección del CHECK de sanciones y revisión de la migración inicial

**Contexto:** segundo bug de la revisión general: el `CHECK` de exclusividad de `sanctions`,
documentado en US2 pero todavía no migrado. Además, el equipo señaló que la migración existente,
`000001_core.up.sql`, no le parecía prolija y pidió revisarla con el mismo criterio.

**Prompt:**

> Corregir el `CHECK` de exclusividad de sanciones en la documentación del modelo de datos, y
> revisar `backend/migrations/000001_core.up.sql` en busca de restricciones incompletas,
> redundancias o decisiones que contradigan lo documentado en US2 y en `AGENTS.md`.

**Resultado:** `backend/migrations/000001_core.up.sql`, `backend/migrations/000001_core.down.sql`,
`backend/internal/transport/auth_integration_test.go`, `docs/sprint1/02-modelo-de-datos.md`,
`AGENTS.md`.

---

### 2026-10-02 — Consistencia entre documentación y código, y pasada de calidad

**Contexto:** con los dos bugs puntuales corregidos, quedaban las contradicciones señaladas en la
revisión general: documentos que se contradecían entre sí o describían comportamiento que el
código no tiene, un `TODO.md` que no reflejaba US4 y US5, y detalles de calidad de código frente a
las reglas de `AGENTS.md`. Se resolvieron antes de empezar US6 para arrancarla sobre una base
coherente.

**Prompt:**

> Resolver las inconsistencias entre la documentación y el código detectadas en la revisión:
> contradicciones entre documentos, afirmaciones que el código no cumple y erratas. Actualizar
> `docs/sprint1/TODO.md` al estado real del proyecto. Hacer además una revisión de calidad del
> backend y de la app contra las reglas de clean code y DRY de `AGENTS.md`, y corregir lo que
> corresponda.

**Resultado:** `docs/sprint1/01-stack.md`, `docs/sprint1/02-modelo-de-datos.md`,
`docs/sprint1/03-arquitectura-roles-seguridad.md`, `docs/sprint1/TODO.md`, `AGENTS.md`,
`README.md`, `compose.yaml`, `.env.example`, `backend/` (config, transport, cmd, db, testdb,
`go.mod`) y `app/` (`src/features`, `src/data`, `src/mocks/catalog.ts` y las pantallas que los
usan).

---

### 2026-10-03 — Gestión de usuarios: backend (US6)

**Contexto:** con la base revisada y verificada contra Postgres, se arrancó US6 por el backend.
Antes de implementar se extrajeron a funciones compartidas la validación de `username`, la
detección de violaciones de unicidad y las conversiones de tipos de `pgtype`, que ya estaban
repetidas en `auth`, `sessions` y `create-admin` (commit propio, separado de la funcionalidad).

**Prompt:**

> Implementar el backend de US6: alta, modificación, baja lógica, listado y búsqueda de usuarios,
> con asignación de un rol válido, solo para el administrador FIA. Respetar el diseño de US3: la
> baja y los cambios de privilegio revocan las sesiones del usuario en la misma transacción.
> Reutilizar las consultas y helpers existentes antes de escribir nuevos, mantener los
> comentarios consistentes con el estilo de cada paquete, y cubrir cada regla con tests de
> integración contra la base real.

**Resultado:** `backend/internal/users/`, `backend/internal/teams/`,
`backend/internal/db/transactor.go`, `backend/internal/domain/errors.go`, `backend/sql/users.sql`,
`backend/sql/teams.sql`, `backend/internal/transport/` (handlers, rutas, errores y tests),
`backend/cmd/api/main.go`, `docs/sprint1/03-arquitectura-roles-seguridad.md`,
`docs/sprint1/TODO.md`.

---

### 2026-10-03 — Gestión de usuarios: pantallas de la app (US6)

**Contexto:** con los endpoints de US6 en la API, la pantalla de usuarios del panel FIA seguía
siendo la maqueta de US4, con datos de ejemplo y los botones deshabilitados. Al probar la app, el
equipo detectó que no se podía modificar un usuario y se pidió conectarla.

**Prompt:**

> Conectar la gestión de usuarios de la app a la API de US6: listado con búsqueda y filtros por
> rol y estado, alta, edición, baja y reactivación, mostrando cada error de la API en el campo
> que corresponde. Mantener la lógica fuera de los componentes de pantalla, reutilizar el design
> system y los helpers existentes, y no duplicar código.

**Resultado:** `app/app/fia/users/` (`_layout.tsx`, `index.tsx`, `new.tsx`, `[id].tsx`),
`app/src/api/users.ts`, `app/src/api/teams.ts`, `app/src/api/formFailure.ts`,
`app/src/auth/loginForm.ts`, `app/src/data/` (`accounts.ts`, `useRemote.ts`,
`useDebouncedValue.ts`, `useFormSubmission.ts`), `app/src/features/RemoteContent.tsx`,
`app/src/features/accounts/`, `app/app/fia/index.tsx`, `docs/sprint1/TODO.md`, `AGENTS.md`.

---

### 2026-10-03 — Cambio de contraseña propia y menú de usuario

**Contexto:** la matriz de permisos de US3 promete que cada usuario puede modificar su propio
perfil, pero ningún usuario podía cambiar su contraseña: solo el administrador FIA, editando su
propia cuenta desde la gestión de usuarios. La herramienta propuso un botón global de "Cambiar
contraseña"; el equipo lo rechazó y definió otra interfaz: un menú desplegable desde un ícono de
usuario arriba a la derecha, con accesos a la cuenta y al cierre de sesión, y el cambio de
contraseña dentro de una pantalla de cuenta.

**Prompt:**

> Permitir que cualquier usuario con sesión cambie su propia contraseña. En la interfaz, reemplazar
> el botón de cerrar sesión por un menú de usuario en la esquina superior derecha, con acceso a una
> pantalla "Mi cuenta" que contenga el cambio de contraseña, y la opción de cerrar sesión.

**Resultado:** `backend/sql/sessions.sql`, `backend/internal/users/service.go`,
`backend/internal/transport/` (handler, ruta y tests), `app/app/account.tsx`,
`app/src/navigation/UserMenu.tsx`, `app/src/navigation/HeaderSessionButton.tsx`,
`app/src/navigation/RoleGate.tsx`, `app/src/api/auth.ts`,
`app/src/features/accounts/passwordConfirmation.ts`, los layouts de `fia`, `team` y `(public)`,
`docs/sprint1/03-arquitectura-roles-seguridad.md`, `docs/sprint1/TODO.md`.

---

### 2026-10-03 — Logos de la FIA y de las escuderías (US4)

**Contexto:** los avatares mostraban iniciales sobre un círculo de color. El equipo propuso usar
los logos reales y consiguió los de la FIA y de las tres escuderías con cuentas demo (Ferrari, Red
Bull y Mercedes), en formatos y tamaños dispares: dos JPG con fondo de color, dos PNG con fondo
transparente, y en total cerca de 1,1 MB.

**Prompt:**

> Usar los logos de la FIA y de las escuderías en los avatares de la app. Adaptar las imágenes
> para que se vean bien en un avatar circular sin aumentar de más el tamaño de la app.

**Resultado:** `app/assets/logos/`, `app/src/logos.ts`, `app/src/components/Avatar.tsx`, el
listado de usuarios, la pantalla de resultados y el menú de usuario.

---

### 2026-10-03 — Base de datos propia para los tests de integración

**Contexto:** al probar la gestión de usuarios, el equipo encontró en el listado decenas de cuentas
con nombres como `t108965040125.1`. Eran usuarios creados por los tests de integración: `make
check` toma `DATABASE_URL` de `.env`, así que los tests corrían contra la misma base que usa la app
y cada corrida dejaba cuentas nuevas.

**Prompt:**

> Explicar de dónde salen las cuentas de prueba que aparecen en la gestión de usuarios y evitar que
> los tests escriban en la base de desarrollo.

**Resultado:** `backend/internal/testdb/testdb.go`, `Makefile`, `.env.example`, `README.md`,
`AGENTS.md`.

---

### 2026-10-03 — Logos como avatar de las cuentas demo y backlog del Sprint 2

**Contexto:** después de que los avatares de las personas volvieran a mostrar iniciales, el
equipo pidió que las cuentas demo de cada organización (`fia.admin`, `ferrari.admin`,
`redbull.admin`, `mercedes.admin`) usen su logo como foto de perfil, y que la idea de que cada
usuario suba su propia foto quede registrada para más adelante.

**Prompt:**

> Asignar el logo de cada organización como foto de perfil de su cuenta demo, sin que el resto de
> las cuentas pierda sus iniciales, y registrar la subida de fotos de perfil como pendiente para
> el Sprint 2.

**Resultado:** `app/src/logos.ts`, el listado de usuarios, el menú de usuario,
`docs/sprint1/TODO.md`.

---

### 2026-10-03 — Documentación de interfaces (US4) y horas reales del sprint

**Contexto:** la única tarea abierta de US4 era documentar las funcionalidades por rol; las
pantallas y el sistema de diseño ya estaban implementados. Además faltaba cargar las horas reales
de US4, US5 y US6, que la cátedra pide comparar con la estimación para el Sprint 2. El equipo
aportó el historial de commits del repositorio como fuente para las horas.

**Prompt:**

> Redactar `docs/sprint1/04-interfaces.md` a partir de lo implementado: qué ve y qué puede hacer
> cada rol, la navegación, el sistema de diseño y las decisiones de interacción, indicando qué
> pantallas ya usan datos reales. Calcular las horas reales de US4, US5 y US6 a partir del
> historial de commits y completar la tabla del sprint.

**Resultado:** `docs/sprint1/04-interfaces.md`, `docs/sprint1/TODO.md`.

---

### 2026-10-04 — Navegación de la gestión de usuarios

**Contexto:** al probar la app, el equipo reportó que, después de cargar o dar de baja un usuario,
no se podía volver a la pantalla de usuarios: "Cancelar" o "Crear usuario" llevaban a otro lado, y
la pestaña Usuarios quedaba mostrando el formulario.

**Prompt:**

> Reproducir el problema reportado al cargar y dar de baja un usuario e intentar salir de la
> pantalla, encontrar la causa y corregirla.

**Respuesta de la IA:** el problema se reprodujo manejando la app real en un navegador automatizado
(Playwright), con la API y la base locales. La baja funcionaba; lo que fallaba era la navegación.
Cuando el formulario de alta se abría desde el acceso rápido del panel, al recargar la página o
entrando por la URL, quedaba como única pantalla de la sección Usuarios, sin el listado debajo. Al
cancelar o crear, la app volvía al panel y la pestaña Usuarios quedaba mostrando el formulario,
sin forma de volver al listado. La corrección declara el listado como base fija de esa sección
(`unstable_settings = { anchor: 'index' }` en su layout) y el acceso rápido del panel navega
incluyendo esa base (`withAnchor`). Se descartaron dos cambios que el equipo también evaluó:
guardar el borrador del formulario al recargar, porque implicaría escribir la contraseña inicial
en el almacenamiento del navegador, y reiniciar la pestaña al salir de ella, porque perdería un
formulario a medio completar; conservar la última pantalla de cada pestaña es el comportamiento
estándar.

**Validación:** antes y después del cambio se corrieron con Playwright los dos recorridos: el
acceso rápido del panel (cancelar, crear y volver a la pestaña Usuarios) y el flujo completo desde
el listado (cancelar, cancelar tras recargar, crear, dar de baja desde la fila, reactivar y dar de
baja desde la ficha, volver). Antes fallaban tres pasos; después, todos llegan al listado. Durante
la prueba se detectó que el servidor de desarrollo de Expo seguía sirviendo el árbol de rutas
viejo: el cambio solo se ve después de reiniciarlo con `npx expo start --clear`. `tsc` en verde.

**Resultado:** `app/app/fia/users/_layout.tsx`, `app/app/fia/index.tsx`.

---

### 2026-10-12 — Modelo de puntajes: la posición se deriva y la auditoría registra puntos (Sprint 2)

**Contexto:** al cerrar la Fase 0 del Sprint 2 quedó migrado un esquema con `score_entries.position` y un `score_change_log` de texto libre. Antes de implementar las US de puntajes, el equipo revisó esas dos decisiones: si hacía falta guardar el orden de llegada por resultado y si un log que no registra los valores sirve como auditoría.

**Prompt:**

> ¿Hace falta registrar la posición en cada resultado? "Las posiciones no son necesarias, se reconstruyen de los puntos." Y un log que no registra lo que cambió no sirve: "ni siquiera sabés qué modificaciones ocurrieron". El orden no debería forzarse al insertar: "debería ser una query nada más".

**Respuesta de la IA:** se replanteó el modelo con tres decisiones encadenadas:

1. **La posición es derivable.** En una clasificación la posición determina los puntos, y no hay un caso real de "posición distinta con los mismos puntos" (las penalizaciones de grilla afectan la largada, no la clasificación; una descalificación mueve juntos posición y puntos). Por eso `score_entries.position` y su `UNIQUE (event_id, position)` —que además impedía empates— se eliminaron: el orden dentro de una carrera y el de las clasificaciones son `ORDER BY points DESC` en la query, y el desempate es una regla del sistema (countback o clave estable), nunca un dato que carga el administrador.
2. **La auditoría registra el dato, no una frase.** `score_change_log` queda como cabecera (evento, actor, fecha) y se agrega `score_change_log_items` con una fila por piloto afectado: `old_points` / `new_points`. El tipo de cambio (`created`/`updated`/`removed`) y la posición se derivan de esos valores, y un `CHECK (old_points IS DISTINCT FROM new_points)` descarta filas sin información.
3. **El estado se deriva siempre.** Nada de "pendiente/confirmado" como booleano guardado: se computa en la lectura.

**Resultado:** `backend/migrations/000006_score_change_items.*` (quita `position`, crea el log estructurado), `backend/seeds/demo.sql`, `backend/internal/db/dbgen/models.go`. Commit `f62041b`.

---

### 2026-10-12 — Escudería del piloto: historial inmutable y equipo del resultado derivado (Sprint 2)

**Contexto:** el diseño de Fase 0 pijaba la escudería de cada resultado con un FK compuesto `(driver_id, team_id) → drivers(id, team_id)`. El equipo señaló que "el piloto puede cambiar de escudería en un momento dado" y que el cambio tendría que ser registrable para trackear los puntos según a qué escudería correspondió en cada período.

**Prompt:**

> Guardar ambos, el team y el piloto, porque el piloto puede cambiar de escudería en un momento dado. El cambio de escudería tendría que ser algo registrable para que los puntos se trackeen según a qué escudería el piloto correspondió hasta la fecha del cambio. ¿El snapshot sería para precalcular la derivación? Las asignaciones viejas no pueden editarse; los suplentes son por escudería realmente.

**Respuesta de la IA:** se probó en la base que el FK compuesto **bloquea** el cambio de escudería (`UPDATE drivers SET team_id …` → violación de FK), y que `ON UPDATE CASCADE` reescribiría los resultados viejos. Se recorrió el debate en tres pasos:

1. **Historial de asignaciones.** Nueva tabla `driver_team_assignments(driver_id, team_id, effective_from)` con períodos **implícitos** y **append-only**: cerrar un período = insertar la fila siguiente; un trigger rechaza `UPDATE`/`DELETE`. Como el historial nunca se edita, la derivación para una fecha pasada nunca cambia.
2. **Dos semánticas del snapshot.** "¿El snapshot precalcula la derivación?" solo es cierto si el resultado es una proyección de la afiliación (opción A, todo derivado). Bajo B, el resultado es un documento publicado: guarda el equipo con el que se publicó. Se eligió **B con equivalencia garantizada**: el `team_id` del resultado **lo completa un trigger** con la asignación vigente en la fecha del evento, así que B ≡ A por construcción; y como los suplentes pertenecen a una escudería, la derivación siempre existe.
3. **Se descarta el FK compuesto.** El equipo del resultado se valida por la historia, no por la fila de `drivers`: se eliminan `score_entries_driver_team_fk`, `drivers_id_team_unique` y `drivers.team_id` (el equipo actual es la asignación abierta, derivable); `score_entries.driver_id` pasa a tener FK propia.

**Resultado:** `backend/migrations/000007_driver_team_assignments.*` (tabla + `assigned_team()` + trigger de inmutabilidad) y `000008_team_snapshot.*` (FKs simples, triggers `set_score_entry_team`/`set_score_change_item_team`, baja de `drivers.team_id`), `backend/seeds/demo.sql` (asignaciones iniciales), `backend/internal/db/dbgen/models.go`. Commit `1ce4953`.

---

### 2026-10-12 — Acuse de recibo por versión publicada (Sprint 2)

**Contexto:** `score_acknowledgements` tenía `UNIQUE (event_id, team_id)`: una escudería acusaba una vez por carrera. Pero las US9/US24 permiten **modificar** puntajes después de publicarlos ("Recibe una notificación cuando se publican o modifican los resultados"), lo que dejaba al acuse apuntando a datos viejos sin forma de volver a confirmar.

**Prompt:**

> Si la FIA corrige los puntajes después de que una escudería acusó, el acuse queda obsoleto y no se puede repetir. Las escuderías afectadas son las afectadas por el cambio; el resto no le incumbe de forma directa. Acuse por versión, que apunte al change log.

**Respuesta de la IA:** el acuse es un hecho (no se deriva), pero su ancla estaba mal: debía apuntar a la **versión publicada**, no al evento. La versión es la fila de `score_change_log`, que ya lleva evento y fecha. Cambios:

1. `score_acknowledgements.event_id` → `change_log_id` con `UNIQUE (change_log_id, team_id)`: cada corrección genera una versión nueva y las escuderías afectadas vuelven a **pendiente** automáticamente.
2. Las escuderías afectadas por un cambio salen de `score_change_log_items.team_id` (equipo derivado con el que se cargó el resultado), de modo que también se notifica a la escudería a la que se le *quitó* un piloto.
3. El estado pendiente/confirmada se deriva de la última versión del evento: confirmada = existe acuse de esa escudería contra ella.

**Resultado:** `backend/migrations/000009_score_acknowledgements_per_version.*`, `backend/internal/db/dbgen/models.go`. Commit `f897de1`.

---

### 2026-10-12 — Notificación = hecho puro: solo usuario, evento y leída (Sprint 2)

**Contexto:** la tabla `notifications` de Fase 0 guardaba `kind`, `title` y `body` junto con `user_id`. Al diseñar las notificaciones de puntajes, el equipo la revisó con tres ideas: la notificación es "de un evento" sin importar su tipo; la mensajería FIA↔escudería es otra cosa y no va en la misma tabla; y el texto hardcodeado en la base no es un hecho, se deriva al leer.

**Prompt:**

> Una notificación es de un evento, no me importa qué tipo. La mensajería no es ortogonal con este tipo de notificaciones, son dos tipos distintos, no tiene sentido mezclarlos en una misma tabla. Droppeá body y title, no necesitamos text hardcodeado en la base de datos. Tampoco necesitamos el kind: tenemos el kind en el evento. Las notificaciones se identifican por el tiempo, el tipo y el usuario.

**Respuesta de la IA:** se aterrizaron las tres ideas en una tabla mínima:

1. **Solo hechos.** `notifications(id, user_id, event_id, read_at)` con `event_id NOT NULL → events ON DELETE RESTRICT`: cada notificación es de un evento, de cualquier tipo (`race | tyre_test | technical_control`), y la base lo garantiza sin CHECK de kinds.
2. **Nada de texto ni kind.** `title`/`body`/`kind` se dropean; el texto mostrado se deriva al leer del evento (nombre, fecha, `event_type_id`) y de la historia en `score_change_log` (publicado vs. corregido). El tipo de evento ya vive en `events.event_type_id`.
3. **Identidad = tiempo + evento + usuario.** La fila (user, event, `created_at`) identifica la notificación: dos publicaciones del mismo evento a distinta hora son dos notificaciones distintas, no duplicados.

**Resultado:** `backend/migrations/000010_notifications_event.*` (up + down: agrega `event_id NOT NULL`, dropea `kind`/`title`/`body` y su CHECK, índice por `event_id`), verificación en la base (INSERT sin `event_id` rechazado; con `event_id`, ok y rollback). Commit `6a2d284`.

---

### 2026-10-12 — Contexto del piloto: número por temporada, rol por período (Sprint 2)

**Contexto:** `drivers` guardaba `number` y `is_starter` como atributos estáticos del piloto. El equipo señaló que el número de coche es de la temporada y que el rol titular/suplente no puede ser una propiedad fija. Se descartó una tabla `race_entries` porque duplicaba la clave `(event_id, driver_id)` de `score_entries`, guardaba `team_id` derivable vía trigger y modelaba un hecho que ninguna US consume.

**Prompt:**

> El is_starter debería estar vinculado a la carrera y no al piloto, y el número debería estar vinculado a la temporada. El modelo de drivers es una mierda, proponé algo más decente y normalizado. ¿El is_starter es por temporada? Pero qué pasa si cambia en el medio de la temporada.

**Respuesta de la IA:** se normalizó por grano de hecho, apoyándose en la historia append-only que ya existía:

1. **`drivers` queda con lo de la persona:** `first_name`, `last_name`, `license_number`, `country_code`. Se dropean `number` e `is_starter` (y su CHECK).
2. **El rol viaja con la pertenencia:** `driver_team_assignments` gana `is_starter`. Un cambio de rol o de equipo a mitad de temporada es una fila nueva con `effective_from` (mecanismo existente, append-only); el rol actual es la fila abierta y el de una fecha pasada, la fila vigente. El `DEFAULT true` es solo mecánica del `ALTER` sobre filas existentes; los seeds y la app pasan el valor explícito.
3. **El número es por (piloto, temporada):** nueva `driver_season_numbers(driver_id, season_id, number)` con `UNIQUE (driver_id, season_id)` y rango 0-999. El número lo asigna el campeonato y sobrevive un cambio de equipo, por eso no tiene historial. La colisión por (temporada, categoría) se valida en el servicio porque la categoría se deriva de la escudería.

**Resultado:** `backend/migrations/000011_driver_context.*` (up + down), `backend/seeds/demo.sql` (asignaciones con rol, números 2026), `backend/internal/db/dbgen/models.go` (Driver sin número/rol, `DriverSeasonNumber`, `DriverTeamAssignment.IsStarter`). Verificado en la base: INSERT de cambio de rol a mitad de año funciona y el UPDATE sigue rechazado por append-only. Commit `5645d7c`.

---

### 2026-10-12 — Jerarquía de eventos: races / tyre_tests / technical_controls (000012)

**Contexto:** El modelo anterior mantenía `event_types`/`event_type_id` como enum plano en `events`. El equipo señaló que cada tipo de evento (carrera, prueba de neumáticos, control técnico) tiene suficientes atributos como para merecer su propia tabla, y que las notificaciones deben cubrir todo a partir del mismo `events.id` (el GP/ronda).

**Prompt:**

> cómo quedaría el modelo si cada tipo de evento tiene su tabla propia en lugar de un enum en events

**Respuesta de la IA:** se normalizó por tipo de hecho: `events` como padre (el GP/ronda oficial), y tres hijas `races`, `tyre_tests`, `technical_controls` con `event_id → events.id`. Los puntajes y notificaciones siguen en `events.id`; las tablas hijas agregan el tipo-specific data. El `event_types` y `event_type_id` se eliminan. La migración `000012` y los seeds se regeneraron; `make check` verde. Verificado: 5 carreras, 1 prueba de neumáticos, 2 controles técnicos bajo los 8 GPs. Commit siguiente.

**Resultado:** `backend/migrations/000012_events_hierarchy.*`, `backend/seeds/demo.sql`/`reference.sql` actualizados, `backend/internal/db/dbgen/models.go` (EventTypeID eliminado, new tables), `docs/prompts-ia.md`. Verificado en la base: FKs events→races/tyre_tests/technical_controls, notificaciones siguen en events.id, setval incluidos. Verificado también que `make db-reset` reproduce todo idempotentemente.


---

### 2026-10-12 — Auditoría del esquema de Sprint 2 y corrección de integridad (000012–000015)

**Contexto:** revisión conjunta del esquema de los commits de Sprint 2. Se detectaron, contra el código y los seeds reales: constraints perdidas al mover columnas de `events` a `races` (000014), `updated_at` sin trigger en las tres hijas de `events`, `tyre_tests.event_id` nullable, timing asimétrico y duplicado en `technical_controls`, migraciones `down` rotas (000012/000014/000015), deriva entre las migraciones y el código generado por sqlc, y un `UNIQUE` de números de piloto que no cubría la categoría.

**Prompt:**

> Auditá las migraciones 000002–000015 una por una y dejá un informe con hallazgos concretos. ¿Y si ponés `UNIQUE (season_id, number)`? ¿Qué se hace en la industria?

**Respuesta de la IA:** la unicidad de un dato derivado que cruza tablas no se resuelve con un `UNIQUE` de una sola tabla. Para `races` se usó el patrón que el propio repositorio ya emplea (`drivers_id_team_unique` + FK compuesta en `score_entries`): se copia `season_id`/`category_id` desde el evento y se protegen con una FK compuesta a `events (id, season_id, category_id)`, habilitando `UNIQUE (season_id, category_id, round)`. Para `season_drivers` la categoría no es atributo de una fila padre sino que se deriva del historial de asignaciones, así que se materializó con un trigger (mismo criterio que `score_entries.team_id` en 000008) y recién ahí el `UNIQUE (season_id, category_id, number)` se sostiene. Un `UNIQUE (season_id, number)` se descartó por incorrecto: F1 y F2 numeran sus coches en forma independiente.

**Resultado:** se corrigieron `000001`/`000002` (baja lógica derivada en `categories`, `teams` y `events`), `000004` (se quita el `summary` de `score_change_log`, redundante con `score_change_log_items`), `000011` (categoría derivada + unicidad por campeonato en los números de piloto), `000012` (`tyre_tests.event_id NOT NULL`, triggers `updated_at` en las hijas, timing solo en `events` con la baja de `started_at/ended_at` de `technical_controls`), `000014`/`000015` (CHECK, FK compuesta y `down` corregidos) y `backend/seeds/demo.sql`. Verificado con `make db-reset`, `make check` en verde, la cadena completa de `down` (15→0) sobre datos sembrados, y pruebas de comportamiento: la FK compuesta rechaza temporada/categoría incoherentes, la ronda se repite en otro campeonato pero no en el mismo, el número de piloto se repite entre categorías pero no dentro de una, y los triggers de `updated_at` y la baja lógica responden. `make sqlc` regeneró `dbgen`.

---

### 2026-10-13 — Calendario: una hija por evento, nombre en el evento y ronda obligatoria (000016)

**Contexto:** antes de implementar US7 (gestión del calendario) se pidió verificar que el modelo no
represente estados inconsistentes. La jerarquía `events` + `races`/`tyre_tests`/`technical_controls`
permitía tres estados inválidos de bajo costo de arreglo: varias hijas del mismo tipo por evento, el
nombre duplicado entre el evento y la hija (podían divergir) y una carrera sin ronda (que además
esquivaba el `UNIQUE` por campeonato cargando `NULL`).

**Prompt:**

> fijate que el modelo de datos sea coherente y no represente estados inconsistentes; evitá la representación de estados inconsistentes para minimizar la cantidad de chequeos

**Respuesta de la IA:** se separaron los hallazgos en *Tier 1* (constraints puntuales, sin tocar la
estructura) y *Tier 2* (el caso "evento sin hija", que no sale con un `UNIQUE`). En esta pasada se
aplicó solo el Tier 1, por ser seguro: `UNIQUE (event_id)` en cada hija (a lo sumo una hija por
evento), se elimina `name` de las tres hijas (el nombre vive en `events`) y `races.round` pasa a
`NOT NULL` con `CHECK (round > 0)`. El índice plano por `event_id` se reemplaza por el único, que ya
cubre las búsquedas.

**Resultado:** `backend/migrations/000016_events_child_integrity.{up,down}.sql`,
`backend/seeds/demo.sql` (las hijas ya no insertan `name`) y `backend/internal/db/dbgen/models.go`
(se van `Name` de `Race`/`TyreTest`/`TechnicalControl`; `Race.Round` pasa a `int32`). Verificado con
`make db-reset`, `make sqlc` y `make check` en verde; el `down` de la 000016 restaura nombre y ronda
opcional y el `up` vuelve a aplicarse; y pruebas de comportamiento en la base: una segunda hija del
mismo evento, una ronda `NULL` y una ronda `0` son rechazadas. Quedan pendientes el Tier 2
("exactamente una hija") y las decisiones de puntajes/acuse y triggers.

---

### 2026-10-13 — Calendario: tipo del evento y exactamente una hija (Tier 2 de 000016)

**Contexto:** continuación de la revisión del modelo de datos. El Tier 1 había dejado un único estado
inconsistente de la jerarquía `events` + hijas: un evento **sin ninguna hija**, que no tiene tipo ni
nombre calculable (y su simétrico, un evento con hijas de **dos tipos distintos**). Se discutió cómo
llevar ese invariante a la base y si la tabla `tyre_tests`, que quedó sin atributos propios, debía
seguir existiendo.

**Prompt:**

> Sigamos con el modelo de datos del calendario. Un `events` sin ninguna fila hija no es de ningún
> tipo: nada lo impide y la app tendría que chequearlo. Y hoy puede colgar una carrera y un control
> técnico del mismo evento. Quiero que la base garantice que cada evento tiene exactamente una hija, la
> de su tipo. Decime cuál es la forma más idiomática de lograrlo en Postgres (trigger diferido, FK
> bidireccional o aceptarlo en el servicio) y confirmá si `tyre_tests` se conserva como tabla marcador.

**Respuesta de la IA:** se evaluaron tres caminos. (a) **Trigger diferido**: `CONSTRAINT TRIGGER ...
DEFERRABLE INITIALLY DEFERRED`; (b) **FK bidireccional** (el patrón *exclusive arc*, con
`CHECK (num_nonnulls(...) = 1)` y una FK compuesta hija→evento); (c) no bloquearlo y validarlo en el
servicio. Se eligió **(a)** porque "al menos un hijo" no es expresable con una FK ni con un `CHECK`
—solo valida filas contra una tabla— y el trigger diferido es la respuesta canónica en Postgres,
además de ser el idioma que el repositorio ya usa (`reject_mutation`, derivaciones con `plpgsql`).
La opción (b) se descartó porque obliga a una FK **cíclica** y estrena `DEFERRABLE` (no hay ninguna en
el proyecto) y complica la futura restricción de puntajes (§B.1), que también se apoya en el tipo del
evento.

Para que la regla sea uniforme ("todo evento tiene exactamente una hija") se conservó `tyre_tests`
como **marcador** (una fila por prueba, hoy solo `event_id`). El tipo se agregó como `events.kind`
(`race | tyre_test | technical_control`), y cada hija lo repite en una columna constante
(`event_kind`) que una FK compuesta contra `events (id, kind)` obliga a coincidir: de un evento de
carreras solo puede colgar una carrera. Con eso queda cerrado el estado "hijos de dos tipos" sin
necesidad de `event_type_id` ni de un `CHECK` de exclusividad.

La mitad "sin hija" se chequea en dos triggers que comparten una función: uno sobre `events`
(`AFTER INSERT OR UPDATE OF kind`) que exige la hija del `kind`, y otro sobre cada hija
(`AFTER DELETE OR UPDATE OF event_id`) que evita dejar huérfano al evento al borrar solo la hija,
ignorando el caso "el padre también se está borrando". Ambos son diferidos para que insertar el
evento y su hija en una misma transacción siga siendo posible (el alta de US7).

**Respuesta de la IA (validación):** el primer intento usó `IF NOT CASE NEW.kind WHEN ... THEN ...
END THEN`, que **no compila**: el parser de `plpgsql` corta la condición en el primer `THEN` de las
ramas del `CASE` y termina con `syntax error at end of input`. La expresión se movió a una asignación
(`has_child := CASE ... END;`) y quedó resuelta.

**Resultado:** `backend/migrations/000016_events_child_integrity.{up,down}.sql` (se fusionó el Tier 2
en la misma migración, que todavía no está en producción), `backend/seeds/demo.sql` (los eventos
declaran su `kind`; las hijas lo heredan por `DEFAULT`), `backend/internal/db/dbgen/models.go`
(`Event.Kind`, `Race/tyreTest/TechnicalControl.EventKind`). Verificado con `make sqlc`, `make db-reset`
y `make check` en verde; el `down`/`up` aplican en ida y vuelta; y 11 pruebas de comportamiento en la
base: evento sin hija rechazado al commit, evento+hija en una transacción aceptado, hija de un tipo
distinto rechazada por FK, segunda hija del mismo tipo rechazada por `UNIQUE`, borrar solo la hija
rechazado y borrar hija+evento aceptado, cambiar el `kind` con una hija viva rechazado, y las
regresiones de Tier 1 (`round` NULL y `round 0`).

---

### 2026-10-13 — Alineación de los comentarios SQL del Sprint 2 al inglés

**Contexto:** la convención del repositorio (AGENTS.md) pide comentarios en inglés; Go, TypeScript y
las migraciones anteriores al Sprint 2 ya lo hacían, pero el DDL del Sprint 2 (000012–000016) y los
comentarios por sección de `seeds/demo.sql` habían quedado en español.

**Prompt:**

> (tras una advertencia de la IA sobre el idioma de los comentarios) sí hacé esto y alineá los estilos
> por favor

**Resultado:** comentarios traducidos al inglés en `backend/migrations/000012`–`000016` (up y down) y
`backend/seeds/demo.sql`. Solo comentarios: ninguna sentencia cambió, así que una base ya en la versión
16 no necesita reseteo.

---

### 2026-10-13 — Equipo del resultado siempre derivado y fecha final de una carrera con resultados (000016)

**Contexto:** cierre del último hallazgo de la revisión del modelo (§C.4/§C.5). El equipo acreditado de
un resultado es una copia denormalizada del historial de pases, y solo se volvía a deducir cuando
cambiaba el piloto o la carrera: una escritura directa sobre el equipo podía dejarlo en desacuerdo con
el historial. Aparte, la fecha de una carrera vive en `events.starts_at` sin ninguna traba una vez
cargados los resultados, y de esa fecha depende la deducción del equipo.

**Prompt:**

> dame el código concreto de lo que vas a hacer para cerrar F4

**Respuesta de la IA:** se evaluó resolverlo en el servicio frente a llevarlo a la base, y se eligió la
base (el repositorio ya traduce errores de la base en errores de dominio). Dos cambios en la 000016:
(1) el trigger que deduce el equipo pasa de `BEFORE INSERT OR UPDATE OF event_id, driver_id` a
`BEFORE INSERT OR UPDATE`, para que se re-deduzca en cada alta y cada edición y no se pueda escribir a
mano; (2) un guardián `BEFORE UPDATE OF starts_at` sobre `events` rechaza mover la fecha si el evento
ya tiene resultados, y deja pasar el caso en que la fecha se reenvía igual junto con otros cambios (la
comparación es `IS DISTINCT FROM`). El rechazo se levanta con un SQLSTATE propio de la clase `RC`
(`RC001`), y `db.GuardViolation` lo reconoce para que el servicio lo traduzca a un `409` cuando se
implemente la edición del calendario (US7). Se descartaron recalcular el equipo al mostrar, reescribir
los resultados y guardar el contrato en vez del equipo.

**Resultado:** `backend/migrations/000016_events_child_integrity.{up,down}.sql` (bloques 7 y 8),
`backend/internal/db/pgconv.go` (`GuardViolation`) y `backend/internal/db/pgconv_test.go`. Verificado
con `make db-reset`, `make sqlc` (sin cambios), `make check` en verde y el `down`/`up` en ida y vuelta;
7 pruebas de comportamiento en la base: la fecha de una carrera con resultados se rechaza con `RC001`,
la misma fecha con otro cambio pasa, un evento sin resultados sí puede moverse, y un equipo escrito a
mano se re-deriva. El lado del servicio (traducir `RC001` a `409` y fijar la fecha en la pantalla)
queda para US7.

---

### 2026-10-13 — Append-only: se conserva el trigger y se difiere la segregación de permisos (F5)

**Contexto:** último punto de la revisión del modelo. F5 preguntaba si los triggers de append-only se
dejan o se reemplazan por un rol de aplicación con `REVOKE`/RLS. La IA había caracterizado el pasaje al
rol como "mitad de un trabajo de DevOps".

**Prompt:**

> por qué decís que es muy complicado el tema del rol?

**Respuesta de la IA:** se midió el cambio en concreto en lugar de sostener la estimación. Son unas
cinco sentencias SQL (crear el rol, `USAGE` del esquema, `GRANT SELECT, INSERT, UPDATE, DELETE ON ALL
TABLES`, el `REVOKE` de las tablas de historial y `ALTER DEFAULT PRIVILEGES`) más un campo de
configuración y una segunda URL de conexión: la caracterización de "complicado" era incorrecta y se
corrigió. Se verificó además que no hay trampas ocultas (las columnas `identity` no piden `USAGE` sobre
la secuencia, las funciones de los triggers corren como el invocador con `EXECUTE` público, los
triggers existentes funcionan bajo un rol sin privilegios y el `REVOKE` sí bloquea la mutación) y el
estado real del proyecto (la API conecta con un usuario **superusuario y `BYPASSRLS`** y dueño de las
19 tablas, no hay despliegue a producción, y el control documentado para "escalada desde la app" es el
RBAC del backend).

**Resultado (decisión del usuario):** se conserva el trigger de append-only, porque cubre a cualquier
escritor (incluidas las migraciones, los seeds y `psql`, que corren como dueño) mientras que el rol
solo cubriría la API, y la segregación de permisos queda anotada como TODO en
`docs/sprint2/TODO.md`. No cambió código. En el mismo archivo quedó anotada, como decisión pendiente,
la cobertura append-only de la auditoría (`score_change_log`, `score_change_log_items`,
`score_acknowledgements`).

---

### 2026-10-13 — Append-only en la auditoría (cierre de F5)

**Contexto:** continuación de la entrada anterior. La cobertura append-only de la auditoría había
quedado anotada como decisión pendiente, no implementada.

**Prompt:**

> agregaste los otros append-only?

**Respuesta de la IA:** no los había implementado; los había dejado anotados porque "dejemos el
trigger" se interpretó como "no agregar nada más". Antes de proponer el cambio se verificó que nada en
el código dependa de poder mutar esas tablas (no tienen queries ni handlers: son de US9 y US24,
todavía sin construir; los únicos `UPDATE` del proyecto son sobre `sessions` y `users`, que sí deben
ser mutables) y que nada borre `login_attempts` (solo se inserta y se lee para el bloqueo de login).

**Resultado (decisión del usuario):** se agregan las guardias en una migración nueva,
`000017_append_only_audit`, en lugar de fusionarlas en `000016`. El motivo por el que §D había elegido
fusionar era el cambio de diseño del calendario en vuelo, que acá no aplica porque el cambio es
puramente aditivo, y la regla por defecto de AGENTS es "solo archivos nuevos en `backend/migrations/"`
(además se aplica sin `db-reset`). El alcance es `score_change_log`, `score_change_log_items`,
`score_acknowledgements` y `login_attempts`, reusando `reject_mutation()`. Se descartó darle un código
propio (`RC002`) para traducirlo a 4xx: la API no debería ofrecer mutar la auditoría nunca, así que un
rechazo ahí es un bug y un 500 es la respuesta correcta.

---

### 2026-10-14 — Calendario: PK compartida, ronda liberada por la baja y catálogo de circuitos (000016)

**Contexto:** antes de implementar US7, revisión de la jerarquía `events` → `races` / `tyre_tests` /
`technical_controls` que había quedado en `000016`, y aplicación de §A.6 (sin nombre almacenado,
catálogo `circuits`), decidido en la revisión del modelo pero no implementado.

**Prompt:**

> Check the event table and children approach, is it idiomatic? Would you improve it in any way?

Seguido de: "¿no perdemos alguna garantía al sacar `kind`?", "¿el trigger es idiomático?" y "decidí
por lo que se usa en producción".

**Respuesta de la IA:** la forma (padre genérico + una tabla por tipo con discriminador) es la
estándar. Encontró tres problemas: las hijas tenían un id propio que nada referenciaba; una carrera
dada de baja **retenía su ronda** (la unicidad vivía en `races` y la baja en `events`), lo que impedía
volver a cargar la misma ronda en US7; y las FK declaraban `ON UPDATE CASCADE` sobre valores que no
pueden cambiar. Propuso mover `round` a `events`; el equipo lo rechazó (la ronda es de la carrera). Se
evaluó también eliminar `kind` de las hijas y de `events`: se demostró en una base descartable que, sin
`kind`, un trigger que cuenta hijas deja pasar dos hijas con transacciones concurrentes salvo que
bloquee la fila del padre, y que lo idiomático es dejar a la FK lo declarativo y al trigger solo "al
menos una". Se conservó `kind`.

**Resultado:** `000016` editada (requiere `make db-reset`): PK de cada hija = `event_id`; `races`
copia `event_is_active` y una única FK compuesta a `events (id, kind, season_id, category_id,
is_active)` mantiene las copias, con la ronda única solo entre carreras activas (índice parcial); FK
de tipo con `ON UPDATE RESTRICT` en las otras hijas; `events.name` eliminado; catálogo `circuits`
(24 circuitos de 2026 en `seeds/reference.sql`) y `races.circuit_id`. Verificado con `make db-reset`,
12 pruebas de comportamiento (baja que libera la ronda, reactivación en conflicto, copia falsificada,
segunda hija, evento sin hija, fecha final, cambio de tipo, cambio de categoría en cascada), ida y
vuelta `down`/`up` con datos y `make check`. Documentado en `docs/sprint2/02-modelo-de-datos.md` §5.

---

### 2026-10-14 — Calendario: campeonato en las hijas y tabla `rounds` (000016)

**Contexto:** revisión del equipo sobre la versión anterior de `000016`, que copiaba temporada,
categoría y estado activo del evento en `races` (FK compuesta de cinco columnas) para que una carrera
dada de baja liberara su ronda.

**Prompt:**

> Drop season_id and category_id from events, and drop events_race_identity_unique (the 5-column
> constraint). Keep only UNIQUE (id, kind) as the FK target for the children. Each child owns its own
> season_id and category_id as native columns, not synced copies. races loses event_is_active and
> keeps event_kind only as the discriminator. Put round on a rounds table (season, category, number,
> circuit, gp_name), with races pointing at it. This lets one weekend hold a sprint and a race, and the
> uniqueness lives on a plain table, so no active-flag copy is needed.

**Respuesta de la IA:** antes de implementar señaló cuatro puntos abiertos: que `races` con campeonato
propio y además apuntando a una ronda permitía que ambos difirieran, qué distingue un sprint de la
carrera en la misma ronda, qué pasa con la ronda vieja si se cancela y se vuelve a armar, y si
`circuits` conserva `gp_name`. Decisiones del equipo: `races` toma el campeonato solo de la ronda;
la consigna no pide sprints, así que la ronda tiene una sola carrera (`UNIQUE (round_id)`); las rondas
tienen baja lógica y el número es único entre las activas (índice parcial); `gp_name` sale de
`circuits`.

**Resultado:** `000016` reescrita (requiere `make db-reset`). `events` queda con tipo, timing y baja
lógica; `tyre_tests` y `technical_controls` con temporada y categoría propias; `races` con `round_id`;
tablas `rounds` y `circuits`. Verificado con `make db-reset`, 15 pruebas de comportamiento (cancelar y
rearmar una ronda, número repetido entre activas, mismo número en otra categoría, reactivación en
conflicto, segunda carrera por ronda, segunda hija, evento sin hija, cambio de tipo, fecha y baja de
una carrera con resultados), ida y vuelta `down`/`up` con datos y `make check`. Queda documentado como
pendiente que la base no impide una carrera activa sobre una ronda dada de baja.

---

### 2026-10-14 — Calendario: evaluación de una revisión externa del modelo (000016)

**Contexto:** el equipo trajo una lista de siete propuestas sobre el esquema del calendario y pidió
evaluarlas antes de aplicar nada.

**Prompt:**

> Evaluate the following: [lista: `control_type` como lookup o enum; `circuit_id` en `tyre_tests`;
> puntajes apuntando a `races`; extender el bloqueo de resultados a la ronda; `ELSE` en
> `assert_event_has_child()`; vista del calendario; trigger de "fecha dentro de la temporada"; quitar
> las categorías si es solo F1]. Don't just apply it, think about it and give me some feedback, which
> ones are worth doing and which ones aren't and why

Seguido de: "Ok, implement the ones worth doing then".

**Respuesta de la IA:** señaló que parte de la lista describía el esquema anterior (las hijas ya tenían
campeonato y `races` ya no tenía circuito ni categoría). Recomendó hacer cuatro: puntajes sobre
`races` (B.1), el bloqueo extendido pero sobre `races.round_id` y `rounds` (donde viven ahora esos
datos, dejando editables el número y el nombre del GP), el `ELSE` explícito (hoy inalcanzable por el
`CHECK`, pero protege al agregar un tipo) y `tyre_tests.circuit_id`. Postergó la vista del calendario
a US7, con `LEFT JOIN` en lugar de `UNION` y sin textos. Descartó el enum/lookup de `control_type`
(la consigna manda al admin a ingresar los tipos), el trigger de fecha dentro de la temporada (cinco
triggers diferidos para una validación de una línea en el servicio) y quitar las categorías (la
consigna pide F1, F2, F3 y F1 Academy).

**Resultado:** `000016` editada con los cuatro cambios (requiere `make db-reset`; una base con pruebas
de neumáticos no migra porque no hay circuito que inferir, y la migración falla entera sin dejar
cambios a medias). Verificado con 19 pruebas de comportamiento, los códigos `RC001` de las guardias
nuevas, ida y vuelta `down`/`up` y `make check`.

---

### 2026-10-14 — Evaluación de una segunda revisión del modelo (000018)

**Contexto:** el equipo trajo una segunda lista de observaciones sobre puntajes, pilotos,
notificaciones, sesiones y nombres, y pidió evaluarlas antes de aplicar nada.

**Prompt:**

> Evaluate the following (same as before): [lista: categoría del equipo vs. de la carrera y fila en
> `season_drivers`; auditoría obligatoria de puntajes; validación del acuse; `season_drivers` al 1 de
> enero; asignaciones sin fin y retroactivas; `UNIQUE` en ítems y notificaciones, notificaciones sin
> tipo; posición y estado; purga de `login_attempts`, índices por IP y por vencimiento; índice de no
> leídas; nombres `driver_season_numbers*`; quitar `is_active`; `ends_at = starts_at`; zona horaria
> del circuito; dumps sin hashes]

Seguido de: "Ok, implement the worth now and leave a note [...] for the ones that should be
implemented in the future".

**Respuesta de la IA:** verificó cada punto contra el esquema y las queries. Separó lo que conviene
hacer ya (barato y cierra un agujero real), lo que es real pero se diseña con su user story, y lo que
no conviene: guardar posición y estado revierte una decisión del equipo (`000006`) sin que la consigna
lo pida; los índices por IP, por vencimiento de sesión y de no leídas no tienen query que los use;
`is_active` es convención del proyecto y se usa; `UNIQUE (user_id, event_id)` en notificaciones
bloquearía avisar una corrección de puntajes.

**Resultado:** migración aditiva `000018` (categoría del equipo = categoría de la carrera con `RC002`,
un ítem por piloto en cada cambio, `ends_at > starts_at`, nombres de `season_drivers`). Verificado
con pruebas de comportamiento, ida y vuelta `down`/`up` y `make check`. Lo diferido quedó como tareas
en `docs/sprint2/sprint.md` (US9, US11, US12, US24) y la purga de `login_attempts` en
`docs/sprint2/TODO.md`. El dump compartido excluye los datos de `users`, `sessions` y `login_attempts`.

---

### 2026-10-15 — Categoría del número de piloto a elección del admin (000019)

**Contexto:** al revisar el esquema apareció que `season_drivers` deriva la categoría del equipo al
1 de enero, un problema ya anotado en `docs/sprint2/sprint.md` (US11): un piloto que se suma a mitad
de temporada no tiene asignación en esa fecha y no puede tener número, y uno que cambia de categoría
queda con la vieja.

**Prompt:**

> we should solve the mid-season rookie problem

**Respuesta de la IA:** propuso seguir el diseño ya escrito en `docs/sprint2/sprint.md` (US11): una
fila por (piloto, temporada, categoría), la categoría elegida por el admin y un chequeo de que el
piloto tenga una asignación en esa categoría durante la temporada. Señaló que la ventana de la
temporada es el año calendario (misma convención que `000011`) y que una asignación vence cuando
empieza la siguiente (períodos implícitos, `000007`), así que el chequeo es el solapamiento de las dos
ventanas. Usó `RC003` para el error de carga, que el servicio responderá con un 4xx.

**Resultado:** migración aditiva `000019` (`UNIQUE (driver_id, season_id, category_id)`, trigger
`assert_season_driver_category` con `RC003`), semilla `demo.sql` con `category_id` explícito y cuatro
pruebas de integración en `backend/internal/db/season_drivers_test.go`. Se confirmó que las pruebas
fallan contra el esquema anterior ("no team assignment for season 1" y la violación del `UNIQUE`
viejo). Nota técnica: `sqlc` no tolera un `DROP FUNCTION` sobre una función renombrada en una
migración previa, así que el trigger se reemplaza con `ALTER FUNCTION ... RENAME` + `CREATE OR
REPLACE`. Verificado con ida y vuelta `down`/`up`, `make sqlc` y `make check`.

---

### 2026-10-15 — Mover los tests a `backend/tests/`

**Contexto:** los tests estaban colocados junto al código (`internal/auth/password_test.go`,
`internal/db/queries_test.go`, `internal/transport/*_integration_test.go`, etc.), lo que mezcla
código de producción y de prueba en el mismo árbol.

**Prompt:**

> should we improve the test structure because now they are everywhere

**Respuesta de la IA:** explicó que Go solo compila los `_test.go` que están en el directorio del
paquete, así que mover los tests afuera solo puede hacerse con paquetes de test externos que
importan al paquete bajo prueba, y por lo tanto solo alcanzan su API exportada. Verificó que los
ocho archivos ya usaban únicamente símbolos exportados, así que el movimiento no exigía cambios de
comportamiento. Propuso un árbol `backend/tests/` espejando `internal/`, con `testdb` como helper
compartido, y comparó con una opción híbrida (solo los tests de integración afuera).

**Resultado:** se movieron los ocho archivos de test y `internal/testdb` a `backend/tests/`, se
renombraron los paquetes a `authtest`, `dbtest`, `sessionstest` y `transporttest`, y se actualizó
`AGENTS.md` con la regla de layout y la consecuencia de la API exportada. `go build ./...`,
`go vet ./...` y `go test ./...` en verde y `internal/` sin ningún `_test.go`. Sin cambios de
comportamiento.

---

### 2026-10-15 — Poda de triggers: se elimina el guardia del lado hijo (000020)

**Contexto:** al decidir cómo cerrar el caso de la ronda cancelada, un `CONSTRAINT TRIGGER` diferido
resultó desproporcionado (tres triggers más una función para un estado borde). Se pidió revisar qué
otros triggers del esquema costaban más de lo que valían, partiendo de la sección de triggers de la
revisión del modelo (§C) y de su regla rectora: llevar los invariantes a la base y no validarlos en
la app.

**Prompt:**

> i think is too expensive this change in triggers, can you identify other places where triggers are
> too much?

**Respuesta de la IA:** precisó la regla rectora —un trigger gana su lugar cuando el estado no debe
existir por ningún escritor (API, seeds, SQL) y su forma es barata; una operación del servicio o una
validación de entrada no lo justifican— e inventarió los 29 triggers vivos. El único que protegía una
operación que el código no hace es el par del lado hijo `assert_event_keeps_child` (tres triggers
diferidos sobre `races`, `tyre_tests` y `technical_controls`, más una función con SQL dinámico):
reapuntar `event_id` no es una operación porque la hija *es* su evento, y el proyecto da de baja con
`deactivated_at` en lugar de borrar. Propuso eliminarlo. Recomendó conservar `set_updated_at`,
`reject_mutation`, `assert_event_has_child`, el trío de finalidad y `assert_season_driver_category`
(§7), y confirmó que `score_entries.team_id` (F4) ya no tiene `drift` porque la finalidad de la
carrera congela `starts_at` al aparecer resultados.

**Resultado:** migración aditiva `000020_drop_child_guard` (elimina la función y los tres triggers;
el `down` los recrea), prueba de integración en `backend/tests/db/child_guard_test.go` que confirma
que borrar una hija se acepta (contra el esquema anterior falla con `SQLSTATE P0001`), y
`docs/sprint2/02-modelo-de-datos.md` §8 con las decisiones. La ronda cancelada queda en el servicio
(US7, anotado en `docs/sprint2/sprint.md`) y la retroactividad de las asignaciones en US11 con un
único `BEFORE INSERT` (`RC004`). Verificado con ida y vuelta `down`/`up` y `make check`.

---

### 2026-10-15 — Tests del invariante de una hija por evento y de los códigos `RC001`/`RC002`

**Contexto:** cerrada la poda de `000020`, la documentación afirmaba "exactamente una hija" cuando el
guardia del lado hijo ya no existía: quedaba media invariante sin test, y la mitad que sí se conserva
(`assert_event_has_child`) solo dispara desde el evento.

**Prompt:**

> add the tests

**Respuesta de la IA:** propuso separar por archivo los invariantes del evento y los códigos de
guardia, reutilizando los helpers existentes (`insertTestDriver`, `assignDriver`, `insertTyreTestEvent`
y las constantes de datos seed). Para `RC001` y `RC002` aprovechó que el evento 1 del seed ya es una
carrera de F1 con resultados: cada aserción corre en una transacción que siempre se revierte, así el
dato seed nunca se modifica. Cubrió: evento sin hija, segunda hija del mismo tipo (violación de PK),
hija de otro tipo (FK compuesta), cambio de `kind` con hija presente, el trío de finalidad de una
carrera disputada (fecha, ronda, campeonato) y `RC002`. Aclaró que la brecha del borrado de la hija
queda como test explícito en `child_guard_test.go` en lugar de quedar implícita.

**Resultado:** `backend/tests/db/event_integrity_test.go`, `race_finality_test.go` y
`score_guard_test.go` nuevos; `child_guard_test.go` ampliado (el evento queda sin hija y un cambio de
`kind` posterior falla al commit); y `02-modelo-de-datos.md` §5/§8 precisados: el invariante es "una
hija por evento verificada al insertar", no "exactamente una hija". `make check` en verde.

---

### 2026-10-15 — Gestión de calendario: backend (US7)

**Contexto:** el esquema ya tenía los tres tipos de evento (`races`, `tyre_tests`,
`technical_controls`) colgando de `events`, pero ningún endpoint los exponía. Se arrancó el Sprint 2
por US7 porque el calendario es la base de los puntajes (US9). Restricción del equipo: los controles
técnicos no se modelan más de lo que ya están, pero US7 igual los da de alta.

**Prompt:**

> Implementar el backend de US7: lectura pública del calendario y gestión (alta, modificación y baja)
> restringida al administrador FIA, para carreras, pruebas de neumáticos y controles técnicos. Una
> carrera tiene exactamente una ronda, así que el servicio escribe evento, carrera y ronda en una
> sola transacción y la baja de la carrera da de baja también la ronda; no hay endpoint de rondas.
> Validar campos obligatorios, `ends_at` posterior a `starts_at`, ronda única por campeonato y fecha
> congelada si la carrera ya tiene resultados (409). La etiqueta del evento se compone en el
> servicio, no se guarda. No modelar más de lo necesario los controles técnicos. Cubrir cada regla
> con tests de integración contra la base real, incluido el acceso denegado a los demás roles.

**Respuesta de la IA:** propuso una vista `calendar_events` (`000021`) como único modelo de lectura,
compartido por listado y detalle, con `LEFT JOIN` hacia la hija para que sqlc tipara las columnas
propias de cada tipo como nulables. Derivó el estado (`scheduled`/`finished`) de `ends_at` contra
`now` en el backend en vez de guardarlo, y compuso la etiqueta por tipo: carrera con `gp_name` y
número de ronda, prueba de neumáticos con la fecha en UTC, control técnico con su `control_type`.
Dejó la ronda cancelada en el servicio (sin `CONSTRAINT TRIGGER`) según la decisión de §C.

**Resultado:** `backend/migrations/000021_calendar_read_model.{up,down}.sql`, `backend/sql/calendar.sql`
y `internal/db/dbgen/calendar.sql.go`, `internal/calendar/{event,validation,service}.go`,
`internal/transport/calendar_handlers.go` y rutas, `internal/db/pgconv.go`,
`tests/transport/calendar_integration_test.go`, `tests/calendar/event_test.go`. `make check` en
verde.

---

### 2026-10-15 — Gestión de calendario: pantallas de la app (US7)

**Contexto:** con los endpoints de US7 en la API, la pantalla de calendario del panel FIA seguía
siendo la maqueta de US4 (datos de ejemplo, botones deshabilitados). Se pidió conectarla sin migrar
todavía la vista pública, que es el primer punto de US8.

**Prompt:**

> Conectar la gestión del calendario de la app a la API de US7: listado con filtros de período y
> categoría, alta, edición y baja con confirmación, mostrando cada error de la API en el campo que
> corresponde. La vista pública del calendario queda como está hasta US8. Mantener la lógica fuera de
> los componentes de pantalla, reutilizar el design system y los helpers existentes, y no duplicar
> código.

**Resultado:** `app/src/api/calendar.ts`, `app/src/data/calendar.ts`,
`app/src/data/useReloadOnFocus.ts` (extraído de `accounts.ts`), `app/src/features/calendar/`
(`eventForm.ts`, `EventFields.tsx`, `EventAdminList.tsx`) y `app/app/fia/calendar/`
(`_layout.tsx`, `index.tsx`, `new.tsx`, `[id].tsx`) en lugar de `app/app/fia/calendar.tsx`. La
respuesta del evento sumó `gpName` para que el formulario de edición pueda precargar el nombre de la
ronda. `make check` en verde.

---

### 2026-10-15 — Visualización del calendario sobre la API (US8)

**Contexto:** US7 dejó la gestión del calendario en la API, pero la vista pública del calendario, el
panel FIA y el panel de escudería seguían leyendo los datos de ejemplo de US4, así que los cambios
del administrador no se reflejaban. Los endpoints de descarga (`/calendar.ics` y `/calendar.pdf`)
ya estaban hechos y probados. Este es el punto de US8 que faltaba: cerrar la brecha entre lo que
carga la FIA y lo que ve el público.

**Prompt:**

> Conectar la vista pública del calendario y los paneles FIA y de escudería a la API de eventos, en
> lugar de los mocks, con filtro por período y categoría y botones para descargar lo que se está
> viendo en .ics y PDF. Los catálogos de estado del evento (etiquetas y colores) se comparten con el
> formulario de gestión en vez de duplicarse, y las banderas toleran cualquier país que llegue de la
> API sin quedar en blanco. Sacar los hooks y los datos de calendario que quedan sin uso. La
> verificación de que el archivo descargado se abra sin conexión queda para hacerla a mano.

**Respuesta de la IA:** propuso unificar los hooks de calendario en `src/data/calendar.ts` (en lugar
de un `calendarHooks.ts` aparte), extraer las etiquetas y tonos de tipo de evento a
`features/eventKind.ts` y darle a `Flag` un reemplazo genérico más el juego de banderas que faltaba
(el país llega como código libre desde la API). Para las fotos por categoría, que también llegan como
texto, agregó un `categoryPhoto` con respaldo en F1 y así sacó los `as any`. Dejó sin marcar los dos
criterios de US8 que dependen de abrir el archivo descargado.

**Resultado:** `app/src/data/calendar.ts` (con `useUpcomingEvents`, `useNextRace` y
`calendarDownloadUrl`), `app/src/features/eventKind.ts`, `app/src/features/{EventList,NextRaceHero}.tsx`,
`app/app/(public)/calendar.tsx`, `app/app/{fia,team}/index.tsx`, `app/src/photos.ts`,
`app/src/illustrations/Flag.tsx` y `app/src/mocks/data.ts` (sin los datos de calendario). `make check`
en verde.

---

### 2026-10-15 — Gestión de pilotos: modelo de las asignaciones (US11)

**Contexto:** el esquema guarda las asignaciones de piloto a escudería como períodos implícitos
append-only (`000007`): cerrar un período es insertar el siguiente, y `assigned_team()` deriva el
equipo de una fecha. Faltaban dos reglas de US11: poder dar de baja a un piloto (sin una fila de
baja, un piloto liberado seguiría en la alineación) y evitar que una asignación retroactiva cambie
el equipo derivado de una carrera ya puntuada, que quedaría desfasado del equipo guardado en
`score_entries`.

**Prompt:**

> Modelar en una migración nueva la baja y la retroactividad de las asignaciones. Una baja es una
> fila de `driver_team_assignments` sin escudería desde una fecha, sin romper el append-only:
> `assigned_team()` tiene que devolver NULL después, y los guardias que necesitan un equipo ya fallan
> solos. Además, una asignación nueva solo puede extender el historial: tiene que empezar después de
> la última del piloto y nunca antes de una carrera en la que ya tenga resultados, porque el puntaje
> ya guardó el equipo con el que corrió. Usar un código de guardia propio que el servicio traduzca a
> 4xx y cubrir cada regla con tests de esquema.

**Respuesta de la IA:** hizo `team_id` nullable y agregó el trigger de extensión como `AFTER INSERT`,
no `BEFORE`: un disparador previo corre antes de resolver el `ON CONFLICT DO NOTHING` del seed
idempotente y lo rompía. Al ser `AFTER`, solo ve las filas realmente insertadas y excluye la propia
fila al leer la última asignación. Código `RC004`.

**Resultado:** `backend/migrations/000022_assignment_lifecycle.{up,down}.sql`,
`backend/tests/db/assignment_lifecycle_test.go`. `make check` en verde.

---

### 2026-10-15 — Gestión de pilotos: backend (US11)

**Contexto:** con el modelo de asignaciones listo, faltaba el backend de US11. El rol `team_admin` ya
llevaba `team_id` en la sesión, pero ningún endpoint gestionaba pilotos: el listado y el alta seguían
siendo maqueta.

**Prompt:**

> Implementar el backend de gestión de pilotos para el administrador de escudería: listar, crear,
> editar y dar de baja pilotos de la propia escudería. Un piloto de otra escudería responde 404, no
> 403, para no revelar que existe. El alta crea la persona y su primera asignación al equipo en una
> transacción y, opcionalmente, el número de la temporada; la categoría del número es la del equipo,
> nunca la elige el cliente. Cambiar el rol titular/suplente es un período nuevo de asignación, no una
> edición. Validar nombre y apellido obligatorios, licencia de 3 a 32 caracteres, país de dos letras y
> número entre 0 y 999, y traducir las violaciones de unicidad y los guardias del esquema a conflictos
> o validaciones. Seguir el patrón handler → service → repo con sqlc y cubrir todo con tests de
> integración, incluido el intento sobre un piloto de otra escudería.

**Respuesta de la IA:** armó el paquete `internal/drivers` con `sql/drivers.sql` y endpoints
`/team/drivers` restringidos a `team_admin`. Para la alineación vigente usó un `JOIN LATERAL` que
toma la última asignación antes de decidir el equipo, para que la fila de baja saque al piloto del
listado; sqlc no infiere la nulabilidad a través de un `LEFT JOIN LATERAL`, así que el número se lee
con un `LEFT JOIN` a la tabla real. El número se archiva con un `upsert` por (piloto, temporada,
categoría).

**Resultado:** `backend/sql/drivers.sql` y `internal/db/dbgen/drivers.sql.go`,
`backend/internal/drivers/{driver,validation,service}.go`,
`backend/internal/transport/drivers_handlers.go` y rutas, `backend/tests/transport/drivers_integration_test.go`,
y el cableado en `cmd/api/main.go`. `make check` en verde.

---

### 2026-10-15 — Gestión de pilotos: pantallas de la app (US11)

**Contexto:** el backend de pilotos ya estaba, pero la pantalla `app/team/drivers.tsx` seguía leyendo
la maqueta de US4 y el botón de alta estaba deshabilitado. El dashboard de escudería mostraba la
alineación desde el mismo mock, así que los cambios del administrador no se veían.

**Prompt:**

> Conectar la gestión de pilotos de la app a la API de US11: listado de la alineación, alta, edición
> con baja confirmada, con el rol y el número en el formulario. Un piloto tocado en la tarjeta abre su
> edición. Validar cada error de la API en el campo que corresponde y reflejar los cambios de
> inmediato. Mantener la lógica fuera de los componentes de pantalla, reutilizar el design system y no
> duplicar código.

**Respuesta de la IA:** creó `src/api/drivers.ts` y `src/data/drivers.ts` sobre `useRemote`, convirtió
`app/team/drivers.tsx` en una carpeta (`index`, `new`, `[id]`, `_layout`) como la de usuarios, y
cambió `DriverCard` al tipo de la API. Al sumarse un tercer `FieldGroup` duplicado, lo extrajo a
`src/components/FieldGroup.tsx` y migró los formularios de cuentas y de eventos.

**Resultado:** `app/src/api/drivers.ts`, `app/src/data/drivers.ts`,
`app/src/features/drivers/{DriverFields.tsx,driverForm.ts}`, `app/src/features/DriverCard.tsx`,
`app/app/team/drivers/{index,new,[id],_layout}.tsx`, `app/app/team/index.tsx`,
`app/src/components/FieldGroup.tsx` y `app/src/{data/hooks.ts,mocks/data.ts}` sin los pilotos de la
maqueta. `make check` en verde.

---

### 2026-10-15 — Visualización de pilotos (US12)

**Contexto:** US11 dejó la alineación vigente por escudería, pero solo para el administrador de esa
escudería (`/team/drivers`). US12 pide que cualquier persona, sin iniciar sesión, consulte la
alineación vigente agrupada por escudería, con búsqueda y filtros, y el detalle de cada piloto.

**Prompt:**

> Exponer y mostrar la alineación vigente de forma pública. Agregar endpoints de solo lectura, sin
> sesión, para listar los pilotos con su escudería y categoría (filtrables por categoría y por nombre)
> y para el detalle de uno. La vista pública arma una pestaña de Pilotos con búsqueda y filtro de
> categoría, agrupa la alineación por escudería y abre el detalle de cada piloto; ningún usuario sin
> permisos puede modificar los datos. Reutilizar el read model de pilotos de US11 y no duplicar la
> consulta de la alineación vigente.

**Respuesta de la IA:** sumó `ListActiveDrivers`/`GetActiveDriver` a `sql/drivers.sql` con el mismo
`JOIN LATERAL` de US11 (la última asignación decide el equipo, así una baja saca al piloto del
listado) y filtros `sqlc.narg` por categoría y por nombre completo, con escape de `LIKE`. En la app
agregó `usePublicDrivers`/`usePublicDriver`, una pestaña de Pilotos con búsqueda debounced y agrupó
la lista plana por equipo en una sola pasada. Compartió un tipo `DriverSummary` para que `DriverCard`
sirva a la pantalla de gestión y al directorio.

**Resultado:** `backend/sql/drivers.sql` y `internal/db/dbgen/drivers.sql.go`,
`backend/internal/drivers/directory.go`, `backend/internal/transport/drivers_handlers.go` y rutas,
`backend/tests/transport/drivers_directory_integration_test.go`,
`app/src/api/drivers.ts`, `app/src/data/drivers.ts`, `app/src/features/DriverCard.tsx`,
`app/app/(public)/drivers/{index,[id],_layout}.tsx` y la pestaña en `app/app/(public)/_layout.tsx`.
`make check` en verde.

---

### 2026-10-15 — Gestión de puntajes (US9)

**Contexto:** el esquema ya tenía `score_entries`, `score_change_log` y sus ítems por piloto, pero
ninguna query los tocaba y la pantalla de resultados del panel FIA seguía siendo la maqueta de US4.
US9 pide cargar y modificar los puntajes de cada carrera, que cada cambio quede auditado con autor y
fecha, y que la clasificación general refleje el cambio. Faltaba, además, un guardia: nada impedía
escribir en `score_entries` sin dejar rastro.

**Prompt:**

> Implementar la gestión de puntajes del administrador FIA. Una carrera tiene una grilla: los
> pilotos de su categoría con sus puntos, que se cargan y se modifican de una vez. Validar negativos,
> campos vacíos y pilotos repetidos, rechazar un piloto que no corre en la categoría de la carrera
> (traducir `RC002` a 4xx) y restringir todo a `fia_admin`. Cada edición tiene que quedar registrada
> en `score_change_log` con su autor y fecha y el detalle por piloto, y la clasificación general
> tiene que reflejarla sin recálculo manual. Como el log no debe poder saltearse, sacar un guardia en
> la base que exija declarar el autor de la escritura.

**Respuesta de la IA:** propuso el armado de la versión en el **servicio** (una fila de
`score_change_log` + un ítem por piloto, en la misma transacción) y dejar en la base solo un guardia
barato: `BEFORE INSERT OR UPDATE OR DELETE` en `score_entries` que exige `SET LOCAL
racecontrol.actor_id` y falla con `RC005`. Descartó —y dejó documentado por qué— hacer que el trigger
armara la versión: agrupar las filas de una edición en un solo log pedía llevar un `change_log_id`
entre filas (un GUC local) o tablas de transición, más maquinaria de la que la auditoría vale. Para
los puntos agregó un override de `numeric` a `float64` en `sqlc.yaml` y un `SetScoreActor` que corre
`set_config(..., true)` dentro de la transacción.

**Resultado:** `backend/migrations/000026_score_audit_guard.{up,down}.sql`, `backend/sql/scores.sql`
y `internal/db/dbgen/scores.sql.go`, `backend/internal/scores/service.go`,
`backend/internal/transport/scores_handlers.go` y rutas, `sqlc.yaml` (override de puntos),
`seeds/demo.sql` (el seed declara un actor), `tests/testdb` y los tests que escriben
`score_entries` directo, `backend/tests/transport/scores_integration_test.go`,
`backend/tests/db/score_guard_test.go` (RC005),
`app/src/api/{scores.ts,client.ts}`, `app/src/data/scores.ts`,
`app/app/fia/results/{index,[id],_layout}.tsx` en lugar de `app/app/fia/results.tsx`. `make check` en
verde.

---

### 2026-10-15 — Puntajes: modelo y backend (US23, US24)

**Contexto:** las tablas de puntajes, del log de cambios y de los acuses existían desde el Sprint 1,
pero ningún endpoint las leía: la clasificación y los acuses de la app eran maqueta. Quedaba abierto
B.2 (quién puede acusar) y la tarea de US24 que pedía agregar tipo y vínculo al cambio en
`notifications`.

**Prompt:**

> Implementar US23 y US24 en el backend. La clasificación y los resultados se calculan con consultas,
> nunca se guardan: la posición es un `RANK()` por puntos (los empates comparten), solo cuentan las
> carreras activas del campeonato y la de escuderías suma el equipo acreditado en cada resultado.
> Endpoints públicos de resultados por carrera y de clasificación. Los avisos de puntajes no se
> guardan: un aviso es una fila de `score_change_log` cuyos ítems incluyen a la escudería, publicado si
> es el primer cambio de la carrera y corregido si no, confirmado si existe su acuse. Migración nueva
> con un trigger que derive la escudería del acuse desde el usuario y rechace con `RC005` un cambio
> que no la afecta. Endpoints de escudería (listar avisos con el contador de pendientes y confirmar,
> con alcance a la propia escudería) y de FIA (quién confirmó cada cambio). Datos de demo con una
> publicación, una corrección y un acuse. Tests de integración de cada endpoint y de base del trigger.

**Respuesta de la IA:** separó la lectura en dos vistas, `race_scores` (`000024`) y `score_notices`
(`000025`), para que resultados, clasificación de pilotos y de escuderías, avisos y seguimiento de la
FIA compartan una sola definición, y embebió `calendar_events` con `sqlc.embed` para reutilizar la
etiqueta de la carrera del calendario. Propuso responder `RC005` con 404 en lugar de 409, porque
distinguir "no existe" de "es de otra escudería" revelaría cambios ajenos, y hacer idempotente el
segundo acuse con `ON CONFLICT DO NOTHING`. Como el log es append-only y no se puede limpiar, los
tests arman un campeonato propio por test (`tests/testdb/championship.go`) y los de base corren en
una transacción que se descarta. La historia de demo va en `seeds/demo_scores.sql`, aplicada después
de crear las cuentas porque las nombra.

**Resultado:** migraciones `000023`–`000025`, `backend/sql/{scores,score_notices}.sql`,
`backend/internal/scores/`, `backend/internal/transport/{scores,score_notices}_handlers.go` y rutas,
`backend/seeds/demo_scores.sql`, tests en `backend/tests/db/acknowledgement_guard_test.go` y
`backend/tests/transport/{scores,score_notices}_integration_test.go`. Decisiones en
`02-modelo-de-datos.md` §9. `make check` en verde.

---

### 2026-10-15 — Puntajes: pantallas de la app (US23, US24)

**Contexto:** con la API de puntajes lista, la pantalla pública de puntajes, el dashboard de escudería
y los comunicados de escudería y de FIA seguían leyendo la maqueta de US4.

**Prompt:**

> Conectar la app a la API de puntajes. Pantalla pública con selector de categoría, clasificación de
> pilotos y de escuderías y resultados por carrera con filtro de carrera. Vista de puntajes para el
> administrador de escudería. En notificaciones de escudería, los avisos de publicación y corrección
> con su estado pendiente o confirmado y un botón "Confirmar recepción", con un indicador de
> pendientes. En la FIA, qué escuderías confirmaron cada cambio. Lógica en hooks o en `src/api`,
> reutilizar los componentes existentes y quitar los mocks reemplazados.

**Respuesta de la IA:** generalizó `StandingsTable` a filas ordenadas genéricas para usarla en
pilotos, escuderías y resultados de una carrera, y armó `ScoreBoard` como bloque compartido entre la
pantalla pública y la de escudería. Para que la insignia de la pestaña, el dashboard y la pantalla de
notificaciones muestren el mismo estado, puso los avisos en un proveedor del layout de escudería que
recarga después de confirmar. Antes de escribir un tercer armado de query string lo extrajo a
`api/client.ts` en un commit de refactor aparte. Dejó el mock de avisos que usa la pantalla de
resultados de la FIA, que es de US9.

**Resultado:** `app/src/api/scores.ts`, `app/src/data/{scores.ts,TeamScoreNoticesProvider.tsx}`,
`app/src/features/scores/`, `app/src/features/StandingsTable.tsx`, las pantallas
`app/app/(public)/{index,standings}.tsx`, `app/app/team/{_layout,index,scores,inbox,notifications}.tsx`
y `app/app/fia/{_layout,index,notifications}.tsx`; se borraron `NotificationList` y los mocks de
clasificación y comunicados. `make check` en verde.

---

### 2026-10-15 — Puntajes: correcciones tras la revisión (US23, US24)

**Contexto:** la revisión del manager y un tester independiente no encontraron bloqueantes en US23 y
US24, pero sí una serie de puntos a corregir antes de cerrar.

**Prompt:**

> Corregir: la seed de demo podía colgar sus ítems y su acuse de un cambio ajeno con el mismo id; la
> pantalla de resultados de la FIA mostraba acuses de maqueta; la insignia de pendientes de la
> escudería solo se actualizaba al entrar a Notificaciones; la categoría de la escudería salía de los
> mocks; extraer los envoltorios repetidos de `pgtype`; agregar tests de un pase de escudería a mitad
> de temporada y del aislamiento por categoría y temporada; y anotar en US9 que la posición se deriva
> de los puntos.

**Respuesta de la IA:** la seed ahora solo escribe sobre sus propios cambios (id, carrera de demo y
`fia.admin`) mediante una vista temporal, y la corrección exige la publicación propia; un test de base
aplica la seed sobre un cambio ajeno dentro de una transacción descartada y verifica que no lo toca.
Los avisos se recargan desde el layout de escudería con `screenListeners` en cada foco de pestaña, sin
sondeo. La categoría se toma de `GET /teams`. `db.Int8`/`db.Int4` reemplazan los literales de
`pgtype` en un commit de refactor aparte.

**Resultado:** `backend/seeds/demo_scores.sql`, `backend/tests/db/demo_scores_seed_test.go`,
`backend/internal/db/pgconv.go`, `backend/tests/testdb/championship.go`,
`backend/tests/transport/scores_integration_test.go`, `app/app/fia/results.tsx`,
`app/app/team/{_layout,index,scores,notifications}.tsx`, `app/src/data/{scores,teams}.ts`, y la tarea
nueva de US9 en `sprint.md`. `make check` en verde.

---

### 2026-10-16 — Revisión de US9 y US12 tras integrar el trabajo del equipo

**Contexto:** después de integrar la carga de puntajes (US9) y el directorio de pilotos (US12), se
revisó que nada se hubiera roto y que la implementación fuera correcta.

**Prompt:**

> Revisar lo que integró mi compañero: que no se haya roto nada y que todo esté bien; corregir lo que
> falle o esté mal.

**Respuesta de la IA:** `make check` en verde; el error de tipos de la app era el `router.d.ts`
generado localmente y desactualizado, no del código. Se encontraron y corrigieron: el guardia de
auditoría (`000026`) reusaba `RC005`, el código del acuse, y pasa a `RC006` con una migración nueva
(`000027`); una carrera cancelada aceptaba puntajes y una futura también, lo que congelaba su fecha
(`RC001`), así que la primera responde `404` y la segunda `409`, y la pantalla de la FIA lista solo
carreras terminadas; un puntaje sobre 9999.99 daba `500` y uno con tres decimales se redondeaba en
silencio (`numeric(6, 2)`), ahora ambos se rechazan; la edición se calcula dentro de la transacción
con la carrera bloqueada, para que dos ediciones simultáneas no registren puntos viejos falsos; y el
número del piloto en la grilla es el de la temporada de la carrera, no el último.

**Resultado:** `backend/migrations/000027_*`, `backend/sql/scores.sql`,
`backend/internal/scores/service.go`, `backend/tests/transport/scores_integration_test.go`,
`backend/tests/db/score_guard_test.go`, `app/app/fia/results/index.tsx`, `app/src/data/scores.ts`,
`docs/sprint2/{sprint,02-modelo-de-datos}.md`. `make check` en verde.

---

### 2026-10-16 — Mensajería interna segura (US17)

**Contexto:** US17 pide que los responsables de escudería y los administradores de la FIA se envíen
mensajes dentro de la plataforma, con acceso solo para los participantes, entradas validadas,
transmisión segura e indicador de no leídos.

**Prompt:**

> Implementar US17: una conversación es un hilo entre una escudería y la FIA con asunto; participan
> los administradores activos de la escudería y todos los de la FIA. Migración `000028` solo DDL con
> mensajes de solo agregado, cursor de lectura por usuario y un guardia `RC007` que rechace a quien no
> participa; el servicio también lo verifica y responde 404. No leídos derivados, sin escribir en
> `notifications`. Bandejas de entrada y enviados, validación en español, seed de demo, pestaña
> "Mensajes" con insignia para escudería y FIA, tests de envío, recepción y acceso no autorizado.

**Respuesta de la IA:** tablas `conversations`, `messages` y `conversation_reads` con la función
`is_conversation_participant` compartida por los dos guardias; agregar un mensaje bloquea la
conversación antes de tomar el id para que el cursor nunca saltee uno que se confirma tarde. Paquete
`internal/messaging` y endpoints `/conversations` (lista con `box`, alta, lectura que marca leído y
respuesta). Seed `demo_messages.sql` con el mismo resguardo que la de puntajes. En la app, pestaña
"Mensajes" con pila (lista, conversación, nueva) y un proveedor de no leídos que actualiza la
insignia; el selector de escudería del formulario de cuentas se extrajo para reusarlo.

**Resultado:** `backend/migrations/000028_*`, `backend/sql/messages.sql`, `backend/internal/messaging/`,
`backend/internal/transport/messaging_handlers.go`, `backend/seeds/demo_messages.sql`,
`backend/tests/db/{messaging_guard,demo_messages_seed}_test.go`,
`backend/tests/transport/messaging_integration_test.go`, `app/app/{team,fia}/messages/`,
`app/src/{api/messages.ts,data/messages.ts,data/UnreadMessagesProvider.tsx,features/messages/,features/TeamPicker.tsx}`,
`docs/sprint2/{sprint,02-modelo-de-datos}.md`. `make check` en verde.

---

### 2026-10-17 — Navegación de conversaciones (US17)

**Pedido:** revisar el sprint y las notas, aplicar correcciones y crear commits por US en español.

**Cambios:** corregir `Screen` para mostrar acciones aun sin título; agregar regreso explícito a
la bandeja según el rol (incluye enlaces directos), recarga manual del hilo y recarga al recuperar
el foco. El problema del botón oculto afectaba también otras pantallas de detalle.

**Validación:** `make check` aprobado con integración en PostgreSQL. La verificación del proxy
HTTPS real permanece pendiente; localhost usa HTTP.

---
