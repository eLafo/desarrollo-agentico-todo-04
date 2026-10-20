# Adaptador: GitHub

La definición dice qué artefactos hay, cómo evolucionan y qué servicios existen. Este adaptador dice cómo se representan, se observan, se proyectan y se materializan en GitHub, en el repositorio Git que aloja y en su copia de trabajo local.

## Representación

| Artefacto | En GitHub | Identidad |
| --- | --- | --- |
| Issue | Una issue. Su tipo es la etiqueta `feature`, `bug` o `chore` | Su URL |
| Environment | La rama del caso, registrada en el comentario de control, y la copia de trabajo local en esa rama | El nombre de la rama |
| Plan | Comentarios en la issue que empiezan por `<!-- adw:plan -->`, uno por revisión. El vigente es el más reciente | La URL del comentario |
| Change | La rama del Environment y su PR, cuya descripción incluye `Closes <URL de la issue>`. Un PR en borrador guarda trabajo sin validar: no está publicado para revisión | La rama y la URL del PR |
| Verification · plan | Comentarios en la issue que empiezan por `<!-- adw:verification-plan -->`, uno por revisión. El vigente es el más reciente | La URL del comentario |
| Verification · acta | Comentarios en la issue que empiezan por `<!-- adw:evidence -->`, uno por intento, con sus capturas y vídeos adjuntos. La vigente es la más reciente | La URL del comentario |
| Verification · ficheros | `.adw/evidence/issue-N/<n>/` en la copia de trabajo local, donde `<n>` es el número del acta en la issue (1 la primera, y así sucesivamente aunque el contador de rechazos vuelva a cero). Los ficheros se numeran `01_<afirmación>-<detalle>.<ext>`, como `02_a1-contador.png`; los registros de los gates se llaman `gate-<gate>.txt` | La ruta |
| Documentation | `docs/history/changes/<tipo>-<N>-<slug>.md`, con el slug de la rama, y su entrada en `docs/history/changes/README.md`; si el humano respondió preguntas con opciones, `docs/history/decisions/<tipo>-<N>-<slug>.md` y su entrada en `docs/history/decisions/README.md`. En la rama del caso | La ruta del documento de cierre |

Para ir del PR a su issue: `gh pr view <URL> --json closingIssuesReferences`. Para ir de la issue a su PR: lee la rama en el comentario de control y usa `gh pr list --repo OWNER/REPO --head <rama> --state all`.

## Eventos

El payload puede venir de un webhook, de GitHub Actions, de la API de eventos o directamente de un humano, en lenguaje natural: «quiero hacer la issue #12», «he añadido un comentario en la issue X». Si viene de un humano, el mensaje es el evento: identifica la issue o el PR que menciona. El trigger del proyecto (`adws/`) envía la foto JSON de una issue abierta y admitida cuyo turno es tuyo: sin `adw:waiting-for-human`, sin stage abierto y sin otro dueño. No dice qué ha cambiado: lo averiguas al reconstruir el estado.

Lo que ocurre en una issue afecta a la Issue y lo que ocurre en un PR, a su Change. Un comentario en la conversación de un PR llega como comentario de una issue que tiene `pull_request`: afecta al Change. Un PR en borrador guarda trabajo sin validar y todavía no está en revisión: lo que el humano escriba en él cuenta como escrito en la issue.

## Comentarios

- Todo comentario del agente empieza por una marca: `<!-- adw:control -->` en el comentario de control, `<!-- adw:plan -->` en los planes, `<!-- adw:verification-plan -->` en los planes de verificación, `<!-- adw:evidence -->` en las actas, `<!-- adw:clarification -->` en las preguntas y las decisiones de `clarify` y `<!-- adw -->` en el resto. El agente y el humano suelen usar la misma cuenta, así que el autor no sirve para distinguirlos. Un comentario con marca lo ha producido el agente; uno sin ella, el humano.
- Lee los comentarios de una issue con `gh api repos/OWNER/REPO/issues/N/comments --paginate` o con `gh issue view <URL> --json comments`. No uses la vista de texto (`gh issue view --comments`): omite sin avisar los comentarios ocultos, y entre ellos están las actas y las revisiones anteriores de las que depende el estado.
- En un PR hay tres tipos de comentario y hay que leerlos todos: conversación y revisiones (`gh pr view <URL> --json comments,reviews`) y comentarios en línea (`gh api repos/OWNER/REPO/pulls/N/comments`).
- Si el humano es el autor del PR, GitHub no le deja pedir cambios ni aprobar: su revisión llega siempre como comentarios. Decide si pide cambios por lo que dice, no por el tipo de revisión.

