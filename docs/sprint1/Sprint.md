# Sprint 1 — Lista de tareas

## Estado

| US | Nombre | SP | Valor | Est. | Tiempo real registrado | Estado |
|----|--------|----|-------|-------|------------------------|--------|
| US1 | Selección de stack y configuración del entorno | 5 | 8 | 6h | 1.5h | ☑ |
| US2 | Diseño del modelo de datos | 5 | 8 | 9h | 3.5h | ☑ |
| US3 | Arquitectura, roles y seguridad | 5 | 8 | 9h | 0.2h | ☑ |
| US4 | Diseño de interfaz por rol | 5 | 8 | 8h | 1.1h | ☑ |
| US5 | Login de usuarios | 3 | 13 | 9h | 1.0h | ☑ |
| US6 | Gestión de usuarios | 8 | 21 | 12h | 1.2h | ☑ |

**Total estimado:** 53h · **Total real registrado:** 8.5h · **Desviación:** −84 %  
El tiempo registrado representa aproximadamente el **16 % del tiempo inicialmente estimado**.

Las horas reales de US4, US5 y US6 se obtuvieron a partir de las marcas de tiempo de los commits: desde el último commit correspondiente a la historia anterior hasta el último commit de la propia historia.

Cuando varias historias se trabajaron durante una misma sesión, como ocurrió con US4 y US5, el tiempo se distribuyó según el momento en que se realizó el commit correspondiente a cada una.

Esta medición debe considerarse **aproximada**, ya que refleja principalmente el tiempo observable de implementación y versionado. No incluye necesariamente todo el tiempo dedicado previamente a análisis, discusión de alternativas, lectura, diseño conceptual ni pruebas manuales realizadas fuera de las sesiones registradas mediante commits.

☑ cerrada · ◐ en curso · ☐ sin empezar

> La cátedra solicita comparar la estimación inicial con el tiempo real registrado por US para utilizar esa información como entrada en la estimación del Sprint 2 con asistencia de IA. La columna **Tiempo real registrado** debe completarse al cerrar cada historia.

---

## Preliminar — Andamiaje del repositorio ✓

- [x] `compose.yaml` + `Makefile` con `podman run` y compatibilidad con Docker Desktop en Windows
- [x] `.env.example` con configuración documentada y sin secretos versionados
- [x] `.gitignore`
- [x] Módulo Go inicializado con `chi`, `pgx` y `golang-migrate`
- [x] `AGENTS.md` con convenciones de desarrollo, principios de Clean Code / DRY y registro del uso de IA
- [x] Lista de tareas del Sprint

---

# US1 — Selección de stack y configuración del entorno

*Est. 6h · Tiempo real registrado: 1.5h · SP 5 · Valor 8*

- [x] Comparar alternativas de frontend, backend, base de datos e infraestructura cloud
- [x] Definir el framework de desarrollo móvil: nativo, híbrido o multiplataforma
- [x] Documentar el stack seleccionado y justificar cada decisión → `docs/sprint1/01-stack.md`
- [x] Crear y dejar accesible el repositorio para el equipo
- [x] Configurar y verificar los entornos de desarrollo (`make db-up`, PostgreSQL 18.6)

## Decisiones tomadas

### Backend

Se seleccionó **Go 1.27 + `chi` v5**.

`chi` fue elegido porque mantiene una integración directa con `net/http`, utilizando handlers compatibles con `http.Handler` y middleware estándar.

Se lo prefirió frente a:

- **Gin**, que introduce un contexto propio y una abstracción mayor sobre `net/http`.
- **`http.ServeMux`**, que elimina dependencias externas, pero ofrece menor comodidad para composición de rutas y middleware.

### Base de datos

Se seleccionó:

- PostgreSQL 18
- `pgx` v5
- `sqlc`

El dominio posee una estructura altamente relacional, por lo que PostgreSQL permite representar adecuadamente relaciones, restricciones y consistencia entre entidades.

