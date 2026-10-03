# Sprint 1 — Lista de tareas

Comisión **Grupo Rojo Ferrari** (Sevenants Antonio, Tourn Felipe) — Enunciado 1 (FIA).

Entrega: **lunes 5/10** · Demo: **martes 6/10**

Leyendo de arriba hacia abajo. Cada US se cierra con sus criterios de éxito verificados,
`make check` en verde y **un commit atómico**.

## Estado

| US | Nombre | SP | Valor | Est. | Real | Estado |
|----|--------|----|-------|-------|-----|--------|
| US1 | Selección de stack y configuración del entorno | 5 | 8 | 6h | 1.5h | ☑ |
| US2 | Diseño del modelo de datos | 5 | 8 | 9h | 3.5h | ☑ |
| US3 | Arquitectura, roles y seguridad | 5 | 8 | 9h | 0.2h | ☑ |
| US4 | Diseño de interfaz por rol | 5 | 8 | 8h | — | ◐ |
| US5 | Login de usuarios | 3 | 13 | 9h | — | ☑ |
| US6 | Gestión de usuarios | 8 | 21 | 12h | — | ☐ |

**Total estimado:** 53h · **Total real:** 5.2h (faltan US4 y US5) · **Desviación:** pendiente Sprint 2

☑ cerrada · ◐ en curso · ☐ sin empezar

> La cátedra pide comparar estimación vs. tiempo real por US para alimentar la estimación del
> Sprint 2 con IA. Completar la columna **Real** al cerrar cada US.

---

## Preliminar — andamiaje del repositorio ✓

- [x] `compose.yaml` + `Makefile` con `podman run` (sin quadlets, corre en Windows con Docker Desktop)
- [x] `.env.example` (configuración documentada, sin secretos versionados)
- [x] `.gitignore`
- [x] Módulo Go inicializado con `chi`, `pgx`, `golang-migrate`
- [x] `AGENTS.md` con convenciones, clean code / DRY y registro de IA
- [x] Este to-do list

---

## US1 — Selección de stack y configuración del entorno

*Est. 6h · Real: 1.5h · SP 5 · Valor 8*

- [x] Comparar alternativas de frontend, backend, base de datos e infraestructura cloud
- [x] Definir el framework de desarrollo móvil (nativo / híbrido / multiplataforma)
- [x] Documentar el stack elegido con justificación por decisión → `docs/sprint1/01-stack.md`
- [x] Repositorio creado y accesible para el equipo
- [x] Entornos de desarrollo configurados y operativos (`make db-up` verificado, Postgres 18.6)

**Decisiones ya tomadas en la sesión de diseño** (documentar el porqué en el doc):

- Backend Go 1.27 + `chi` v5 — elegido por ser `net/http` nativo (handlers `http.Handler`,
  middleware estándar) frente a gin (contexto propio, más magia) y `http.ServeMux` (cero
  deps pero sin composición de middleware).
- PostgreSQL 18 + `pgx` v5 + `sqlc` — dominio altamente relacional; `sqlc` da tipado en tiempo
  de compilación sin ocultar el SQL.
- Frontend Expo + expo-router + expo-secure-store — una sola base de código para móvil y web
  (`expo start --web`), cubre el requisito de app móvil + sistema web sin duplicar UI.
- Auth: sesiones opacas del lado del servidor (hash del token en la BD) en vez de JWT puro,
  porque US6 exige desactivar usuarios y un JWT no se puede revocar antes de su expiración.
- Contenedores con `podman`, compose opcional.

**Criterios de éxito**

- [x] El stack queda documentado con justificación por decisión
- [x] Repositorios accesibles para el equipo
- [x] Entornos configurados y operativos

**Decisión confirmada:** `expo-router` como router de la app; US4 y US5 se implementaron sobre
él. Queda pendiente completar en el informe final la matriz ponderada de frameworks de backend.

---

## US2 — Diseño del modelo de datos

*Est. 9h · Real: 3.5h · SP 5 · Valor 8*

