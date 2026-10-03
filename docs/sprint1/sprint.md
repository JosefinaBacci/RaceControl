# Sprint 1 — Lista de tareas

## Estado

| US | Nombre | SP | Valor | Est. | Tiempo real registrado | Estado |
|----|--------|----|-------|-------|------------------------|--------|
| US1 | Selección de stack y configuración del entorno | 5 | 8 | 6h | 1.5h | ☑ |
| US2 | Diseño del modelo de datos | 5 | 8 | 9h | 3.5h | ☑ |
| US3 | Arquitectura, roles y seguridad | 5 | 8 | 9h | 0.2h | ☑ |
| US4 | Diseño de interfaz por rol | 5 | 8 | 8h | 1.1h | ☑ |
| US5 | Login de usuarios | 3 | 13 | 9h | 1.0h | ☑ |
| US6 | Gestión de usuarios | 8 | 21 | 12h | 1.2h | ☑ |

**Total estimado:** 53h · **Total real registrado:** 8.5h · **Desviación:** −84 %  
El tiempo registrado representa aproximadamente el **16 % del tiempo inicialmente estimado**.

Las horas reales de US4, US5 y US6 se obtuvieron a partir de las marcas de tiempo de los commits: desde el último commit correspondiente a la historia anterior hasta el último commit de la propia historia.

Cuando varias historias se trabajaron durante una misma sesión, como ocurrió con US4 y US5, el tiempo se distribuyó según el momento en que se realizó el commit correspondiente a cada una.

Esta medición debe considerarse **aproximada**, ya que refleja principalmente el tiempo observable de implementación y versionado. No incluye necesariamente todo el tiempo dedicado previamente a análisis, discusión de alternativas, lectura, diseño conceptual ni pruebas manuales realizadas fuera de las sesiones registradas mediante commits.

☑ cerrada · ◐ en curso · ☐ sin empezar

---

# Análisis de estimación del Sprint 1

El Sprint presenta una diferencia considerable entre las estimaciones iniciales y el tiempo real registrado.

No obstante, no resulta correcto interpretar esta desviación como evidencia de que todas las historias futuras deben estimarse un **84 % por debajo**.

Los datos muestran comportamientos diferentes según el tipo de tarea.

## Historias de diseño y documentación

US1, US2 y US3 fueron significativamente sobreestimadas.

Parte de esta diferencia se explica porque:

- Varias decisiones se tomaron en paralelo
- Algunas decisiones de historias posteriores surgieron durante historias anteriores
- El tiempo basado en commits no captura todo el proceso de análisis
- La documentación resultó más rápida de producir que lo previsto

## Historias de implementación

US5 y US6 también tuvieron tiempos inferiores a los esperados.

Una causa relevante fue que, al momento de implementarlas, ya se encontraba disponible:

- Una arquitectura definida
- Una base de datos preparada
- Un sistema de capa
- Infraestructura de autenticación
- Convenciones de desarrollo
- Componentes reutilizables

Esto redujo considerablemente el coste marginal de agregar nuevas funcionalidades.
