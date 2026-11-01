# Sprint 3 — Lista de tareas

Consigna: `docs/Sprint3_GrupoRojoFerrari.pdf`. Los pendientes que no pertenecen a ninguna US siguen en
`docs/sprint2/TODO.md`; las decisiones de modelo, en `docs/sprint2/02-modelo-de-datos.md`.

Cada US se cierra con sus criterios de éxito verificados (prueba independiente incluida), `make check`
en verde y un commit por cambio lógico.

## Estado

| US | Nombre | SP | Valor | Est. | Tiempo real registrado | Estado |
|----|--------|----|-------|------|------------------------|--------|
| US21 | Registro de controles técnicos | 3 | 8 | 2.5h | 0.6h | ☑ |
| US22 | Visualización de controles técnicos | 3 | 8 | 1.5h | 0.4h | ☑ |
| US19 | Registro de prueba de neumáticos | 3 | 8 | 2.5h | 0.5h | ☑ |
| US20 | Visualización de prueba de neumáticos | 3 | 8 | 1.5h | 0.3h | ☑ |
| US15 | Registro de sanciones | 5 | 8 | 4h | 0.6h | ☑ |
| US25 | Visualización de sanciones (público) | 5 | 13 | 2h | 0.3h | ☑ |
| US26 | Visualización y confirmación de sanciones (escudería) | 5 | 13 | 2.5h | 0.4h | ☑ |
| US18 | Disponibilidad de funcionalidades en aplicación móvil | 5 | 13 | 16h | 1.5h | ☑ |

**Total estimado:** 32.5h · **Total real registrado:** 4.6h

El tiempo real sale de los registros de la sesión de IA (31/10 17:42 – 01/11 00:14, hora
argentina): el tiempo activo de cada equipo de implementación más la parte proporcional del
tiempo del manager (coordinación, merges, prueba de aceptación y de estrés), repartidos entre
las historias de cada equipo según su estimación. Los equipos trabajaron en paralelo, así que el
tiempo de calendario fue menor que la suma.

☑ cerrada · ◐ en curso · ☐ sin empezar

El orden de la tabla es el de implementación. Los controles técnicos y las pruebas de neumáticos ya
existían como tipos de evento del calendario (`000012`, `000016`): el sprint agregó el registro de sus resultados.
Las sanciones tienen tablas propias, y US26 reutiliza el patrón de notificación y acuse de US24.
US18 va al final porque es la paridad de todo lo anterior en móvil.

---

## Prueba de aceptación del sprint (manager, 2026-10-31)

Hecha sobre `main` con las tres ramas integradas, por un agente que no escribió el código:

- **API (221 verificaciones):** permisos de cada ruta (401 sin sesión, 403 con otro rol), validaciones
  campo por campo (tipos equivocados, fechas imposibles, fuera del evento o futuras, eventos cancelados o
  sin empezar, textos vacíos o de más de 2000 caracteres, cuerpos mal formados o de 5 MB), filtros con
  bordes inclusivos, ausencia de nombres de usuario en las respuestas públicas y rutas de solo lectura.
- **Concurrencia:** 40 altas simultáneas del mismo resultado → una sola 201; 40 correcciones simultáneas
  → una sola 201; 30 revocaciones simultáneas de una sanción → una sola 200; 20 modificaciones
  simultáneas → 21 versiones; 30 acuses simultáneos → una sola fila.
- **Carga:** 3000 lecturas públicas por historial con 64 clientes: ~2500–3100 req/s, p95 < 40 ms, sin
  errores.
- **Base de datos:** `UPDATE`/`DELETE` directos sobre resultados, versiones y acuses rechazados por las
  guardas append-only; la seed es idempotente.
- **Interfaz:** las 36 rutas de los tres roles recorridas en Chromium headless a 375 px y 1366 px sin
  excepciones ni desbordes, y los flujos completos de registrar/corregir resultados y de registrar,
  confirmar desde un teléfono y revocar sanciones.

