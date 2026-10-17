# Revisión de notas — 2026-10-17

Se contrastaron las notas de revisión con la consigna del Sprint 2 y el código existente.
Tras la confirmación del usuario se implementaron también las ampliaciones de las notas.

| Nota | Resultado |
|---|---|
| 1. Validaciones | Corregidas horas imposibles, fechas reales, números de piloto inválidos y formato/rango decimal de puntajes. La API conserva las validaciones obligatorias y de permisos. |
| 2. Circuitos difíciles de recorrer | Selector desplegable con búsqueda y desplazamiento vertical. |
| 3. Eventos en el pasado | Alta y reprogramación rechazadas por API; aviso previo en el alta de la app. |
| 4 y 10. Detalle público de carrera | Nuevo detalle de solo lectura con circuito, horarios y los mismos resultados públicos. Clasificación de escuderías ya disponible en Puntajes. |
| 5. Sitio público escondido | Acceso visible desde los menús de FIA y escudería. |
| 6. No se puede volver del chat | Corregida la causa compartida en Screen y regreso explícito a la bandeja por rol. |
| 7. Puntos iniciales | Cero para una carrera nueva. Se mantiene carga manual: el proyecto decidió no persistir orden de llegada, y la consigna no define una escala de puntos por categoría. |
| 8. Carreras sin resultados | Distintivo «Sin puntajes» en la lista de carga de resultados, calculado por temporada. Además hay aviso y acceso directo en el panel FIA; no requiere notificaciones push. |
| 9. Listas extensas | Circuitos y resultados por carrera usan desplegable con búsqueda. |
| 11. Número opcional | Número de temporada obligatorio en API y formulario. Se rechazan valores inválidos y borrados silenciosos; los registros históricos no se modifican. |
| 12. Historial de asignaciones | Historial global y filtrado para FIA y movimientos propios para escuderías. Incluye bajas y transferencias; probado contra acceso ajeno. |
| 13. Buscar otros pilotos | Acceso directo al directorio de solo lectura desde la gestión de pilotos de la escudería. |
| 14. Diseño de mensajería | Bandeja con búsqueda, selector de destinatario, burbujas propias y ajenas diferenciadas, autor, horario y estado de entrega; navegación corregida. |
| 15. Ticks de visto | Enviado y Leído visibles al remitente. Solo cuenta la lectura de la parte destinataria; el chat activo actualiza cada diez segundos. |
| 16. Mensajes entre escuderías | Implementados con migración 000029. Solo las dos escuderías acceden; FIA y terceros quedan excluidos. |
| 17. Cifrado | HTTPS directo o mediante proxy confiable, con prueba TLS real y rechazo de HTTP en producción. Cifrado en tránsito; no se usa la contraseña como clave ni se afirma cifrado de extremo a extremo o de almacenamiento. |

## Verificación

- `make check` con PostgreSQL en `racecontrol_test`: vet, formato Go, tests unitarios e integración,
  typecheck y regresiones de formularios. Docker existente iniciado; base de desarrollo sin modificar.
- Expo SDK 57: rutas generadas actualizadas por el servidor de desarrollo antes del typecheck.
- Navegador: calendario público, apertura de carrera sin sesión, mensaje de resultados pendientes
  y regreso al calendario; selector de resultados visible.
- Exportaciones: PDF e ICS filtrados por F1 descargados de la API de pruebas. PDF abierto y
  renderizado localmente, fechas y nombres legibles; 19 eventos en los dos archivos.
- Se probó TLS real en integración y la interfaz de elección de destinatarios con una cuenta temporal.
- No se probó la app en un dispositivo físico ni se publicó a un servidor externo.
- Migración 000029 aplicada en racecontrol_test. La base local de desarrollo sigue en versión 1;
  debe actualizarse con `make migrate` y `make seed` antes de usar la nueva API contra ella.
