# Registro de prompts de IA

La cátedra (**APS 2026, Enunciado 1**) exige registrar los prompts utilizados con IA para generar
código o decisiones de diseño, porque "la correcta utilización de las herramientas de IA es parte
de la evaluación del proyecto".

Este archivo es el registro de la **comisión implementadora** (Grupo Rojo Ferrari). Cada entrada
declara el contexto, el prompt utilizado, la respuesta de la herramienta y qué se validó o
modificó de la salida.

> Los prompts usados por la comisión de Análisis y Management se registran en su propio
> repositorio.

---

## Regla de trabajo

Antes de dar por cerrada cualquier tarea que haya usado IA:

1. Agregar la entrada acá con el prompt utilizado.
2. Revisar la salida: leer el código generado, ejecutar los tests, verificar que no haya
   introducido dependencias no deseadas ni secretos.
3. Anotar en el campo **Validación** qué se verificó y qué se corrigió a mano.

Las entradas se registran en el momento de la tarea, nunca a posteriori. Los prompts se redactan
de forma clara y autocontenida, conservando la intención, el alcance y las restricciones del
pedido original. Cuando el pedido se apoyó en el contexto previo de la conversación, ese contexto
se explicita en el campo **Contexto**.

Todas las entradas siguen la misma estructura: **Contexto**, **Prompt**, **Respuesta de la IA**
(cuando aporta algo que no queda reflejado en el resultado), **Validación** y **Resultado**.

---

## Entradas

### 2026-10-02 — Definición del stack tecnológico (US1)

**Contexto:** la comisión implementadora tenía que decidir backend, base de datos, framework de
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

**Validación:** se ajustó la recomendación inicial. La IA propuso sesiones en servidor por
seguridad, pero el criterio definitivo se fijó con el equipo: **JWT puro se descarta** porque US6
exige desactivar usuarios y un JWT no se puede revocar antes de su expiración, lo que dejaría
acceso a cuentas desactivadas. También se acotó el árbol de dependencias del frontend a tres
paquetes (`expo`, `expo-router`, `expo-secure-store`) justificando que `expo-router` reemplaza a
`react-navigation` en lugar de sumarse. La comparación de frameworks de backend (chi / gin / echo /
fiber / stdlib) se contrastó con la documentación de cada proyecto antes de escribir
`01-stack.md`.

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

**Validación:** la estructura se ajustó a mano para cumplir las reglas de clean code del curso
(funciones cortas, sin singletons, errores envueltos con contexto). Se descartó la propuesta de
generar los modelos de dominio por reflexión a partir de las structs de sqlc, y quedó
documentado que los errores de dominio se definen en el service y se chequean con `errors.Is`.

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

**Validación:** la salida se revisó y se corrigió en cuatro puntos.

1. La primera propuesta modelaba el sujeto de una sanción como `(subject_type, subject_id)`, el
   patrón polimórfico. Se descartó: no tiene clave foránea ni integridad referencial. Se adoptó
   dos FK opcionales con `CHECK ((driver_id IS NULL) <> (team_id IS NOT NULL))`.
2. Se dudó guardar `team_id` en `score_entries` por redundante. Se confirmó como
   denormalización **deliberada y correcta**: un puntaje registra para qué escudería se contabilizó
   en el momento de la carrera, y derivar el equipo actual mostraría un histórico falso.
3. Se evaluó usar el tipo `ENUM` de PostgreSQL para roles y estados. Se rechazó: agregar un valor
   exige `ALTER TYPE` con lock sobre la tabla. Se usaron `CHECK`, que hace que agregar un rol sea
   una migración aditiva.
4. La recomendación inicial de claves UUID para evitar conflictos al fusionar ramas se evaluó y se
   descartó por decisión del equipo: el problema real no son las claves sino las filas. Se adoptó
   `bigint GENERATED ALWAYS AS IDENTITY` (que impide asignar ids a mano y desincronizar la
   secuencia) más seeds con ids fijos y `setval`, con `make db-reset` como mecanismo de resolución.

También se verificó contra la base real que las ocho restricciones `CHECK` rechazan el dato
inválido, que `make migrate` es idempotente, que `make migrate-down` revierte, y que
`make db-reset` reproduce los mismos ids de seed.

