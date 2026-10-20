---
description: Start the TODO development environment and verify that the application responds
allowed-tools: Bash, Read
---

# Arranca el entorno de desarrollo

Levanta la aplicación desde la raíz de esta copia de trabajo: el checkout principal o un worktree de `bin/tree` (`trees/issue-N`). No mates procesos existentes ni elimines volúmenes.

## Modos

- Usa el modo **host** por defecto.
- Usa el modo **Docker completo** si `$ARGUMENTS` contiene `--docker` o si Go/Buffalo no están disponibles en el host.

## Puertos

Cada copia de trabajo tiene los suyos (ver `docs/procedures/development.md`, «Varios entornos a la vez»):

1. Lee los de la sesión con `printenv APP_PORT DB_PORT`.
2. Si no están, mira si la copia de trabajo tiene `.env` con `git status --short --ignored .env` (lo muestra como `!! .env`).
   - Si lo tiene, es un worktree con el entorno sin cargar: explica que hay que cargarlo con `set -a; . ./.env; set +a` antes de abrir la sesión y detente. No uses los puertos del checkout principal.
   - Si no lo tiene, es el checkout principal: la aplicación usa el `3000` y PostgreSQL, el `5432`.

Escribe los puertos tal cual en los comandos, no `$APP_PORT`: los permisos no aprueban un comando con variables.

## Comprobaciones previas

1. Confirma que `docker compose` está disponible y que el daemon responde.
2. Comprueba si los puertos de la aplicación y de PostgreSQL ya están ocupados (`lsof -nP -iTCP:<puerto> -sTCP:LISTEN`).
   - Si pertenecen a esta copia de trabajo (`docker compose ps` muestra sus servicios en marcha, o el proceso tiene su directorio de trabajo en esta raíz) y el servicio responde, reutilízalos.
   - Si pertenecen a otro proceso, informa del conflicto con el puerto y quién lo ocupa, y detente. No lo finalices.

## Modo host

1. Comprueba que `go` y `buffalo` están disponibles.
   - Si falta alguno, explica el bloqueo y continúa mediante Docker completo.
2. Ejecuta `docker compose up -d db`.
3. Espera a que PostgreSQL figure como saludable en `docker compose ps`.
4. Ejecuta `go mod download`. Si falta `node_modules/`, ejecuta también `npm ci` y `npx playwright install chromium`: los necesitan los tests de navegador.
5. Inicia `PORT=<puerto de la aplicación> buffalo dev` en segundo plano desde la raíz y conserva su salida para poder diagnosticar el arranque.
6. Espera hasta que `http://localhost:<puerto de la aplicación>/healthz` responda correctamente.

## Modo Docker completo

1. Ejecuta `docker compose up --build -d`.
2. Espera a que `db` y `app` figuren como saludables en `docker compose ps`.
3. Si falta `node_modules/`, ejecuta `npm ci` y `npx playwright install chromium`: los tests de navegador se ejecutan en el host.
4. Comprueba `http://localhost:<puerto de la aplicación>/healthz`.

## Resultado

Informa de:

- modo utilizado;
- servicios iniciados;
- URL de la aplicación;
- respuesta de `/healthz`;
- cualquier fallback, advertencia o proceso que deba mantenerse activo.
