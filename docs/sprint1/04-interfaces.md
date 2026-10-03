# US4 — Diseño de interfaz por rol

## Objetivo

Definir qué ve y qué puede hacer cada rol, construir un sistema de diseño común a las tres
interfaces y entregar una primera versión navegable de cada una, con datos de ejemplo donde la
funcionalidad todavía no existe en el backend.

## 1. Una sola interfaz, tres puntos de entrada

La app es una sola base de código Expo que corre en iOS, Android y web (ver `01-stack.md`). Los
tres roles no son tres apps: son tres grupos de rutas con su propia navegación, y la sesión decide
a cuál se entra.

| Rol | Rutas | Cómo se llega |
|-----|-------|---------------|
| Público | `app/(public)/` | Sin sesión. Es la portada de la app |
| Administrador FIA | `app/fia/` | Login con una cuenta `fia_admin` |
| Administrador de escudería | `app/team/` | Login con una cuenta `team_admin` |

Comparten tres pantallas: el login (`app/login.tsx`), "Mi cuenta" (`app/account.tsx`) y el menú de
usuario del encabezado.

**El acceso lo controla el backend, no la navegación.** `RoleGate` redirige a quien entra a una
sección que no le corresponde (un `team_admin` que escribe `/fia` en la barra vuelve a `/team`),
pero eso es ergonomía: los datos de cada sección los protege la API, que responde `401` o `403`
según la matriz de permisos de US3.

## 2. Funcionalidades por rol

