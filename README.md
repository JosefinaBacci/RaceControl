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

Los tests de integración del backend corren contra la base de `DATABASE_URL`; sin esa variable
se omiten, así que conviene correr `make check` con la base levantada.