**Formato.** Cada comentario del agente tiene dos lectores. Arriba, a la vista y en lo que cabe en una pantalla, va lo que necesita el humano: qué es el comentario, su resultado y qué se espera de él. Debajo, plegado, va el detalle para los agentes y la auditoría, sin recortarlo: los agentes leen el markdown en bruto, así que plegarlo no les quita nada.

- Cada sección plegada es `<details><summary>…</summary>`, con una línea en blanco después de la del `<summary>` y otra antes de `</details>`: sin ellas, GitHub no interpreta el markdown de dentro. El `<summary>` no interpreta markdown: usa `<code>` en lugar de comillas invertidas.
- El `<summary>` dice el resultado de la sección, no «Ver más»: «Funcional · 14 de 14 cumplidas, con vídeos y capturas». Así el comentario se entiende sin abrir nada.
- Nunca se pliega lo que pide algo al humano ni lo que explica un resultado negativo: las preguntas, los defectos de un rechazo, el motivo de un fallo y lo que tiene que hacer para seguir.
- Como mucho, dos niveles de plegado.

**Lo superado se oculta.** Al publicar un comentario que sustituye a otro, oculta el anterior como obsoleto. Ocultar no borra: el comentario queda a un clic, en su sitio del historial, y la API lo sigue devolviendo.

| Al publicar | Oculta |
| --- | --- |
| Una revisión del plan o del plan de verificación | La revisión anterior del mismo plan |
| Un acta | El acta anterior |
| El workflow de un evento | Los comentarios `<!-- adw -->` anteriores en la misma issue o el mismo PR: los workflows de eventos pasados y los que cedían el turno o contaban un fallo, que ya se han resuelto |

Nunca se ocultan el comentario de control, las preguntas y decisiones de `clarify`, que son el registro de lo que decidió el humano, ni ningún comentario del humano. Para ocultar, `gh api graphql -f query='mutation($id: ID!) { minimizeComment(input: {subjectId: $id, classifier: OUTDATED}) { minimizedComment { isMinimized } } }' -f id=<node_id>`, con el `node_id` que da la API REST del comentario. Si falla, sigue: no es un fallo del servicio.

## Control

- **Admisión:** la issue lleva la etiqueta `adw`. Si una petición directa admite la issue, pon la etiqueta.
- **Propiedad:** la asignación de la issue. El dueño es la cuenta de GitHub con la que trabaja el sistema (`@me`), y sin asignar significa libre. Para reclamarla, `gh issue edit <URL> --add-assignee @me` y relee: si hay otro asignado además de ti, quítate con `--remove-assignee @me` y no actúes. Dos máquinas con la misma cuenta cuentan como el mismo dueño. El humano libera el caso desasignando la issue.
- **Turno:** la issue lleva `adw:waiting-for-human` mientras el turno es del humano. Al cederlo, publica primero un comentario que diga qué esperas del humano y que quite esa etiqueta cuando termine, y después pon la etiqueta. El humano devuelve el turno quitándola. Si lo devuelve con una petición directa, quítala tú.
- **Trabajo sin validar:** lo hay si quedan cambios sin registrar, o si la rama tiene commits respecto a su base y su cabeza no es la que dice «validada en» un PR publicado. Pasa tras un rechazo, y también cuando el trabajo ya estaba registrado, por ejemplo al recuperar una ejecución que murió. Al ceder el turno con trabajo sin validar, si quedan cambios sin registrar, haz primero commit con `wip: <título de la issue> (sin validar)`: así la copia de trabajo queda limpia aunque falle lo demás. Después, con o sin ese commit, `git push -u origin <rama>`. Si el caso no tiene PR, ábrelo en borrador con `gh pr create --draft --repo OWNER/REPO --base <base> --head <rama> --body-file <fichero>`, con `Closes <URL de la issue>` y una descripción que diga que el trabajo no está validado y enlace la última acta, si la hay. Si ya tiene PR y no está en borrador (`gh pr view <URL> --json isDraft`), pásalo a borrador con `gh pr ready <URL> --undo` y comenta en él lo mismo. Si el repositorio no admite borradores (`gh pr create --draft` falla), no abras PR y enlaza la rama; si el PR ya existía, déjalo como está y comenta en él que su cabeza no está validada. Menciona el PR o la rama en el comentario con el que cedes el turno. Si el push falla, dilo en ese comentario: el commit queda en local.
- **Comentario de control:** un único comentario en la issue que empieza por `<!-- adw:control -->`. Recoge la rama del caso y su base. Lo crea el dueño la primera vez que lo necesita y después lo edita en el sitio, sin publicar otro. Solo lo escribe el dueño, así que no hay escrituras concurrentes.