`sqlc` permite generar código Go fuertemente tipado a partir de SQL explícito, manteniendo control sobre las consultas sin introducir una capa ORM que oculte su funcionamiento.

### Frontend

Se seleccionó:

- Expo
- `expo-router`
- `expo-secure-store`

Expo permite mantener una única base de código para aplicaciones móviles y web mediante `expo start --web`, cubriendo los requisitos de aplicación móvil y sistema web sin duplicar la interfaz.

### Autenticación

Se seleccionaron **sesiones opacas administradas por el servidor** en lugar de JWT puro.

El token de sesión se genera aleatoriamente y únicamente su hash es almacenado en la base de datos.

Esta decisión permite revocar inmediatamente las sesiones ante eventos como desactivación de un usuario, modificación del rol, modificación de la escudería y cambio de contraseña.

JWT puro dificultaría la revocación inmediata antes de su expiración sin introducir mecanismos adicionales de invalidación.

### Infraestructura

Se decidió utilizar contenedores mediante `podman`, manteniendo compatibilidad con Compose de forma opcional.

## Criterios de éxito

- [x] El stack queda documentado con una justificación para cada decisión
- [x] Los repositorios son accesibles para el equipo
- [x] Los entornos de desarrollo se encuentran configurados y operativos

**Decisión confirmada:** `expo-router` se utiliza como router principal de la aplicación. US4 y US5 fueron implementadas sobre esta arquitectura.

**Pendiente para el informe final:** completar la matriz ponderada utilizada para comparar frameworks de backend.

---

# US2 — Diseño del modelo de datos

*Est. 9h · Tiempo real registrado: 3.5h · SP 5 · Valor 8*

- [x] Identificar entidades del dominio y sus atributos
- [x] Identificar relaciones y dependencias
- [x] Justificar el uso de una base relacional frente a alternativas NoSQL
- [x] Documentar el modelo final → `docs/sprint1/02-modelo-de-datos.md`
- [x] Incluir ERD mediante Mermaid
- [x] Crear migraciones DDL en `backend/migrations/`
- [x] Implementar `000001_core` con las seis tablas correspondientes al alcance del Sprint 1
- [x] Separar los datos de referencia en `backend/seeds/`
- [x] Mantener los seeds idempotentes y fuera de `schema_migrations`
- [x] Definir queries en `backend/sql/`
- [x] Ejecutar `make sqlc`
- [x] Verificar que `sqlc` v1.31.1 procese correctamente el esquema de las migraciones
- [x] Implementar tests de integración contra PostgreSQL real

## Entidades identificadas durante el diseño general

- `users`
- `teams`
- `drivers`
- `categories`
- `seasons`
- `events`
- `score_entries`
- `technical_controls`
- `sanctions`
- `notifications`
- `sessions`
- `login_attempts`

El modelo conceptual contempla además otras entidades auxiliares relacionadas con resultados, reconocimientos y notificaciones.

## Criterios de éxito

- [x] El tipo de base de datos se encuentra justificado según las características del dominio
- [x] Las entidades se encuentran identificadas y relacionadas
- [x] El modelo permite incorporar nuevas categorías y escuderías sin realizar cambios estructurales relevantes

## Decisiones tomadas

### Claves primarias

Se utilizan:

`bigint GENERATED ALWAYS AS IDENTITY`

La opción `ALWAYS` impide asignar identificadores manualmente de forma accidental y evita desincronizaciones entre los valores insertados y la secuencia interna.

### Roles y estados

Los roles y estados se representan mediante restricciones `CHECK` en lugar de `ENUM`.

Agregar un nuevo valor requiere únicamente una migración aditiva sobre la restricción, evitando una dependencia fuerte de tipos enumerados específicos de PostgreSQL.

### Usuarios públicos

No existen cuentas para usuarios públicos.

Los únicos roles persistidos son:

- `fia_admin`
- `team_admin`

El público se representa como la ausencia de una sesión autenticada.

### Migraciones y ramas

Ante conflictos entre ramas:

