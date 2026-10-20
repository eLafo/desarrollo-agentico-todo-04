# Trigger de issues

Este lanzador consulta las issues abiertas con la etiqueta configurada, de más recientes a más antiguas por **fecha de creación**, y elige las que tienen el **turno del agente**. Para cada una invoca Claude con la foto JSON de la issue, en el **worktree del caso** (`trees/issue-N`), con su propio entorno. Lanza hasta `max_parallel` sesiones a la vez y no espera a que terminen para volver a consultar. Cada consulta recoge hasta 100 issues y borra los worktrees de las issues cerradas; después espera el intervalo configurado y vuelve a consultar.

## Preparación

Necesitas Go (la versión del proyecto), Bash, Git, GitHub CLI (`gh`) 2.99 o posterior y Claude Code (`claude`) v2.1.259+ en `PATH`, autenticados, y el checkout del proyecto ya disponible. `bin/tree` usa además `lsof` y `pgrep`. El lanzador reutiliza la librería YAML existente.

La verificación necesita además:

- Node.js 20 o posterior y [`playwright-cli`](https://playwright.dev/agent-cli/introduction) con su navegador: `npm install -g @playwright/cli@latest` y `playwright-cli install-browser chromium`. El navegador no se descarga solo en el primer uso. La configuración del proyecto está en `.playwright/cli.config.json` (Chromium sin ventana y viewport fijo de 1280x800) y la skill de la CLI, en `.claude/skills/playwright-cli/`; si actualizas la CLI, vuelve a instalar la skill con `playwright-cli install --skills`.
- Las dependencias de los tests de navegador, en cada worktree: `npm ci` y `npx playwright install chromium`, que ejecuta `env:start` si faltan. Los tests están en `e2e/` y `playwright.config.ts` arranca la aplicación para ellos con `GO_ENV=test` en el puerto de los tests de navegador del worktree, que tiene que estar libre.
- `soda` para las migraciones de test (`go install github.com/gobuffalo/pop/v6/soda@v6.4.1`), en `PATH` o en `$(go env GOPATH)/bin`.
- Docker: cada worktree levanta su propio PostgreSQL, en su puerto (`env:start`).

El dev container (`.devcontainer/`) ya instala todo esto.

Antes de arrancar el trigger, abre Claude Code una vez de forma interactiva en el checkout y acepta el diálogo de confianza. Si el directorio no es de confianza, `claude -p` ignora los permisos de `.claude/settings.json` («this workspace has not been trusted») y la skill no podrá hacer su trabajo sin preguntar. Basta con hacerlo en el checkout: las sesiones en los worktrees de `trees/` usan los permisos y los hooks de su copia de `.claude/settings.json`.

Este starter todo-06 incluye la skill **`/adw`**, que reutiliza `/feature`, `/bug`, `/chore` y `/implement` para planificar e implementar, y `check`, `verify`, `judge` y `document` para verificar y documentar. La skill no tiene el flujo escrito: su especificación describe artefactos con sus ciclos de vida, servicios y asociaciones, y ante cada evento compone el workflow que necesita ese caso. El trigger solo decide cuándo invocarla.

El humano está **fuera del chat** e interviene solo en GitHub, en dos puntos de control: responder las preguntas del agente en la issue, si no está clara, y revisar el PR. Si el humano comenta el PR, el agente atiende la revisión en el mismo PR; el flujo termina cuando el humano lo integra o lo cierra. Antes de publicar, el agente verifica el cambio con evidencias y documenta el caso. Si la verificación lo rechaza, lo corrige solo hasta un tope de rechazos; si lo alcanza, o si algo falla, lo cuenta en la issue y se detiene.

La issue también lleva el control del caso. La etiqueta `adw` la admite en el sistema; la asignación indica qué cuenta la gestiona, y una issue asignada a otra persona no la toca ningún otro sistema; `adw:waiting-for-human` indica que es tu turno, y se devuelve quitándola. Ver [control](../.claude/skills/adw/references/control.md).

El estado de cada caso es visible en la issue mediante etiquetas: `stage:<stage>` mientras un stage está abierto y `<artifact>:<milestone>` por cada milestone alcanzado (por ejemplo, `issue:clarified`, `plan:ready` o `verification:validated`). La skill crea las que falten. Los ciclos de vida están en [los artefactos](../.claude/skills/adw/references/artifacts/) y el modelo que los rige, en [GSM](../.claude/skills/adw/references/gsm.md).

## Verificación y evidencias

El camino de una issue clara hasta el PR es: planificar la verificación y el cambio, implementar, validar, documentar y publicar.

- **Plan de verificación.** En cuanto la issue está clara, y sin mirar el plan de implementación, el agente convierte sus criterios en afirmaciones refutables: qué se comprueba, en qué dimensión, con qué método, qué resultado se espera, qué lo refutaría y qué evidencia quedará. Para cada criterio fija además su prueba para el revisor: el vídeo, la captura o la salida que lo demuestra a la vista. Lo publica en la issue (`<!-- adw:verification-plan -->`). Si al hacerlo encuentra un caso que la issue no resuelve, pregunta antes de implementar.
- **Validación.** Tres verificadores independientes comprueban el cambio en paralelo: **estructural** (gates y convenciones de `docs/procedures/conventions.md`), **funcional** (tests de Go y de navegador, test nuevo en rojo sobre la base y en verde sobre el cambio, HTTP, valores medidos en el DOM con capturas y vídeos, base de datos) y **fidelidad** (todo lo pedido, solo lo pedido y la documentación al día). Lo que se puede automatizar se automatiza; lo demás lo juzga un LLM con una rúbrica fijada de antemano, y cada hallazgo grave pasa por un intento de refutación antes de contar.
- **Acta.** Cada intento deja un acta en la issue (`<!-- adw:evidence -->`). A la vista, el veredicto, lo que falla si rechaza, la prueba de cada criterio (capturas, vídeos o extractos) y lo que tienes que decidir. Plegado, el árbol de comprobaciones por dimensión, como la lista de checks de un PR: cada comprobación con su resultado, quién la decide (⚙️ una herramienta, 🤖 un agente que juzga o 👤 tú) y su detalle. Las capturas y los vídeos se suben con `gh … --attach`. Los ficheros quedan en `.adw/evidence/issue-N/<n>/`, ignorado por git.
- **Documentación.** La que describe el sistema (`docs/knowledge/` y `docs/procedures/`) viaja con el cambio: el plan dice qué documentos deja desfasados, la implementación los actualiza y la validación rechaza si alguno sigue contando lo de antes. Con el cambio validado, el agente escribe además la historia del caso: `docs/history/changes/<tipo>-<N>-<slug>.md`, con su entrada en el índice condicional, y, si respondiste preguntas con opciones, tus decisiones en `docs/history/decisions/`. Todo viaja en el PR. Las skills de planificación y `/prime` leen la documentación desde `docs/README.md`, que dice qué leer siempre y qué según la tarea.
- **PR.** Es donde revisas la validación sin tener que buscar la evidencia. Su descripción tiene un bloque por criterio de la issue, con su prueba a la vista (un vídeo, una captura o dos lado a lado) y una frase con lo medido, y el commit en que se validó. Después vienen las preguntas que solo puedes decidir tú y, plegado, cómo está hecho.

Todo lo que publica el agente sigue la misma regla: arriba, lo que necesitas ver o hacer; debajo, plegado, el detalle completo, en viñetas y sin identificadores internos. El estado del caso está en un solo comentario, el panel del caso, que el agente edita en su sitio: qué está haciendo, si te toca y dónde está cada cosa, con el historial de ejecuciones y paradas. Solo publica un comentario nuevo cuando tienes algo nuevo que leer, como un plan o un acta, o cuando te toca a ti («Te toca: …»). Al publicar una revisión o un acta oculta como obsoleta la anterior, y al retomar el caso oculta los «te toca» ya resueltos. Ocultar no borra: el comentario queda a un clic y el estado del caso lo sigue leyendo.

## Bucle de corrección

Si la validación rechaza el cambio, el rechazo vuelve al plan como una petición de cambio, con los defectos de todas las dimensiones, y el agente replanifica, implementa y valida de nuevo, sin cederte el turno. El plan de verificación no cambia por un rechazo: lo que se pidió sigue siendo lo que se pidió. Solo tú puedes cambiarlo, con una petición en el PR o una aclaración en la issue.

El bucle admite `max_rejections` rechazos seguidos sin intervención humana (ver [Configurar](#configurar)). Al alcanzarlo, la issue queda con `verification:failed` y el turno pasa a ti, con el resumen de los rechazos y sus actas. El trabajo sin validar no se pierde: se registra en la rama del caso y queda en un PR en borrador. Cuando devuelves el turno, el contador vuelve a cero y el último rechazo entra al plan junto con lo que hayas escrito.

## Cuándo lanza la skill

El trigger no sigue eventos uno a uno: mira el estado de control de cada issue y lanza la skill cuando el turno es del agente. Una issue tiene el turno del agente si:

- está abierta y lleva la etiqueta configurada (`adw`);
- no lleva `adw:waiting-for-human`: nadie está esperando al humano;
- no lleva ninguna etiqueta `stage:*`: no hay otra ejecución trabajando en el caso;
- está sin asignar o asignada solo a la cuenta de `gh` con la que trabaja el trigger.

Así no se forman ciclos. Mientras la skill trabaja, la etiqueta `stage:*` bloquea el caso, y al terminar le cede el turno al humano con `adw:waiting-for-human`, aunque no haya tenido nada que hacer. Los comentarios del agente no despiertan al trigger, y los del humano tampoco hasta que devuelve el turno quitando la etiqueta. Por eso, para que el agente atienda tus respuestas o tu revisión del PR, escribe lo que quieras y después quita `adw:waiting-for-human` de la issue.

## Un worktree por caso

Cada caso trabaja en su worktree, `trees/issue-N`, con su propio PostgreSQL, sus contenedores y sus puertos (ver [Varios entornos a la vez](../docs/procedures/development.md#varios-entornos-a-la-vez)). Así varios casos avanzan a la vez, y tú puedes trabajar en el checkout principal mientras tanto.

- **Crear.** Antes de lanzar un caso, el trigger ejecuta `bin/tree create N`: si el worktree no existe, lo crea desde la rama por defecto de `origin`, con su `.env`; si existe, entra sin más. Lanza la sesión dentro, con las variables del `.env` exportadas. La skill sitúa la rama del caso y arranca el entorno.
- **En paralelo.** Lanza hasta `max_parallel` sesiones a la vez y no vuelve a lanzar un caso mientras su sesión sigue en marcha. Si una sesión falla, no lanza más casos, espera a las que siguen en marcha y termina con error.
- **Saltar.** Si el worktree de un caso tiene cambios sin registrar, lo dice y salta el caso; los demás siguen.
- **Borrar.** En cada consulta, si la issue de un worktree está cerrada y su sesión no está en marcha, ejecuta `bin/tree remove N`: para lo que siga en marcha y borra los contenedores, el volumen, la imagen, el worktree y la rama local. Si hay cambios sin registrar o commits que no están en ningún remoto, no lo borra y te lo dice una vez: revísalo y bórralo tú con `bin/tree remove N --force`. Los logs se conservan: están en el checkout principal.
- **Salida.** Cada línea de una sesión va precedida de `[#N]`.

Los agentes no borran worktrees: el proyecto les deniega `docker compose down`, y `bin/tree` no está en sus permisos.

Para invocar `/adw` a mano sobre un caso, hazlo desde su worktree:

```bash
bin/tree create 12
cd trees/issue-12
set -a; . ./.env; set +a
claude
```

Desde el checkout principal también funciona, pero la rama del caso se queda en él y git no deja usarla después en el worktree del caso.

## Configurar

Edita **`config/adw.yml`**. No necesitas editar el código ni establecer variables de entorno:

```yaml
repo: ""
label: adw
interval_seconds: 20
max_rejections: 3
max_parallel: 3
prompt: >-
  /adw {{payload}}
```

- `repo`: opcional. Vacío u omitido, la consulta no pasa `--repo` y `gh` usa su contexto de repositorio habitual. Con `OWNER/REPO`, se consulta ese repositorio explícitamente.
- `label`: etiqueta que admite las issues en el sistema. Debe coincidir con la que usa la skill (`adw`).
- `interval_seconds`: espera entre consultas. Las sesiones lanzadas siguen en marcha mientras tanto.
- `max_rejections`: rechazos seguidos de la verificación, sin intervención humana, tras los que el caso se da por fallido y se te cede el turno. Entero mayor o igual que 1; si se omite, 3. Con 3 hay dos correcciones automáticas; con 1 no hay ninguna. Lo lee la skill directamente de este fichero, así que vale igual cuando la lanza el trigger que cuando escribes `/adw quiero hacer la issue #12`. Cada caso usa el valor de su rama, que lo hereda de la rama por defecto al crearse: para cambiarlo en un caso en curso, cámbialo en su rama. El trigger solo lo valida al arrancar.
- `max_parallel`: sesiones de `claude -p` a la vez, cada una en el worktree de su caso. Entero mayor o igual que 1; si se omite, 1, de una en una. Cada sesión tiene su PostgreSQL, su aplicación y su navegador: la máquina tiene que dar para tantos entornos.
- `prompt`: texto que recibe Claude. Puede invocar tu propia skill o comando; `{{payload}}` se sustituye por la foto JSON de la issue (`url`, `number`, `title`, `labels`, `assignees` y `updatedAt`). No se evalúa como código de shell.

En forks o con varios remotos, comprueba qué repositorio resolverá `gh` con `gh repo view --json nameWithOwner --jq .nameWithOwner`; no presupongas que será tu fork. El trigger no cambia el repositorio predeterminado ni la configuración Git. Si `gh` no puede resolverlo o falla la consulta, termina sin lanzar Claude.

El bloque `>-` permite escribir un prompt largo en varias líneas: YAML las une con espacios.

## Arrancar

Desde la raíz del repositorio:

```bash
bash adws/trigger-github.sh
```

Para consultar y procesar la lista una sola vez:

```bash
bash adws/trigger-github.sh --once
```

También puedes ejecutar `go run ./adws` directamente. El script sitúa la ejecución en este checkout, el principal; los casos corren en sus worktrees. Ambos admiten `--config ruta/al/config.yml` para elegir otro fichero. `--once` ejecuta Claude: no es una simulación.

Cada invocación equivale a `CLAUDE_CODE_PRINT_BG_WAIT_CEILING_MS=0 claude -p "<prompt configurado con la foto de la issue>" --permission-prompts none`, ejecutado en `trees/issue-N` con las variables de su `.env`. La variable quita el tope de espera a los subagentes en segundo plano: sin ella, si el coordinador termina su turno mientras un subagente sigue trabajando, `claude -p` corta la sesión a los 10 minutos y el caso se queda a medias. El flag elimina herramientas que necesitan respuesta humana, como `AskUserQuestion`, y deniega solicitudes no autorizadas en vez de esperar; **no amplía permisos ni los desactiva**. Requiere Claude Code v2.1.259+. Confirma que el modo del cliente carga las skills/comandos del proyecto; no uses uno que los oculte.

La skill continúa automáticamente hasta abrir el PR, con el cambio validado y documentado. Si la issue no está clara, deja sus preguntas en ella (`issue:info-requested`) y te cede el turno. Cuando respondes y quitas `adw:waiting-for-human`, la siguiente consulta lanza la skill y retoma el trabajo. Con el PR publicado ocurre lo mismo con tus comentarios de revisión: el agente amplía el plan de verificación con lo que pides, revalida y actualiza el documento antes de volver a publicar. Los errores y el tope de rechazos se reportan en la issue y dejan la etiqueta `<artifact>:failed`; quitar `adw:waiting-for-human` es reintentar. Detén el sondeo con `Ctrl+C`. Si `gh`, `git`, `bin/tree create` o una sesión de `claude` devuelven error, el lanzador deja de lanzar casos, espera a las sesiones en marcha y termina.

## Ver qué hace el agente

Cada invocación de `/adw` deja un log en `.adw/logs/<issue>/<iteración>_<sesión>.jsonl` del checkout principal, también si la sesión corre en un worktree, con una línea JSON por cada evento: el prompt, cada llamada a una herramienta con su resultado o su fallo, los subagentes que empiezan y terminan, y el final de la sesión. Las llamadas de los subagentes también quedan ahí, identificadas con `agent_id` y `agent_type`. Para seguir una issue en directo:

```bash
tail -f .adw/logs/12/*.jsonl
```

La sesión escribe en `.adw/logs/unknown-<n>/` hasta que la skill sabe a qué issue afecta y la asocia con `.claude/skills/adw/scripts/log.sh case`. Si el evento no afecta a ningún caso, el log se queda ahí. Lo escriben los hooks de `.claude/settings.json`, que no registran nada en las sesiones que no invocan `adw`. `.adw/` está en `.gitignore`, así que los logs no ensucian el checkout.

## Límites de esta versión

- El lanzador no modifica etiquetas ni cierra issues, y solo recuerda las sesiones que tiene en marcha: en cada consulta vuelve a mirar el estado de control de las issues. Abrir un PR no cierra la issue por sí solo.
- El lanzador crea el worktree de cada caso y lo borra cuando se cierra su issue, pero no prepara su rama ni su entorno: eso lo hace la skill, con la rama registrada en el comentario de control de la issue y `env:start`. Antes de lanzar un caso comprueba que su worktree está limpio (`git status --porcelain`). El checkout principal no lo toca.
- El lanzador no interpreta comentarios ni gestiona reintentos o recuperación: eso lo decide la skill al reconstruir el estado.
- No interpreta el resultado de Claude. Código cero no demuestra que haya PR: el estado real está en las etiquetas de la issue.
- `--permission-prompts none` impide esperar o interrogar al humano, pero las operaciones siguen necesitando permisos previamente configurados.
- Si Claude no arranca o muere a mitad, **no hay garantía de comentario** y la issue puede quedarse con una etiqueta `stage:*`, que el trigger respeta como candado. Para reintentarla, quita esa etiqueta. Si además dejó cambios sin registrar, ni el trigger ni la skill arrancan con el worktree sucio: regístralos en la rama del caso, desde su worktree, y súbelos (`git commit` y `git push`), y la siguiente ejecución seguirá desde donde estaba; o descártalos y quita también `change:implemented`, para que vuelva a implementar.
- Las evidencias completas viven en el worktree del caso, en la máquina que validó (`.adw/evidence/`), y se borran con él. En GitHub queda lo que publica el acta: capturas y vídeos adjuntos y extractos de texto.
- Los puertos salen del número de la issue: el `trees/issue-22` de dos clones distintos usa los mismos, y `env:start` para el segundo con el conflicto.
- `--attach` solo funciona en github.com y GitHub Enterprise Cloud y requiere permiso de push. Admite imágenes de hasta 10 MB y vídeos de hasta 10 MB en el plan gratuito o 100 MB en los de pago: los vídeos de la verificación son cortos.
- `docker build` es informativo en la verificación local; en la CI sí bloquea. El agente no vigila la CI del PR.
- Un acta vale para el commit en que se validó. Si alguien sube commits a la rama del PR fuera del flujo de revisión, la evidencia deja de corresponder a la cabeza del PR.

## Comprobar el lanzador

```bash
go test -p 1 ./adws
```

Los tests usan `gh`, `git` y `claude` simulados: no consultan GitHub ni ejecutan un agente real. Los tests usan además un `bin/tree` simulado. Comprueban la configuración YAML, `max_rejections` (3 por defecto) y `max_parallel` (1 por defecto), con valores válidos e inválidos, repo opcional o explícito, paradas ante errores de `gh`, la consulta, el filtro por turno del agente, que cada sesión corre en su worktree con las variables de su `.env` y con su salida prefijada, el orden secuencial con `max_parallel: 1`, las sesiones a la vez y su tope, que un caso en marcha no se relanza y uno terminado sí, que un worktree sucio salta su caso sin parar a los demás, el borrado de los worktrees de issues cerradas (no el de una sesión en marcha, y el aviso, una sola vez, si no se pueden borrar), el flag sin prompts de permiso, la espera sin tope a los subagentes, la sustitución literal del prompt y las paradas por error, esperando a las sesiones en marcha. No sustituyen un ensayo completo del workflow hasta el PR.
