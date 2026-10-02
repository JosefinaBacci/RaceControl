# AGENTS.md

Instrucciones para agentes de IA que trabajan en este repositorio.

## Contexto del proyecto

RaceControl es la plataforma para la FIA (Enunciado 1 de APS 2026): gestión de calendario,
escuderías, pilotos, controles técnicos, sanciones, puntajes, pruebas de neumáticos y
comunicación entre FIA y escuderías. Roles de usuario: administrador FIA, administrador de
escudería y público general.

Este repositorio contiene la **implementación** (backend + app) y la documentación técnica de
diseño. La comisión de Análisis y Management trabaja en otro repositorio.

## Stack (decidido en US1, no cambiar sin actualizar `docs/sprint1/01-stack.md`)

- **Backend:** Go 1.27, `net/http` + `chi` v5
- **Base de datos:** PostgreSQL 18, `pgx` v5, queries tipados con `sqlc`
- **Migraciones:** SQL embebido en el binario vía `golang-migrate`
- **App:** Expo + expo-router + expo-secure-store (una sola base de código para móvil y web)
- **Auth:** sesiones opacas del lado del servidor (hash del token en la BD), contraseñas con
  **Argon2id**, revocación inmediata

## Estructura

```
backend/
  cmd/api           API HTTP
  cmd/migrate       Runner de migraciones
  internal/config   Configuración desde el entorno
  internal/db       Pool de pgx, runner de migraciones, queries generadas por sqlc
  internal/db/dbgen Código generado por sqlc; nunca editar a mano
  internal/domain   Entidades y errores de dominio (sin dependencias de infraestructura)
  internal/auth     Hashing de contraseñas, middleware RBAC
  internal/sessions Creación, rotación y revocación de sesiones
  internal/users    Servicio de gestión de usuarios
  internal/transport Handlers, middlewares, routing (chi)
  migrations/       Archivos .sql numerados + embed.go
  sql/              Queries que sqlc convierte en Go
  sqlc.yaml         Configuración de sqlc
app/
  app/              Rutas (expo-router)
  src/api           Cliente HTTP
  src/components    Componentes reutilizables
  src/theme         Design system: colores, tipografía, espaciado
docs/
  sprint1/          Documentación de diseño por user story
  prompts-ia.md     Registro de prompts de IA (exigido por la cátedra)
```

## Reglas de trabajo

1. **Comentarios en inglés, y solo cuando el código no se explica solo.** Nada de comentarios
   que repitan la línea siguiente, nada de comentarios de tipo "TODO: hacer". Prefiere nombres
   explícitos y funciones cortas.
2. **Comentarios de documentación en handlers** solo cuando el contrato (status codes, errores)
   no sea evidente.
3. La documentación de cara al profesor va en **español**; los comentarios del código, en inglés.
4. **Commits atómicos:** un commit por user story o por cambio lógico coherente. Nunca mezcles
   refactors con funcionalidad en el mismo commit. El historial sirve como evidencia de avance
   del sprint.
5. Antes de dar por terminada cualquier tarea: `make check` (lint + typecheck + tests) en verde.
6. **Registro de IA:** toda ayuda de IA usada para generar código o decisiones de diseño se
   agrega a `docs/prompts-ia.md`. La cátedra evalúa el uso de IA.

## Clean code y DRY

Aplicá Clean Code y DRY de forma estricta. No es una preferencia: es parte del criterio de
evaluación.

- **Una responsabilidad por función.** Si el nombre de la función necesita un "y" para
  describirla, dividila.
- **Funciones cortas.** Si no cabe en una pantalla, es demasiado larga. Extraé bloques con
  nombre en lugar de agregar comentarios que expliquen bloques.
- **DRY:** la duplicación de la tercera ocurrencia se elimina, no la segunda. Al copiar código,
  dejá una nota de que hay que extraerlo; al notar el patrón, extraé una función o tabla.
  Tablas de datos repetidas (roles, estados, tipos) van como datos, no como `switch` repartidos.
- **Nombres explícitos.** Sin abreviaturas crípticas ni letras sueltas salvo Receiver/Context.
  Los booleanos se leen como oraciones (`isActive`, `hasPermission`).
- **Argumentos.** Cuatro o más parámetros es señal de que hace falta un struct.
- **Comentarios:** explicá el **por qué**, nunca el **qué**. El código ya dice qué hace.
  Un comentario que se desactualiza es peor que ningún comentario.
- **Antipatrones a evitar:** `flag` de bool que cambia el comportamiento, funciones que mutan
  argumentos, singletons globales, `else` anidados profundos (preferí `return` temprano),
  números mágicos (constantes con nombre).
- **Errores:** no los ignores con `_`. Envolvé con contexto y agregá el mensaje que ayuda a
  diagnosticar. Un error sin contexto obliga a adivinar.

## Convenciones de código

### Go

- Layout: `handler` (HTTP) → `service` (reglas de negocio) → `repo` (SQL). El handler no
  contiene lógica de negocio ni SQL.
- Errores: se devuelven envueltos con `fmt.Errorf("...: %w", err)`; el servicio define los
  errores de dominio con `errors.Is`. Los errores esperados (no encontrado, conflicto,
  no autorizado) se traducen a respuestas HTTP en una sola capa.
- Queries: todo el SQL vive en `backend/sql/`. Después de tocarlo, correr `make sqlc`.
- El paquete de HTTP se llama `transport`, no `http`: dentro de él `http` siempre
  significa `net/http`.
- Migraciones: solo archivos nuevos en `backend/migrations/`, nunca edites una ya aplicada.
- Timeouts y contextos: los handlers heredan el `*http.Request` context; las queries usan
  contexto.
- Nada de `panic` fuera de `main` y de los tests.

### Base de datos

- Nombres en `snake_case` y en plural las tablas (`users`, `teams`, `drivers`).
- Toda tabla de negocio lleva `created_at` y `updated_at`.
- Claves foráneas con `ON DELETE` explícito; nunca borrados en cascada silenciosos sobre datos
  históricos de sanctiones o puntajes.
- Idempotencia: los nombres únicos donde aplique (`users.username`, `users.email`).
- **Claves primarias:** `bigint GENERATED ALWAYS AS IDENTITY`. `ALWAYS` impide que
  cualquier código asigne un id a mano, que es lo que desincroniza la secuencia.
- **Datos de ejemplo con ids fijos** en las migraciones de seed, con `setval` al final.
  Por eso los conflictos entre ramas se resuelven con `make db-reset`, nunca fusionando
  filas: las migraciones se fusionan, los datos se regeneran.
- **Baja lógica siempre:** `is_active` + `deactivated_at`. Nunca `DELETE` sobre una fila
  referenciada, así los ids nunca se reciclan.

### App (TypeScript / React Native)

- El código comparte componentes entre iOS, Android y web: nada específico de plataforma fuera
  de un archivo con sufijo `.web.tsx` / `.native.tsx`.
- La lógica de negocio va en hooks o en `src/api`, nunca dentro de los componentes de pantalla.
- Nunca guardes el token de sesión en `AsyncStorage`: usá `expo-secure-store`.

## Comandos

```bash
make db-up        # Postgres en contenedor
make migrate      # Aplicar migraciones
make sqlc         # Regenerar queries tras editar backend/sql
make backend-run  # Levantar la API
make backend-test # Tests del backend
make check        # Lint + typecheck + tests de todo el proyecto
```

## Estado del sprint

El progreso por user story se lleva en `docs/sprint1/TODO.md`. Leelo antes de empezar a
trabajar y actualizá las casillas al terminar.