- las migraciones se integran manualmente;
- los datos de prueba y referencia se regeneran mediante `make db-reset`.

## Revisión del esquema

Antes de desplegarse, la migración `000001_core` fue ajustada.

En `users`:

- `is_active` pasó a calcularse como una columna generada a partir de `deactivated_at`;
- se eliminaron un índice y una restricción redundantes.

De esta forma se evita mantener dos representaciones independientes del mismo estado.

Cada integrante del equipo debe ejecutar:

`make db-reset`

para regenerar correctamente el esquema local.

## Entidades implementadas durante Sprint 1

- `categories`
- `seasons`
- `teams`
- `users`
- `sessions`
- `login_attempts`

## Diseño preparado para Sprint 2

Las siguientes tablas se encuentran **diseñadas conceptualmente y documentadas, pero no implementadas durante Sprint 1**:

- `drivers`
- `event_types`
- `events`
- `score_entries`
- `score_acknowledgements`
- `technical_controls`
- `technical_control_results`
- `sanctions`
- `sanction_acknowledgements`
- `notifications`

Su implementación queda fuera del alcance del Sprint 1.

---

# US3 — Arquitectura, roles y seguridad

*Est. 9h · Tiempo real registrado: 0.2h · SP 5 · Valor 8*

- [x] Definir arquitectura general
- [x] Definir capas, servicios, módulos y comunicación
- [x] Definir el modelo de permisos para FIA, escuderías y público
- [x] Establecer estándares de autenticación y seguridad
- [x] Documentar las decisiones → `docs/sprint1/03-arquitectura-roles-seguridad.md`

## Arquitectura

Se adopta una separación en tres capas principales:

`handler → service → repo`

### Handler

Responsable de:

- Recibir requests HTTP
- Validar aspectos de transporte
- Convertir datos de entrada
- Delegar la ejecución al service
- Generar la respuesta HTTP

No contiene lógica de negocio ni consultas SQL.

### Service

Contiene:

- Reglas de negocio
- Autorización de operaciones
- Coordinación entre repositorios
- Validaciones propias del dominio

### Repository

Responsable del acceso a los datos persistidos.

---

## Estándares de seguridad

- [x] Contraseñas almacenadas mediante Argon2id
- [x] Parámetros de Argon2id ajustables según hardware
- [x] Salt independiente para cada contraseña
- [x] Coste de hashing almacenado junto con el hash
- [x] Sesiones opacas mediante tokens aleatorios de 32 bytes
- [x] Únicamente SHA-256 del token persistido en la base de datos
- [x] Nueva sesión generada en cada login
- [x] Revocación de sesiones ante cambios sensibles
- [x] Expiración por inactividad
- [x] Expiración absoluta
- [x] Revocación inmediata al desactivar un usuario
- [x] Revocación ejecutada dentro de la misma transacción que modifica la cuenta
- [x] Cookie `httpOnly`, `Secure` y `SameSite=Strict` para web
- [x] `expo-secure-store` para almacenamiento seguro del token en móvil
- [x] Bloqueo temporal ante múltiples intentos fallidos de login
- [x] RBAC implementado mediante middleware
- [x] Auditoría de intentos de autenticación mediante `login_attempts`

La rotación del token en cada request fue descartada, ya que dos peticiones concurrentes podrían invalidarse mutuamente.

---

## Decisiones que se apartan del enunciado

### Sin JWT puro

No se utiliza JWT puro debido a la necesidad de revocación inmediata ante cambios en la cuenta.

La solución seleccionada prioriza sesiones administradas en servidor.

### Público sin rol persistido

El público no se representa mediante un valor de `users.role`.

Se considera público a cualquier cliente sin una sesión autenticada.

### Escudería derivada de la sesión

El `team_id` de un `team_admin` se obtiene a partir de la sesión autenticada y nunca se acepta desde el cuerpo de la petición.

Esto evita que un administrador de escudería intente operar sobre datos pertenecientes a otra organización.

---