La columna **Datos** indica si la pantalla ya usa la API o muestra datos de ejemplo. Las pantallas
con datos de ejemplo lo avisan con un cartel ("Versión inicial: los datos de esta pantalla son de
ejemplo"), para que nadie confunda la maqueta con información real.

### 2.1 Público

| Pantalla | Qué muestra | Datos |
|----------|-------------|-------|
| Inicio | Próxima carrera con cuenta regresiva, mapa del circuito y datos técnicos; los cuatro primeros del campeonato de F1; las últimas sanciones | Ejemplo |
| Carreras | Calendario filtrable por categoría y por próximos o pasados: carreras, pruebas de neumáticos y controles técnicos | Ejemplo |
| Puntajes | Clasificación del campeonato por categoría, con el líder destacado | Ejemplo |
| Sanciones | Sanciones publicadas a pilotos y escuderías, filtrables por categoría | Ejemplo |

Desde cualquier pantalla pública, el botón "Ingresar" del encabezado lleva al login. Con una
sesión abierta, ese lugar lo ocupa el menú de usuario, que incluye "Mi panel".

### 2.2 Administrador FIA

| Pantalla | Qué permite | Datos |
|----------|-------------|-------|
| Panel | Resumen: próximos eventos, sanciones y puntajes sin acuse, cuentas activas, accesos rápidos | Cuentas activas reales; el resto, ejemplo |
| Usuarios | Listar, buscar y filtrar cuentas por rol y estado; crear, editar, desactivar y reactivar (US6) | **Reales** |
| Calendario | Calendario completo con el botón de nuevo evento | Ejemplo |
| Resultados | Puntajes publicados con el acuse de cada escudería, y sanciones emitidas | Ejemplo |
| Comunicados | Mensajes a las escuderías con su confirmación de lectura | Ejemplo |

### 2.3 Administrador de escudería

| Pantalla | Qué permite | Datos |
|----------|-------------|-------|
| Mi escudería | Ficha de la escudería, sus pilotos, puntos, acuses pendientes y próximos eventos | Ejemplo |
| Pilotos | Titulares y suplentes de su escudería, con el botón de agregar piloto | Ejemplo |
| Acuses | Sanciones y puntajes que la escudería tiene que confirmar, con el botón "Acusar recibo" | Ejemplo (el acuse funciona en la sesión, no se guarda) |
| Comunicados | Avisos y reglamentos enviados por la FIA | Ejemplo |

Todas las pantallas de escudería muestran solo los datos de **su** escudería, que sale de la
sesión y no de un parámetro, igual que en el backend.

### 2.4 Pantallas comunes

| Pantalla | Qué permite | Datos |
|----------|-------------|-------|
| Login | Usuario y contraseña, validación antes de enviar, errores por campo, aviso de bloqueo tras cinco intentos, acceso al sitio público sin cuenta (US5) | **Reales** |
| Mi cuenta | Datos de la cuenta y cambio de la contraseña propia, pidiendo la actual | **Reales** |
| Menú de usuario | Nombre, rol y escudería; accesos a "Mi panel", "Mi cuenta" y "Cerrar sesión" | **Reales** |

## 3. Navegación

- **Pestañas por rol.** Cada rol tiene su barra de pestañas con íconos y texto.
- **Barra inferior en el teléfono, lateral en la web.** En pantallas web de 1024 px o más, las
  pestañas pasan a una barra lateral de 240 px: en un monitor ancho, una barra inferior queda
  lejos y desperdicia espacio. En el teléfono la barra inferior queda al alcance del pulgar.
- **Encabezado común:** logo de RaceControl a la izquierda y, a la derecha, "Ingresar" o el menú
  de usuario.
- **Pantallas apiladas** dentro de una sección cuando hace falta: el alta y la edición de usuarios
  se abren sobre el listado y tienen su propia URL en la web (`/fia/users/new`,
  `/fia/users/[id]`).

## 4. Sistema de diseño

Vive en `app/src/theme` (valores) y `app/src/components` (componentes). No se usa ningún kit de
UI externo: los componentes salen de los primitivos de React Native, para que el diseño sea el
mismo en las tres plataformas.

### 4.1 Por qué un tema oscuro

La referencia visual son las transmisiones y los gráficos de Fórmula 1: fondo oscuro, datos en
blanco y un único color de acento. El tema oscuro hace que el rojo del acento y las fotos de las
categorías resalten, y se lee bien en pantallas de boxes o paddock con poca luz. Se eligió un solo
tema para el Sprint 1; un tema claro es un cambio de valores en `colors.ts`, no de componentes.

### 4.2 Colores

| Token | Valor | Uso |
|-------|-------|-----|
| `background` | `#0B0B0F` | Fondo de la app |
| `surface` / `surfaceRaised` / `surfaceMuted` | `#15151B` / `#1D1D25` / `#26262F` | Tarjetas, campos, estados presionados |
| `border` | `#2A2A33` | Bordes y separadores |
| `text` / `textMuted` | `#F5F5F7` / `#9C9CA8` | Texto principal y secundario |
| `accent` | `#E10600` | Acciones principales, selección, marca |
| `success` / `warning` / `danger` / `info` | `#3DDC84` / `#FFB020` / `#FF5A52` / `#5AA9FF` | Estados; cada uno con una variante translúcida para fondos |

Contraste medido según WCAG:

| Combinación | Contraste | Nivel |
|-------------|-----------|-------|
| `text` sobre `background` | 18,0:1 | AAA |
| `textMuted` sobre `background` | 7,2:1 | AAA |
| `textMuted` sobre `surface` | 6,7:1 | AA |
| Texto blanco sobre `accent` (botones) | 5,0:1 | AA |
| `danger` sobre `surface` | 5,9:1 | AA |
| `accent` sobre `background` | 4,0:1 | AA solo para texto grande |

La última fila es una deuda registrada: los enlaces "Ver todas" usan el rojo de acento en texto
de 12 px, por debajo del 4,5:1 que pide AA para texto chico (ver sección 7).

### 4.3 Tipografía, espaciado y tamaños

- **Escala tipográfica** de ocho variantes con nombre (`hero`, `display`, `title`, `heading`,
  `body`, `bodyStrong`, `caption`, `overline`), de 34 a 11 px. Las pantallas eligen una variante,
  nunca un tamaño suelto.
- **Espaciado** en una escala de 4 a 32 px (`xs` a `xxl`) y radios de 6, 10 y 16 px más el de píldora.
- **Ancho máximo de contenido** de 960 px, centrado: en un monitor ancho las líneas no se estiran.
- **Área táctil mínima de 44 px** en botones y campos, la recomendada para el uso con el dedo.

### 4.4 Componentes

| Componente | Para qué |
|------------|----------|
| `Screen` | Estructura de cada pantalla: título, subtítulo, acción del encabezado y desplazamiento |
| `Card`, `CardLink` | Agrupar contenido y enlazar a la pantalla completa |
| `Button`, `IconButton` | Acción principal, secundaria o discreta; estado de carga y deshabilitado |
| `TextField` | Campo con ícono, error debajo y botón para mostrar la contraseña |
| `ChipGroup` | Filtros y selección única: pestañas subrayadas o píldoras |
| `Badge` | Estado corto con tono (activo, sin acuse, sanción) |
| `Notice` | Avisos de información, éxito, advertencia o error |
| `ListItem` | Fila con elemento inicial, título, subtítulo y elemento final |
| `Avatar` | Iniciales con el color de la escudería, o una imagen |
| `StatTile`, `HeroCard`, `Countdown` | Métricas, fichas destacadas con foto y cuenta regresiva |

Encima de estos, `app/src/features` arma bloques que comparten varios roles: listas de eventos,
de sanciones y de comunicados, la tabla de posiciones, la tarjeta de piloto, el contenido que se
carga desde la API (`RemoteContent`) y los campos de una cuenta (`AccountFields`).

### 4.5 Imágenes

- **Fotos por categoría** (F1, F2, F3, F1 Academy) de Wikimedia Commons, con sus licencias en
  `app/assets/photos/CREDITS.md`. Los degradados que aseguran la legibilidad del texto se aplican
  en la app, sin modificar las fotos.
- **Ilustraciones SVG propias** de banderas, trazados de circuitos y un auto, que escalan sin
  perder calidad y no suman peso de imágenes.
- **Logos** de la FIA y de las escuderías junto al rol de cada cuenta y en la pantalla de
  resultados; las cuentas demo de cada organización los usan como avatar. Las demás cuentas
  muestran sus iniciales. Créditos en `app/assets/logos/CREDITS.md`.

## 5. Decisiones de interacción

| Decisión | Motivo |
|----------|--------|
| Los errores de la API se muestran en el campo que corresponde | Un "ya hay un usuario con ese nombre" debajo del usuario se corrige sin buscar qué falló |
| La desactivación se confirma en el lugar (en la fila o en la ficha), no con un diálogo | El `Alert` de React Native no funciona en la web; confirmar en el lugar funciona igual en las tres plataformas |
| Botones de acción críticos protegidos sobre la propia cuenta | Un administrador no ve el botón de baja ni el selector de rol en su cuenta; el backend lo rechaza igual |
| Búsqueda con 300 ms de espera | Evita un pedido a la API por cada tecla |
| Cartel de "datos de ejemplo" en las pantallas sin backend | La demo no se confunde con datos reales |
| El formulario de login valida antes de enviar | Los errores de formato se ven al instante, sin ida y vuelta al servidor |

## 6. Accesibilidad

- Roles y etiquetas accesibles en botones, pestañas, menús y campos (`accessibilityRole`,
  `accessibilityLabel`), incluidos los botones que son solo un ícono ("Editar ferrari.admin").
- Los errores y avisos se anuncian a los lectores de pantalla (`accessibilityLiveRegion`).
- Contraste AA o superior en todo el texto, salvo el caso registrado en la sección 7.
- Área táctil mínima de 44 px.

## 7. Deuda y pendientes

| Pendiente | Cómo se resuelve |
|-----------|------------------|
| Las pantallas de calendario, puntajes, sanciones, pilotos y comunicados usan datos de ejemplo | Se conectan a la API a medida que el Sprint 2 implemente esos módulos; los hooks de `src/data` son el único punto que cambia |
| El rojo de acento como texto chico queda en 4,0:1 | Usar una variante más clara del rojo solo para texto, sin cambiar el de los botones |
| Logos de cinco escuderías sin cargar | Backlog del Sprint 2 en `TODO.md` |
| Foto de perfil propia | Backlog del Sprint 2 en `TODO.md` |

## Criterios de éxito

| Criterio | Estado |
|----------|--------|
| Interfaz clara, intuitiva y fácil de usar | Navegación por pestañas por rol, un mismo sistema de diseño en todas las pantallas y errores en contexto; se valida con el profesor en la demo |
| La versión inicial muestra las funcionalidades mockeadas de cada rol | ✅ Sección 2: las tres interfaces completas, con datos reales donde ya hay backend |