- [x] Identificar entidades del dominio y sus atributos
- [x] Identificar relaciones y dependencias
- [x] Justificar relacional vs. NoSQL frente al stack de US1
- [x] Documentar el modelo final → `docs/sprint1/02-modelo-de-datos.md` (con ERD mermaid)
- [x] Migraciones solo DDL en `backend/migrations/`: `000001_core` (6 tablas)
- [x] Datos de referencia en `backend/seeds/` (idempotentes, fuera de `schema_migrations`)
- [x] Queries en `backend/sql/` + `make sqlc` (sqlc v1.31.1 lee el esquema de las migraciones)
- [x] Tests de integración contra Postgres real

**Entidades del dominio (borrador inicial)**

`users`, `teams` (escuderías), `drivers`, `categories` (F1, F2, F3, Academy), `seasons`,
`events` (calendario: carreras, pruebas de neumáticos, controles técnicos), `score_entries`,
`technical_controls`, `sanctions`, `notifications_ack`, `sessions`, `login_attempts`.

**Criterios de éxito**

- [x] Tipo de BD justificado según el dominio
- [x] Entidades definidas y relacionadas
- [x] El modelo escala a nuevas categorías/escuderías sin cambios estructurales

**Decisiones tomadas**

- PK `bigint GENERATED ALWAYS AS IDENTITY`; `ALWAYS` impide asignar ids a mano y desincronizar
  la secuencia.
- Roles y estados con `CHECK`, no `ENUM`: agregar un valor es una migración aditiva.
- Sin cuentas para el público: solo `fia_admin` y `team_admin`, el público lee sin autenticarse.
- Conflictos entre ramas: las migraciones se fusionan, los datos se regeneran con `make db-reset`.

**Pendiente para el Sprint 2:** 10 tablas más del dominio (`drivers`, `event_types`, `events`,
`score_entries`, `score_acknowledgements`, `technical_controls`, `technical_control_results`,
`sanctions`, `sanction_acknowledgements`, `notifications`). Ya están diseñadas en el doc.

**Revisión del esquema:** `000001_core` se editó antes de desplegarse (en `users`, `is_active`
pasó a ser una columna generada a partir de `deactivated_at`; se quitaron un índice y una
restricción redundantes). Cada integrante tiene que correr `make db-reset`.

**Entidades del dominio implementadas (corte Sprint 1)**

`categories`, `seasons`, `teams`, `users`, `sessions`, `login_attempts`

---

## US3 — Arquitectura, roles y seguridad

*Est. 9h · Real: 0.2h · SP 5 · Valor 8*

- [x] Definir arquitectura general: capas, servicios, módulos, comunicación
- [x] Definir el modelo de permisos por rol (FIA / escudería / público)
- [x] Establecer estándares de seguridad: autenticación, cifrado, políticas
- [x] Documentar → `docs/sprint1/03-arquitectura-roles-seguridad.md`

**Capas:** `handler → service → repo`. El handler no tiene lógica de negocio ni SQL.

**Estándares de seguridad documentados**

- [x] Contraseñas con Argon2id (parámetros por hardware, salt por usuario, coste versionado en el hash)
- [x] Sesiones opacas: token 32 bytes aleatorios, en la BD solo el SHA-256
- [x] Sesión nueva en cada login y revocación de todas las sesiones al cambiar la cuenta (US6);
      la rotación por request se descartó porque dos peticiones simultáneas se invalidarían
      mutuamente
- [x] Expiración por inactividad y absoluta
- [x] Revocación inmediata al desactivar un usuario, en la misma transacción
- [x] Cookie `httpOnly; Secure; SameSite=Strict` en web; `expo-secure-store` en móvil

**Decisiones que se apartan del enunciado por criterio propio**

- Sin JWT: la revocación que pide el enunciado exige consultar la base en cada request
- El público no es un valor de `users.role`: es la ausencia de sesión
- El `team_id` del `team_admin` sale de la sesión, nunca del cuerpo de la petición

**Hallazgo para la estimación del Sprint 2:** las historias de diseño (US1, US2, US3) se resolvieron
muy por debajo de la estimación, y la desviación viene de la redacción, no del diseño. La matriz de
permisos de la US3 son 13 filas y una tabla; el trabajo real fue decidir y justificar, y dejar por
escrito lo que se descartó y por qué. Para el Sprint 2 conviene distinguir dentro de la estimación
entre "escribir el documento" y "elegir la solución".
- [x] Bloqueo por intentos fallidos de login (implementado en US5)
- [x] RBAC en middleware: ningún endpoint sensible sin chequeo de rol (implementado en US5)
- [x] Auditoría de eventos de autenticación: cada intento queda en `login_attempts`
- [ ] Purga periódica de `login_attempts` (no urgente con el volumen actual)