## Interpretación del tiempo real registrado

El valor de **0.2h** no representa el tiempo total necesario para definir desde cero la arquitectura y las políticas de seguridad.

Gran parte de estas decisiones se habían definido incrementalmente durante US1 y US2.

El tiempo registrado en US3 representa principalmente el período observable correspondiente a su **consolidación, documentación y formalización** como una historia independiente.

Este caso demuestra que las historias de diseño pueden contener trabajo intelectual realizado antes de que la historia sea cerrada formalmente, por lo que el tiempo basado únicamente en commits debe interpretarse con cautela.

---

## Criterios de éxito

- [x] La arquitectura de autenticación y roles se encuentra documentada
- [x] La matriz de permisos especifica las operaciones permitidas para cada rol
- [x] Los endpoints sensibles se encuentran protegidos mediante autorización por rol

## Deuda técnica

- [ ] Purga periódica de `login_attempts`

No se considera urgente dado el volumen actual de datos.

---

# US4 — Diseño de interfaz por rol

*Est. 8h · Tiempo real registrado: 1.1h · SP 5 · Valor 8*

- [x] Definir el alcance funcional de la interfaz pública
- [x] Definir el alcance funcional del panel FIA
- [x] Definir el alcance funcional del panel de escudería
- [x] Crear un sistema de diseño común
- [x] Definir colores, tipografías y componentes reutilizables
- [x] Centralizar el sistema visual en `app/src/theme` y `app/src/components`
- [x] Implementar una primera versión de las interfaces utilizando datos mockeados
- [x] Documentar funcionalidades disponibles por rol → `docs/sprint1/04-interfaces.md`

## Criterios de éxito

- [x] La interfaz utiliza un sistema de diseño consistente entre pantallas
- [x] La navegación permite acceder claramente a las funcionalidades correspondientes a cada rol
- [x] La versión inicial representa mediante datos mockeados las funcionalidades previstas
- [x] Las interfaces se encuentran separadas de acuerdo con los permisos definidos para público, FIA y escudería

La percepción de facilidad de uso e intuición se validará adicionalmente durante la demostración del Sprint.

---

# US5 — Login de usuarios

*Est. 9h · Tiempo real registrado: 1.0h · SP 3 · Valor 13*

- [x] Implementar interfaz de login con usuario y contraseña
- [x] Validar datos ingresados en la aplicación
- [x] Validar nuevamente los datos en backend
- [x] Identificar el rol del usuario autenticado
- [x] Redireccionar al usuario a la interfaz correspondiente según su rol
- [x] Mostrar un mensaje ante credenciales inválidas

## Backend

Se implementaron:

- `POST /auth/login`
- `POST /auth/logout`
- `GET /auth/me`

Además:

- Argon2id para validación de contraseñas
- bloqueo ante intentos fallidos
- registro de intentos en `login_attempts`
- cookie `httpOnly` para web
- token persistido mediante `expo-secure-store` en móvil

## Criterios de éxito

- [x] El acceso se permite con credenciales válidas
- [x] El acceso se rechaza correctamente con credenciales inválidas
- [x] Se respetan los estándares definidos en US3
- [x] La plataforma identifica correctamente el rol del usuario
- [x] La navegación posterior al login corresponde al rol autenticado

---

# US6 — Gestión de usuarios

*Est. 12h · Tiempo real registrado: 1.2h · SP 8 · Valor 21*

- [x] Alta de usuarios
- [x] Modificación de usuarios existentes
- [x] Baja mediante desactivación
- [x] Reactivación de usuarios
- [x] Listado de usuarios
- [x] Búsqueda de usuarios
- [x] Filtrado por rol
- [x] Filtrado por estado
- [x] Asignación de roles válidos
- [x] Asignación de escudería cuando corresponda
- [x] Pantalla de gestión conectada a la API

## Backend

Se implementaron:

- `GET /users`
- `POST /users`
- `GET /users/{id}`
- `PATCH /users/{id}`
- `POST /users/{id}/deactivate`
- `POST /users/{id}/reactivate`
- `GET /teams`

