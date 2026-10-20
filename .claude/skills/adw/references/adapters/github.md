# Adaptador: GitHub

La definición dice qué artefactos hay, cómo evolucionan y qué servicios existen. Este adaptador dice cómo se representan, se observan, se proyectan y se materializan en GitHub, en el repositorio Git que aloja y en su copia de trabajo local.

## Representación

| Artefacto | En GitHub | Identidad |
| --- | --- | --- |
| Issue | Una issue. Su tipo es la etiqueta `feature`, `bug` o `chore` | Su URL |
| Environment | La rama del caso, registrada en el comentario de control, y la copia de trabajo del caso en esa rama: el worktree `trees/issue-N` del clon local, con su entorno aislado | El nombre de la rama |
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

**Formato.** Cada comentario del agente tiene dos lectores. Arriba, a la vista, va lo que necesita el humano: qué es el comentario, su resultado, lo que tiene que ver y lo que se espera de él. Debajo, plegado, va el detalle para los agentes y la auditoría, sin recortarlo: los agentes leen el markdown en bruto, así que plegarlo no les quita nada.

- **Viñetas, no prosa.** Cada dato va en una viñeta con su etiqueta en negrita, como `- **Esperado:** 422 y el mensaje`. Un párrafo solo cuando la explicación no se puede partir.
- **Sin identificadores internos.** Lo que ve el humano, a la vista o plegado, no usa los identificadores del sistema ni cita lo que no puede abrir. Eso incluye las afirmaciones (`A4`), los criterios (`C1`), las rúbricas (`E3`, `F1`, `D2`), los nombres cortos de los datos y las recetas del plan, las severidades (`nit`, `should-fix`, `blocker`), los stages (`Verification.Validation`), las huellas del árbol y las rutas de `.adw/`. Dilo con palabras: «el criterio 1», «los textos de la interfaz», «una sugerencia menor», «debería corregirse», «bloquea». Lo que los agentes necesitan para trazar va en un comentario HTML, como `<!-- A4 · E3 · 02_a4-r1.json -->`, que GitHub no muestra y los agentes leen en bruto. Los commits, las ramas, los PR y los ficheros del repositorio sí se citan, porque el humano puede abrirlos.
- **Marcas.** ✅ cumple, ⚠️ no bloquea pero hay algo que mirar, ❌ bloquea y ⚪ no se pudo comprobar. Quién lo decide: ⚙️ una herramienta, 🤖 un agente que juzga y 👤 el humano. Un comentario que usa los iconos de quién lo decide los explica en una línea, encima de donde aparecen por primera vez.
- **Lo que el humano tiene que ver o hacer no se pliega.** Eso son las pruebas para el revisor (capturas, vídeos, extractos), lo que decide, las preguntas, los defectos de un rechazo, el motivo de un fallo y lo que tiene que hacer para seguir. Va en su propia sección, a la vista y fuera de los plegados. Dentro de un plegado, como mucho, se remite a esa sección.
- Cada sección plegada es `<details><summary>…</summary>`, con una línea en blanco después de la del `<summary>` y otra antes de `</details>`: sin ellas, GitHub no interpreta el markdown de dentro. El `<summary>` no interpreta markdown: usa `<b>` y `<code>` en lugar de asteriscos y comillas invertidas.
- El `<summary>` dice el resultado de la sección, no «Ver más»: «✅ <b>Funcional</b> · 17 comprobaciones: 15 ✅ · 2 ⚠️». Así el comentario se entiende sin abrir nada.
- **Anidar.** Un plegado puede contener otros, con los niveles que haga falta. Para que cada nivel se vea sangrado, los hijos van dentro de `<dl><dd>`, con una línea en blanco después de `<dd>` y otra antes de `</dd></dl>`:

  ```markdown
  <details><summary>✅ <b>Estructural</b> · 2 comprobaciones: 2 ✅</summary>
  <dl><dd>

  <details><summary>✅ ⚙️ Formato · <code>gofmt</code></summary>
  <!-- A13 · gate-gofmt.txt -->

  - **Resultado:** ningún fichero sin formatear

  </details>
  <details><summary>✅ ⚙️ Compilación · <code>go build</code></summary>
  …
  </details>

  </dd></dl>
  </details>
  ```