**Resultado:** `backend/migrations/000001_core.*`, `backend/sqlc.yaml`, `backend/sql/users.sql`,
`docs/sprint1/02-modelo-de-datos.md`

> Correcciones posteriores:
>
> - El seed `000002_seed_reference` de esta entrada fue eliminado en el commit *Separa esquema y
>   datos de referencia* (ver la entrada siguiente).
> - El `CHECK` del punto 1 estaba invertido: rechazaba los dos casos válidos y aceptaba los dos
>   inválidos. Se reemplazó por `num_nonnulls(driver_id, team_id) = 1` (ver la entrada
>   *Corrección del CHECK de sanciones y revisión de la migración inicial*).

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

**Implementación:** se borró `000002_seed_reference.*`; se creó `backend/seeds/reference.sql`
(idempotente, con `setval` sobre el `max(id)` vivo), `backend/seeds/embed.go`,
`backend/internal/seed/` y `backend/cmd/seed/`. Se agregó `APP_ENV` y `SEED_DEMO_PASSWORD` a la
configuración, con una validación que **impide arrancar** si `SEED_DEMO_PASSWORD` está presente
con `APP_ENV=production`. `make seed` pasó a ser `go run ./cmd/seed` y `make db-reset` quedó como
drop + migrate + seed.

**Validación:** se comprobó contra la base real que `make db-reset` reproduce `categories` 1–4 y
`teams` 1–8, que `make seed` corrido dos veces no duplica filas, y que no hay retroceso de
secuencia: tras insertar un equipo desde la aplicación (id 9) y reejecutar el seed, el siguiente
insert recibió el id 10.

**Corrección de la IA (asumida explícitamente):** la IA predijo que `make migrate-down 1` dejaría
las filas intactas. Fue **incorrecto**: con migraciones solo DDL ese comando tira el esquema
entero, tablas y filas incluidas. El comportamiento real verificado es que `make migrate` no
inserta datos y la recuperación es `make migrate` + `make seed`, que devuelve los mismos ids. El
beneficio real del cambio no es "el rollback no borra filas" sino que desaparece el estado
intermedio confuso del diseño anterior, donde quedaba un esquema válido pero sin datos de
referencia.

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

**Respuesta de la IA:** el documento propone un monolito modular, una matriz de permisos de 13
filas, estándares de Argon2id y sesiones opacas, un modelo de amenazas y una tabla de qué está
implementado contra qué llega en US5 y US6. Decisiones que se apartan de lo obvio, y por qué:

1. **Sesiones opacas en vez de JWT.** Un JWT se valida sin tocar la base, que es su ventaja y
   también el problema: la revocación queda diferida hasta que expira. El enunciado pide que
   desactivar un usuario lo saque del sistema de inmediato, y eso obliga a consultar la base en cada
   request. El costo se aceptó a propósito.
2. **La rotación del token no es en cada request.** Es la recomendación habitual, pero dos
   peticiones simultáneas con la misma sesión (un formulario que dispara dos requests, dos
   pestañas) se invalidarían mutuamente y cerrarían la sesión del usuario sin que hiciera nada. Se
   rota en login, en cambio de privilegio y ante acción sensible. Esta decisión **corrige lo que
   el propio TODO del equipo daba por sentado**, que decía "rotación en cada uso".
3. **El público no es un valor de `users.role`.** Es la ausencia de sesión. Guardarlo como rol
   obligaría a aceptar usuarios sin contraseña y volvería laxa la restricción `CHECK` del rol.
4. **El `team_id` de un `team_admin` sale de la sesión, nunca del cuerpo de la petición.** Si el
   alcance fuera un parámetro, cambiar un campo del JSON bastaría para leer datos de otra escudería.
5. **`last_used_at` se amortigua a 5 minutos** para que una sesión activa no convierta cada lectura
   en un `UPDATE`, aceptando que una sesión viva hasta 5 minutos más que el TTL.
6. **Política de contraseñas sin reglas de símbolos**, solo longitud mínima de 12 y rechazo si se
   parece al `username`. Las reglas de símbolos empujan a patrones predecibles.

**Validación:** se contrastó cada afirmación contra el código ya escrito, no contra memoria: el
formato de `username`, el `CHECK` de rol, la exclusividad de `team_id`, las tablas `sessions` y
`login_attempts` y los parámetros de Argon2id se verificaron en `migrations/000001_core.up.sql`,
`internal/config/config.go` y `internal/transport/router.go`.