Los endpoints de administración de usuarios son accesibles exclusivamente para `fia_admin`.

`GET /teams` se mantiene público porque es utilizado por formularios y componentes que requieren información de las escuderías.

## Reglas de negocio

Un cambio de:

- Rol
- Escudería;
- Contraseña;
- Estado de activación

Provoca la revocación de las sesiones correspondientes dentro de la misma transacción.

Además:

- Un administrador FIA no puede desactivarse a sí mismo
- Un administrador FIA no puede modificar su propio rol
- Se evita de esta manera dejar al sistema sin al menos un administrador activo

## Perfil propio

Se implementó:

`POST /account/password`

Disponible para cualquier usuario autenticado.

El endpoint:

- Solicita la contraseña actual
- Aplica la política vigente de contraseñas
- Actualiza la credencial
- Revoca el resto de las sesiones del usuario
- Mantiene activa únicamente la sesión que realizó el cambio

## Aplicación

La gestión de usuarios se encuentra implementada en:

`app/fia/users/`

Incluye:

- Listado
- Búsqueda
- Filtros
- Alta
- Edición
- Desactivación
- Reactivación
- Confirmaciones antes de acciones sensibles

La implementación utiliza:

- `src/api/users.ts`
- hooks definidos en `src/data/accounts.ts`

El panel FIA muestra además la cantidad real de cuentas activas.

## Criterios de éxito

- [x] El administrador FIA puede crear usuarios
- [x] El administrador FIA puede modificar usuarios
- [x] El administrador FIA puede desactivar y reactivar usuarios
- [x] El administrador FIA puede asignar roles válidos
- [x] El administrador FIA puede buscar, filtrar y listar usuarios
- [x] Los cambios sensibles revocan sesiones de manera consistente
- [x] Un usuario sin sesión recibe `401`
- [x] Un `team_admin` intentando utilizar funcionalidades FIA recibe `403`

---

# Pendientes de infraestructura

- [x] Crear `README.md` con instrucciones de puesta en marcha
- [x] Crear `docs/prompts-ia.md` con registro de prompts utilizados con IA
- [ ] Definir estrategia de despliegue de la aplicación web
- [ ] Resolver el problema de cookies entre el frontend desplegado en Vercel y la API
- [ ] Alojar backend y PostgreSQL
- [ ] Ejecutar `make migrate` durante el despliegue
- [ ] Ejecutar `make seed` según corresponda durante la inicialización del entorno
- [x] Ejecutar `make check` utilizando `DATABASE_URL`
- [x] Verificar los 16 tests de integración contra PostgreSQL 18
- [ ] Registrar en `docs/prompts-ia.md` los prompts utilizados durante US4, US5 y las pruebas de despliegue en Vercel

---

# Backlog para Sprint 2

Se registran aquí funcionalidades y mejoras identificadas durante Sprint 1 que quedan fuera de su alcance.

## Foto de perfil propia

Permitir que cada usuario pueda cargar, modificar y eliminar una foto de perfil desde "Mi cuenta".

Actualmente la aplicación muestra iniciales, excepto en las cuentas demo que utilizan el logo correspondiente a su organización.

La funcionalidad requiere:

- tabla `user_avatars`;
- `PUT /account/avatar`;
- `DELETE /account/avatar`;
- `GET /users/{id}/avatar`;
- validación del contenido real del archivo y no únicamente de su extensión;
- límite de tamaño;
- re-codificación de la imagen;
- reducción a aproximadamente 256 px;
- evitar servir directamente el archivo subido por el usuario;
- integración de `expo-image-picker`.

Dependencias adicionales previstas:

- `golang.org/x/image`
- `expo-image-picker`

**Estimación inicial:** 3h incluyendo tests.

## Logos de escuderías

Agregar los logos de:

- McLaren
- ART
- PREMA
- Campos
- Ferrari Driver Academy

