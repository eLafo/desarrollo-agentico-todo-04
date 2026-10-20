---
name: check
description: Ejecuta en local los gates de la CI (formato, vet, migraciones de test, tests, compilación, tests de navegador e imagen de Docker) y devuelve un informe estructurado con un registro por gate. Úsala para validar un cambio antes de publicarlo o para obtener la evidencia de los gates. Solo comprueba e informa; no arregla nada.
allowed-tools: Read, Bash(bash .claude/skills/check/scripts/check.sh), Bash(bash .claude/skills/check/scripts/check.sh *)
---

# Gates de la CI en local

Réplica de `.github/workflows/ci.yaml` con un script de comandos fijos:

```bash
bash .claude/skills/check/scripts/check.sh [--out <directorio en .adw/>] [gate ...]
```

| Gate | Comando | Bloquea | Dimensión |
| --- | --- | --- | --- |
| `gofmt` | `gofmt -l` sobre los `.go` que se registrarían (`git ls-files -co --exclude-standard`, sin lo ignorado ni lo borrado): lo que vería la CI | Sí | Estructural |
| `vet` | `go vet ./...` | Sí | Estructural |
| `migrate` | `soda drop -e test`, `soda create -e test` y `soda migrate -e test`: una base de test vacía, como en la CI | Sí | Funcional |
| `test` | `GO_ENV=test go test -p 1 ./...` | Sí | Funcional |
| `build` | `go build ./...` | Sí | Estructural |
| `e2e` | `npx playwright test`: los tests de navegador de `e2e/`, con la aplicación arrancada por `playwright.config.ts` con `GO_ENV=test` en `E2E_PORT` (3001 en el checkout principal) | Sí | Funcional |
| `docker` | `docker build -t todo .` | No: informativo en local; en la CI sí bloquea | Estructural |

## Uso

- Sin gates, ejecuta todos en el orden de la CI. Con gates, solo esos: `check.sh gofmt vet build docker` para la dimensión estructural y `check.sh migrate test e2e` para la funcional. No ejecutes `test` ni `e2e` en dos procesos a la vez sobre la misma copia de trabajo: comparten su base de datos de test. En worktrees distintos sí, porque cada uno tiene la suya.
- No se para en el primer fallo, para recoger todos los defectos de una vez. Si falla `migrate`, `test` y `e2e` quedan como `skip`.
- Con `--out`, deja en ese directorio un registro por gate, `gate-<gate>.txt`, con el comando, el código de salida, la duración y la salida recortada. Es la evidencia de los gates. El directorio tiene que estar dentro de `.adw/`.
- En un worktree de `bin/tree`, el script carga su `.env`: los gates usan sus puertos (`DB_PORT` y `E2E_PORT`) y su proyecto de Compose. En el checkout principal, 5432 y 3001.
- `migrate`, `test` y `e2e` necesitan el servicio `db` de esta copia de trabajo en marcha (`env:start`): si su puerto es de otro proyecto, no lo tocan y dan `error`. `e2e` necesita además las dependencias de Node (`npm ci`), Chromium (`npx playwright install chromium`) y su puerto libre; si falta algo, da `error`. Con `DB_HOST`, usan ese servidor. También necesitan `soda` (`go install github.com/gobuffalo/pop/v6/soda@v6.4.1`). `docker` necesita Docker.

## Resultado

Una línea por gate y el veredicto:

```text
gofmt	pass	exit=0	.adw/evidence/issue-12/1/gate-gofmt.txt
vet	pass	exit=0	.adw/evidence/issue-12/1/gate-vet.txt
build	fail	exit=1	.adw/evidence/issue-12/1/gate-build.txt
docker	pass	exit=0	.adw/evidence/issue-12/1/gate-docker.txt
GATES: RED (build)
```

| Estado | Significado |
| --- | --- |
| `pass` | El gate pasa |
| `fail` | El gate falla: es un defecto del cambio |
| `skip` | No se ejecuta porque falló un gate del que depende |
| `error` | Falta una herramienta, una dependencia de Node o el navegador, el puerto de los tests de navegador está ocupado, o la base de datos de test no es la de este proyecto o falla la conexión con ella: no es un defecto del cambio, sino del entorno |

El script sale con 0 en verde, 1 si falla un gate bloqueante, 2 si el uso es incorrecto y 3 si hay un `error` en un gate bloqueante.

Informa de cada gate con su estado y, si falla, de los ficheros y líneas que da la salida, con el mensaje mínimo útil y el comando para reproducirlo. No pegues la salida completa: está en el registro.

## Límites

- No arregles los fallos ni cambies el código: informa.
- No publiques nada ni cambies el estado del workflow: eso lo decide quien te invoca.