Hallazgos corregidos: horas de las sanciones en UTC como el resto (`27ef8b0`), filtro de fechas apilado
en teléfonos (`7527501`) y `decodeJSON` que nombra el campo con tipo inválido y rechaza datos sobrantes
(`b166f2a`).

## Punto de partida antes de implementar el sprint

- Eventos `technical_control` y `tyre_test` con fecha, campeonato y tipo de control
  (`technical_controls.control_type`), alta/edición/baja desde el calendario de US7.
- Diseño de `sanctions` con `CHECK (num_nonnulls(driver_id, team_id) = 1)` y de
  `sanction_acknowledgements` (`docs/sprint1/02-modelo-de-datos.md` §2 y §4); todavía sin migración.
- Notificaciones, acuses append-only (`000017`) y la vista de seguimiento de acuses de US24.
- Pantallas de sanciones (pública, panel FIA y escudería) sobre mocks desde US4
  (`app/src/features/SanctionList.tsx`, `useSanctions` en `app/src/data/hooks.ts`).
- La app es una sola base Expo para iOS, Android y web: US18 es sobre todo verificación de paridad.

## US21 — Registro de controles técnicos

*Est. 2.5h · SP 3 · Valor 8*

- [x] Modelo del resultado del control (`technical_control_results`: escudería, aprobado, observaciones,
      autor, fecha) en una migración nueva
- [x] Queries `sqlc` y servicio con validaciones (resultado obligatorio, fecha válida, control existente)
- [x] Endpoint restringido a `fia_admin`
- [x] Formulario de registro en el panel FIA (recorrido en Chromium por la prueba de aceptación: errores
      sin veredicto y sin observaciones, alta y aparición en la pestaña pública)
- [x] Tests de integración (incluido acceso denegado a otros roles)

**Criterios de éxito**

- [x] El administrador puede registrar un control técnico con fecha, tipo y resultado
- [x] El sistema valida los campos obligatorios y rechaza fechas inválidas
- [x] El control registrado queda guardado y disponible para su consulta
- [x] Solo el administrador de la FIA puede registrar controles técnicos

## US22 — Visualización de controles técnicos

*Est. 1.5h · SP 3 · Valor 8*

- [x] Endpoint público de lectura del historial
- [x] Listado del historial con filtros por fecha y tipo de control (pestaña pública "Técnica", recorrida
      en Chromium a 375 px y 1366 px)
- [x] Acceso de solo lectura para escudería y público; verificar que coincida con lo cargado

**Criterios de éxito**

- [x] Los responsables de escudería y el público pueden consultar el historial de controles técnicos
- [x] Cada control muestra su fecha, tipo y resultado
- [x] Se puede filtrar el historial por fecha y tipo
- [x] Los datos mostrados coinciden con los cargados por el administrador
- [x] Los usuarios no pueden modificar ni eliminar los controles

## US19 — Registro de prueba de neumáticos

*Est. 2.5h · SP 3 · Valor 8*

- [x] Modelo: tipo y resultado de la prueba sobre `tyre_tests` (hoy solo tiene campeonato y circuito),
      con autor y fecha, en una migración nueva (`000032` catálogo de compuestos, `000033`
      `tyre_test_results` append-only con correcciones; ver `docs/sprint2/02-modelo-de-datos.md`)
- [x] Queries `sqlc` y servicio con validaciones
- [x] Endpoint restringido a `fia_admin`
- [x] Formulario de registro en el panel FIA (recorrido en Chromium por la prueba de aceptación: alta de un
      compuesto y aparición en el historial público)
- [x] Tests de integración (incluido acceso denegado a otros roles)

**Criterios de éxito**

- [x] El administrador puede registrar una prueba de neumático con fecha, tipo y resultado
- [x] El sistema valida los campos obligatorios y rechaza fechas inválidas
- [x] La prueba registrada queda guardada y disponible para su consulta
- [x] Solo el administrador de la FIA puede registrar pruebas de neumáticos