Los archivos deberán utilizar como nombre el código correspondiente definido en los seeds y almacenarse en:

`app/assets/logos/`

## Seguridad

- Purga periódica de `login_attempts`
- Evaluar bloqueo por IP utilizando la información registrada en `login_attempts`

## Documentación

- Completar la matriz ponderada de frameworks de backend prometida durante US1

## Deploy

Definir la estrategia de despliegue para que la aplicación web pueda utilizar correctamente la cookie de sesión contra la API.

---

# Análisis de estimación del Sprint 1

El Sprint presenta una diferencia considerable entre las estimaciones iniciales y el tiempo real registrado.

No obstante, no resulta correcto interpretar esta desviación como evidencia de que todas las historias futuras deben estimarse un **84 % por debajo**.

Los datos muestran comportamientos diferentes según el tipo de tarea.

## Historias de diseño y documentación

US1, US2 y US3 fueron significativamente sobreestimadas.

Parte de esta diferencia se explica porque:

- Varias decisiones se tomaron en paralelo
- Algunas decisiones de historias posteriores surgieron durante historias anteriores
- El tiempo basado en commits no captura todo el proceso de análisis
- La documentación resultó más rápida de producir que lo previsto

Para Sprint 2 conviene distinguir explícitamente:

- Tiempo de análisis
- Tiempo de decisión
- Tiempo de documentación
- Tiempo de implementación

## Historias de implementación

US5 y US6 también tuvieron tiempos inferiores a los esperados.

Una causa relevante fue que, al momento de implementarlas, ya se encontraba disponible:

- Una arquitectura definida
- Una base de datos preparada
- Un sistema de capa
- Infraestructura de autenticación
- Convenciones de desarrollo
- Componentes reutilizables

Esto redujo considerablemente el coste marginal de agregar nuevas funcionalidades.

## Story Points y horas

Los Story Points no deben recalcularse utilizando directamente las horas reales.

Los SP representan principalmente:

- Complejidad relativa
- Incertidumbre
- Cantidad de trabajo
- Riesgos
- Dependencias

Por este motivo, que una historia de 8 SP haya requerido 1.2h no significa que todas las historias futuras de 8 SP deban estimarse en 1.2h.

Los datos obtenidos durante este Sprint todavía no constituyen una muestra suficiente para establecer una conversión estable entre Story Points y horas.

---

# Lecciones aprendidas del Sprint 1

1. **Las historias de diseño y documentación fueron sobreestimadas.**  
   El trabajo de análisis fue menor al previsto y varias decisiones se resolvieron incrementalmente entre historias.

2. **Definir correctamente la arquitectura al comienzo redujo el esfuerzo de las historias posteriores.**  
   La reutilización de capas, autenticación, repositorios y componentes permitió implementar nuevas funcionalidades con menor coste.

3. **La medición basada en commits debe interpretarse como una aproximación.**  
   No representa necesariamente todo el tiempo dedicado a análisis, lectura, discusión y pruebas realizadas fuera de las sesiones de desarrollo registradas.

4. **Los Story Points no deben convertirse directamente a horas.**  
   Deben continuar utilizándose como medida de complejidad relativa, mientras que las horas se ajustarán utilizando la evidencia acumulada entre Sprints.

5. **El uso de IA permitió reducir principalmente tareas repetitivas.**  
   Se utilizó como asistencia para generación de scaffolding, consultas, tests, documentación y revisión, mientras que las decisiones arquitectónicas y de negocio fueron analizadas y validadas por el equipo.

6. **Para Sprint 2 conviene estimar de forma más granular.**  
   Cada historia debería separar, cuando corresponda análisis, diseño, implementación, integración, pruebas y documentación.

7. **No se aplicará directamente la desviación global de −84 % al Sprint 2.**  
   Las funcionalidades que incorporen nuevas integraciones, procesamiento de archivos, infraestructura, concurrencia o reglas de negocio desconocidas conservan un nivel de incertidumbre superior al observado en las historias de Sprint 1.