## Estado

**Proyección.** Se proyecta en la issue con etiquetas. Todas van en kebab-case: minúsculas y palabras separadas por guiones.

- Cada stage abierto es `stage:` más el nombre del stage (`ReviewResponse` → `stage:review-response`). Es también el candado del caso. Si hay varios stages abiertos a la vez, como `VerificationPlanning` y `Planning`, lleva una etiqueta por cada uno.
- Cada milestone alcanzado es el artefacto, dos puntos y el milestone (`changes_requested` del Change → `change:changes-requested`).
- El tipo de la issue es `feature`, `bug` o `chore`. Si la issue ya trae una, respétala.
- Los milestones terminales se observan sin etiqueta: `issue:closed` es la issue cerrada, `change:merged` es el PR integrado y `change:discarded`, el PR cerrado sin integrar.
- `environment:ready` y `environment:released` no se proyectan: describen la máquina, no el caso. Se observan en local: la rama actual, que esté al día con `origin` y si `http://localhost:3000/healthz` responde.

**Catálogo.** Son las únicas etiquetas que usa el adaptador. No inventes otras ni cambies su forma.

| Familia | Etiquetas |
| --- | --- |
| Control | `adw`, `adw:waiting-for-human` |
| Tipo | `feature`, `bug`, `chore` |
| Stages | `stage:clarification`, `stage:setup`, `stage:teardown`, `stage:planning`, `stage:implementation`, `stage:verification-planning`, `stage:validation`, `stage:documenting`, `stage:publication`, `stage:review-response` |
| Issue | `issue:clarified`, `issue:info-requested` |
| Plan | `plan:ready`, `plan:question-raised` |
| Change | `change:implemented`, `change:published`, `change:review-answered`, `change:changes-requested` |
| Verification | `verification:planned`, `verification:question-raised`, `verification:validated`, `verification:rejected` |
| Documentation | `documentation:documented` |
| Fallos | `issue:failed`, `environment:failed`, `plan:failed`, `change:failed`, `verification:failed`, `documentation:failed` |

**Observación.** Las etiquetas son el último estado proyectado. Contrástalas con lo observable: comentarios humanos posteriores al último comentario del agente, comentarios de plan, commits sin subir, si el PR contiene la cabeza de la rama. Para los artefactos de este adaptador:

- **`change:published`:** el PR existe, no está en borrador (`gh pr view <URL> --json isDraft,headRefOid,body`), contiene la cabeza de la rama y su descripción dice «validada en» ese commit, con el enlace al acta vigente.

