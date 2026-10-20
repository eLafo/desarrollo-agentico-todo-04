---
name: chore
description: Investiga una tarea de mantenimiento y escribe un plan autocontenido en plans/
allowed-tools: Bash, Read, Glob, Grep, Write
---

# Planifica un chore

Recibe en `$ARGUMENTS` una tarea de mantenimiento. Tu único propósito es investigar y escribir un plan prescriptivo; no la implementes.

1. Comprueba `git status --short --branch` y localiza las convenciones, versiones y automatizaciones relevantes.
2. Lee `docs/README.md` y recorre sus índices: lo que se lee siempre y, de «Según la tarea», solo lo que encaje con el encargo, como el documento de la funcionalidad afectada en `docs/knowledge/` o los documentos de cierre de casos anteriores sobre el mismo código.
3. Verifica el estado actual en el repositorio. No supongas herramientas, paquetes ni scripts que no existan.
4. Identifica compatibilidad, configuración, documentación y posibles efectos operativos.
5. Delimita el cambio mínimo y ordena sus pasos para evitar modificaciones incidentales.
6. Escribe un único plan Markdown autocontenido en `plans/`. Usa un nombre kebab-case descriptivo, salvo que el argumento indique una ruta concreta dentro de `plans/`; en ese caso usa exactamente esa ruta.

El plan es el único archivo que puedes crear o modificar. No cambies código de aplicación, tests, configuración ni dependencias.

El plan debe incluir:

- `# Chore: …`
- `## Encargo y resultado esperado`
- `## Estado actual y restricciones`
- `## Alcance` (incluido y excluido)
- `## Archivos relevantes` (ruta y responsabilidad)
- `## Implementación paso a paso`
- `## Documentación afectada` (los documentos de `docs/knowledge/` y `docs/procedures/` que cambian y qué cambia en cada uno, o «Ninguno» con el motivo)
- `## Tests y validación`
- `## Riesgos y recuperación`

Deriva de `$ARGUMENTS` y de la evidencia del repositorio los archivos concretos, los puntos de ruptura u otros límites de comportamiento relevantes y las validaciones aplicables. El plan debe preservar el comportamiento observable que no forme parte explícita del encargo y explicar cómo comprobarlo antes y después del cambio.

Detalla comandos reproducibles y el orden de los cambios. Usa solo validaciones disponibles y pertinentes, como `gofmt -l .`, `go vet ./...`, `go test -p 1 ./...`, `go build ./...`, `npx playwright test` o `docker build -t todo .`. No añadas una herramienta ni una dependencia salvo que el encargo lo requiera explícitamente y el plan justifique el coste.

Si el cambio deja desfasado un documento de `docs/knowledge/` o `docs/procedures/` (un comportamiento, una regla, un texto o un comando que describe), actualizarlo es un paso más de la implementación, y si añade una funcionalidad que ningún documento describe, el paso crea el suyo. No planifiques los documentos de `docs/history/`: los escribe `document` al cerrar el caso.

Al terminar, informa solo de la ruta del plan, sus decisiones principales y las incertidumbres pendientes. Sugiere `/implement <ruta>` como siguiente paso.
