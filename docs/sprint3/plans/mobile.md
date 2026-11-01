# Plan: equipo mobile (US18)

## Alcance

La app es **una sola base de código** Expo + expo-router para iOS, Android y web (decisión de US1).
Cada pantalla, hook y cliente HTTP es el mismo en las tres plataformas: no hay una "app móvil"
separada que construir. US18 consiste en **verificar** que todo lo de la web funciona en un teléfono
y **cerrar las brechas** que aparezcan. No hay modelo de datos, migraciones ni endpoints nuevos.

## Matriz de paridad (entregable de la tarea "Definir el alcance y la paridad")

Columna "Teléfono": disponible en la app con la misma pantalla y el mismo endpoint que la web.
"Nav" indica dónde se llega en el teléfono después de este cambio (pestaña o menú **Más**).

### Público (sin sesión)

| Funcionalidad web | Ruta | Teléfono | Nav en teléfono |
|---|---|---|---|
| Inicio: próxima carrera, cuenta regresiva, resúmenes | `/` | Sí | Pestaña Inicio |
| Calendario con filtros, detalle de evento | `/calendar`, `/events/[id]` | Sí | Pestaña Carreras |
| Descarga del calendario `.ics` / PDF | `CalendarDownloads` | Sí (abre la URL con `Linking`, el sistema la descarga o la ofrece a Calendario) | Pestaña Carreras |
| Pilotos y ficha del piloto | `/drivers`, `/drivers/[id]` | Sí | Más → Pilotos |
| Clasificación (puntajes) | `/standings` | Sí | Pestaña Puntajes |
| Sanciones con filtros y detalle (US25) | `/sanctions`, `/sanctions/[id]` | Sí | Pestaña Sanciones |
| Controles técnicos y pruebas de neumáticos con filtros (US20, US22) | `/technical` | Sí | Más → Técnica |
| Inicio de sesión | `/login` | Sí (token en `expo-secure-store`) | Botón "Ingresar" del encabezado |

### Administrador de escudería (`team_admin`)

| Funcionalidad web | Ruta | Teléfono | Nav en teléfono |
|---|---|---|---|
| Panel de la escudería | `/team` | Sí | Pestaña Mi escudería |
| Pilotos: alta, edición, baja (US11) | `/team/drivers`, `new`, `[id]` | Sí | Pestaña Pilotos |
| Puntajes de la escudería | `/team/scores` | Sí | Más → Puntajes |
| Sanciones recibidas y acuse (US26) | `/team/inbox` | Sí, con insignia de pendientes | Pestaña Sanciones |
| Notificaciones de puntajes y acuse (US24) | `/team/notifications` | Sí, con insignia | Más → Notificaciones (la insignia se suma en "Más") |
| Mensajes con la FIA | `/team/messages`, `new`, `[id]` | Sí, con insignia de no leídos | Pestaña Mensajes |
| Historial de asignaciones | `/team/assignments` | Sí | Más → Asignaciones |
| Sitio público | `/team/public` | Sí | Más → Sitio público (y menú de usuario) |
| Mi cuenta, cerrar sesión | `/account` | Sí | Menú de usuario del encabezado |

### Administrador FIA (`fia_admin`)

| Funcionalidad web | Ruta | Teléfono | Nav en teléfono |
|---|---|---|---|
| Panel FIA | `/fia` | Sí | Pestaña Panel |
| Usuarios: alta, edición, baja, bloqueo (US6) | `/fia/users`, `new`, `[id]` | Sí | Más → Usuarios |
| Calendario: alta, edición, baja de eventos (US7, US8) | `/fia/calendar`, `new`, `[id]` | Sí | Pestaña Calendario |
| Resultados de carrera (US23) | `/fia/results`, `/fia/results/[id]` | Sí | Pestaña Resultados |
| Registro de controles técnicos (US21) | `/fia/results/controls/[id]` | Sí | Pestaña Resultados |
| Registro de pruebas de neumáticos (US19) | `/fia/results/tyres/[id]` | Sí | Pestaña Resultados |
| Sanciones: alta, detalle, seguimiento de acuses (US15, US26) | `/fia/sanctions`, `new`, `[id]` | Sí | Pestaña Sanciones |
| Seguimiento de acuses de puntajes (US24) | `/fia/notifications` | Sí | Más → Acuses |
| Mensajes con escuderías | `/fia/messages`, `new`, `[id]` | Sí, con insignia | Más → Mensajes (la insignia se suma en "Más") |
| Historial de asignaciones | `/fia/assignments` | Sí | Más → Asignaciones |
| Sitio público, Mi cuenta, cerrar sesión | `/fia/public`, `/account` | Sí | Más / menú de usuario |

**Conclusión:** no hay ninguna funcionalidad de la web ausente en el teléfono. La única brecha real es
de **navegación**: en el teléfono la barra inferior tenía 6 (público), 8 (escudería) y 9 (FIA)
pestañas, y a 320–375 px las etiquetas se cortan ("Mi es…", "Sanci…", "Notifi…").

## Riesgos propios de las plataformas nativas (revisados)

