# Registro de prompts de IA

La cátedra (**APS 2026, Enunciado 1**) exige registrar los prompts utilizados con IA para generar
código o decisiones de diseño, porque "la correcta utilización de las herramientas de IA es parte
de la evaluación del proyecto".

Este archivo es el registro de la **comisión implementadora** (Grupo Rojo Ferrari). Cada entrada
declara el contexto, el prompt literal y qué se validó o modificó de la salida.

> Los prompts usados por la comisión de Análisis y Management se registran en su propio
> repositorio.

---

## Regla de trabajo

Antes de dar por cerrada cualquier tarea que haya usado IA:

1. Agregar la entrada acá con el prompt literal.
2. Revisar la salida: leer el código generado, ejecutar los tests, verificar que no haya
   introducido dependencias no deseadas ni secretos.
3. Anotar en la columna **Validación** qué se corrigió a mano.

Ningún prompt se registra "a posteriori" ni de forma resumida: se copia textual.

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

**Validación:** se ajustó de la recomendación inicial. La IA propuso sesiones en servidor por
seguridad, pero el criterio definitivo se fijó con el equipo: **JWT
puro se descarta** porque US6 exige desactivar usuarios y un JWT no se puede revocar antes de su
expiración, lo que dejaría acceso a cuentas desactivadas. También se acotó el árbol de
dependencias del frontend a tres paquetes (`expo`, `expo-router`, `expo-secure-store`)
justificando que `expo-router` reemplaza a `react-navigation` en lugar de sumarse. La comparación
de frameworks de backend (chi / gin / echo / fiber / stdlib) se contrastó con la
documentación de cada proyecto antes de escribir `01-stack.md`.

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

> Modelar en PostgreSQL el dominio de una plataforma de gestión de tags de la Fórmula 1:
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

> Corrección posterior: el seed `000002_seed_reference` de esta entrada fue eliminado en el
> commit *Separar esquema y datos de referencia* (ver la entrada de abajo).

### 2026-10-02 — Separación de esquema y datos de referencia

**Prompt (literal):**

> Why did you put the seeds in the migrations directory?

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
`equipos` 1–8, que `make seed` corrido dos veces no duplica filas, y que no hay retroceso de
secuencia: tras insertar un equipo desde la aplicación (id 9) y reejecutar el seed, el siguiente
insert recibió el id 10.

**Corrección de la IA (asumida explícitamente):** la IA predijo que
`make migrate-down 1` dejaría las filas intactas. Fue **incorrecto**: con migraciones solo DDL ese
comando tira el esquema entero, tablas y filas incluidas. El comportamiento real verificado es que
`make migrate` no inserta datos y la recuperación es `make migrate` + `make seed`, que devuelve los
mismos ids. El beneficio real del cambio no es "el rollback no borra filas" sino que desaparece el
estado intermedio confuso del diseño anterior, donde quedaba un esquema válido pero sin datos de
referencia.

**Resultado:** `backend/seeds/*`, `backend/internal/seed/seed.go`, `backend/cmd/seed/main.go`,
`backend/internal/config/config.go`, `Makefile`, `.env.example`, `AGENTS.md`,
`docs/sprint1/02-modelo-de-datos.md`, `docs/sprint1/TODO.md`
