# Sprint 1 — Lista de tareas

Comisión **Grupo Rojo Ferrari** (Sevenants Antonio, Tourn Felipe) — Enunciado 1 (FIA).

Entrega: **lunes 5/10** · Demo: **martes 6/10**

Leyendo de arriba hacia abajo. Cada US se cierra con sus criterios de éxito verificados,
`make check` en verde y **un commit atómico**.

## Estado

| US | Nombre | SP | Valor | Est. | Real | Estado |
|----|--------|----|-------|-------|-----|--------|
| US1 | Selección de stack y configuración del entorno | 5 | 8 | 6h | 1.5h | ☑ |
| US2 | Diseño del modelo de datos | 5 | 8 | 9h | — | ☐ |
| US3 | Arquitectura, roles y seguridad | 5 | 8 | 9h | — | ☐ |
| US4 | Diseño de interfaz por rol | 5 | 8 | 8h | — | ☐ |
| US5 | Login de usuarios | 3 | 13 | 9h | — | ☐ |
| US6 | Gestión de usuarios | 8 | 21 | 12h | — | ☐ |

**Total estimado:** 53h · **Total real:** 1.5h · **Desviación:** pendiente Sprint 2

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

**Decisión pendiente de confirmación del equipo:** `expo-router` como router de la app.
Alternativa con la misma cantidad de dependencias: `react-navigation` nativo. La decisión de
Go + chi ya está firme. Pendiente también completar en el informe final la matriz ponderada
de frameworks de backend.

---

## US2 — Diseño del modelo de datos

*Est. 9h · Real: — · SP 5 · Valor 8*

- [ ] Identificar entidades del dominio y sus atributos
- [ ] Identificar relaciones y dependencias
- [ ] Justificar relacional vs. NoSQL frente al stack de US1
- [ ] Documentar el modelo final → `docs/sprint1/02-modelo-de-datos.md` (con ERD mermaid)
- [ ] Migraciones en `backend/migrations/`
- [ ] Queries en `backend/sql/` + `make sqlc`
- [ ] Tests del repo contra Postgres real

**Entidades del dominio (borrador inicial)**

`users`, `teams` (escuderías), `drivers`, `categories` (F1, F2, F3, Academy), `seasons`,
`events` (calendario: carreras, pruebas de neumáticos, controles técnicos), `score_entries`,
`technical_controls`, `sanctions`, `notifications_ack`, `sessions`, `login_attempts`.

**Criterios de éxito**

- [ ] Tipo de BD justificado según el dominio
- [ ] Entidades definidas y relacionadas
- [ ] El modelo escala a nuevas categorías/escuderías sin cambios estructurales

---

## US3 — Arquitectura, roles y seguridad

*Est. 9h · Real: — · SP 5 · Valor 8*

- [ ] Definir arquitectura general: capas, servicios, módulos, comunicación
- [ ] Definir el modelo de permisos por rol (FIA / escudería / público)
- [ ] Establecer estándares de seguridad: autenticación, cifrado, políticas
- [ ] Documentar → `docs/sprint1/03-arquitectura-roles-seguridad.md`

**Capas:** `handler → service → repo`. El handler no tiene lógica de negocio ni SQL.

**Estándares de seguridad a documentar**

- [ ] Contraseñas con Argon2id (parámetros por hardware, salt por usuario)
- [ ] Sesiones opacas: token 32 bytes aleatorios, en la BD solo el SHA-256; rotación en cada uso
- [ ] Expiración por inactividad y absoluta
- [ ] Revocación inmediata al desactivar o eliminar un usuario
- [ ] Cookie `httpOnly; Secure; SameSite=Strict` en web; `expo-secure-store` en móvil
- [ ] Rate limit y bloqueo por intentos fallidos de login
- [ ] RBAC en middleware: ningún endpoint sensible sin chequeo de rol
- [ ] Auditoría de eventos de autenticación

**Criterios de éxito**

- [ ] Arquitectura de autenticación y roles documentada
- [ ] La matriz de permisos dice qué puede hacer cada rol sobre cada componente

---

## US4 — Diseño de interfaz por rol

*Est. 8h · Real: — · SP 5 · Valor 8*

- [ ] Definir alcance de la interfaz de cada rol
- [ ] Sistema de diseño común: colores, tipografías, componentes reutilizables
- [ ] Versión inicial de cada interfaz (con datos mockeados)
- [ ] Documentar funcionalidades por rol → `docs/sprint1/04-interfaces.md`

**Criterios de éxito**

- [ ] Interfaz clara, intuitiva, fácil de usar
- [ ] La versión inicial muestra las funcionalidades mockeadas de cada rol

---

## US5 — Login de usuarios

*Est. 9h · Real: — · SP 3 · Valor 13*

- [ ] Interfaz de login con usuario y contraseña
- [ ] Validación de datos ingresados
- [ ] Identificación del rol del usuario
- [ ] Redirección a la interfaz del rol correspondiente (US4)
- [ ] Notificación de ingreso incorrecto (datos inválidos o credenciales incorrectas)

**Backend:** hasher Argon2id, `POST /auth/login`, `POST /auth/logout`, `GET /auth/me`,
rate limit + lockout, redirección por rol en el cliente.

**Criterios de éxito**

- [ ] Acceso permitido con datos correctos
- [ ] Se cumplen los estándares de US3
- [ ] Interfaz clara y fácil de usar
- [ ] La plataforma identifica correctamente el rol al iniciar sesión

---

## US6 — Gestión de usuarios

*Est. 12h · Real: — · SP 8 · Valor 21*

- [ ] Alta de usuarios
- [ ] Modificación de datos de un usuario existente
- [ ] Baja (eliminación o desactivación) de un usuario
- [ ] Listado y búsqueda de usuarios (solo administrador FIA)
- [ ] Asignación de un rol válido

**Criterios de éxito**

- [ ] El admin FIA crea, modifica y elimina usuarios, les asigna un rol válido, los busca y lista
- [ ] Los cambios se reflejan de manera consistente
- [ ] Un usuario sin permisos de rol no puede acceder ni usar la funcionalidad

---

## Pendientes de infraestructura

- [ ] `README.md` con instrucciones de arranque
- [ ] `docs/prompts-ia.md` — registro de prompts usados con IA (exigido por la cátedra)