| Riesgo | Estado |
|---|---|
| ¿Empaqueta para Android e iOS? | `npx expo export --platform android` y `--platform ios` terminan sin errores (bundles Hermes de 3.3 MB y 3 MB). Se vuelve a correr al final. |
| APIs solo web (`window`, `document`, `localStorage`) fuera de `.web.tsx` | Ninguna. El único uso de `document` está en `src/theme/webStyles.web.ts`; la versión nativa (`webStyles.ts`) es vacía. `Platform.OS` solo en el login (teclado iOS) y en el corte de la barra lateral. |
| Descargas del calendario (`.ics`/PDF) | `Linking.openURL` sobre un endpoint público: en el teléfono abre el navegador del sistema, que descarga el archivo (o lo ofrece a Calendario en iOS). No necesita token. |
| URL de la API desde un dispositivo | `EXPO_PUBLIC_API_URL` (en `src/api/client.ts`), por defecto `http://localhost:8080`, que **no** sirve desde un teléfono: hay que poner la IP de la máquina (`http://192.168.x.x:8080`). Fuera de `__DEV__` el cliente exige HTTPS. Ya está documentado en `app/.env.example` (IP de la red local; `10.0.2.2` en el emulador de Android). La API escucha en todas las interfaces (`HTTP_ADDR=:8080`). |
| Token de sesión | Nativo: `expo-secure-store` (`tokenStore.ts`), enviado como `Authorization: Bearer`; el backend solo devuelve el token en el cuerpo si llega `X-Client-Platform: native` (cubierto por `tests/transport/auth_integration_test.go`). Web: cookie `HttpOnly` (`tokenStore.web.ts` no guarda nada). Nunca `AsyncStorage`. CORS no aplica en nativo. |
| Áreas seguras (notch, barra de gestos) | El encabezado y la barra inferior de `Tabs` (react-navigation) ya respetan `react-native-safe-area-context`; las pantallas van dentro del navegador, así que heredan los márgenes. |
| Teclado tapando formularios | Android ajusta la ventana (modo `resize` por defecto de Expo). En iOS `ScrollView` no se corre solo: se agrega `automaticallyAdjustKeyboardInsets` al `Screen` compartido, así todos los formularios (sanciones, usuarios, pilotos, mensajes) dejan ver el campo enfocado. El login ya usa `KeyboardAvoidingView`. |

## Cambio de navegación (único cambio de comportamiento, y solo de navegación)

- Una tabla de datos por rol (`src/navigation/roleTabs.ts`): nombre de ruta, título, ícono y si es
  **principal en el teléfono**. Un único componente `RoleTabs` dibuja las pestañas de los tres grupos
  (público, escudería, FIA) a partir de esa tabla: DRY, sin `Tabs.Screen` repetidos por layout.
- En el teléfono (todo lo nativo, y la web por debajo de 1024 px) se ven como máximo **5 pestañas**:
  4 principales + **Más**. "Más" es una ruta (`more.tsx` en cada grupo) que lista el resto con su
  insignia; la insignia de "Más" es la suma de las insignias que esconde, y "Más" queda resaltada
  cuando la pantalla activa es una de las que lista.
- En la barra lateral de escritorio (web ≥ 1024 px) no cambia nada: se ven todas las pestañas en el
  mismo orden y "Más" se oculta.
- Principales: público Inicio, Carreras, Puntajes, Sanciones · escudería Mi escudería, Pilotos,
  Sanciones, Mensajes · FIA Panel, Calendario, Resultados, Sanciones.
- Cada etiqueta de la barra del teléfono tiene como máximo 10 caracteres (lo que entra a 320 px):
  "Mi escudería" se muestra como "Escudería" solo en esa barra (`phoneLabel`).
- El filtro de rango de fechas que se corta a 375 px lo arregla el equipo controls: no se toca.

## Verificación

- **No hay dispositivos reales ni emuladores disponibles.** La evidencia es: (1) las exportaciones
  nativas `expo export --platform android|ios` sin errores, y (2) un recorrido automático con
  Playwright (Chromium) en modo táctil por **todas las rutas de los tres roles** a 320×640,
  375×740, 414×896, 740×360 (teléfono apaisado), 768×1024 (tableta) y 1366×900 (escritorio):
  errores de consola, excepciones, respuestas 4xx/5xx de la API, desbordamiento horizontal y
  **etiquetas de pestañas cortadas**. Se recorre también "Más" y la navegación desde ahí.
- Test de la app (`app/tests/navigation.test.cjs`): cada rol tiene como máximo 4 principales, los
  nombres de la tabla existen como rutas en `app/app/<grupo>/`, y la insignia de "Más" suma solo
  las pestañas que esconde.
- Login y permisos por rol en el camino nativo (token Bearer) ya los cubren los tests de integración
  del backend; se repite a mano con `curl` y `X-Client-Platform: native`.

## Resultado de la verificación

- Exportaciones nativas después del cambio: Android (Hermes, 3.3 MB) e iOS (Hermes, 3 MB) sin errores
  ni advertencias.
- Recorrido de 44 rutas (11 públicas, 21 FIA, 12 escudería) en 6 tamaños: 0 excepciones, 0 respuestas
  4xx/5xx de la API, 0 desbordamientos horizontales, 0 etiquetas de pestaña cortadas. Lo único que
  aparece es el 401 esperado de `/auth/me` cuando no hay sesión.
- Antes del cambio, a 320 px: 8 etiquetas cortadas en la barra FIA, 8 en la de escudería y
  "Sanciones" en la pública.
- "Más" en el teléfono: cada destino oculto se abre tocando, para los tres roles; en escritorio la
  barra lateral sigue con todas las pestañas y sin "Más".
- Reflejo de cambios: la FIA abre una conversación por la API mientras la escudería está en el
  teléfono; al cambiar de pestaña sube la insignia y la conversación aparece en la lista.
- `app.json` fija la orientación vertical en iOS y Android; el apaisado (740×360) se probó en la web.
- Pendiente fuera de US18: el filtro de fechas se corta a 375 px (lo arregla el equipo controls) y,
  a 320 px, la insignia "Corregido" de una tarjeta de control técnico sobresale unos píxeles del
  borde de la tarjeta.

## Preguntas abiertas

Ninguna que bloquee: el cambio es solo de navegación y no quita ninguna pantalla.
