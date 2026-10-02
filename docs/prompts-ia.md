# Registro de prompts de IA

La cátedra (**APS 2026, Enunciado 1**) exige registrar los prompts utilizados con IA para generar
código o decisiones de diseño, porque "la correcta utilización de las herramientas de IA es parte
de la evaluación del proyecto".

Este archivo es el registro de la **comisión implementadora** (Grupo Rojo Ferrari). Cada entrada
declara el contexto, el prompt literal y qué se validó o modificó de la salida.

> Los prompts usados por la comisión de Análisis y Management se registran en su propio
> repositorio.

---

## Regla de trabajo

Antes de dar por cerrada cualquier tarea que haya usado IA:

1. Agregar la entrada acá con el prompt literal.
2. Revisar la salida: leer el código generado, ejecutar los tests, verificar que no haya
   introducido dependencias no deseadas ni secretos.
3. Anotar en la columna **Validación** qué se corrigió a mano.

Ningún prompt se registra "a posteriori" ni de forma resumida: se copia textual.

---

## Entradas

### 2026-10-02 — Definición del stack tecnológico (US1)

**Contexto:** la comisión implementadora tenía que decidir backend, base de datos, framework de
app móvil y mecanismo de autenticación para el Enunciado 1 (plataforma FIA). Restricción del
enunciado: la misma funcionalidad debe existir como sistema web y como app móvil.

**Prompt:**

> Comparar alternativas para el stack de una plataforma web + móvil de gestión deportiva
> (calendario, escuderías, pilotos, controles técnicos, sanciones, puntajes). Requisitos: roles
> FIA / escudería / público, la misma funcionalidad en web y móvil, y foco en seguridad de
> autenticación porque hay que poder desactivar usuarios. El equipo viene de TypeScript. Evaluar
> frameworks de backend en Go, base de datos, framework de app móvil multiplataforma, y
> mecanismos de autenticación (JWT vs. sesiones). Recomendar una opción por área justificando el
> porqué de los descartes.

**Validación:** se ajustó de la recomendación inicial. La IA propuso sesiones en servidor por
seguridad, pero el criterio definitivo se fijó con el equipo: **JWT
puro se descarta** porque US6 exige desactivar usuarios y un JWT no se puede revocar antes de su
expiración, lo que dejaría acceso a cuentas desactivadas. También se acotó el árbol de
dependencias del frontend a tres paquetes (`expo`, `expo-router`, `expo-secure-store`)
justificando que `expo-router` reemplaza a `react-navigation` en lugar de sumarse. La comparación
de frameworks de backend (chi / gin / echo / fiber / stdlib) se contrastó con la
documentación de cada proyecto antes de escribir `01-stack.md`.

**Resultado:** `docs/sprint1/01-stack.md`

---

### 2026-10-02 — Estructura de repositorio y convenciones (AGENTS.md)

**Contexto:** repositorio vacío salvo los PDF del enunciado. Necesitábamos la estructura de
carpetas, el layout de capas del backend y las convenciones de código antes de implementar US5.

**Prompt:**

> Proponer la estructura de carpetas para un monorepo con backend Go (chi + pgx + sqlc +
> golang-migrate) y una app Expo con expo-router. Necesito un layout en capas
> handler → service → repo, un módulo de configuración desde entorno, y las convenciones de
> nomenclatura de tablas y columnas. Incluir qué carpetas corresponden a qué user story del
> Sprint 1.

**Validación:** la estructura se ajustó a mano para cumplir las reglas de clean code del curso
(funciones cortas, sin singletons, errores envueltos con contexto). Se descartó la propuesta de
generar los modelos de dominio por reflexión a partir de las structs de sqlc, y quedó
documentado que los errores de dominio se definen en el service y se chequean con `errors.Is`.

**Resultado:** `AGENTS.md`, `backend/`, `app/`, `docs/sprint1/TODO.md`