**Criterios de éxito**

- [x] Arquitectura de autenticación y roles documentada
- [x] La matriz de permisos dice qué puede hacer cada rol sobre cada componente

---

## US4 — Diseño de interfaz por rol

*Est. 8h · Real: — · SP 5 · Valor 8*

- [x] Definir alcance de la interfaz de cada rol (pantallas pública, FIA y escudería)
- [x] Sistema de diseño común: colores, tipografías, componentes reutilizables (`app/src/theme`,
      `app/src/components`)
- [x] Versión inicial de cada interfaz (con datos mockeados)
- [ ] Documentar funcionalidades por rol → `docs/sprint1/04-interfaces.md`

**Criterios de éxito**

- [ ] Interfaz clara, intuitiva, fácil de usar (a validar en la demo)
- [x] La versión inicial muestra las funcionalidades mockeadas de cada rol

---

## US5 — Login de usuarios

*Est. 9h · Real: — · SP 3 · Valor 13*

- [x] Interfaz de login con usuario y contraseña
- [x] Validación de datos ingresados (en la app y, de nuevo, en el backend)
- [x] Identificación del rol del usuario
- [x] Redirección a la interfaz del rol correspondiente (US4)
- [x] Notificación de ingreso incorrecto (datos inválidos o credenciales incorrectas)

**Backend:** hasher Argon2id, `POST /auth/login`, `POST /auth/logout`, `GET /auth/me`, bloqueo
por intentos, cookie `httpOnly` en web y token en `expo-secure-store` en móvil.

**Criterios de éxito**

- [x] Acceso permitido con datos correctos
- [x] Se cumplen los estándares de US3
- [x] Interfaz clara y fácil de usar
- [x] La plataforma identifica correctamente el rol al iniciar sesión

**Pendiente:** cargar la columna **Real** de US4 y US5 en la tabla de estado.

---

## US6 — Gestión de usuarios

*Est. 12h · Real: — · SP 8 · Valor 21*

- [x] Alta de usuarios
- [x] Modificación de datos de un usuario existente
- [x] Baja (eliminación o desactivación) de un usuario
- [x] Listado y búsqueda de usuarios (solo administrador FIA)
- [x] Asignación de un rol válido
- [ ] Pantalla de gestión de usuarios de la app conectada a la API

**Backend:** `GET /users` (búsqueda por usuario o email, filtros por rol y estado),
`POST /users`, `GET /users/{id}`, `PATCH /users/{id}`, `POST /users/{id}/deactivate`,
`POST /users/{id}/reactivate`, todas solo para `fia_admin`; `GET /teams` público para el
formulario de asignación. Un cambio de rol, escudería o contraseña y la baja revocan las sesiones
del usuario en la misma transacción. Un administrador no puede desactivarse ni cambiar su propio
rol, así que siempre queda al menos uno activo.

**Criterios de éxito**

- [ ] El admin FIA crea, modifica y elimina usuarios, les asigna un rol válido, los busca y lista
- [x] Los cambios se reflejan de manera consistente (sesiones revocadas en la misma transacción)
- [x] Un usuario sin permisos de rol no puede acceder ni usar la funcionalidad (`401` sin sesión,
      `403` para `team_admin`, probado)

---

## Pendientes de infraestructura

- [x] `README.md` con instrucciones de arranque
- [x] `docs/prompts-ia.md` — registro de prompts usados con IA (exigido por la cátedra)
- [ ] Decidir el deploy de la app web: en Vercel la cookie de sesión no llega a la API (ver la
      deuda técnica de `03-arquitectura-roles-seguridad.md`)
- [ ] Alojar el backend con Postgres (`make migrate` + `make seed` como pasos del deploy)
- [x] Correr `make check` con `DATABASE_URL`: los 16 tests de integración pasan contra Postgres 18
- [ ] Registrar en `docs/prompts-ia.md` los prompts usados en US4, US5 y el deploy en Vercel