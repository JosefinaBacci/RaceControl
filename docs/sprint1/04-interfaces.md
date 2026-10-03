# US4 — Diseño de interfaz por rol

## Objetivo

Definir qué ve y qué puede hacer cada rol, armar un sistema de diseño común y tener una primera
versión navegable de cada interfaz. Donde el backend todavía no existe, las pantallas usan datos de
ejemplo.

## 1. Estructura

Es una sola app Expo para iOS, Android y web. Cada rol tiene su grupo de rutas y la sesión decide a
cuál se entra:

| Rol | Rutas | Acceso |
|-----|-------|--------|
| Público | `app/(public)/` | Sin sesión, es la portada |
| Administrador FIA | `app/fia/` | Cuenta `fia_admin` |
| Administrador de escudería | `app/team/` | Cuenta `team_admin` |

El login, "Mi cuenta" y el menú de usuario son comunes a todos.

`RoleGate` redirige a quien entra a una sección que no le corresponde, pero eso es solo comodidad:
quien protege los datos es la API, que responde `401` o `403` según la matriz de permisos de US3.

## 2. Pantallas por rol

Las pantallas con datos de ejemplo muestran un aviso para que no se confundan con datos reales.

**Público**

| Pantalla | Contenido | Datos |
|----------|-----------|-------|
| Inicio | Próxima carrera con cuenta regresiva y circuito, top 4 de F1, últimas sanciones | Ejemplo |
| Carreras | Calendario filtrable por categoría, próximas y pasadas | Ejemplo |
| Puntajes | Clasificación por categoría | Ejemplo |
| Sanciones | Sanciones publicadas, filtrables por categoría | Ejemplo |

**Administrador FIA**

| Pantalla | Contenido | Datos |
|----------|-----------|-------|
| Panel | Resumen y accesos rápidos | Cuentas activas reales, el resto de ejemplo |
| Usuarios | Listar, buscar, crear, editar, desactivar y reactivar cuentas (US6) | Reales |
| Calendario | Calendario completo y alta de eventos | Ejemplo |
| Resultados | Puntajes y sanciones con el acuse de cada escudería | Ejemplo |
| Comunicados | Mensajes a las escuderías y su lectura | Ejemplo |

**Administrador de escudería**

| Pantalla | Contenido | Datos |
|----------|-----------|-------|
| Mi escudería | Ficha, pilotos, puntos, acuses pendientes y próximos eventos | Ejemplo |
| Pilotos | Titulares y suplentes | Ejemplo |
| Acuses | Sanciones y puntajes para confirmar | Ejemplo (el acuse no se guarda) |
| Comunicados | Avisos de la FIA | Ejemplo |

La escudería sale de la sesión, no de un parámetro, así que cada administrador solo ve la suya.

**Comunes**

| Pantalla | Contenido | Datos |
|----------|-----------|-------|
| Login | Usuario y contraseña, validación, errores por campo y aviso de bloqueo (US5) | Reales |
| Mi cuenta | Datos de la cuenta y cambio de contraseña | Reales |
| Menú de usuario | Nombre, rol, "Mi panel", "Mi cuenta" y "Cerrar sesión" | Reales |

## 3. Navegación

Cada rol tiene su barra de pestañas. En el celular va abajo y en web, a partir de 1024 px, pasa a
una barra lateral. El encabezado muestra el logo y, a la derecha, "Ingresar" o el menú de usuario.
El alta y la edición de usuarios se abren sobre el listado, con su propia URL
(`/fia/users/new`, `/fia/users/[id]`).

## 4. Sistema de diseño

Los valores están en `app/src/theme` y los componentes en `app/src/components`. No usamos ningún kit
de UI externo, así se ve igual en las tres plataformas.

Elegimos un tema oscuro con rojo como acento, inspirado en las transmisiones de Fórmula 1. Hacer
un tema claro sería cambiar los valores de `colors.ts`, sin tocar componentes.

| Token | Valor |
|-------|-------|
| `background` | `#0B0B0F` |
| `surface` | `#15151B` |
| `text` / `textMuted` | `#F5F5F7` / `#9C9CA8` |
| `accent` | `#E10600` |
| `success` / `warning` / `danger` / `info` | `#3DDC84` / `#FFB020` / `#FF5A52` / `#5AA9FF` |

El texto principal y el secundario cumplen el contraste AAA sobre el fondo, y los botones rojos con
texto blanco cumplen AA. La excepción es el rojo usado como texto chico en los enlaces "Ver todas"
(4:1), que queda como pendiente.

También definimos una escala tipográfica con nombre (de `hero` a `overline`), una escala de
espaciado de 4 a 32 px, un ancho máximo de contenido de 960 px y un área táctil mínima de 44 px.

Componentes principales: `Screen`, `Card`, `Button`, `TextField`, `ChipGroup`, `Badge`, `Notice`,
`ListItem`, `Avatar`, `StatTile`, `HeroCard` y `Countdown`. Los bloques que comparten varios roles
(listas de eventos y sanciones, tabla de posiciones, tarjeta de piloto) están en `app/src/features`.

Imágenes:

- Fotos por categoría de Wikimedia Commons, con los créditos en `app/assets/photos/CREDITS.md`. Los
  degradados para que el texto se lea se aplican en la app, sin editar las fotos.
- Ilustraciones SVG propias de banderas, circuitos y un auto.
- Logos de la FIA y de las escuderías en los roles, los resultados y las cuentas demo. Créditos en
  `app/assets/logos/CREDITS.md`.

## 5. Decisiones

- Los errores de la API se muestran en el campo que corresponde.
- La desactivación de una cuenta se confirma en la misma fila o ficha, porque el `Alert` de React
  Native no funciona en web.
- Un administrador no ve el botón de baja ni el cambio de rol sobre su propia cuenta (el backend
  igual lo rechaza).
- La búsqueda de usuarios espera 300 ms antes de consultar la API.
- El login valida el formato antes de enviar.

En accesibilidad: botones, pestañas y campos tienen rol y etiqueta (también los botones que son solo
un ícono), y los errores se anuncian a los lectores de pantalla.

## 6. Pendientes

- Conectar calendario, puntajes, sanciones, pilotos y comunicados a la API en el Sprint 2. Solo
  cambian los hooks de `src/data`.
- Usar un rojo más claro para el texto chico.
- Cargar los logos de cinco escuderías.
- Foto de perfil propia.

## Criterios de éxito

| Criterio | Estado |
|----------|--------|
| Interfaz clara, intuitiva y fácil de usar | Pestañas por rol, un mismo sistema de diseño y errores en contexto. Se valida en la demo |
| La versión inicial muestra las funcionalidades de cada rol | ✅ Las tres interfaces completas, con datos reales donde ya hay backend |
