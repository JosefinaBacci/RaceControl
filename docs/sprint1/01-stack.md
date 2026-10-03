# US1 — Selección de stack y configuración del entorno

**Comisión:** Grupo Rojo Ferrari (Sevenants Antonio, Tourn Felipe)
**Story points:** 5 · **Business value:** 8
**Estado:** decisión tomada en sesión de diseño, documentada acá.

## Objetivo

Definir las tecnologías del sistema y dejar el entorno de desarrollo operativo, de modo que el
equipo tenga una base tecnológica común antes de escribir la primera línea de código de US5.

## Resumen de decisiones

| Área | Decisión | Alternativas descartadas |
|------|----------|--------------------------|
| Backend | Go 1.27 + `chi` v5 | gin, echo, fiber, `http.ServeMux` pelado |
| Base de datos | PostgreSQL 18 + `pgx` v5 + `sqlc` | MySQL, MongoDB, Redis |
| App | Expo + expo-router + expo-secure-store | React Native CLI, Ionic, Flutter, dos apps separadas |
| Autenticación | Sesiones opacas en servidor, Argon2id | JWT puro, OAuth externo |
| Entorno | Contenedores `podman`, compose opcional | Quadlets/systemd, servidor remoto |

---

## 1. Backend: Go 1.27 + chi

### Candidatos evaluados

| Framework | Punto a favor | Punto en contra |
|-----------|---------------|-----------------|
| **chi** v5 | `net/http` nativo: handlers `http.Handler`, middleware `func(http.Handler) http.Handler`. Sin contexto propio, sin binding implícito. | Hay que escribir el enrutado de params y el decode de JSON (poco código) |
| gin v1.12 | El más popular, curva de aprendizaje más corta para quien viene de Express | `gin.Context` propio que no es `*http.Request`; auto-binding y validación que se superponen con lo que igual hay que escribir; historial de CVEs en el manejo de SSL/TLS |
| echo v4 | Similar a gin, con bindings más sobrios | Mismo problema de contexto propio, ecosistema más chico |
| fiber v2 | Muy rápido, API amplia | Corre sobre `fasthttp`, no sobre `net/http`: deja de ser idiomático y complica el uso de middleware de la stdlib |
| `http.ServeMux` (stdlib) | Cero dependencias; Go 1.22+ ya resuelve métodos y wildcards (`GET /users/{id}`) | No compone middleware: hay que escribir a mano el encadenado y el manejo de contexto |

### Decisión y justificación

**Elegimos chi.** El criterio fue "lo más nativo y fácil que se pueda sostener en el tiempo".

chi *es* `net/http` con un router encima: cualquier handler, middleware o test escrito contra la
librería estándar funciona igual. Eso importa para el resto del proyecto, no solo hoy:

- El middleware de RBAC (US3) y el de autenticación son funciones `func(http.Handler) http.Handler`,
  exactamente igual que cualquier middleware de la documentación de Go.
- Los tests usan `httptest.NewServer` y `http.NewRequest` sin adaptadores.
- No hay una capa de abstracción propia que deserialice por su cuenta y pueda desincronizarse de la
  stdlib.

Descartamos `fiber` por la misma razón: reemplazar `net/http` por `fasthttp` es una decisión
irreversible a nivel arquitectura y no aporta al problema que estamos resolviendo.

Descartamos la stdlib pelada porque, aunque sea lo más nativo posible, la composición de
middleware es una pieza estructural que vamos a necesitar de todas formas y chi la da resuelta,
verificada y con una sola dependencia muy chica.

### Justificación del lenguaje

Go fue elegido por sobre Node/TypeScript y Java/Spring porque:

- El perfil de concurrencia del dominio (muchas lecturas públicas simultáneas del calendario,
  los puntajes y las sanciones) es trivial en Go y pide muy poca configuración en los otros.
- Despliegue simple: un binario estático, que corre igual en el contenedor de desarrollo y en
  producción.