- **`verification:planned`:** hay un plan de verificación posterior a la última vez que la Issue perdió `clarified` y que recoge todas las peticiones de cambio del humano. Una petición posterior a él lo invalida.
- **Acta vigente:** la última `adw:evidence`. Su veredicto da `validated` o `rejected`, que valen mientras el Change conserve `implemented`. `publish` ata el acta al commit que crea, y la descripción del PR dice «validada en `<sha>`». La evidencia sigue valiendo mientras ese commit sea la cabeza del PR. Los ficheros del caso en `docs/history/` no la invalidan.
- **Rechazos seguidos:** las actas con veredicto «rechazado» posteriores a la última intervención humana, que es lo más reciente entre el último comentario humano en la issue o el PR y la última vez que se quitó `adw:waiting-for-human` (`gh api repos/OWNER/REPO/issues/N/timeline --paginate`, eventos `unlabeled`). El intento en curso es ese número más uno.
- **Rechazo pendiente:** si la última acta es un rechazo y ninguna revisión del plan posterior la cita, el rechazo todavía no ha llegado al plan. Pasa cuando el humano devuelve el turno tras agotarse el tope.
- **`documentation:documented`:** el documento y su entrada en el índice, y las decisiones con la suya si hubo preguntas con opciones, están en la rama respecto a la base, registrados o en la copia de trabajo (`git diff --name-only origin/<base>` y `git status --porcelain`), y el documento enlaza el acta vigente. Si enlaza una anterior, describe otro diff y no cuenta.

**Escritura.** Cambia etiquetas con `gh issue edit <URL> --add-label … --remove-label …`, sin reemplazar el resto. Si una etiqueta no existe, créala con `gh label create`. Al abrir un stage, pon su etiqueta antes de nada. En lo demás, publica primero y cambia las etiquetas al final: quien vea una etiqueta nueva ya encontrará en GitHub lo que anuncia.

**Reintentos.** Si un stage falla o agota su tope, proyecta `<artifact>:failed`, comenta el fallo y cede el turno. El humano reintenta devolviéndolo: al quitar `adw:waiting-for-human`, se invalida `failed`, quitas su etiqueta y el contador de rechazos vuelve a cero. Si el fallo fue por agotar el tope, proyecta además `verification:rejected`: el último rechazo llega al plan. Si una ejecución se interrumpió y dejó una etiqueta `stage:*`, el humano la quita para reintentar.

## Acta

El acta es la evidencia que ve el humano. Los ficheros de `.adw/evidence/` se quedan en la máquina que validó; lo que el acta no publique no lo puede comprobar nadie más. Sigue el formato de los comentarios: arriba, el veredicto y una línea por criterio con su prueba; debajo, plegado, el detalle de cada dimensión.

```markdown
<!-- adw:evidence -->
### ✅ Validado · acta de verificación, intento k de <max_rejections>
Veredicto: **validado**. <Lo que importa al humano, en una frase>

| Qué se pedía | Resultado | Prueba |
| --- | --- | --- |
| <Criterio o petición, con las palabras de la issue> | ✅ | [vídeo](.adw/evidence/issue-12/1/10_a2-completar.webm) · [captura](.adw/evidence/issue-12/1/19_a2-recarga.png) · A2 |
| Sin regresiones: <lo que se comprobó, en pocas palabras> | ✅ | A4–A12 |

**Para ti, en la propuesta:** <los puntos para el humano, en una línea>

Árbol `<huella corta>` · [plan de verificación, revisión r](<enlace>) · [plan de implementación, revisión r](<enlace>)

<details><summary>Contexto · cambio validado, entorno, verificadores y preparación</summary>

Cambio validado: `<rama>` @ `<commit corto>` con los cambios sin registrar · árbol `<huella corta>` · regla de decisión aplicada en cada dimensión · cómo se comprobó que la aplicación servía la copia · modelos de cada verificador

</details>

<details><summary>Funcional · <c> de <n> cumplidas, con vídeos y capturas</summary>

| # | Afirmación | Método | Esperado | Observado | Veredicto | Evidencia | Reproducir |
| --- | --- | --- | --- | --- | --- | --- | --- |

Capturas, vídeos y extractos de texto

</details>

<details><summary>Fidelidad · <criterios cubiertos> · <contención> · <documentación></summary>

Matriz criterio → afirmaciones y test asociado · cambios sin justificar · criterios sin cubrir · documentos revisados y desfasados

</details>

<details><summary>Estructural · <gates> · <juicios></summary>

Automatizado: <gate · código de salida> · Juicios: regla · cita · razonamiento · refutación · confianza · severidad

</details>

<details><summary>Para el humano · <n> puntos</summary>

`should-fix`, juicios de confianza baja, verificaciones humanas y observaciones fuera de rúbrica

</details>
```

