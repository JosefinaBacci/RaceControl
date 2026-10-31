# Sprint 3 — Lista de tareas

Consigna: `docs/Sprint3_GrupoRojoFerrari.pdf`. Los pendientes que no pertenecen a ninguna US siguen en
`docs/sprint2/TODO.md`; las decisiones de modelo, en `docs/sprint2/02-modelo-de-datos.md`.

Cada US se cierra con sus criterios de éxito verificados (prueba independiente incluida), `make check`
en verde y un commit por cambio lógico.

## Estado

| US | Nombre | SP | Valor | Est. | Tiempo real registrado | Estado |
|----|--------|----|-------|------|------------------------|--------|
| US21 | Registro de controles técnicos | 3 | 8 | 2.5h | — | ☐ |
| US22 | Visualización de controles técnicos | 3 | 8 | 1.5h | — | ☐ |
| US19 | Registro de prueba de neumáticos | 3 | 8 | 2.5h | — | ☐ |
| US20 | Visualización de prueba de neumáticos | 3 | 8 | 1.5h | — | ☐ |
| US15 | Registro de sanciones | 5 | 8 | 4h | — | ☐ |
| US25 | Visualización de sanciones (público) | 5 | 13 | 2h | — | ☐ |
| US26 | Visualización y confirmación de sanciones (escudería) | 5 | 13 | 2.5h | — | ☐ |
| US18 | Disponibilidad de funcionalidades en aplicación móvil | 5 | 13 | 16h | — | ☐ |

**Total estimado:** 32.5h · **Total real registrado:** —

☑ cerrada · ◐ en curso · ☐ sin empezar

El orden de la tabla es el de implementación. Los controles técnicos y las pruebas de neumáticos ya
existen como tipos de evento del calendario (`000012`, `000016`): falta registrar su **resultado**.
Las sanciones necesitan tabla nueva, y US26 reutiliza el patrón de notificación y acuse de US24.
US18 va al final porque es la paridad de todo lo anterior en móvil.

---

## Lo que ya está y se reutiliza

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
- [ ] Formulario de registro en el panel FIA (hecho, compila y empaqueta para web; falta recorrerlo a mano
      en el navegador)
- [x] Tests de integración (incluido acceso denegado a otros roles)

**Criterios de éxito**

- [x] El administrador puede registrar un control técnico con fecha, tipo y resultado
- [x] El sistema valida los campos obligatorios y rechaza fechas inválidas
- [x] El control registrado queda guardado y disponible para su consulta
- [x] Solo el administrador de la FIA puede registrar controles técnicos

## US22 — Visualización de controles técnicos

*Est. 1.5h · SP 3 · Valor 8*

- [x] Endpoint público de lectura del historial
- [ ] Listado del historial con filtros por fecha y tipo de control (pestaña pública "Técnica"; filtros
      verificados en la API, falta recorrer la pantalla a mano)
- [x] Acceso de solo lectura para escudería y público; verificar que coincida con lo cargado

**Criterios de éxito**

- [x] Los responsables de escudería y el público pueden consultar el historial de controles técnicos
- [x] Cada control muestra su fecha, tipo y resultado
- [x] Se puede filtrar el historial por fecha y tipo
- [x] Los datos mostrados coinciden con los cargados por el administrador
- [x] Los usuarios no pueden modificar ni eliminar los controles

## US19 — Registro de prueba de neumáticos

*Est. 2.5h · SP 3 · Valor 8*

- [ ] Modelo: tipo y resultado de la prueba sobre `tyre_tests` (hoy solo tiene campeonato y circuito),
      con autor y fecha, en una migración nueva
- [ ] Queries `sqlc` y servicio con validaciones
- [ ] Endpoint restringido a `fia_admin`
- [ ] Formulario de registro en el panel FIA
- [ ] Tests de integración (incluido acceso denegado a otros roles)

**Criterios de éxito**

- [ ] El administrador puede registrar una prueba de neumático con fecha, tipo y resultado
- [ ] El sistema valida los campos obligatorios y rechaza fechas inválidas
- [ ] La prueba registrada queda guardada y disponible para su consulta
- [ ] Solo el administrador de la FIA puede registrar pruebas de neumáticos

## US20 — Visualización de prueba de neumáticos

*Est. 1.5h · SP 3 · Valor 8*