- El tipado estático y la explicititud hacen que la lógica de permisos sea fácil de auditar, que es
  un requisito del proyecto.

> Nota: la comparación completa de alternativas de backend, con la matriz de criterios
> ponderados, se completa en el informe final.

---

## 2. Base de datos: PostgreSQL 18 + pgx + sqlc

### Por qué relacional

El dominio es relacional en su definición, no por conveniencia:

- Un **piloto pertenece a una escudería** dentro de una **categoría** y una **temporada**.
- Un **puntaje** identifica a piloto, escudería, carrera y posición: es una tabla de unión con
  atributos propios.
- Una **sanción** aplica sobre un piloto o una escudería, referencia un evento y debe conservar su
  historial aunque el evento se modifique.
- El enunciado pide explícitamente que puntajes, sanciones y controles técnicos sean consistentes
  y disponibles de inmediato para tres audiencias distintas.

Eso pide claves foráneas, `UNIQUE`, claves foráneas compuestas y consultas de agregación
(`SUM` por escudería, posición en el campeonato). MongoDB nos obligaría a simular todo eso con
documentos embebidos y agregaciones manuales, y nos deja sin integridad referencial declarativa.

### Por qué PostgreSQL y no MySQL

MySQL cubre el modelo relacional básico, pero PostgreSQL da cuatro cosas que el proyecto usa:

- **DDL transaccional.** Una migración que falla a mitad de camino se revierte entera; en MySQL
  cada `CREATE TABLE` confirma implícitamente y deja el esquema a medias.
- **Restricciones expresivas.** `CHECK` con expresiones regulares, columnas generadas y funciones
  como `num_nonnulls` permiten que la base haga cumplir las reglas del dominio (formato de
  `username`, exactamente un sancionado por sanción) y no solo el código.
- **Tipos nativos** como `inet` para las IP de `login_attempts` y `jsonb` por si más adelante hay
  que guardar datos poco estructurados (un resultado de control técnico, por ejemplo).
- **Mejor soporte de `sqlc`**, cuyo motor más maduro es el de PostgreSQL y usa su propio parser.

Se elige PostgreSQL. Los roles y estados se modelan con `CHECK` en lugar del tipo `ENUM`, por el
motivo que se detalla en `02-modelo-de-datos.md`.

### Capa de acceso: pgx + sqlc

Consideramos tres opciones:

| Opción | Evaluación |
|--------|------------|
| **pgx + sqlc** | Escribimos SQL explícito (en `backend/sql/`) y `sqlc generate` produce structs Go y funciones tipadas. Errores de columnas y firmas detectados **en compilación**. Cero reflection. Se descarta GORM: el mapeo objeto-relacional esconde el SQL, complica las consultas de reporting y agrega una dependencia grande. |
| pgx crudo | Lo más simple y totalmente idiomático | Se descarta: el mapeo de filas a structs se repite en cada query y el refactor se vuelve manual |
| GORM | CRUD muy rápido | Se descarta por opacidad: el enunciado exige integridad y consultas sobre puntajes/sanciones, y un ORM estorba ahí |

El único costo de `sqlc` es un paso de generación (`make sqlc`) que hay que correr después de
tocar un `.sql`. Es aceptable porque el Makefile lo automatiza y `AGENTS.md` lo exige.

### Migraciones

SQL embebido en el binario con `golang-migrate`, aplicado con un comando propio
(`make migrate`, que corre `cmd/migrate`) y no en el arranque de la API: así el deploy decide
cuándo cambia el esquema y dos réplicas de la API nunca compiten por migrar. El esquema queda
versionado en el repositorio como código y cada entorno aplica exactamente los mismos archivos. **Nunca se edita una migración ya aplicada**: los cambios son migraciones nuevas.

---

## 3. App: Expo + expo-router + expo-secure-store

### El requisito que condiciona todo

El enunciado pide dos cosas: un **sistema web integral** y que **toda la funcionalidad esté
disponible desde una app móvil**. Esa duplicación de UI es el riesgo de costo más grande del
proyecto.

