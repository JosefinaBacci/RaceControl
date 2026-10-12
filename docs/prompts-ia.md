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