**Lo superado se oculta.** Al publicar un comentario que sustituye a otro, oculta el anterior como obsoleto. Ocultar no borra: el comentario queda a un clic, en su sitio del historial, y la API lo sigue devolviendo.

| Cuándo | Oculta |
| --- | --- |
| Al publicar una revisión del plan o del plan de verificación | La revisión anterior del mismo plan |
| Al publicar un acta | El acta anterior |
| Al empezar una ejecución después de que el humano devuelva el turno | Los comentarios `<!-- adw -->` anteriores en la issue y en su PR: los [«te toca»](#te-toca), que ya se han resuelto y quedan en el historial del [panel](#panel-del-caso) |

Nunca se ocultan el panel del caso, las preguntas y decisiones de `clarify`, que son el registro de lo que decidió el humano, ni ningún comentario del humano. Para ocultar, `gh api graphql -f query='mutation($id: ID!) { minimizeComment(input: {subjectId: $id, classifier: OUTDATED}) { minimizedComment { isMinimized } } }' -f id=<node_id>`, con el `node_id` que da la API REST del comentario. Si falla, sigue: no es un fallo del servicio.

## Control

- **Admisión:** la issue lleva la etiqueta `adw`. Si una petición directa admite la issue, pon la etiqueta.
- **Propiedad:** la asignación de la issue. El dueño es la cuenta de GitHub con la que trabaja el sistema (`@me`), y sin asignar significa libre. Para reclamarla, `gh issue edit <URL> --add-assignee @me` y relee: si hay otro asignado además de ti, quítate con `--remove-assignee @me` y no actúes. Dos máquinas con la misma cuenta cuentan como el mismo dueño. El humano libera el caso desasignando la issue.
- **Turno:** la issue lleva `adw:waiting-for-human` mientras el turno es del humano. Al cederlo, publica primero un [«te toca»](#te-toca) que diga qué esperas del humano y que quite esa etiqueta cuando termine, actualiza el [panel](#panel-del-caso) y después pon la etiqueta. El humano devuelve el turno quitándola. Si lo devuelve con una petición directa, quítala tú.
- **Trabajo sin validar:** lo hay si quedan cambios sin registrar, o si la rama tiene commits respecto a su base y su cabeza no es la que dice «validada en» un PR publicado. Pasa tras un rechazo, y también cuando el trabajo ya estaba registrado, por ejemplo al recuperar una ejecución que murió. Al ceder el turno con trabajo sin validar, si quedan cambios sin registrar, haz primero commit con `wip: <título de la issue> (sin validar)`: así la copia de trabajo queda limpia aunque falle lo demás. Después, con o sin ese commit, `git push -u origin <rama>`. Si el caso no tiene PR, ábrelo en borrador con `gh pr create --draft --repo OWNER/REPO --base <base> --head <rama> --body-file <fichero>`, con `Closes <URL de la issue>` y una descripción que diga que el trabajo no está validado y enlace la última acta, si la hay. Si ya tiene PR y no está en borrador (`gh pr view <URL> --json isDraft`), pásalo a borrador con `gh pr ready <URL> --undo` y comenta en él lo mismo. Si el repositorio no admite borradores (`gh pr create --draft` falla), no abras PR y enlaza la rama; si el PR ya existía, déjalo como está y comenta en él que su cabeza no está validada. Menciona el PR o la rama en el comentario con el que cedes el turno. Si el push falla, dilo en ese comentario: el commit queda en local.
- **Comentario de control:** un único comentario en la issue que empieza por `<!-- adw:control -->`: el [panel del caso](#panel-del-caso). Recoge la rama del caso y su base. Lo crea el dueño la primera vez que actúa sobre el caso y después lo edita en el sitio, sin publicar otro. Solo lo escribe el dueño, así que no hay escrituras concurrentes.

## Panel del caso

El comentario de control es el sitio al que mira el humano para saber cómo va el caso: qué está pasando, si le toca y dónde está cada cosa. Sustituye a los avisos de cada ejecución: no se publica un comentario por evento.

````markdown
<!-- adw:control -->
### ⏳ Trabajando: <lo que estás haciendo, en pocas palabras>
- **Después:** <la cadena prevista, en palabras: «documentar y publicar»>
- **Se espera llegar a:** <la propuesta lista para revisión | preguntas en la issue | un fallo tras agotar los intentos>
- **Planes:** [cómo se prueba](<enlace>) · [cómo se hace](<enlace>)
- **Acta:** [✅ validado, intento k de <max_rejections>](<enlace>)
- **Propuesta:** #<n>, <en borrador | lista para revisión>
- **Rama:** `<rama>` · **base:** `<base>`

<details><summary>Historial · <n> ejecuciones, <m> paradas</summary>

- <AAAA-MM-DD hh:mm> ⏳ <el evento, en una línea> → <la cadena prevista> <!-- pasos: 1. <Artifact>.<Stage> · <servicio> · <quién lo realiza y con qué modelo> · <milestone esperado>; 2. … · intento k de <max_rejections> -->
- <hh:mm> ⛔ <el fallo, en una línea> · [aviso](<enlace al «te toca»>)
- <hh:mm> ✅ <lo que se alcanzó: «validado», «propuesta publicada»>

</details>
````

- **El título dice el estado del caso.** Es «⏳ Trabajando: <…>» mientras ejecutas, el título del último «te toca» cuando el turno es del humano («👀 Te toca: revisar la propuesta #23») y «✅ Terminado: <propuesta integrada o descartada>» al cerrarse el caso. Con el turno del humano, las líneas «Después» y «Se espera llegar a» se cambian por «**Qué necesito:** [<lo que necesito>](<enlace al «te toca»>)».
- **Las líneas sin contenido no aparecen.** Sin plan de verificación no hay «Planes», y sin acta no hay «Acta».
- **El historial** tiene una entrada por ejecución y por parada, de la más antigua a la más reciente, con la fecha y la hora en UTC. Cada ejecución lleva en un comentario oculto los pasos, los servicios y los modelos del workflow compuesto, y cada recomposición, su motivo. Cuando el humano devuelve el turno, la parada dice cómo se resolvió: «→ liberaste los puertos a las 07:01».
- **La línea «Rama»** es el registro de la rama del caso y de su base que leen `setup_environment` y la búsqueda del PR.
- **Cuándo se edita:** al empezar cada ejecución, cada vez que se abre un stage o se recompone la cadena, y al terminar, al ceder el turno o al alcanzar un milestone terminal. Edítalo con `gh api -X PATCH repos/OWNER/REPO/issues/comments/<id> -F body=@<fichero>`.
- **Vive en la issue,** aunque el evento ocurra en el PR.
- Editar un comentario no notifica al humano: lo que exige su atención va en un «te toca».

## Te toca

Cada vez que cedes el turno publicas un comentario nuevo, que notifica al humano. Va donde ha ocurrido el evento, en la issue o en su PR, y empieza por lo que tiene que hacer:

````markdown
<!-- adw -->
### ⛔ Te toca: <lo que necesito, en pocas palabras>
1. <El primer paso, concreto: el comando o el sitio>
2. Quita la etiqueta `adw:waiting-for-human`. <Lo que harás entonces: «sigo desde preparar el entorno»>

**Por qué:**
- <El motivo, en una o dos viñetas>

**Cómo queda:**
- <Lo que está hecho y lo que no se pierde: los planes, la rama, el PR en borrador>

<details><summary>Detalle · <lo que contiene: «qué proceso usa cada puerto»></summary>

<Los registros del fallo, el comando denegado, el resumen de los rechazos con el enlace a cada acta>

</details>
````

| Icono | Cuándo |
| --- | --- |
| ⛔ | Un stage ha fallado o ha agotado su tope (`report_failure`) |
| 👀 | La propuesta está lista para revisión: «👀 Te toca: revisar la propuesta #<n>». Los pasos dicen dónde está la prueba de cada criterio y cuántas preguntas hay, que si quiere cambios los comente en el PR y que, si no, lo integre él. Sin «Por qué» |
| 👉 | El caso no tiene nada que hacer, por ejemplo porque ya ha terminado |

Las preguntas de `clarify` también ceden el turno, pero son su propio comentario (`<!-- adw:clarification -->`), con el título «❓ Te toca: responder <n> preguntas».

## Estado

**Proyección.** Se proyecta en la issue con etiquetas. Todas van en kebab-case: minúsculas y palabras separadas por guiones.

- Cada stage abierto es `stage:` más el nombre del stage (`ReviewResponse` → `stage:review-response`). Es también el candado del caso. Si hay varios stages abiertos a la vez, como `VerificationPlanning` y `Planning`, lleva una etiqueta por cada uno.
- Cada milestone alcanzado es el artefacto, dos puntos y el milestone (`changes_requested` del Change → `change:changes-requested`).
- El tipo de la issue es `feature`, `bug` o `chore`. Si la issue ya trae una, respétala.
- Los milestones terminales se observan sin etiqueta: `issue:closed` es la issue cerrada, `change:merged` es el PR integrado y `change:discarded`, el PR cerrado sin integrar.
- `environment:ready` y `environment:released` no se proyectan: describen la máquina, no el caso. Se observan en local, en el worktree del caso: la rama actual, que esté al día con `origin` y si `http://localhost:<puerto>/healthz` responde, con el puerto de `printenv APP_PORT` (3000 si la sesión corre en el checkout principal).

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

## Plan de verificación

El plan de verificación se publica para informar, no para que lo aprueben. Arriba, a la vista, lo que el humano querría corregir antes de que se pruebe nada: cómo se le va a demostrar cada criterio, lo que decidirá él y lo que el plan da por hecho. Debajo, plegado, el árbol de comprobaciones con el mismo orden que tendrá el acta, y la preparación.

````markdown
<!-- adw:verification-plan -->
### ADW · plan de verificación (revisión r)
- **Recoge:** la [issue #N](<enlace>), con sus <n> criterios<, y las peticiones de [este comentario](<enlace>)>
- **Respecto a la revisión anterior:** <solo si hay una: qué cambia>

⚙️ lo comprobará una herramienta · 🤖 lo juzgará un agente · 👤 lo decidirás tú

#### Cómo te lo voy a demostrar

- **Criterio 1** · <el criterio, en pocas palabras>: <su prueba para el revisor, como «dos capturas del mensaje, al añadir y al editar»>
- **Sin regresiones:** <lo que se comprobará, en pocas palabras>

#### Lo que decidirás tú

- 👤 **<La verificación humana, como pregunta>** <Por qué no lo decide una herramienta ni un agente>. No bloquea.

#### Supuestos

- <Lo que el plan da por hecho y la issue no dice, como «las tareas ya guardadas con más de 120 caracteres no se tocan»>. Si no estás de acuerdo, dilo en la issue.

#### Comprobaciones · <n>: <a> ⚙️ · <b> 🤖 · <c> 👤

<details><summary><b>Funcional</b> · <n> comprobaciones, <m> bloquean</summary>
<dl><dd>

<details><summary>Criterio 1 · <el criterio, en pocas palabras> · <n> comprobaciones</summary>
<!-- C1 · A1 A4 A6 A9 · test asociado: los de A1 -->
<dl><dd>

<details><summary>⚙️ <enunciado de la afirmación> · bloquea</summary>
<!-- A4 -->

- **Cómo:** <el procedimiento, con los datos de partida y las recetas por su nombre>
- **Esperado:** <el valor concreto>
- **Falla si:** <la condición de refutación>
- **Evidencia:** <lo que quedará, como «la respuesta de cada petición y la consulta de la base antes y después»>

</details>

</dd></dl>
</details>
<details><summary>Sin regresiones · <n> comprobaciones</summary>
…
</details>

</dd></dl>
</details>
<details><summary><b>Fidelidad</b> · <n> comprobaciones, <m> bloquean</summary>
…
</details>
<details><summary><b>Estructural</b> · <n> comprobaciones, <m> bloquean</summary>
…
</details>
<details><summary>Preparación · lo que hay en la base, el entorno, los datos de partida y las recetas</summary>
<dl><dd>

<details><summary>Datos de partida · <lo que deja en la base, como «dos tareas pendientes»></summary>

```bash
<el comando>
```

</details>
<details><summary>Receta · <lo que hace, como «escribir en un campo como el usuario»></summary>
…
</details>

</dd></dl>
</details>
<details><summary>Orden de la validación</summary>

1. <Qué se ejecuta en cada unidad y en qué orden>

</details>
````

- **El enunciado de cada afirmación** es el nombre con el que la verá el humano en el acta, así que se entiende sin el plan: «El servidor rechaza 121 caracteres al crear y al editar». El `<summary>` lleva delante quién la decide y detrás si bloquea.
- **Las comprobaciones se agrupan** por dimensión y, dentro de la funcional, por criterio y petición, más «Sin regresiones». Lo que no sale de un criterio, como los textos de la interfaz, va al final de su dimensión.
- **Los datos de partida y las recetas** llevan por nombre lo que hacen, nunca una letra, y se citan por ese nombre en el procedimiento de cada afirmación.
- **«Lo que decidirás tú» y «Supuestos»** solo aparecen si tienen algo.

## Acta

El acta es la evidencia que ve el humano. Los ficheros de `.adw/evidence/` se quedan en la máquina que validó: lo que el acta no publique no lo puede comprobar nadie más, y lo que no pueda abrir el humano no se le cita. Arriba, a la vista, va el veredicto y lo que el humano tiene que ver o decidir. Debajo, plegado, el árbol de comprobaciones, como la lista de checks de un PR.

````markdown
<!-- adw:evidence -->
### ✅ Validado · acta de verificación, intento k de <max_rejections>
Veredicto: **validado**

⚙️ lo comprueba una herramienta · 🤖 lo juzga un agente · 👤 lo decides tú

#### Pruebas

✅ **Criterio 1** · <el criterio, en pocas palabras>

| <Estado A> | <Estado B> |
| --- | --- |
| ![<Lo que muestra>](.adw/evidence/issue-12/1/adjuntos/criterio-1-antes.png) | ![<Lo que muestra>](.adw/evidence/issue-12/1/adjuntos/criterio-1-despues.png) |

✅ **Criterio 2** · <el criterio, en pocas palabras>

.adw/evidence/issue-12/1/adjuntos/criterio-2-completar.webm

✅ **Sin regresiones** · <lo que muestra la prueba, si el plan fijó una>

#### Lo que decides

No bloquea. Te lo pregunto también en la propuesta.

1. 👤 **<La verificación humana, como pregunta>**
   - <Su contexto, con la prueba en la que mirarlo>
2. ⚠️ 🤖 **<La sugerencia de un juicio, como pregunta: «¿Renombro dos tests del modelo?»>**
   - <Qué dice el juicio y dónde>

#### Comprobaciones

<details><summary>✅ <b>Funcional</b> · <n> comprobaciones: <a> ✅ · <b> ⚠️</summary>
<dl><dd>

<details><summary>✅ Criterio 1 · <el criterio, en pocas palabras> · <a> ✅</summary>
<!-- C1 · A1 A4 A6 -->
<dl><dd>

<details><summary>✅ ⚙️ <enunciado de la afirmación></summary>
<!-- A4 · 02_a4-r1.json … 06_a4-r5.json · 10_a4-db-despues.txt -->

- **Casos:** <las entradas probadas>
- **Esperado:** <el valor del plan>
- **Observado:** <el valor medido>
- **Prueba:** las capturas del criterio 1, en «Pruebas»

```json
<un extracto recortado de la evidencia>
```

</details>

</dd></dl>
</details>
<details><summary>✅ Sin regresiones · <a> ✅</summary>
…
</details>

</dd></dl>
</details>
<details><summary>✅ <b>Fidelidad</b> · <n> comprobaciones: <a> ✅</summary>
…
</details>
<details><summary>✅ <b>Estructural</b> · <n> comprobaciones: <a> ✅ · <b> ⚠️</summary>
<dl><dd>

<details><summary>✅ ⚙️ Formato · <code>gofmt</code></summary>
<!-- A13 · gate-gofmt.txt -->

- **Resultado:** ningún fichero sin formatear

</details>
<details><summary>⚠️ 🤖 Nombres · <n> criterios: <a> ✅ · <b> ⚠️</summary>
<!-- A19 · E3 · 65_a19-acta-e3.md · <modelo> -->
<dl><dd>

<details><summary>✅ <enunciado del criterio, como «El comentario de <code>TitleMaxLength</code> sigue siendo cierto»></summary>

- **Cita:** `ruta:línea` «cita literal»
- **Razonamiento:** <una frase>
- **Confianza:** alta

</details>
<details><summary>⚠️ <enunciado del criterio> · sugerencia menor</summary>

- **Cita:** `ruta:línea` «cita literal»
- **Razonamiento:**
  - <cada paso del razonamiento, en una viñeta>
- **Confianza:** baja: <por qué>

</details>

</dd></dl>
</details>

</dd></dl>
</details>
<details><summary>Contexto · qué se validó y dónde</summary>
<!-- árbol <huella completa> · base <commit> · evidencias .adw/evidence/issue-N/<n>/ · modelo de cada verificador -->

- **Commit validado:** [`<commit corto>`](<enlace>), <sin cambios sin registrar | con cambios sin registrar>
- **Aplicación:** <cómo se comprobó que servía la copia de trabajo>
- **Planes:** [cómo se prueba](<enlace>) · [cómo se hace](<enlace>)

</details>
<details><summary>Observaciones · <n>, no bloquean</summary>

- <Lo que los juicios vieron fuera de su rúbrica>

</details>
````

Un rechazo usa la misma plantilla. La diferencia es que, antes de «Pruebas», va lo que falla:

````markdown
### ❌ Rechazado · acta de verificación, intento k de <max_rejections>
Veredicto: **rechazado**

⚙️ lo comprueba una herramienta · 🤖 lo juzga un agente · 👤 lo decides tú

#### Qué falla

1. ❌ ⚙️ **<Enunciado de la comprobación>** (<dimensión>, criterio <n>)
   - **Esperado:** …
   - **Observado:** …
2. …

**Siguiente:** corrijo el plan con estos defectos (intento k+1 de <max_rejections>).
````

Después van «Pruebas», con las que se pudieron producir, y el árbol, con ❌ en lo que falla, cerrado igual: el fallo ya está a la vista en «Qué falla».

- **La línea `Veredicto:`** va siempre a la vista, porque la observación del estado la lee.
- **Cada comprobación** es `<marca> <quién> <enunciado>`. Quién la decide sale del método que fijó el plan de verificación: ⚙️ si es automatizada, 🤖 si es un juicio y 👤 si es humana. Una comprobación automatizada que ejecuta un agente con herramientas (peticiones, medidas del DOM, consultas) es ⚙️, porque su veredicto sale de comparar lo medido con el valor esperado, y las dos cosas están en su detalle.
- **Las filas que agrupan** (dimensión, criterio, juicio) no llevan icono, sino el recuento de sus marcas. Una dimensión lleva ❌ si falla algo que bloquea, y ✅ si no; sus ⚠️ se ven en el recuento.
- **La funcional se agrupa** por criterio y petición, con sus palabras en pocas palabras, más «Sin regresiones». Lo que no sale de un criterio, como los textos de la interfaz, va al final.
- **Cada juicio se despliega** en un plegado por criterio de la rúbrica, con el enunciado del acta del juicio como título. Si cumple, lleva la cita y el razonamiento en una frase. Si no, lleva la cita literal, el razonamiento en viñetas, la confianza y, si pasó por el refutador, el resultado de la refutación con su motivo. Las actas de juicio no se citan: se publican así.
- **Pruebas.** Una entrada por criterio y petición, con la prueba para el revisor que fijó el plan de verificación: un vídeo si es un recorrido, una captura si es un estado, dos capturas en una tabla si se comparan dos estados, o un extracto corto en un bloque de código si no hay interfaz. En el árbol, la comprobación que la produjo remite a ella, sin repetirla.
- **Lo que decides** recoge las verificaciones humanas (👤), los hallazgos que deberían corregirse, las sugerencias menores y los juicios de confianza baja (⚠️ 🤖), cada uno como una pregunta que el humano pueda contestar. Si no hay nada, la sección no aparece.
- **Los comentarios ocultos** de cada comprobación y cada grupo llevan lo que necesitan los agentes: los identificadores de las afirmaciones, los criterios y las rúbricas, los ficheros de evidencia y la procedencia de cada juicio.
- **Adjuntos con nombre legible.** Antes de adjuntar, copia cada captura y cada vídeo a `.adw/evidence/issue-N/<n>/adjuntos/` con un nombre sin identificadores, como `criterio-3-escribir-de-mas.webm`: GitHub enseña el nombre del vídeo encima del reproductor. Cítalos con esa ruta, la misma que se pasa a `--attach`, y con la sintaxis de markdown, `![<lo que muestra>](ruta)`, también dentro de una tabla: `gh` sube cada fichero una vez y reescribe todas sus referencias en markdown con la URL publicada, pero no la ruta de una etiqueta `<img>`. Un vídeo en una línea propia queda como reproductor.
- **La evidencia textual** (registros, transcripciones, consultas) va recortada en un bloque de código dentro de su comprobación.
- **Si el rechazo agota el tope,** en lugar de «Siguiente» el acta dice «Tope alcanzado (`max_rejections`: <valor>): el caso pasa al humano.».
- **Un comentario de GitHub admite 65 536 caracteres.** Si el acta se acerca, recorta primero los extractos de texto y después los razonamientos de los criterios que cumplen. Los registros completos están en `.adw/evidence/`.

## Servicios

El plan, el plan de verificación y el acta viven en la issue, no en la rama. `plans/issue-N.md` es solo una copia de trabajo para las skills: nunca se commitea. `.adw/` está en `.gitignore`: lo que se escribe ahí no ensucia el checkout.

| Servicio | En GitHub |
| --- | --- |
| `clarify` | Publica las preguntas, o las decisiones sobre dudas recibidas, como comentario en la issue que empiece por `<!-- adw:clarification -->`. Si son preguntas, su título es «❓ Te toca: responder <n> preguntas». Cada pregunta va a la vista, numerada, con sus opciones en viñetas y lo que implica cada una; el análisis que la motiva, plegado. Si la clasificas, pon la etiqueta del tipo. |
| `setup_environment` | La sesión ya corre en el worktree del caso, `trees/issue-N`, con su entorno cargado: lo crea el trigger con `bin/tree create N` antes de lanzarla (ver `adws/README.md`). Si el panel del caso no tiene rama, asígnala como `<tipo>/issue-N-<slug>` (el slug sale del título, en minúsculas y con guiones), con la rama por defecto del repositorio como base (`gh repo view OWNER/REPO --json defaultBranchRef`), y regístrala en su línea «Rama». Si la rama no existe en `origin`, `git fetch origin <base>` y `git switch -c <rama> origin/<base>`. Si existe, `git fetch origin <rama>`, `git switch <rama>` y `git merge --ff-only origin/<rama>`. Si git no deja cambiar a la rama porque está en otra copia de trabajo (el checkout principal u otro worktree), es un fallo del entorno: dilo con la ruta que da git. Después ejecuta `env:start`. |
| `teardown_environment` | Ejecuta `env:stop`. |
| `write_plan` | El plan se escribe en `plans/issue-N.md`. Publícalo en la issue como comentario que empiece por `<!-- adw:plan -->` y por `### ADW · plan de implementación (revisión r)`. A la vista, en viñetas: «Recoge», con las peticiones y su origen y el enlace al acta si es un rechazo de validación; «Respecto a la revisión anterior», solo si hay una; «Enfoque», en dos o tres viñetas (en un bug, la causa y el arreglo), y «Supuestos», lo que el plan decide y el humano podría querer cambiar, como no migrar una columna. Debajo, plegados: «Ficheros · <n>», con una tabla de qué cambia en cada uno; «Pasos · <n>», con sus títulos, y cada sección `##` del plan, con su título como resumen. Oculta la revisión anterior y borra el fichero. |
| `implement` | Antes de lanzarlo, escribe el plan vigente en `plans/issue-N.md`. Bórralo cuando termine: no forma parte del Change. |
| `plan_verification` | Lee el código de la base, no de la copia de trabajo, que puede estar ya en la rama del caso o ir por detrás: `git show origin/<base>:<ruta>`. El plan de verificación se escribe en `.adw/evidence/issue-N/verification-plan.md`. Publícalo en la issue según la plantilla de [Plan de verificación](#plan-de-verificación). Oculta la revisión anterior. |
| `validate` | Usa `.adw/evidence/issue-N/<n>/`. Si ese directorio existe de un intento que no llegó a publicar su acta, vacíalo antes. Antes de empezar, escribe ahí el plan de verificación vigente (`verification-plan.md`) y el plan de implementación vigente (`plan.md`), que solo lee la fidelidad. El cambio validado es la rama (`git branch --show-current`), el commit de partida (`git rev-parse --short HEAD`) y la huella completa del árbol de la copia de trabajo, cambios sin registrar incluidos (`bash .claude/skills/adw/scripts/tree.sh`). Antes de verificar, comprueba que la aplicación sirve la copia de trabajo: en modo host, `buffalo dev` recompila solo al cambiar el código: espera a que su salida muestre una compilación posterior al último cambio, o reinícialo; en modo Docker completo (`docker compose ps` muestra `app`), reconstrúyela con `docker compose up --build -d app` y espera a que `/healthz` responda. El navegador usa la sesión `-s=issue-N`, que se cierra al terminar. El acta se escribe en `acta.md`, según la plantilla de [Acta](#acta), y se publica con `gh issue comment <URL> --body-file .adw/evidence/issue-N/<n>/acta.md --attach <fichero> …`, con un `--attach` por cada captura o vídeo, copiados antes a `adjuntos/` con un nombre legible. Después, oculta el acta anterior. |
| `document` | Escribe `docs/history/changes/<tipo>-<N>-<slug>.md` y su entrada en `docs/history/changes/README.md` en la copia de trabajo. Las preguntas con opciones son las de los comentarios `<!-- adw:clarification -->`, y sus respuestas, los comentarios del humano que las contestan: si las hay, escribe también `docs/history/decisions/<tipo>-<N>-<slug>.md` y su entrada en `docs/history/decisions/README.md`. No hace commit. |
| `publish` | Hace commit con un mensaje convencional según el tipo (`feat:`, `fix:` o `chore:`), documento incluido. Comprueba con `git diff --name-only <huella del acta> HEAD` que solo difieren `docs/history/changes/<tipo>-<N>-<slug>.md`, `docs/history/changes/README.md` y, si existen, `docs/history/decisions/<tipo>-<N>-<slug>.md` y `docs/history/decisions/README.md`. Después, `git push -u origin <rama>`. Si no hay PR, lo abre con `gh pr create --repo OWNER/REPO --base <base> --head <rama> --body-file <fichero>`, con `Closes <URL de la issue>` y el enlace al plan vigente en la descripción. El enlace llega al humano en el «te toca» 👀 con el que cedes el turno. Si ya existe, el push lo actualiza, `gh pr edit <URL> --body-file <fichero>` renueva la descripción y, si estaba en borrador, `gh pr ready <URL>` lo deja listo para revisión. La descripción sigue la plantilla de [Descripción del PR](#descripción-del-pr). |
| `answer_review` | Responde con comentarios en el PR. |
| `report_failure` | Publica un [«te toca»](#te-toca) ⛔: lo que el humano tiene que hacer para seguir, con quitar `adw:waiting-for-human` como último paso; por qué ha fallado, nombrando el paso con palabras («preparar el entorno»), y cómo queda el caso. Si se agotó el tope, lo dice e indica el valor de `max_rejections`. El detalle, plegado: los registros del fallo o el resumen de los rechazos de la racha, con el enlace a cada acta. |

El workflow compuesto no se publica como comentario: se refleja en el [panel del caso](#panel-del-caso). Al empezar una ejecución después de que el humano devuelva el turno, oculta los comentarios `<!-- adw -->` anteriores (ver [Comentarios](#comentarios)).

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

1. <👤 | ⚠️ 🤖> **<Pregunta concreta>**
   - <Su contexto, con la prueba si hace falta>

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
- «Necesito que decidas» son las preguntas de la sección «Lo que decides» del acta vigente, con sus mismos iconos. Si el acta no tiene esa sección, tampoco la tiene la descripción. Las observaciones sin severidad van plegadas.
- La descripción sigue las reglas de [Comentarios](#comentarios): viñetas, y ni identificadores internos ni severidades.

## Límites de `--attach`

`gh` 2.99 o posterior sube imágenes y vídeos con `--attach` en `gh issue comment`, `gh pr create` y `gh pr edit`, hasta 50 ficheros por comando. Acepta PNG, JPEG, GIF, WebP, SVG, MP4, MOV y WebM. Las imágenes admiten hasta 10 MB; los vídeos, hasta 10 MB en el plan gratuito y 100 MB en los de pago, así que los vídeos deben ser cortos: comprueba su tamaño antes de adjuntarlos. Solo funciona en github.com y GitHub Enterprise Cloud, y requiere permiso de push en el repositorio. Si una subida falla, `gh` termina con error: no es un rechazo del cambio, pero el acta queda incompleta y `validate` falla.