### Candidatos evaluados

| Opción | Web | Móvil | Veredicto |
|--------|-----|-------|-----------|
| **Expo + react-native-web** | `npx expo start --web` renderiza los mismos componentes en el navegador | iOS, Android y web desde un solo código | Elegida |
| React Native CLI | No hay target web sin cableado manual | iOS y Android, configuración nativa por plataforma | Descartada: obliga a dos UIs o a un setup manual de web |
| Flutter | Web y móvil, muy consistente | Igual | Descartada por falta de experiencia del equipo en Dart |
| Dos apps (Next.js + RN) | Nativa y cómoda | Nativa | Descartada: dos bases de código, dos deploys, el doble de mantenimiento |

### Por qué Expo y no React Native pelado

Expo no es un framework alternativo a React Native: es React Native más un conjunto de módulos
nativos ya compilados y una capa de herramientas. Sobre eso aporta exactamente lo que el
proyecto necesita:

1. **Web con el mismo código.** `react-native-web` renderiza los componentes en el navegador.
   Una pantalla escrita una vez corre en los tres destinos.
2. **`expo-secure-store`** es la pieza de seguridad clave: guarda el token de sesión en el
   Keychain de iOS y el Keystore de Android. Sin Expo habría que escribir un módulo nativo por
   plataforma.
3. **Builds de demo rápidos**, sin Android Studio ni Xcode instalados.

### Dependencias: tres decisiones, el resto del SDK

Surge una pregunta razonable: ¿no estamos agregando dependencias? El objetivo fue minimizar el
árbol. Las decisiones de arquitectura son tres paquetes:

- `expo` — toolchain de desarrollo y build.
- `expo-router` — routing por archivos. **No es una dependencia extra**: reemplaza a
  react-navigation, que habríamos tenido que instalar igual.
- `expo-secure-store` — almacenamiento seguro del token.

El resto de `app/package.json` son módulos del propio SDK de Expo, en versiones que el SDK fija:
dependencias que `expo-router` necesita (`react-native-screens`, `react-native-safe-area-context`,
`expo-linking`, `expo-constants`) y tres que se sumaron en US4 para la interfaz:
`@expo/vector-icons` (iconos), `expo-linear-gradient` (degradés sobre las fotos) y
`react-native-svg` (banderas, circuitos y el auto ilustrados).

No se suma ningún UI kit: los componentes salen de los primitivos de React Native (`Pressable`,
`StyleSheet`) y del design system propio en `app/src/theme`. Menos peso, más control visual y un
look consistente entre las tres plataformas.

### Lo que se acepta como costo

- El **bundle web** es más pesado que el de una SPA con React plano. Se acepta: el enunciado
  prioriza una sola base de código y el volumen de datos del dominio es bajo (sin video ni
  streams). Si el performance resultara un problema, se evalúa cachear con `Static` rendering en
  las páginas públicas.
- `expo-router` impone convención de rutas por archivo. Es una restricción aceptada a cambio de
  no mantener react-navigation.

---

## 4. Autenticación: sesiones opacas + Argon2id

Este es el punto donde la decisión es puramente de **seguridad**, y es la razón por la que la
señalamos como decisión de diseño y no como detalle de implementación.

### Por qué no JWT puro

El enunciado, en US6, pide **eliminar o desactivar** usuarios. Si la autenticación fuera un JWT,
un token emitido antes de la desactivación seguiría siendo válido hasta su expiración: el
usuario desactivado conservaría acceso, y probablemente el período de gracia sería de horas.
Revocar un JWT exige, además, o una lista de revocación consultada en cada request (que es
justamente el costo de una sesión en servidor) o una vida tan corta que obliga a refrescar
constantemente.

Con sesiones en servidor la desactivación es inmediata: el token deja de existir y el usuario
pierde acceso en el request siguiente.

### Mecanismo elegido

