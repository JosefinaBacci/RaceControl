# RaceControl

Plataforma de gestión para la FIA (APS 2026, Enunciado 1): calendario, escuderías, pilotos,
controles técnicos, sanciones, puntajes y comunicación entre la FIA y las escuderías. Una sola
base de código sirve la app web y la app móvil.

| Parte | Tecnología |
|-------|------------|
| Backend | Go 1.27, `net/http` + `chi`, sesiones opacas con Argon2id |
| Base de datos | PostgreSQL 18, `pgx`, `sqlc`, `golang-migrate` |
| App | Expo + `expo-router` (iOS, Android y web) |

La justificación de cada decisión está en [`docs/sprint1/`](docs/sprint1/) y las convenciones de
trabajo en [`AGENTS.md`](AGENTS.md).

## Requisitos

- Go 1.27
- Node.js 22 y npm
- Podman o Docker, para la base de datos
- `make` (en Windows, desde Git Bash)

## Puesta en marcha

```bash
cp .env.example .env              # completar SEED_DEMO_PASSWORD (12+ caracteres) para tener cuentas demo
make db-up                        # Postgres 18 en un contenedor
make migrate                      # esquema
make seed                         # categorías, escuderías y cuentas demo
make backend-run                  # API en http://localhost:8080
```

En otra terminal:

```bash
cp app/.env.example app/.env
make app-install
make app-web                      # app en el navegador; `make app-start` para Expo Go en el teléfono
```

### Cuentas demo

`make seed` crea estas cuentas, todas con la contraseña de `SEED_DEMO_PASSWORD`. Nunca se crean
con `APP_ENV=production`.

| Usuario | Rol |
|---------|-----|
| `fia.admin` | Administrador FIA |
| `ferrari.admin` | Administrador de Scuderia Ferrari |
| `redbull.admin` | Administrador de Red Bull Racing |
| `mercedes.admin` | Administrador de Mercedes-AMG Petronas |

En producción, el primer administrador se crea con
`make create-admin ADMIN_USERNAME=<usuario>`.

## Comandos útiles

| Comando | Qué hace |
|---------|----------|
| `make db-reset` | Borra la base y la recrea: migrate + seed |
| `make sqlc` | Regenera el código Go después de editar `backend/sql/` |
| `make check` | Lint, typecheck y tests de todo el proyecto |
| `make help` | Lista todos los comandos |

Los tests de integración del backend corren contra una base propia, `racecontrol_test`, en el
mismo contenedor: `make check` la crea y la actualiza sola si `TEST_DATABASE_URL` está definida
en `.env` (viene en `.env.example`). Así los tests nunca dejan usuarios en la base de desarrollo.
Sin esa variable, los tests de integración se omiten.

## HTTPS en producción (US17)

Los mensajes viajan cifrados mediante HTTPS. Hay dos configuraciones admitidas:

- **TLS directo:** definir `TLS_CERT_FILE` y `TLS_KEY_FILE` con rutas al certificado y a su clave
  privada. La API sirve HTTPS (TLS 1.2 o posterior); no versionar esos archivos.
- **Proxy HTTPS:** terminar TLS en el proxy, configurar `TRUST_PROXY_HEADERS=true` y sobrescribir
  `X-Forwarded-Proto` con `https`. La API debe quedar accesible únicamente desde ese proxy mediante
  la red privada o firewall; nunca confiar en encabezados de clientes directos.

Usar `APP_ENV=production`, `SEED_DEMO_PASSWORD=` y `CORS_ALLOWED_ORIGINS` con el origen HTTPS de la
app. La configuración rechaza producción sin TLS ni proxy confiable y certificados incompletos.
Las solicitudes HTTP planas se rechazan con 426; las cookies de producción son Secure y HttpOnly.
Configurar `EXPO_PUBLIC_API_URL=https://...` antes de compilar la app: una compilación de producción
rechaza URLs HTTP. El entorno de desarrollo conserva HTTP local.

El cifrado es **en tránsito**. No es cifrado de extremo a extremo ni cifrado de la base de datos.
No se reutilizan contraseñas como claves: la sesión y los permisos determinan quién puede acceder.
La suite verifica una conexión TLS real con certificado confiado por el cliente de prueba.

## Historial y conversaciones entre escuderías

Aplicar `make migrate` antes de iniciar la API actualizada (migración 000029).
La pestaña Asignaciones muestra el historial global y filtrado a la FIA, y solo los movimientos
propios a cada escudería, incluidas sus bajas y transferencias. No se reescribe el historial.

En Mensajes se puede elegir FIA u otra escudería. Un hilo entre dos escuderías es privado para
ambas: la FIA y las demás escuderías no pueden consultarlo ni escribir. Los participantes no se
pueden modificar. «Leído» significa que al menos un administrador de la parte destinataria abrió
el mensaje; leerlo desde otra cuenta de la propia escudería no cuenta. El chat activo se actualiza
cada diez segundos y también permite actualización manual.