## US20 — Visualización de prueba de neumáticos

*Est. 1.5h · SP 3 · Valor 8*

- [x] Endpoint público de lectura del historial
- [x] Listado del historial con filtros por fecha y tipo de prueba
- [x] Acceso de solo lectura para escudería y público; verificar que coincida con lo cargado

**Criterios de éxito**

- [x] Los responsables de escudería y el público pueden consultar el historial de pruebas de neumáticos
- [x] Cada prueba muestra su fecha, tipo y resultado
- [x] Se puede filtrar el historial por fecha y tipo
- [x] Los datos mostrados coinciden con los cargados por el administrador
- [x] Los usuarios no pueden modificar ni eliminar las pruebas

## US15 — Registro de sanciones

*Est. 4h · SP 5 · Valor 8*

- [x] Migración de `sanctions` (piloto **o** escudería, motivo, penalización, fecha, autor) según el
      diseño de `docs/sprint1/02-modelo-de-datos.md` §2
- [x] Queries `sqlc` y servicio con validaciones: motivo obligatorio, destinatario y penalización válidos
- [x] Endpoint restringido a `fia_admin`
- [x] Formulario de registro en el panel FIA
- [x] Tests de registro y de acceso no autorizado

**Criterios de éxito**

- [x] El administrador puede registrar una sanción a un piloto o a una escudería
- [x] Cada sanción incluye obligatoriamente el motivo y la penalización correspondiente
- [x] El sistema rechaza sanciones sin motivo, sin destinatario o con datos inválidos
- [x] Cada sanción queda registrada con fecha y autor
- [x] Solo el administrador de la FIA puede registrar sanciones

## US25 — Visualización de sanciones (público)

*Est. 2h · SP 5 · Valor 13*

- [x] Endpoint público de sanciones
- [x] Listado sobre la API (reemplaza el mock de `useSanctions`)
- [x] Vista de detalle con motivo, penalización y destinatario
- [x] Filtros por piloto, escudería y fecha. Los selectores se arman con los destinatarios del historial,
      así que también incluyen pilotos desvinculados y escuderías anteriores.

**Criterios de éxito**

- [x] Cualquier persona puede ver las sanciones aplicadas sin iniciar sesión
- [x] Cada sanción muestra el destinatario, el motivo, la penalización y la fecha
- [x] Se puede filtrar por piloto, escudería y fecha
- [x] Los datos mostrados coinciden con los registrados por el administrador
- [x] El público no puede modificar la información

## US26 — Visualización y confirmación de sanciones (escudería)

*Est. 2.5h · SP 5 · Valor 13*

- [x] Vista de sanciones para el rol escudería
- [x] `sanction_acknowledgements` siguiendo el modelo de acuse de US24 (usuario, sanción, fecha y hora),
      append-only
- [x] Notificación al registrar o modificar una sanción
- [x] Botón de confirmación de recibo y su estado (pendiente / confirmada)
- [x] Vista de seguimiento de confirmaciones para el administrador

**Criterios de éxito**

- [x] El responsable de escudería puede ver las sanciones aplicadas
- [x] Recibe una notificación cuando se registra o modifica una sanción
- [x] Puede confirmar la recepción, y la confirmación queda registrada con fecha y hora
- [x] El estado de la notificación (pendiente o confirmada) es visible
- [x] El administrador puede ver qué escuderías confirmaron cada sanción

## US18 — Disponibilidad de funcionalidades en aplicación móvil

*Est. 16h · SP 5 · Valor 13*

La app es una sola base Expo para iOS, Android y web, así que la paridad se verifica en lugar de
construirse dos veces. Matriz de paridad por rol, riesgos nativos y método de prueba en
[`plans/mobile.md`](plans/mobile.md). **No hubo dispositivos reales ni emuladores disponibles:** la
evidencia son las exportaciones nativas (`npx expo export --platform android` y `--platform ios`, sin
errores) y un recorrido automático con Playwright en modo táctil por todas las rutas de los tres roles
a 320, 375, 414, 740×360 (apaisado), 768 y 1366 px de ancho.