- [ ] Endpoint público de lectura del historial
- [ ] Listado del historial con filtros por fecha y tipo de prueba
- [ ] Acceso de solo lectura para escudería y público; verificar que coincida con lo cargado

**Criterios de éxito**

- [ ] Los responsables de escudería y el público pueden consultar el historial de pruebas de neumáticos
- [ ] Cada prueba muestra su fecha, tipo y resultado
- [ ] Se puede filtrar el historial por fecha y tipo
- [ ] Los datos mostrados coinciden con los cargados por el administrador
- [ ] Los usuarios no pueden modificar ni eliminar las pruebas

## US15 — Registro de sanciones

*Est. 4h · SP 5 · Valor 8*

- [ ] Migración de `sanctions` (piloto **o** escudería, motivo, penalización, fecha, autor) según el
      diseño de `docs/sprint1/02-modelo-de-datos.md` §2
- [ ] Queries `sqlc` y servicio con validaciones: motivo obligatorio, destinatario y penalización válidos
- [ ] Endpoint restringido a `fia_admin`
- [ ] Formulario de registro en el panel FIA
- [ ] Tests de registro y de acceso no autorizado

**Criterios de éxito**

- [ ] El administrador puede registrar una sanción a un piloto o a una escudería
- [ ] Cada sanción incluye obligatoriamente el motivo y la penalización correspondiente
- [ ] El sistema rechaza sanciones sin motivo, sin destinatario o con datos inválidos
- [ ] Cada sanción queda registrada con fecha y autor
- [ ] Solo el administrador de la FIA puede registrar sanciones

## US25 — Visualización de sanciones (público)

*Est. 2h · SP 5 · Valor 13*

- [ ] Endpoint público de sanciones
- [ ] Listado sobre la API (reemplaza el mock de `useSanctions`)
- [ ] Vista de detalle con motivo, penalización y destinatario
- [ ] Filtros por piloto, escudería y fecha

**Criterios de éxito**

- [ ] Cualquier persona puede ver las sanciones aplicadas sin iniciar sesión
- [ ] Cada sanción muestra el destinatario, el motivo, la penalización y la fecha
- [ ] Se puede filtrar por piloto, escudería y fecha
- [ ] Los datos mostrados coinciden con los registrados por el administrador
- [ ] El público no puede modificar la información

## US26 — Visualización y confirmación de sanciones (escudería)

*Est. 2.5h · SP 5 · Valor 13*

- [ ] Vista de sanciones para el rol escudería
- [ ] `sanction_acknowledgements` siguiendo el modelo de acuse de US24 (usuario, sanción, fecha y hora),
      append-only
- [ ] Notificación al registrar o modificar una sanción
- [ ] Botón de confirmación de recibo y su estado (pendiente / confirmada)
- [ ] Vista de seguimiento de confirmaciones para el administrador

**Criterios de éxito**

- [ ] El responsable de escudería puede ver las sanciones aplicadas
- [ ] Recibe una notificación cuando se registra o modifica una sanción
- [ ] Puede confirmar la recepción, y la confirmación queda registrada con fecha y hora
- [ ] El estado de la notificación (pendiente o confirmada) es visible
- [ ] El administrador puede ver qué escuderías confirmaron cada sanción

## US18 — Disponibilidad de funcionalidades en aplicación móvil

*Est. 16h · SP 5 · Valor 13*

- [ ] Definir el alcance y la paridad de funcionalidades respecto a la web
- [x] Proyecto móvil y navegación principal (Expo + expo-router desde US1)
- [x] Autenticación y roles contra el backend (token en `expo-secure-store` desde US5)
- [ ] Pantallas de consulta: calendario, puntajes y pilotos ya compartidas; faltan sanciones, pruebas de
      neumáticos y controles técnicos (dependen de US20, US22 y US25)
- [ ] Funciones propias de cada rol (escudería y administrador) verificadas en móvil
- [ ] Pruebas en distintos dispositivos y tamaños de pantalla

**Criterios de éxito**

- [ ] El usuario puede acceder desde la app móvil a las mismas funcionalidades que en la web
- [ ] El inicio de sesión y los permisos por rol funcionan igual que en la web
- [ ] Los datos mostrados en la app coinciden con los de la plataforma web
- [ ] La app se visualiza y funciona correctamente en distintos dispositivos y tamaños de pantalla
- [ ] Los cambios hechos en la plataforma se reflejan en la app
