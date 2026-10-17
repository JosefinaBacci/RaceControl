# Revisión de notas — 2026-10-17

Se contrastaron las notas de revisión con la consigna del Sprint 2 y el código existente.
Se priorizaron las US y los errores reproducibles; las propuestas adicionales quedan explícitas.

| Nota | Resultado |
|---|---|
| 1. Validaciones | Corregidas horas imposibles, fechas reales, números de piloto inválidos y formato/rango decimal de puntajes. La API conserva las validaciones obligatorias y de permisos. |
| 2. Circuitos difíciles de recorrer | Selector desplegable con búsqueda y desplazamiento vertical. |
| 3. Eventos en el pasado | Alta y reprogramación rechazadas por API; aviso previo en el alta de la app. |
| 4 y 10. Detalle público de carrera | Nuevo detalle de solo lectura con circuito, horarios y los mismos resultados públicos. Clasificación de escuderías ya disponible en Puntajes. |
| 5. Sitio público escondido | Acceso visible desde los menús de FIA y escudería. |
| 6. No se puede volver del chat | Corregida la causa compartida en Screen y regreso explícito a la bandeja por rol. |
| 7. Puntos iniciales | Cero para una carrera nueva. Se mantiene carga manual: el proyecto decidió no persistir orden de llegada, y la consigna no define una escala de puntos por categoría. |
| 8. Carreras sin resultados | Distintivo «Sin puntajes» en la lista de carga de resultados, calculado por temporada. No es una notificación push. |
| 9. Listas extensas | Circuitos y resultados por carrera usan desplegable con búsqueda. |
| 11. Número opcional | El modelo actual permite alta sin número de temporada. No se impuso obligatoriedad a partir de una pregunta de las notas; sí se corrigen valores inválidos y borrados silenciosos. |
| 12. Historial de asignaciones | El historial se conserva en la base; falta exponerlo por escudería y globalmente. Ampliación pendiente. |
| 13. Buscar otros pilotos | Acceso directo al directorio de solo lectura desde la gestión de pilotos de la escudería. |
| 14. Diseño de mensajería | Corregidas navegación y actualización; rediseño visual integral pendiente. |
| 15. Ticks de visto | La app tiene no leídos para el receptor; falta exponer la lectura al remitente. Ampliación pendiente. |
| 16. Mensajes entre escuderías | El modelo de US17 es escudería–FIA. Extenderlo exige migración y revisión de participantes/permisos; pendiente. |
| 17. Cifrado | La arquitectura prevé HTTPS en el proxy. Pendiente validarlo en un despliegue real; no se agregó cifrado con la contraseña ni se afirma cifrado de extremo a extremo. |

## Verificación

- `make check` con PostgreSQL en `racecontrol_test`: vet, formato Go, tests unitarios e integración,
  typecheck y regresiones de formularios. Docker existente iniciado; base de desarrollo sin modificar.
- Expo SDK 57: rutas generadas actualizadas por el servidor de desarrollo antes del typecheck.
- Navegador: calendario público, apertura de carrera sin sesión, mensaje de resultados pendientes
  y regreso al calendario; selector de resultados visible.
- Exportaciones: PDF e ICS filtrados por F1 descargados de la API de pruebas. PDF abierto y
  renderizado localmente, fechas y nombres legibles; 19 eventos en los dos archivos.
- No se probó la app en un dispositivo iOS/Android ni un despliegue HTTPS.