- [x] Definir el alcance y la paridad de funcionalidades respecto a la web (matriz en `plans/mobile.md`:
      ninguna funcionalidad web falta en el teléfono; la brecha era la barra de pestañas)
- [x] Proyecto móvil y navegación principal (Expo + expo-router desde US1); en el teléfono la barra
      inferior muestra 4 pestañas + **Más** para los tres roles, armada desde una tabla por rol
      (`src/navigation/tabSpecs.ts`); la barra lateral de escritorio no cambia
- [x] Autenticación y roles contra el backend (token en `expo-secure-store` desde US5; camino nativo
      con `Authorization: Bearer` probado con `curl`: 200 en lo propio, 403 en lo de otro rol, 401
      tras cerrar sesión)
- [x] Pantallas de consulta: calendario, puntajes, pilotos, sanciones, pruebas de neumáticos y
      controles técnicos, recorridas en todos los anchos sin errores de consola, sin respuestas 4xx/5xx
      y sin desbordamiento horizontal
- [x] Funciones propias de cada rol (escudería y administrador): todas sus pantallas y formularios
      cargan en el teléfono y cada destino de **Más** se abre tocando (prueba automática); los
      formularios no se enviaron desde un dispositivo real
- [x] Pruebas en distintos tamaños de pantalla (los seis anchos de arriba); en dispositivos reales no

**Criterios de éxito**

- [x] El usuario puede acceder desde la app móvil a las mismas funcionalidades que en la web
- [x] El inicio de sesión y los permisos por rol funcionan igual que en la web
- [x] Los datos mostrados en la app coinciden con los de la plataforma web (mismo cliente HTTP y
      mismos endpoints en las tres plataformas; ningún dato se transforma por plataforma)
- [x] La app se visualiza y funciona correctamente en distintos dispositivos y tamaños de pantalla
      (recorrido del manager sobre `main` a 320, 375, 414, 768 y 1366 px y apaisado, en modo táctil:
      sin desbordes ni etiquetas cortadas; filtro de fechas apilado en `7527501` e insignias del
      historial que bajan de línea en `2736a44`. Sin dispositivo real: la evidencia es el recorrido y los
      paquetes nativos de Android e iOS que compilan)
- [x] Los cambios hechos en la plataforma se reflejan en la app (prueba automática: la FIA abre una
      conversación por la API y el teléfono de la escudería la muestra, con la insignia actualizada,
      al cambiar de pestaña, sin recargar)

## Revisión funcional del 1 de noviembre

- Se recorrieron 47 combinaciones de rol y ruta en Edge: acceso público y FIA a 1280 px y escudería a 375 px. Se verificaron la carga de pantallas y detalles, los destinos, las respuestas de la API y el desbordamiento horizontal. No se detectaron excepciones de la interfaz en el recorrido.
- En una base de QA separada se probó el login de FIA y escudería, se renombró una escudería y se verificó el dato en el panel y la cuenta. Desde dos sesiones se envió, recibió y respondió un mensaje. La FIA registró una sanción, se comprobó el motivo en el detalle público, la escudería confirmó la notificación y se verificó el acuse desde la API de seguimiento de la FIA. Se descargaron PDF e ICS filtrados por categoría.
- En Android se abrió la app con Expo Go SDK 57 en un emulador API 36. Se verificó la carga de los filtros del historial de sanciones a 720 × 1280 px. Esta comprobación nativa fue parcial: no se repitieron allí todos los flujos web ni se probó en iOS o en un teléfono físico.
- Se ejecutó `make check` con PostgreSQL: autenticación, roles y cuentas, calendario y exportaciones, pilotos y asignaciones, resultados y clasificaciones, mensajería, controles técnicos, neumáticos, sanciones y acuses; también TypeScript y las 33 pruebas de la app. Se incluyó la regresión del acuse ante una lectura fallida y su verificación concurrente.
