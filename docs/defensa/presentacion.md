---
title: "Proyecto Fly Away"
subtitle: "Informe final"
author: "Equipo Celeste (grupo planificador): Franco Jarque · Giuliano Taormina · Josefina Bacci"
institute: "Administración de Proyectos de Software 2026 · DCIC · Universidad Nacional del Sur"
date: "Noviembre 2026"
lang: es
theme: metropolis
aspectratio: 169
slide-level: 2
---

# El proyecto

## Qué es Fly Away

Un sistema que administra los vuelos de una aerolínea:

- Los **empleados** hacen sus tareas desde un mismo lugar, cada uno según su rol.
- Los **clientes** lo usan para comprar pasajes y recibir notificaciones.

\medskip

Las metas: digitalizar la gestión de vuelos, optimizar la venta de pasajes, mejorar la experiencia de los pasajeros y reducir los errores de los empleados.

\medskip

Lo desarrolló un equipo implementador de dos estudiantes. Nosotros, como grupo planificador, escribimos el backlog, planificamos los sprints y revisamos cada entrega.

## Los objetivos del POS

:::: {.columns}
::: {.column width="50%"}
1. Consultar y reservar vuelos desde casa
2. Registrar un vuelo en menos de 15 minutos
3. Avisar de un cambio en menos de 15 minutos
4. Compra con ticket y factura en menos de 5 minutos
5. Reporte de ocupación por vuelo, clase y fecha
6. Consultar el estado de un vuelo desde mostrador
7. Evitar la doble venta de un asiento
:::
::: {.column width="46%"}
Los siete están cubiertos por lo implementado y **ninguno fue incumplido**.

\medskip

Todavía no los damos por cumplidos del todo:

- Varios dependen de lo que falta: la factura ante ARCA y la selección de asientos.
- Los tiempos (15 y 5 minutos) no los medimos formalmente.
:::
::::

# Los sprints

## Sprint 1: vuelos y autenticación

:::: {.columns}
::: {.column width="52%"}
::: {.footnotesize}
| US   | Historia              | Estimadas | Reales |
| :--- | :-------------------- | --------: | -----: |
| US22 | Autenticar al usuario |        23 |     20 |
| US01 | Crear vuelos          |        43 |     40 |
| US02 | Modificar vuelos      |        15 |     12 |
| US03 | Cancelar vuelos       |        15 |      8 |
| US04 | Buscar vuelos         |        25 |     20 |
| US05 | Consultar información |        10 |      8 |
|      | **Total**             |   **131** | **108** |
:::
:::
::: {.column width="44%"}
Tardaron un 17,56 % menos de lo estimado. Habíamos dejado margen porque el equipo son estudiantes, y sobró.

\medskip

Nos dio la primera medida de su capacidad: **108 horas por sprint**, unas 27 horas semanales por persona.

\medskip

Las historias que reutilizaban lo ya hecho, como cancelar y buscar, salieron más rápido de lo esperado.
:::
::::

## Sprint 2: reserva, pago y facturación

:::: {.columns}
::: {.column width="52%"}
::: {.footnotesize}
| US   | Historia                    | Estimadas | Reales |
| :--- | :-------------------------- | --------: | -----: |
| US06 | Reservar pasajes            |        22 |      5 |
| US07 | Pagar la reserva            |        31 |      7 |
| US08 | Confirmar la transacción    |        18 |      4 |
| US09 | Generar el ticket           |        17 |      3 |
| US10 | Generar y enviar la factura |        21 |      6 |
| US23 | Cerrar sesión               |         5 |      1 |
|      | **Total**                   |   **114** | **26** |
:::
:::
::: {.column width="44%"}
Tardaron un 77,19 % menos. Dividimos de más las tareas simples, fuimos holgados con la documentación y no tuvimos en cuenta cuánto ayuda la IA con las pasarelas de pago.

::: {.caution title="Error nuestro"}
::: {.footnotesize}
La reserva salió sin selección de asientos porque no lo especificamos. Dejamos la reserva básica y sumamos cinco historias al backlog (US24 a US28).
:::
:::
:::
::::

## Sprint 3: concurrencia, ocupación y notificaciones

:::: {.columns}
::: {.column width="52%"}
::: {.footnotesize}
| US   | Historia                    | Original | Corregida |
| :--- | :-------------------------- | -------: | --------: |
| US11 | Compras concurrentes        |       25 |         8 |
| US12 | Consultar ocupación         |       12 |         4 |
| US13 | Reporte de ocupación        |       13 |         3 |
| US14 | Estado desde mostrador      |       14 |         4 |
| US15 | Avisar cambios de horario   |       16 |         6 |
| US16 | Avisar cancelaciones        |       14 |         4 |
|      | **Total estimado**          |   **94** |    **29** |
|      | **Total real**              |   **14** |    **14** |
:::
:::
::: {.column width="44%"}
Planificamos antes de hacer la retrospectiva del sprint 2, así que la estimación salió sin los ajustes que habíamos acordado. La corregimos después.

\medskip

Aun corregida, tardaron un 51,72 % menos.

\medskip

US11 era la historia de más riesgo técnico y llevó muy pocas horas.
:::
::::

## Horas estimadas y reales

| Sprint                  | Estimadas | Reales |    Desvío |
| :---------------------- | --------: | -----: | --------: |
| 1                       |       131 |    108 |  −17,56 % |
| 2                       |       114 |     26 |  −77,19 % |
| 3 (estimación original) |        94 |     14 |  −85,11 % |
| 3 (re-estimación)       |        29 |     14 |  −51,72 % |

\medskip

::: {.footnotesize}
\muted{Desvío = (reales − estimadas) / estimadas. Negativo es sobreestimación.}
:::

\medskip

Se completaron todas las historias de los tres sprints. Para uno o dos sprints más quedan la integración con ARCA, la reserva con asientos y la aplicación móvil.

# Conclusiones

## Qué aprendimos en cada sprint

**Sprint 1.** Lo que ya está construido abarata lo que sigue, y nuestro criterio de Story Points no era parejo: US02 tenía 5 puntos y llevó 12 horas, mientras que otras del mismo tamaño llevaron cerca de 20.

\medskip

**Sprint 2.** Si la especificación es ambigua, lo que se implementa es distinto de lo esperado, y la responsabilidad es de quien la escribió.

\medskip

**Sprint 3.** Lo que decide una retrospectiva tiene que pasar al sprint siguiente antes de planificarlo. Y con tres sprints no llegamos a tener historial suficiente para estimar con rigor.

## Sobre la forma de trabajo

- La comunicación con el equipo implementador fue fluida, por WhatsApp y, cuando hizo falta, por Discord.
- Cuidamos no sobrecargarnos: todos somos estudiantes con otras materias.
- Scrum nos dejó corregir sobre resultados concretos y no sobre suposiciones.
- Lo que limitó el alcance de cada sprint no fue implementar, sino el tiempo que teníamos para revisar antes de la demo.

\medskip

**Uso de IA.** Le bajó mucho el tiempo al equipo implementador, sobre todo en pasarelas de pago, documentación e investigación. Creemos que explica buena parte de nuestra sobreestimación.

## {.standout}

¿Preguntas?