- Token de **32 bytes aleatorios** (`crypto/rand`) generado al iniciar sesión.
- En la base de datos se guarda **solo el SHA-256 del token**, nunca el token. Una filtración de
  la base de datos no permite autenticarse.
- **Sesión nueva en cada login**: el token nunca se reutiliza entre inicios de sesión. La
  rotación en cada request se descartó porque dos peticiones simultáneas se invalidarían
  mutuamente; el detalle está en US3.
- **Expiración por inactividad** (12 h) y **expiración absoluta** (168 h), de modo que un token
  robado tiene un techo de vida acotado.
- **Contraseñas con Argon2id**, la primera recomendación de OWASP. El hash incluye sal por
  usuario y los parámetros (memoria, iteraciones, paralelismo) se configuran por entorno para
  poder subirlos cuando el hardware lo permita: las contraseñas existentes siguen funcionando y
  se rehashean con el coste nuevo en el siguiente login correcto.
- **Transporte según plataforma**: cookie `httpOnly; Secure; SameSite=Strict` en web, y
  `expo-secure-store` en móvil. La cookie `httpOnly` es la que impide que un XSS lea el token;
  `SameSite=Strict` corta el CSRF sobre el endpoint de sesión.
- **Bloqueo por intentos fallidos** (cinco seguidos bloquean el usuario 15 minutos) y registro de
  cada intento en `login_attempts`.

El detalle completo de la matriz de permisos y de los estándares de seguridad se desarrolla en
US3 (`03-arquitectura-roles-seguridad.md`).

---

## 5. Entorno de desarrollo

### Contenedores con podman

| Requisito | Decisión |
|-----------|----------|
| Funcionar en Windows | Se usa `podman` + un `compose.yaml` estándar. En Windows, `podman machine` o Docker Desktop levantan el mismo compose. **Nada de systemd ni Quadlets**, que son específicos de Linux. |
| Sin dependencias del host | Postgres 18 corre en un contenedor. El equipo no necesita instalar un servidor de base de datos ni configurar usuarios locales. |
| Reproducible | El `Makefile` expone `make db-up`, `make db-down`, `make db-logs`. Las credenciales de desarrollo están fijadas en el Makefile y documentadas; nunca son secretos reales. |
| Aislamiento | El volumen nombrado `racecontrol-db-data` conserva los datos entre reinicios. `make db-down` no borra datos. |

### Configuración

- Toda la configuración se lee del entorno (`.env`), versionando solo `.env.example`.
- **Ningún secreto en el repositorio.** `.env` está en `.gitignore` desde el primer commit.
- Los parámetros que hay que justificar (Argon2id, TTL de sesión) están declarados en
  `.env.example` con su valor por defecto y su motivo, para que un revisor pueda ver de un
  vistazo qué está configurado.

### Verificación del entorno

```
$ make db-up
$ psql $DATABASE_URL -tAc 'select version();'
PostgreSQL 18.6 ...
```

El contenedor levanta en ~3 segundos y responde `pg_isready` correctamente.

---

## Criterios de éxito

| Criterio | Estado |
|----------|--------|
| El stack tecnológico queda documentado y aprobado por el equipo | ✅ este documento |
| Los repositorios están creados y son accesibles para el equipo | ✅ `RaceControl` en Git, rama `main` |
| Los entornos están configurados y operativos | ✅ Postgres 18.6 en contenedor, verificado |

## Consecuencias aceptadas

1. **Costo del bundle web** por usar React Native para web. Se acepta a cambio de una sola base
   de código.
2. **Un paso de generación** (`make sqlc`) que hay que recordar después de tocar un `.sql`.
   Mitigado con el target del Makefile y con la instrucción en `AGENTS.md`.
3. **Enums como CHECK en vez de tipos ENUM de PostgreSQL**, para que agregar un valor sea una
   migración y no un lock de reescritura de tabla.
4. **El equipo tiene que aprender Expo**, que es una capa adicional sobre React Native. Se
   mitiga con que la documentación de Expo es mayoritaria y el routing por archivos elimina
   configuración.