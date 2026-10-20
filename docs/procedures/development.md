# Desarrollo

Cómo se arranca la aplicación, cómo se prueba y qué comprueba la CI.

## Requisitos

- Docker, para PostgreSQL y, si se quiere, para la aplicación.
- Para trabajar en el host: Go 1.26, la CLI de Buffalo (`go install github.com/gobuffalo/cli/cmd/buffalo@v0.18.14`) y `soda` (`go install github.com/gobuffalo/pop/v6/soda@v6.4.1`).
- Para los tests de navegador: Node.js LTS, `npm ci` y `npx playwright install chromium`.

El devcontainer (`.devcontainer/devcontainer.json`) lo instala todo y arranca la base de datos.

## Arrancar y parar

`/env:start` arranca el entorno y comprueba que responde; `/env:stop` lo para sin borrar datos. A mano:

| Modo | Arrancar | Parar |
| --- | --- | --- |
| Host (por defecto) | `docker compose up -d db` y `buffalo dev`, que recompila al cambiar el código Go | Parar `buffalo dev` y `docker compose stop db` |
| Docker completo | `docker compose up --build -d` | `docker compose stop` |

En el checkout principal, la aplicación queda en http://localhost:3000, y `/healthz` dice si responde. Usa los puertos 3000 y 5432: si los ocupa otro proyecto, hay que pararlo antes o trabajar en un worktree (ver [Varios entornos a la vez](#varios-entornos-a-la-vez)).

## Varios entornos a la vez

Cada caso puede tener su worktree, `trees/issue-N`, con su propio PostgreSQL, sus contenedores y sus puertos. Así trabajan a la vez varios casos, o un humano y el agente, sin pisarse.

| Orden | Qué hace |
| --- | --- |
| `bin/tree create <N>` | Crea `trees/issue-N` desde la rama por defecto de `origin`, con su `.env`, e imprime su ruta. Si ya existe, solo imprime la ruta |
| `bin/tree remove <N>` | Para lo que siga en marcha en el worktree y borra sus contenedores, su volumen, su imagen, el worktree y su rama local. Se niega si hay cambios sin registrar o commits que no están en ningún remoto; `--force` los descarta |

El `.env` del worktree fija el proyecto de Compose, `<repositorio>-issue-N`, y los puertos, que salen del número de la issue con n = N mod 1000:

| Variable | Para qué | En `trees/issue-N` | En el checkout principal |
| --- | --- | --- | --- |
| `APP_PORT` | La aplicación en desarrollo | 13000+n | 3000 |
| `E2E_PORT` | La aplicación de los tests de navegador | 14000+n | 3001 |
| `DB_PORT` | PostgreSQL | 15000+n | 5432 |

Compose y la aplicación leen el `.env` solos; `go test` y Playwright, no. Antes de trabajar en un worktree, también antes de abrir `claude`, carga su entorno:

```bash
cd trees/issue-22
set -a; . ./.env; set +a
```

- Los gates (`check.sh`) lo cargan solos.
- `buffalo dev` necesita el puerto: `PORT=$APP_PORT buffalo dev`. `PORT` no va en el `.env`, porque la aplicación lo carga por encima de las variables del proceso y pisaría el que Playwright da a la aplicación de los tests.
- Si otro proceso ocupa un puerto del worktree, `/env:start` lo dice y para.

El workflow `adw` crea el worktree de cada caso y lo borra cuando se cierra su issue (ver [adws/README.md](../../adws/README.md)).

## Bases de datos

| Base | Para qué | Cómo se crea |
| --- | --- | --- |
| `todo_development` | La aplicación en desarrollo | La crea `compose.yaml` |
| `todo_test` | Los tests de Go y de navegador | `soda create -e test` |

Cada worktree tiene las mismas bases en su propio PostgreSQL. Al arrancar, la aplicación aplica las migraciones pendientes de su entorno. Los tests de Go necesitan la base de test migrada antes: `soda migrate -e test`.

## Tests de Go

```bash
GO_ENV=test go test -p 1 ./...
```

`-p 1` porque todos los paquetes comparten la base de test. Las suites están en `actions/` (peticiones y pantallas) y `models/` (reglas del modelo).

## Tests de navegador

```bash
npx playwright test
npx playwright test e2e/tareas/anadir-una-tarea.spec.ts
```

Playwright arranca la aplicación con `GO_ENV=test` en `127.0.0.1:3001` (otro puerto con `E2E_PORT`), contra `todo_test` (en `DB_PORT`), y la para al terminar. Cada test empieza con la lista vacía y crea las tareas que necesita con la fixture `tasks` de `e2e/fixtures.ts`. Si un test falla, el informe queda en `playwright-report/` y la traza en `test-results/`.

## Gates

```bash
bash .claude/skills/check/scripts/check.sh
```

Ejecuta en local lo mismo que la CI, en su orden: `gofmt`, `go vet`, migraciones (borra y vuelve a crear `todo_test`), tests de Go, `go build`, tests de navegador y `docker build`. El último solo informa: en local no cambia el veredicto. Se puede pedir solo algunos gates, por ejemplo `check.sh migrate test`.

## CI

`.github/workflows/ci.yaml` se ejecuta en cada PR y en cada push a `main`, con tres jobs:

| Job | Qué hace |
| --- | --- |
| `test` | Formato, `go vet`, migraciones, tests de Go y build, con PostgreSQL como servicio |
| `e2e` | Los tests de navegador. Si fallan, sube el informe como artefacto |
| `docker` | Construye la imagen |

## El workflow `adw`

El trabajo sobre issues de GitHub lo puede hacer el workflow `adw`, que planifica, implementa, verifica con evidencias, documenta y abre el PR. Cómo se configura y se lanza está en [adws/README.md](../../adws/README.md).
