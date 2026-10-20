---
name: verify
description: Ejecuta las afirmaciones automatizadas de un plan de verificación contra el sistema real y guarda su evidencia en ficheros (tests de Go y de navegador en rojo sobre la base y en verde sobre el cambio, transcripciones HTTP, valores del DOM con capturas y vídeos de playwright-cli, consultas a la base de datos). Úsala para comprobar un cambio con evidencia reproducible. Solo verifica e informa; no arregla nada.
allowed-tools: Read, Glob, Grep, Write, Bash(bash .claude/skills/check/scripts/check.sh *), Bash(bash .claude/skills/verify/scripts/red-green.sh *), Bash(npx playwright test *), Bash(playwright-cli -s=issue-*), Bash(curl -sS -i http://localhost:*), Bash(printenv APP_PORT), Bash(docker compose exec -T db psql *), Bash(tee .adw/*), Bash(git diff *), Bash(git merge-base *), Bash(ls *)
---

# Verifica con evidencia

Recibes afirmaciones de un plan de verificación, cada una con su procedimiento, su resultado esperado, su condición de refutación y la evidencia prevista. Para cada una, ejecútala contra el sistema real, guarda la salida en un fichero y decide comparando el valor observado con el esperado.

## Entradas

- Las afirmaciones que te tocan, con su identificador (`A1`, `A2`…).
- El directorio de evidencias, dentro de `.adw/`, como `.adw/evidence/issue-12/1/`.
- El número de la issue, para la sesión del navegador.
- La base del cambio, para los tests nuevos: el commit en que la rama se separó de su base (`git merge-base HEAD origin/<base>`).
- La aplicación en marcha en `http://localhost:<puerto>`, con el puerto de la copia de trabajo: `printenv APP_PORT` en un worktree de `bin/tree`, 3000 en el checkout principal. Escríbelo tal cual en los comandos, no `$APP_PORT`: los permisos no aprueban un comando con variables.

## Reglas de la evidencia

- **Auténtica:** la salida real, guardada en un fichero del directorio de evidencias. Una captura que solo has visto, una medición que solo cuentas o una salida que no has guardado no cuentan.
- **Con nombre:** `NN_<afirmación>-<detalle>.<ext>`, numerados en el orden en que los produces: `01_a1-contador.json`, `02_a1-contador.png`. Usa siempre la opción de nombre de cada herramienta (`--filename`, `tee`, la ruta de `video-start`); sin ella, el fichero acaba en `.playwright-cli/` y no es evidencia.
- **Reproducible:** anota el comando o los pasos exactos que la producen.
- **Discriminante:** muestra que la prueba podía fallar: rojo en la base y verde en el cambio, el estado antes y después, el caso negativo.
- **Medida:** la decisión sale del valor medido. La captura es para el humano.
- **Proporcionada:** recorta a lo relevante y no guardes secretos.

## Procedimiento

Ejecuta cada afirmación con el método que le corresponda.

### Tests

- La suite completa, las migraciones y los tests de navegador, con la skill `check`: `bash .claude/skills/check/scripts/check.sh --out <dir> migrate test e2e`. Hazlo antes que cualquier test suelto: deja la base de datos de test recién migrada.
- Un test nuevo de Go, en rojo sobre la base y en verde sobre el cambio, con el script de esta skill:

  ```bash
  bash .claude/skills/verify/scripts/red-green.sh --out <dir> --name 04_a4-rojo-verde \
    --base <commit base> --run 'Test_ActionSuite/Test_TasksIndex_PendingCount' ./actions actions/tasks_test.go
  ```

  Crea un worktree de la base en `.adw/base/`, copia los ficheros de test del cambio, ejecuta los tests seleccionados en los dos lados, guarda las dos ejecuciones y borra el worktree. Los tests de una suite se seleccionan como `Suite/Test`; selecciona solo los tests nuevos, porque todos tienen que fallar en la base. Si los tests usan ficheros nuevos que no son de test, como un escenario de `fixtures/`, pásalos con `--extra fixtures/tasks.toml`. Solo vale el veredicto `DISCRIMINANTE`. Si es `DISCRIMINANTE (compilación)`, la base no compila con los tests: vale como rojo solo si lo que le falta a la base es el comportamiento que pide el criterio, y lo dices. Si el cambio trae migraciones, la base se ejecuta con el esquema nuevo: tenlo en cuenta al leer su rojo.
- Un test nuevo de navegador (`e2e/*.spec.ts`), igual, con el mismo script. Cada lado arranca su propia aplicación con `GO_ENV=test` en el puerto de los tests de navegador (`E2E_PORT`, 3001 en el checkout principal) y ejecuta el test en Chromium: en la base, la aplicación de la base con el test del cambio. Selecciona con `--run` (es el `--grep` de Playwright) o pasando solo los ficheros nuevos:

  ```bash
  bash .claude/skills/verify/scripts/red-green.sh --out <dir> --name 05_a5-rojo-verde \
    --base <commit base> e2e/tareas/completar-sigue-tras-recargar.spec.ts
  ```

  A la base se le copian también `playwright.config.ts` y `e2e/fixtures.ts`, que son el arnés de los tests. Si el test usa otro fichero nuevo, pásalo con `--extra`. Es la prueba para lo que depende del navegador, como el JavaScript de `public/assets/`: un test de Go no ejecuta ese código y no puede fallar sobre la base.

### HTTP

Para peticiones sin estado, guarda la transcripción completa:

```bash
curl -sS -i http://localhost:<puerto>/<ruta> | tee <dir>/05_a5-http.txt
```

Los formularios que cambian estado necesitan el token CSRF y la cookie de sesión: compruébalos con el navegador o con un test de acciones.

### Interfaz

Usa `playwright-cli` (ver su skill en `.claude/skills/playwright-cli/`) con la sesión del caso, sin ventana y con el viewport fijo de `.playwright/cli.config.json`:

```bash
playwright-cli -s=issue-12 open http://localhost:<puerto>/
playwright-cli -s=issue-12 snapshot                     # referencias de los elementos
playwright-cli -s=issue-12 eval "() => document.querySelector('.list-count').textContent" --filename=<dir>/01_a1-contador.json
playwright-cli -s=issue-12 screenshot --filename=<dir>/02_a1-contador.png
```

Cada paso puede ser un comando suelto (`fill`, `type`, `click`, `press`, `select`, `check`, `eval`, `screenshot`) o, para un recorrido largo o un vídeo, un script con `run-code --filename=<dir>/NN_<afirmación>-recorrido.js`, como propone la skill de `playwright-cli`. Guarda el script en el directorio de evidencias: es la forma de repetir el recorrido. Su código corre aislado, solo con la página.

- **Estado:** mide el valor en el DOM con `eval … --filename=…` y haz una captura.
- **Recorrido:** graba un vídeo corto, con un capítulo por afirmación y el valor medido antes y después:

  ```bash
  playwright-cli -s=issue-12 video-start <dir>/03_a3-completar.webm
  playwright-cli -s=issue-12 video-show-actions
  playwright-cli -s=issue-12 video-chapter "A3 · Completar una tarea" --description="El contador pasa de 2 a 1" --duration=1500
  playwright-cli -s=issue-12 eval "() => document.querySelector('.list-count').textContent" --filename=<dir>/04_a3-antes.json
  playwright-cli -s=issue-12 click <ref>
  playwright-cli -s=issue-12 eval "() => document.querySelector('.list-count').textContent" --filename=<dir>/05_a3-despues.json
  playwright-cli -s=issue-12 video-stop
  ```

  El vídeo es WebM y se ajusta a 800x800. Que dure segundos, no minutos: GitHub no admite vídeos de más de 10 MB en el plan gratuito. Comprueba su tamaño con `ls -l` antes de darlo por bueno.
- Los formularios con `data-confirm` abren un diálogo de confirmación: acéptalo con `dialog-accept`.
- Revisa también la consola (`console error`): un error de JavaScript es un hallazgo.
- Cierra la sesión al terminar: `playwright-cli -s=issue-12 close`.

### Persistencia

Consulta la base de datos de desarrollo antes y después de la acción, y guarda las dos salidas. `docker compose exec` va al PostgreSQL de esta copia de trabajo:

```bash
docker compose exec -T db psql -U postgres -d todo_development -c "SELECT title, completed FROM tasks ORDER BY created_at" | tee <dir>/06_a6-db-antes.txt
```

Cada afirmación que mide la interfaz o los datos empieza desde un estado conocido: la base de desarrollo es del entorno del caso y arrastra datos de intentos anteriores. Vacía la tabla y crea solo lo que necesita la afirmación (tres tareas, una completada), por la interfaz o con SQL:

```bash
docker compose exec -T db psql -U postgres -d todo_development -c "DELETE FROM tasks"
```

Anota esa preparación en «Reproducir»: forma parte de la evidencia.

### Casos negativos y límite

Los mismos métodos, con la entrada inválida o el caso límite. Comprueba también que el error se muestra como dice la afirmación.

### Sin comportamiento observable

Declara la afirmación como no verificable por ejecución y explica por qué.

## Resultado

Para cada afirmación:

| Campo | Contenido |
| --- | --- |
| Afirmación | Su identificador y su texto |
| Esperado | El valor que fijaba el plan |
| Observado | El valor medido, literal |
| Veredicto | `cumplida`, `no cumplida` o `no verificable`, con su motivo |
| Evidencia | Los ficheros, verificados con `ls` |
| Reproducir | El comando o los pasos exactos |

Añade las afirmaciones nuevas que hayas necesitado (por ejemplo, una regresión cerca del código tocado), marcadas como añadidas. Nunca quites ni suavices una afirmación del plan.

Si un método no puede ejecutarse (no arranca el navegador, no hay base de datos de test), dilo como fallo del método, no como afirmación no cumplida.

## Límites

- No cambies el código, los tests ni la configuración: verifica e informa.
- Escribe solo en el directorio de evidencias. El script de rojo y verde usa `.adw/base/` y lo deja limpio.
- No publiques nada ni cambies el estado del workflow: eso lo decide quien te invoca.
