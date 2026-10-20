---
name: feature
description: Investiga una funcionalidad y escribe un plan autocontenido en plans/
allowed-tools: Bash, Read, Glob, Grep, Write
---

# Planifica una feature

Recibe en `$ARGUMENTS` la descripción de una funcionalidad. Tu único propósito es investigar y escribir un plan ejecutable; no implementes la funcionalidad.

1. Comprueba `git status --short --branch` y recorre el flujo existente desde rutas Buffalo y acciones hasta modelos, migraciones, templates Plush y assets cuando sean relevantes.
2. Lee `docs/README.md` y recorre sus índices: lo que se lee siempre y, de «Según la tarea», solo lo que encaje con el encargo, como el documento de la funcionalidad afectada en `docs/knowledge/` o los documentos de cierre de casos anteriores sobre el mismo código.
3. Convierte la petición en comportamiento observable y criterios de aceptación. Distingue requisitos explícitos de inferencias.
4. Busca patrones existentes y define el cambio vertical mínimo; no inventes una arquitectura paralela.
5. Identifica los tests que deben demostrar el comportamiento y las condiciones necesarias para ejecutarlos.
6. Escribe un único plan Markdown autocontenido en `plans/`. Usa un nombre kebab-case descriptivo, salvo que el argumento indique una ruta concreta dentro de `plans/`; en ese caso usa exactamente esa ruta.

El plan es el único archivo que puedes crear o modificar. No cambies código de aplicación, tests, configuración ni dependencias.

El plan debe incluir:

- `# Feature: …`
- `## Encargo y criterios de aceptación`
- `## Comportamiento actual y propuesta`
- `## Alcance` (incluido y excluido)
- `## Diseño por capas` (solo las capas afectadas)
- `## Archivos relevantes` (ruta y responsabilidad)
- `## Implementación paso a paso`
- `## Documentación afectada` (los documentos de `docs/knowledge/` y `docs/procedures/` que cambian y qué cambia en cada uno, o «Ninguno» con el motivo)
- `## Tests y validación`
- `## Riesgos y recuperación`

Cubre persistencia y migraciones, modelo, acción y ruta, renderizado y JavaScript o CSS solo cuando el caso los requiera. Respeta los patrones observados en el repositorio. Lo que depende del navegador, como el JavaScript de `public/assets/` o un recorrido que solo se ve en la página, se prueba con un test de navegador en `e2e/`; lo demás, con tests de Go. Las validaciones habituales son comprobar formato con `gofmt -l .`, ejecutar `go vet ./...`, `go test -p 1 ./...`, `go build ./...` y, si hay tests de navegador, `npx playwright test`; los tests de aplicación necesitan PostgreSQL y las migraciones de test preparadas. Incluye únicamente las comprobaciones pertinentes a la funcionalidad.

Si el cambio deja desfasado un documento de `docs/knowledge/` o `docs/procedures/` (un comportamiento, una regla, un texto o un comando que describe), actualizarlo es un paso más de la implementación, y si añade una funcionalidad que ningún documento describe, el paso crea el suyo. No planifiques los documentos de `docs/history/`: los escribe `document` al cerrar el caso.

Al terminar, informa solo de la ruta del plan, sus decisiones principales y las incertidumbres pendientes. Sugiere `/implement <ruta>` como siguiente paso.