Un rechazo usa la misma plantilla, con estas diferencias:

```markdown
### ❌ Rechazado · acta de verificación, intento k de <max_rejections>
Veredicto: **rechazado**. <Lo que importa al humano, en una frase>

#### Qué falla
1. **<Dimensión> · <afirmación> (<criterio>):** se esperaba <esperado> y se observó <observado> · [evidencia](<ruta>)
2. …

**Siguiente:** el plan se revisa con estos defectos (intento k+1 de <max_rejections>).
```

Después, la misma tabla, con ❌ en lo que falla, y las mismas secciones plegadas.

- La línea `Veredicto:` va siempre a la vista: la observación del estado la lee.
- La tabla tiene una fila por criterio y por petición del humano, con sus palabras y no con su identificador, y una fila para las regresiones. La columna «Prueba» enlaza la prueba para el revisor que fijó el plan de verificación y termina con el identificador de la afirmación: si el enlace a un vídeo se queda solo en su celda, GitHub lo convierte en un reproductor dentro de la tabla.
- Las capturas y los vídeos se citan con la misma ruta que se pasa a `--attach`, como imagen (`![A1 · contador con una tarea completada](.adw/evidence/issue-12/1/02_a1-contador.png)`) o como enlace (`[vídeo](.adw/evidence/issue-12/1/03_a1-completar.webm)`): `gh` sube cada fichero una vez y reescribe todas sus referencias con la URL publicada. Un vídeo citado como imagen queda como su URL en una línea propia, que GitHub muestra como reproductor. Lo que no se cite se añade al final, en el orden de los `--attach`.
- La evidencia textual (registros, transcripciones, consultas) va recortada en bloques de código, dentro de su sección plegada, citando el fichero del que sale.
- Si el rechazo agota el tope, en lugar de «Siguiente» el acta dice «Tope alcanzado (`max_rejections`: <valor>): el caso pasa al humano.».
- Un comentario de GitHub admite 65 536 caracteres. Si el acta se acerca, recorta los extractos de texto: los registros completos están en `.adw/evidence/`.

## Servicios

El plan, el plan de verificación y el acta viven en la issue, no en la rama. `plans/issue-N.md` es solo una copia de trabajo para las skills: nunca se commitea. `.adw/` está en `.gitignore`: lo que se escribe ahí no ensucia el checkout.

