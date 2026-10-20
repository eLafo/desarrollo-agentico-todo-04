# Asociaciones

Cablean los contratos de los artefactos: a qué artefacto concreto corresponde cada entrada. Es el único sitio donde un artefacto conoce a otro.

## Relaciones

Los artefactos se referencian por identidad, nunca por su funcionamiento interno.

- Cada Issue tiene como mucho un Environment, un Plan, un Change, una Verification y una Documentation. Juntos forman un caso.
- El Environment, el Plan, el Change, la Verification y la Documentation conocen la identidad de su Issue.

## Cableado

| Entrada | Se conecta con |
| --- | --- |
| `Plan.specification` | La Issue del caso. Está clara mientras la Issue tenga `clarified`. |
| `Plan.change_request` | El Change y la Verification del caso. Llega una petición, con los comentarios que la motivan, cada vez que el Change alcanza `changes_requested`, y otra, con el acta y los defectos de cada dimensión, cada vez que la Verification alcanza `rejected`. Si el rechazo llega al devolver el humano el turno tras agotarse el tope, trae también los comentarios humanos posteriores a esa acta. |
| `Change.specification` | El Plan del caso. Está listo mientras el Plan tenga `ready`. |
| `Change.verification` | La Verification del caso. Está validado mientras tenga `validated`. Trae el acta y el plan de verificación vigentes. |
| `Change.documentation` | La Documentation del caso. Está escrita mientras tenga `documented`. |
| `Issue.questions` | El Plan y la Verification del caso. Llega una duda cada vez que uno de ellos alcanza `question_raised`. |
| `Verification.specification` | La Issue y el Change del caso. Está clara mientras la Issue tenga `clarified`. Llega una petición de cambio del humano, con los comentarios que la motivan, cada vez que el Change alcanza `changes_requested`, y cuando el humano devuelve el turno tras agotarse el tope con comentarios que piden algo. |
| `Verification.change` | El Change del caso. Tiene `implemented` mientras el Change lo tenga. |
| `Verification.environment` | El Environment del caso. Está listo mientras tenga `ready`. |
| `Verification.rationale` | El Plan del caso: su revisión vigente. |
| `Documentation.work_item` | La Issue del caso: su identidad, tipo, título y criterios, y las preguntas con opciones que respondió el humano. |
| `Documentation.change` | El Change y la Verification del caso. Está listo mientras el Change tenga `implemented` y la Verification `validated`. El acta vigente alimenta la sección de pruebas. |
| `Documentation.environment` | El Environment del caso. La copia de trabajo está en la rama del caso mientras tenga `ready` o `released`. |
| `Environment.work_item` | La Issue del caso: su identidad, tipo y título. Está clara mientras la Issue tenga `clarified`. |
| `Plan.environment` | El Environment del caso. Está listo mientras tenga `ready`. |
| `Change.environment` | El Environment del caso. Está listo mientras tenga `ready` y liberado mientras tenga `released`. Su rama es la del Change. |
| `Environment.demand` | El Plan, el Change y la Verification del caso. Hay demanda mientras el Plan tenga la `specification` clara y no tenga `ready`, el Change tenga la `specification` lista y no tenga `implemented`, o el Change tenga `implemented` y la Verification no tenga `validated` ni `rejected`: mientras quede trabajo que el entorno en marcha pueda desbloquear. La validación funcional necesita la aplicación en marcha. También hay demanda si la Verification tiene `validated`, el Change no tiene `published` y el Environment no tiene ni `ready` ni `released`: hay que volver a la rama del caso para documentar y publicar; `Setup` la recupera y `Teardown` para el entorno a continuación. |

## Eventos

Un evento afecta al artefacto sobre el que ocurre: lo que pasa en la petición de trabajo afecta a la Issue, y lo que pasa en la propuesta de código o en su revisión, al Change. Desde ese artefacto se llega al caso por las relaciones. Un evento que no afecta a ningún artefacto de un caso no tiene efecto.

## Cascadas

Se derivan del cableado. Están aquí para leer de un vistazo cómo se propaga una invalidación.

- **Camino feliz.** Al alcanzar `Issue.clarified` se abren `VerificationPlanning` y, con el entorno listo, `Planning`, que no dependen entre sí. Después, `Implementation` y `Validation`. Con `validated` deja de haber `demand`: `Teardown` para el entorno mientras `Documenting` escribe el documento, y solo entonces se abre `Publication`.
- **La validación rechaza.** `Verification.rejected` llega al Plan como `change_request` e invalida `Plan.ready`. La `specification` del Change deja de estar lista, lo que invalida `implemented` y, con él, `rejected`. Se reabren `Planning`, con una revisión que cita el rechazo, e `Implementation`, y después `Validation`. El plan de verificación no cambia por el rechazo. Si al replanificar aparece una duda, va al humano por `question_raised`. Si un rechazo alcanza el tope, `Validation` se cierra con `failed` y el turno pasa al humano; cuando lo devuelve, el último rechazo llega al Plan junto con lo que haya escrito, y si lo escrito pide algo, también a la Verification.
- **Una revisión pide cambios.** `Change.changes_requested` llega al Plan como `change_request` e invalida `Plan.ready`, y a la Verification como petición de cambio del humano, lo que invalida `planned`. La `specification` del Change deja de estar lista, lo que invalida `implemented` y, con él, `validated`, `documented` y `published`. Vuelve a haber `demand`: `Setup` arranca el entorno, lo que invalida `released`. Se reabren `Planning` y `VerificationPlanning`, que añade afirmaciones para lo pedido; después, `Implementation` y `Validation`. Al validar, `Teardown` para el entorno, `Documenting` actualiza el documento y `Publication` actualiza la propuesta. Al volver a alcanzar `published`, `ReviewResponse` responde explicando lo que ha cambiado.
- **Una duda funcional.** El planificador o la verificación alcanzan `question_raised`, que llega a la Issue por `questions` e invalida `clarified`. `Clarification` pregunta en la issue. Sin `specification` clara no hay `demand`, así que `Teardown` para el entorno y el caso queda quieto hasta que responde el humano. Cuando la Issue vuelve a `clarified`, se invalida `question_raised` y se reabren `Planning` y `VerificationPlanning`, porque los dos perdieron su `specification` clara.
- **Implementación terminada.** Al alcanzar `Change.implemented` sigue habiendo `demand` mientras falte validar: el entorno sigue en marcha para `Validation`. Al alcanzar `validated`, `Teardown` para el entorno (`released`) y solo entonces se abre `Publication`: el entorno se para justo antes de publicar.
- **Otro caso usa la copia de trabajo.** Se invalidan `Environment.ready` y `Environment.released`. No se invalida nada más: los guards esperan, y `Setup` se reabre en cuanto vuelve a haber `demand`. Si el caso ya estaba validado y sin publicar, esa misma demanda lo devuelve a su rama.
- **El humano cambia el alcance de la issue.** Se invalida `Issue.clarified`. La `specification` del Plan y la de la Verification dejan de estar claras, lo que invalida `Plan.ready` y `Verification.planned`, y la cascada sigue por el Change.
