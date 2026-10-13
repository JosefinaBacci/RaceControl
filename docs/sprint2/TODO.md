# TODO del Sprint 2

Pendientes que no pertenecen a ninguna user story del sprint, pero que conviene resolver antes de un
despliegue real.

## Segregar los permisos de base de datos (rol de aplicación)

**Estado:** pendiente. **Decisión relacionada (F5 de la revisión del modelo):** se conserva el
trigger de append-only.

Hoy la API conecta con el usuario `racecontrol`, que en el contenedor de desarrollo es
**superusuario, `BYPASSRLS` y dueño de las 19 tablas**. Consecuencia: `REVOKE` y RLS no protegen
nada, porque un dueño no puede quitarse sus propios privilegios y un superusuario saltea RLS.

Trabajo pendiente, para cuando exista un despliegue con la base en serio:

1. Crear un rol de login dedicado (`racecontrol_app`), `NOSUPERUSER NOBYPASSRLS`, que no sea dueño de
   ninguna tabla.
2. `GRANT USAGE ON SCHEMA public`, `GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA
   public` y `ALTER DEFAULT PRIVILEGES ... ON TABLES` para que las tablas nuevas hereden las grants.
3. `REVOKE UPDATE, DELETE` sobre las tablas de historial y auditoría (`driver_team_assignments`,
   `score_change_log`, `score_change_log_items`, `score_acknowledgements`, `login_attempts`), como
   segunda capa además del trigger.
4. Una segunda URL de conexión: la API usa el rol nuevo; las migraciones, los seeds y `create-admin`
   siguen con el dueño.
5. Documentar el cambio en `docs/sprint1/03-arquitectura-roles-seguridad.md`.

Detalles ya verificados, para no re-descubrirlos:

- Las columnas `GENERATED ALWAYS AS IDENTITY` **no** exigen `USAGE` sobre la secuencia: alcanza con el
  `INSERT` sobre la tabla.
- Las funciones de los triggers se ejecutan con los privilegios del invocador (`SECURITY INVOKER`) y
  `EXECUTE` es `PUBLIC` por defecto, así que un `GRANT SELECT` alcanza para que sigan funcionando.
- Las grants son **por base**: se pierden en cada `make db-reset` (base nueva). El bloque tiene que
  vivir en una migración o en el target, no en el init del contenedor, que corre una sola vez al
  crear el *cluster*.
- `CREATE ROLE` es de *cluster* y no de base: desde una migración que corre en dev y en test necesita
  un guard `IF NOT EXISTS`.
- El rol **no reemplaza** al trigger. Las migraciones, los seeds y `psql` corren como dueño, y a esos
  solo los cubre el trigger; por eso F5 decidió conservarlo y dejar el rol como capa adicional.

**Por qué no ahora:** no hay despliegue a producción (el `compose.yaml` se declara solo para
desarrollo) y el control documentado para "escalada de privilegios desde la app" es el RBAC del
backend, no los privilegios de base. Es trabajo de infraestructura, no de una user story del sprint.

### Append-only ya resuelto

Los hechos congelados ya tienen su guardia (`reject_mutation()`): `driver_team_assignments` (desde
`000007`) y `score_change_log`, `score_change_log_items`, `score_acknowledgements` y `login_attempts`
(desde `000017`). `notifications` queda fuera a propósito: tiene estado "leída" y se actualiza.