| Servicio | En GitHub |
| --- | --- |
| `clarify` | Publica las preguntas, o las decisiones sobre dudas recibidas, como comentario en la issue que empiece por `<!-- adw:clarification -->`. Cada pregunta va a la vista, con sus opciones y lo que implica cada una; el análisis que la motiva, plegado. Si la clasificas, pon la etiqueta del tipo. |
| `setup_environment` | Si el comentario de control no tiene rama, asígnala como `<tipo>/issue-N-<slug>` (el slug sale del título, en minúsculas y con guiones), con la rama por defecto del repositorio como base (`gh repo view OWNER/REPO --json defaultBranchRef`), y regístrala. Si la rama no existe en `origin`, `git fetch origin <base>` y `git switch -c <rama> origin/<base>`. Si existe, `git fetch origin <rama>`, `git switch <rama>` y `git merge --ff-only origin/<rama>`. Después ejecuta `env:start`. |
| `teardown_environment` | Ejecuta `env:stop`. |
| `write_plan` | El plan se escribe en `plans/issue-N.md`. Publícalo en la issue como comentario que empiece por `<!-- adw:plan -->` y por `### ADW · plan de implementación (revisión r)`. A la vista: «Recoge:», con las peticiones y su origen y el enlace al acta si es un rechazo de validación; qué cambia respecto a la revisión anterior, en una o dos frases; el enfoque (en un bug, la causa y el arreglo); una tabla con los ficheros que toca y qué cambia en cada uno, y los títulos de los pasos. Debajo, cada sección `##` del plan plegada, con su título como resumen. Oculta la revisión anterior y borra el fichero. |
| `implement` | Antes de lanzarlo, escribe el plan vigente en `plans/issue-N.md`. Bórralo cuando termine: no forma parte del Change. |
| `plan_verification` | Lee el código de la base, no de la copia de trabajo, que puede estar en la rama de otro caso: `git show origin/<base>:<ruta>`. El plan de verificación se escribe en `.adw/evidence/issue-N/verification-plan.md`. Publícalo en la issue como comentario que empiece por `<!-- adw:verification-plan -->` y por `### ADW · plan de verificación (revisión r)`. A la vista: «Recoge:», con la especificación de la issue y las peticiones del humano y su enlace; qué cambia respecto a la revisión anterior, en una o dos frases; una tabla con cada criterio y petición y cómo se va a probar, empezando por su prueba para el revisor; las verificaciones que le tocarán al humano, y el número de afirmaciones: bloqueantes, no bloqueantes y sustituidas. Debajo, plegadas, las secciones del plan: criterios y preparación, el resumen de afirmaciones, cada dimensión, la verificación humana, la matriz y el orden de la validación, lo previsto para el humano y las dudas. Oculta la revisión anterior. |
| `validate` | Usa `.adw/evidence/issue-N/<n>/`. Si ese directorio existe de un intento que no llegó a publicar su acta, vacíalo antes. Antes de empezar, escribe ahí el plan de verificación vigente (`verification-plan.md`) y el plan de implementación vigente (`plan.md`), que solo lee la fidelidad. El cambio validado es la rama (`git branch --show-current`), el commit de partida (`git rev-parse --short HEAD`) y la huella completa del árbol de la copia de trabajo, cambios sin registrar incluidos (`bash .claude/skills/adw/scripts/tree.sh`). Antes de verificar, comprueba que la aplicación sirve la copia de trabajo: en modo host, `buffalo dev` recompila solo al cambiar el código: espera a que su salida muestre una compilación posterior al último cambio, o reinícialo; en modo Docker completo (`docker compose ps` muestra `app`), reconstrúyela con `docker compose up --build -d app` y espera a que `/healthz` responda. El navegador usa la sesión `-s=issue-N`, que se cierra al terminar. El acta se escribe en `acta.md`, según la plantilla de [Acta](#acta), y se publica con `gh issue comment <URL> --body-file .adw/evidence/issue-N/<n>/acta.md --attach <fichero> …`, con un `--attach` por cada captura o vídeo. Después, oculta el acta anterior. |
| `document` | Escribe `docs/history/changes/<tipo>-<N>-<slug>.md` y su entrada en `docs/history/changes/README.md` en la copia de trabajo. Las preguntas con opciones son las de los comentarios `<!-- adw:clarification -->`, y sus respuestas, los comentarios del humano que las contestan: si las hay, escribe también `docs/history/decisions/<tipo>-<N>-<slug>.md` y su entrada en `docs/history/decisions/README.md`. No hace commit. |
| `publish` | Hace commit con un mensaje convencional según el tipo (`feat:`, `fix:` o `chore:`), documento incluido. Comprueba con `git diff --name-only <huella del acta> HEAD` que solo difieren `docs/history/changes/<tipo>-<N>-<slug>.md`, `docs/history/changes/README.md` y, si existen, `docs/history/decisions/<tipo>-<N>-<slug>.md` y `docs/history/decisions/README.md`. Después, `git push -u origin <rama>`. Si no hay PR, lo abre con `gh pr create --repo OWNER/REPO --base <base> --head <rama> --body-file <fichero>`, con `Closes <URL de la issue>` y el enlace al plan vigente en la descripción, y comenta el enlace en la issue. Si ya existe, el push lo actualiza, `gh pr edit <URL> --body-file <fichero>` renueva la descripción y, si estaba en borrador, `gh pr ready <URL>` lo deja listo para revisión. La descripción sigue la plantilla de [Descripción del PR](#descripción-del-pr). |
| `answer_review` | Responde con comentarios en el PR. |
| `report_failure` | Comenta en la issue, a la vista, qué stage ha fallado, por qué y que el humano quite `adw:waiting-for-human` para reintentar. Si se agotó el tope, lo dice e indica el valor de `max_rejections`. El detalle, plegado: los registros del fallo o el resumen de los rechazos de la racha, con el enlace a cada acta. |

El workflow compuesto se publica como comentario en la issue o en el PR, según dónde haya ocurrido el evento. A la vista, el evento, lo siguiente y adónde se espera llegar; los pasos, plegados. Al publicarlo, oculta los comentarios `<!-- adw -->` anteriores (ver [Comentarios](#comentarios)).

## Descripción del PR

La descripción es donde el humano revisa la validación: tiene que ver la evidencia sin buscarla. Va por criterio, con la prueba a la vista y una frase con lo medido; el detalle está en el acta.

````markdown
Closes <URL de la issue>

<Qué cambia, en una o dos frases>

## ✅ Validado por el agente

Validada en `<sha>` contra la aplicación en marcha, la base de datos y la suite de tests. Cada criterio de la issue, con su prueba:

### <Criterio, con las palabras de la issue>

<URL del vídeo, en una línea propia>

<Lo medido, en una frase: «Antes había 5 tareas, 2 completadas. Después quedan las 3 pendientes, intactas en la base.»>

### <Criterio que compara dos estados>

| <Estado A> | <Estado B> |
| --- | --- |
| <img src="<URL>" width="420" alt="…"> | <img src="<URL>" width="420" alt="…"> |

### Además

- **Tests:** <cuántos nuevos y si discriminan; la suite>
- **CI en local:** <gates>
- **Seguridad:** <casos negativos, si los hay>
- **Código:** <rúbricas y alcance>

Cada medida y cómo repetirla: [acta de verificación](<enlace>) · [plan de verificación](<enlace>)

## Necesito que decidas

1. **<Pregunta concreta>** <Contexto en una o dos frases, con la prueba si hace falta>

<details><summary>Cómo está hecho</summary>

- **Capas y ficheros:** …
- **Tests:** …
- **Documentación:** los documentos de `docs/knowledge/` y `docs/procedures/` que actualiza el cambio, el documento de cierre y, si las hay, las decisiones
- **Plan de implementación:** [revisión r](<enlace>)
- **Informe de `implement`:** los comandos que fallaron, si los hubo
- **Observaciones menores:** …

</details>
````

- «Validada en `<sha>`», con el commit que acaba de crear `publish`, y el enlace al acta vigente: la observación de `change:published` los lee.
- La prueba de cada criterio es la que fijó el plan de verificación y publicó el acta: un vídeo si es un recorrido, una captura si es un estado, dos capturas lado a lado si se comparan dos estados, o un extracto corto de salida en un bloque de código si no hay interfaz. Las URLs se toman del acta (`gh api repos/OWNER/REPO/issues/comments/<id> --jq .body`): no se vuelven a subir.
- «Necesito que decidas» recoge los `should-fix`, los juicios de confianza baja y las verificaciones humanas, cada uno como una pregunta que el humano pueda contestar. Las observaciones sin severidad van plegadas.

## Límites de `--attach`

`gh` 2.99 o posterior sube imágenes y vídeos con `--attach` en `gh issue comment`, `gh pr create` y `gh pr edit`, hasta 50 ficheros por comando. Acepta PNG, JPEG, GIF, WebP, SVG, MP4, MOV y WebM. Las imágenes admiten hasta 10 MB; los vídeos, hasta 10 MB en el plan gratuito y 100 MB en los de pago, así que los vídeos deben ser cortos: comprueba su tamaño antes de adjuntarlos. Solo funciona en github.com y GitHub Enterprise Cloud, y requiere permiso de push en el repositorio. Si una subida falla, `gh` termina con error: no es un rechazo del cambio, pero el acta queda incompleta y `validate` falla.