**Desviación registrada:** la US3 se estimaba en 9 h y se resolvió en 0.2 h. El hallazgo para el
Sprint 2, anotado en el TODO, es que las historias de diseño de este sprint se desvían por el
trabajo de decidir y justificar, no por el de escribir: conviene separar ambos en la estimación.

**Nota sobre el proceso:** la redacción larga en español produjo, de forma repetida, caracteres CJK
o cirílicos en líneas concretas. Se detectó con un escaneo de todo el archivo por carácter fuera del
rango latino, y se corrigió en tres documentos de este commit. Es un fallo recurrente de la
herramienta y conviene escanear siempre después de escribir documentación en español.

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

**Validación:** los hallazgos se verificaron contra el código y no contra lo que dicen los
documentos. El bug 2 se reprodujo apuntando `DATABASE_URL` a un puerto sin servicio. El bug 1 se
comprobó con la tabla de verdad de los cuatro casos. El bug 3 se reconstruyó a partir de
`app/vercel.json`, `transport/auth_handlers.go`, `app/src/api/tokenStore.web.ts` y la §3.4 de
US3, y queda pendiente de una decisión del equipo sobre dónde se aloja la API. La omisión de
los tests se confirmó con `go test -v`, que los reporta como `SKIP`.

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

**Respuesta de la IA:** el cambio es de una línea por rama. El error original pasa a ser el que
se envuelve, y el error de cierre se suma con `errors.Join`, que descarta los `nil`:

```go
// antes
return nil, errors.Join(fmt.Errorf("create migration driver: %w", db.Close()))
// después
return nil, errors.Join(fmt.Errorf("create migration driver: %w", err), db.Close())
```

Lo mismo en la rama `create migrator`. `go vet` no lo detectaba porque el código original es Go
válido: `%w` con un `nil` no falla, imprime `%!w(<nil>)`.

**Validación:** se reprodujo el caso antes y después del cambio, con la base inaccesible:

```
antes:   migrate: create migration driver: %!w(<nil>)
después: migrate: create migration driver: failed to connect to `user=x database=none`:
         127.0.0.1:1 (127.0.0.1): dial error: dial tcp 127.0.0.1:1: connect: connection refused
```

`go vet`, `gofmt` y `go test ./...` en verde (los tests de integración se omiten sin
`DATABASE_URL`). En este registro se unificó el encabezado **Prompt:**, todas las entradas
pasaron a la estructura Contexto → Prompt → Respuesta → Validación → Resultado, se agregaron los
separadores que faltaban, se corrigieron erratas y la regla de trabajo ahora explica cómo se
redactan los prompts.

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

**Respuesta de la IA:** para las sanciones se propuso
`CHECK (num_nonnulls(driver_id, team_id) = 1)`, que se lee igual que la regla ("exactamente un
sancionado") y elimina la comparación de booleanos que había llevado al error. Sobre la migración
inicial, los hallazgos fueron:

1. **`users_deactivation_timestamped` cubre un solo sentido.** `is_active OR deactivated_at IS NOT
   NULL` impide desactivar sin fecha, pero permite un usuario activo con `deactivated_at` cargado.
   Además, `is_active` se deriva de `deactivated_at`, lo que contradice la regla de US2 de no
   repetir datos derivables.
2. **Índice redundante.** `teams_category_id_idx` duplica el prefijo del índice único
   `(category_id, code)`, que PostgreSQL ya usa para filtrar por categoría y para validar la FK.
3. **Restricción redundante.** `users_username_lowercase` ya está implicada por
   `users_username_format`, cuya expresión regular solo admite minúsculas.
4. **Legibilidad.** Los triggers de `updated_at` están agrupados al final, lejos de sus tablas, y
   el alineado de columnas es irregular en `users`. En el `down`, los `DROP TRIGGER` sobran:
   `DROP TABLE` ya elimina los triggers de la tabla.

**Decisión del equipo:** editar `000001` en lugar de agregar una `000002`. La regla de
`AGENTS.md` que prohíbe editar migraciones aplicadas protege bases desplegadas, y esta todavía no
corre en ningún entorno compartido: el costo es que cada integrante corra `make db-reset`. Se
aplicaron los cuatro puntos: `is_active` pasó a ser una columna generada
(`GENERATED ALWAYS AS (deactivated_at IS NULL) STORED`), se eliminaron el índice y la restricción
redundantes, cada trigger quedó junto a su tabla y el `down` se redujo a los `DROP TABLE` y la
función.

**Validación:** la tabla de verdad de las sanciones se comprobó a mano para los cuatro casos.
`sqlc generate`, que analiza el esquema con el parser de PostgreSQL, aceptó la migración y
produjo exactamente el mismo código que antes. Para eso `is_active` conserva su posición en la
tabla: con la columna movida, sqlc dejaba de mapear las consultas al struct `User`. El test que
desactivaba usuarios escribiendo `is_active` se ajustó para marcar solo `deactivated_at`.
`go vet`, `gofmt` y `go test ./...` en verde.

Verificación posterior contra Postgres 18 (`make db-reset` y `make check` con `DATABASE_URL`):
la migración y el seed se aplican sin errores y los 16 tests de integración pasan. Además se
probaron las restricciones directamente con `psql`: `is_active` pasa de `true` a `false` al cargar
`deactivated_at` y PostgreSQL rechaza escribirla a mano; un `username` en mayúsculas y un
`team_admin` sin escudería se rechazan; y el `CHECK` de sanciones, probado sobre una tabla
temporal, acepta exactamente los dos casos válidos de los cuatro.

**Resultado:** `backend/migrations/000001_core.up.sql`, `backend/migrations/000001_core.down.sql`,
`backend/internal/transport/auth_integration_test.go`, `docs/sprint1/02-modelo-de-datos.md`,
`AGENTS.md`, y la nota de corrección en la entrada de US2 de este registro.

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

**Respuesta de la IA:** en la documentación, el criterio fue que cada afirmación coincida con el
código o quede marcada como pendiente:

1. **Rotación del token.** `01-stack.md` decía "rotación en cada uso" y `03` describía una
   rotación con detección de reuso que no estaba implementada. Se unificó el diseño en lo que se
   construye: sesión nueva en cada login y revocación de todas las sesiones ante un cambio de rol,
   escudería o estado (US6). La rotación por request sigue descartada por el problema de las
   peticiones simultáneas.
2. **Migraciones.** `01-stack.md` decía que se aplicaban al arrancar la API; se aplican con
   `make migrate`. Se documentó por qué es mejor así: el deploy decide cuándo cambia el esquema y
   dos réplicas nunca compiten por migrar.
3. **Por qué PostgreSQL.** `01-stack.md` lo justificaba por los `ENUM`, que el proyecto decidió no
   usar. Se reemplazó por los motivos reales: DDL transaccional, restricciones expresivas
   (`CHECK` con regex, columnas generadas, `num_nonnulls`), tipos como `inet` y el soporte de
   `sqlc`. También se quitó la mención a un CI que no existe.
4. **Dependencias de la app.** "Exactamente tres" ya no era cierto: se distinguieron las tres
   decisiones de arquitectura de los módulos del SDK que exige `expo-router` y de los tres que se
   sumaron en US4 para la interfaz.
5. **Modelo de datos (Sprint 2).** Las FK en plural (`teams_id`) contradecían la convención y se
   pasaron a singular. `drivers.category_id` se eliminó del diseño: la categoría se deriva de la
   escudería, y guardarla dos veces rompía la 3FN que el documento declara. Se quitó del diagrama
   una relación `categories → users` que no existe.
6. **Seguridad.** `03` afirmaba HSTS y logs con `RequestID`; la API no enviaba HSTS ni tiene logger
   de requests. Se agregó la cabecera y se corrigió la descripción de los logs. La purga de
   `login_attempts` y el conflicto del deploy en Vercel quedaron registrados como pendientes en
   la tabla de deuda técnica.

En el código, los cambios fueron:

- **Logout idempotente.** Exigía una sesión válida, así que con una sesión vencida respondía
  `401` y nunca borraba la cookie `httpOnly`, que el navegador no puede borrar por su cuenta.
  Ahora responde `204` siempre y limpia la cookie.
- **Un solo canal por plataforma.** El login web recibe solo la cookie y el móvil solo el token.
- **IP del cliente.** `middleware.RealIP` confiaba en `X-Forwarded-For` de cualquier cliente. Ahora
  solo se activa con `TRUST_PROXY_HEADERS=true`, detrás de un proxy que sobrescriba esa cabecera.
- **`config.Load`** pasó de una función de 60 líneas a funciones que devuelven valores, sin mutar
  argumentos (antipatrón listado en `AGENTS.md`).
- **Código muerto y duplicado.** Se eliminó el wrapper `internal/db/queries.go`, que solo usaban
  los tests, y su test dejó de duplicar `internal/testdb`. El test que decía verificar la
  secuencia del seed no la verificaba; ahora inserta una escudería y comprueba que su id supera
  los ids fijos del seed.
- **Detalles:** constantes con nombre para los timeouts, `pgerrcode.UniqueViolation` en lugar del
  literal `"23505"`, el `Retry-After` como dato de la tabla de errores y no como caso especial por
  status, la ayuda del flag `-down` corregida y `go.mod` ordenado con `go mod tidy`.
- **App.** El badge de acuse estaba copiado en tres lugares y se extrajo a
  `AcknowledgementBadge`. `SanctionList` recibía un booleano que cambiaba su comportamiento y
  ahora recibe qué mostrar a la derecha de cada fila. Los conteos que calculaban las pantallas
  del panel FIA y de escudería pasaron a hooks de `src/data`, como pide `AGENTS.md`.

**Validación:** `make check` en verde (`go vet`, `gofmt`, `go test` y `tsc`), y `expo export -p
web` compila el bundle sin errores. Se agregaron tests de integración para el logout sin sesión y
para el login móvil sin cookie; como el resto de los tests de integración, se omiten sin
`DATABASE_URL`; corridos después contra Postgres 18, los 16 pasan. Se escanearon todos los
documentos modificados en busca de caracteres fuera del rango latino.

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

**Respuesta de la IA:** un módulo `internal/users` con su service y un `internal/teams` mínimo
para el formulario de asignación. Decisiones que vale la pena justificar:

1. **Reutilizar antes que escribir** (a partir de la primera corrección del equipo, ver
   **Validación**). El cambio de contraseña usa la consulta existente
   `UpdateUserPasswordHash` y la revocación usa `RevokeUserSessions`; la nueva `UpdateAccount`
   solo toca email, rol y escudería. `GetUserByID` quedó sin uso y se eliminó en favor de
   `GetAccount`, que además trae el nombre de la escudería. También se eliminó
   `sessions.RevokeAllForUser`: la revocación ocurre dentro de la transacción de US6, así que el
   método quedaba sin uso.
2. **Transacciones explícitas.** Un `db.Transactor` agrupa la actualización y la revocación de
   sesiones: o se aplican las dos, o ninguna.
3. **La escudería sigue al rol.** Pasar una cuenta a `fia_admin` le quita la escudería sin que el
   cliente tenga que indicarlo, y pasarla a `team_admin` exige una escudería existente.
4. **Ningún administrador puede bloquearse a sí mismo.** No puede desactivar su cuenta ni cambiar
   su propio rol. Como cada administrador solo puede dar de baja a otros, siempre queda al menos
   uno activo.
5. **Qué revoca sesiones.** Un cambio de rol, escudería o contraseña y la baja revocan todas las
   sesiones; un cambio de email no, porque no altera privilegios.
6. **Búsqueda segura.** El texto de búsqueda se escapa antes de usarse en `LIKE`, porque los
   nombres de usuario admiten `_`, que `LIKE` interpretaría como comodín.
7. **Errores con campo.** Un `username` o email repetido responde `409` indicando el campo, para
   que la app marque el input correcto.

**Validación:** el equipo corrigió dos veces a la herramienta durante la implementación.

- **Código duplicado.** La primera versión del service escribía su propia validación de
  `username`, su propia detección de unicidad y sus propias conversiones de `pgtype`, y la consulta
  de actualización reescribía también la contraseña aunque ya existía `UpdateUserPasswordHash`. El
  equipo frenó la implementación para exigir que se revisara lo existente antes de escribir nada
  nuevo. De esa revisión salieron los helpers compartidos (en un commit de refactor propio), la
  reutilización de `UpdateUserPasswordHash` y `RevokeUserSessions`, y la eliminación de
  `GetUserByID` y `sessions.RevokeAllForUser`.
- **Comentarios inconsistentes.** El equipo preguntó si los comentarios nuevos seguían el estilo
  del código existente. No lo hacían: los paquetes de dominio, servicios y transporte no tienen
  comentarios, y la herramienta había agregado bloques explicativos en esos paquetes. Se
  eliminaron o se redujeron a una línea donde explican un porqué no evidente, y en el paquete
  `db`, que sí documenta cada función exportada, se completaron los que faltaban.

12 tests de integración nuevos contra Postgres 18 cubren los permisos (`401` sin
sesión, `403` para `team_admin`), el alta con normalización del email, siete casos de validación,
los duplicados, la búsqueda y sus filtros (incluido el `%` literal), la revocación de sesiones por
cambio de rol, contraseña y baja en dos dispositivos a la vez, que un cambio de email no las
revoca, el bloqueo de uno mismo y el catálogo de escuderías. `make check` en verde con los 25
tests de integración de `transport` corriendo, sin omitidos.

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

**Respuesta de la IA:** la pantalla pasó a una carpeta `app/fia/users/` con su propio stack
(listado, `new` y `[id]`), para que el alta y la edición se abran sobre el listado y tengan URL
propia en la web. Decisiones relevantes:

1. **Un helper de errores compartido.** El login ya traducía los errores de la API a mensajes por
   campo. En lugar de copiarlo, se extrajo a `src/api/formFailure.ts` (en un commit de refactor
   propio) y lo usan el login y los formularios de cuentas.
2. **Estados de carga en un solo lugar.** `useRemote` resuelve carga, error y recarga, y descarta
   respuestas viejas si el usuario cambia el filtro antes de que llegue la anterior.
   `RemoteContent` muestra el indicador de carga o el error con un botón de reintentar.
3. **Lógica fuera de las pantallas**, como pide `AGENTS.md`: qué cambió en el formulario se calcula
   en `src/features/accounts/accountChanges.ts`, y el formulario solo envía lo modificado.
4. **Búsqueda con espera de 300 ms**, para no hacer un pedido por cada tecla.
5. **Confirmación de la baja en la misma pantalla.** `Alert` de React Native no funciona en la web,
   así que la desactivación pide confirmar en el lugar: en la fila del listado o en la ficha de la
   cuenta. La lógica de baja y reactivación vive en un solo hook, `useStatusToggle`, que usan las
   dos vistas.
6. **La propia cuenta, protegida también en la interfaz.** El administrador no ve el selector de
   rol ni el botón de baja sobre su propia cuenta; el backend lo rechaza de todas formas.
7. **Sin datos de ejemplo de usuarios.** Se eliminaron los usuarios simulados y el panel FIA muestra
   la cantidad real de cuentas activas.

**Validación:** al revisar la primera versión, el equipo notó que la herramienta había quitado del
listado los botones de editar y de dar de baja que tenía la maqueta de US4, dejando solo la fila
navegable. Se restituyeron: el lápiz abre la ficha, la papelera desactiva la cuenta previa
confirmación en la misma fila, y una cuenta desactivada muestra en su lugar el botón de
reactivar. Para no duplicar la lógica entre la fila y la ficha se extrajo `useStatusToggle`.

`tsc` en verde y `expo export -p web` compila el bundle. Se verificó contra la API
real, con una sesión por cookie como la de la app web, el flujo completo: login como `fia.admin`,
alta con el email normalizado, búsqueda, cambio de rol que quita la escudería, el `409` con el
campo indicado ante un usuario repetido, baja y reactivación. También se comprobó que el
preflight de CORS permite `PATCH` desde el origen de la app.

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

**Respuesta de la IA:**

1. **Backend: `POST /account/password`.** Pide la contraseña actual, para que alguien frente a una
   sesión abierta ajena no pueda apropiarse de la cuenta. Si es incorrecta responde `400` sobre ese
   campo y no `401`, para que la app no interprete que la sesión se cerró. La nueva contraseña
   tiene que cumplir la política y ser distinta de la actual.
2. **Se cierran las demás sesiones, no la actual.** En lugar de escribir una consulta nueva, la
   existente `RevokeUserSessions` recibió un parámetro opcional con la sesión a conservar; las
   revocaciones de US6 la siguen usando sin él. La escritura del hash quedó en un helper común
   para no repetirla entre la edición de la FIA y el cambio propio.
3. **Menú de usuario.** Un avatar abre un menú con el nombre, el rol, "Mi panel", "Mi cuenta" y
   "Cerrar sesión". No se agregó "Configuración" porque todavía no hay nada que configurar. El
   botón del encabezado dejó de recibir un modo público o privado: muestra el menú si hay sesión
   y "Ingresar" si no la hay.
4. **Pantalla "Mi cuenta"** para los dos roles, con los datos de la cuenta y el formulario de cambio
   de contraseña con confirmación. Para permitir los dos roles se generalizó `RoleGate` a una
   lista de roles en lugar de escribir otra guarda.

**Validación:** 3 tests de integración nuevos contra Postgres 18: el cambio conserva la sesión que
lo hizo y cierra las de otro dispositivo, la contraseña vieja deja de funcionar y la nueva sí;
se rechazan la contraseña actual incorrecta, la nueva igual a la actual y la nueva demasiado
corta; y sin sesión responde `401`. `make check` en verde, `tsc` en verde y `expo export -p web`
compila el bundle.

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

**Respuesta de la IA:** cada logo se llevó a un PNG de 256 × 256 px: los de fondo transparente se
recortaron a su contenido y se centraron con margen, el de Red Bull se completó con su propio
color de fondo, y el de la FIA se recortó al emblema con el mismo margen que el resto. El total
bajó a unos 130 KB. Un registro `src/logos.ts`, con el mismo criterio que el de fotos, asocia
cada logo con el id fijo de su escudería en el seed. `Avatar` acepta una imagen opcional, así que
las escuderías sin logo siguen mostrando sus iniciales.

**Validación:** antes de integrarlos se generó una vista previa de los cuatro logos como avatares
sobre el fondo oscuro de la app; el de la FIA tocaba el borde del círculo y se volvió a recortar
con margen. `tsc` en verde y `expo export -p web` incluye los cuatro logos optimizados. Los logos
son marcas registradas: se documentaron en `app/assets/logos/CREDITS.md` con la misma aclaración de
uso académico que las fotos.

**Resultado:** `app/assets/logos/`, `app/src/logos.ts`, `app/src/components/Avatar.tsx`, el
listado de usuarios, la pantalla de resultados y el menú de usuario.

> Corrección posterior del equipo: con los logos en el avatar, cualquier cuenta mostraba el logo
> de su organización en lugar de algo propio, incluidas las cuentas de prueba. Un logo representa
> a una organización, no a una persona. Se decidió que los avatares de las personas vuelvan a
> mostrar sus iniciales, y que el logo aparezca como ícono junto al rol ("Administrador de
> escudería · Scuderia Ferrari") en el listado de usuarios y en el menú, mediante un componente
> `RoleIcon` que usa el ícono genérico cuando la organización no tiene logo. La pantalla de
> resultados conserva los logos porque sus filas son escuderías.

---

### 2026-10-03 — Base de datos propia para los tests de integración

**Contexto:** al probar la gestión de usuarios, el equipo encontró en el listado decenas de cuentas
con nombres como `t108965040125.1`. Eran usuarios creados por los tests de integración: `make
check` toma `DATABASE_URL` de `.env`, así que los tests corrían contra la misma base que usa la app
y cada corrida dejaba cuentas nuevas.

**Prompt:**

> Explicar de dónde salen las cuentas de prueba que aparecen en la gestión de usuarios y evitar que
> los tests escriban en la base de desarrollo.

**Respuesta de la IA:** los tests pasan a usar una base propia, `racecontrol_test`, en el mismo
contenedor de Postgres, configurada con una variable aparte, `TEST_DATABASE_URL`. El cambio de
código es uno solo, en `testdb.Pool`, que usan todos los tests de integración. Un target nuevo,
`make test-db`, crea esa base si no existe y le aplica las migraciones y los datos de referencia
(sin cuentas demo); `make backend-test` y `make check` lo ejecutan solos cuando la variable está
definida. Sin la variable, los tests de integración se omiten, nunca caen sobre la base de la app.

**Validación:** se contó la cantidad de usuarios de la base de desarrollo antes y después de
correr `make check`: quedó igual (247), mientras los usuarios de los tests se crearon en
`racecontrol_test`. Todos los tests en verde. Las cuentas que ya habían quedado en la base de
desarrollo se eliminan con `make db-reset`.

**Resultado:** `backend/internal/testdb/testdb.go`, `Makefile`, `.env.example`, `README.md`,
`AGENTS.md`.
