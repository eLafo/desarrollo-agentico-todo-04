# Verification

La prueba de que el Change cumple la especificación. Validar es decidir si el cambio está bien; recopilar evidencias es demostrar que se probó, cómo, qué se esperaba, qué se observó y por qué el veredicto se sigue de eso, de forma que alguien que no estuvo delante pueda comprobarlo. Una validación sin evidencia es una afirmación que nadie puede comprobar, y una evidencia sin plan previo no demuestra nada, porque nadie dijo antes qué debía mostrar. Por eso el plan de verificación se fija antes de probar.

## Información

- El plan de verificación, por revisión: afirmaciones refutables, cada una con su dimensión, su método, su procedimiento, el resultado esperado, la condición de refutación, la evidencia prevista y si es bloqueante, y, para cada criterio y petición, su prueba para el revisor. Lo describe [plan_verification](../services/plan_verification.md). Las revisiones anteriores se conservan.
- Un acta por intento: el veredicto de cada afirmación, de cada dimensión y de cada criterio, con su prueba para el revisor, sus evidencias, la regla de decisión aplicada y el cambio validado.
- Los ficheros de evidencia de cada intento.
- Los rechazos seguidos sin intervención humana.

## Dimensiones

| | Estructural | Funcional | Fidelidad |
| --- | --- | --- | --- |
| **Pregunta** | ¿Está bien construido? | ¿Funciona lo que hay? | ¿Es lo pedido, y solo lo pedido? |
| **Referencia** | Las convenciones del proyecto: capas, patrones del repositorio y principios de diseño | El comportamiento esperado, sin regresiones | La `specification`. El plan de implementación solo explica por qué existe un cambio; nunca es la referencia |
| **Regla de decisión** | Un `blocker` rechaza; un `should-fix` va a la propuesta como deuda; un `nit` solo se informa | Si falla una afirmación bloqueante, rechaza | Si queda un criterio sin cubrir, hay un cambio no pedido o un documento desfasado, rechaza |

Fidelidad tiene dos caras: **cobertura**, cada cosa pedida está hecha, y **contención**, cada cambio del diff responde a algo pedido, sin ampliaciones de alcance, refactors no pedidos ni extras. Los cambios mecánicos necesarios (formato, tests, migraciones que exige un criterio, la actualización de la documentación que el cambio deja desfasada) son legítimos si se trazan a un criterio a través de un paso del plan. Y comprueba además que la **documentación sigue al día**: si un documento que describe el sistema sigue contando lo de antes del cambio, el cambio no está terminado.

Funcional y fidelidad se tocan. La funcional prueba que funciona, incluido lo que la especificación no menciona (casos límite, errores, regresiones). La fidelidad no vuelve a probar: cita las evidencias de las otras dimensiones y añade lo suyo, los criterios que nadie cubre y los cambios que ningún criterio justifica.

Cada afirmación termina cumplida, no cumplida o no verificable, con su motivo. Cada hallazgo tiene una severidad, `blocker`, `should-fix` o `nit`, independiente de ese estado. Las verificaciones humanas y las declaraciones de que algo no tiene comportamiento observable nunca son bloqueantes: van a la propuesta para el humano. Un criterio está cubierto cuando se cumplen todas sus afirmaciones bloqueantes y las demás están declaradas con su motivo.

El juicio de un LLM nunca anula un resultado automatizado. Un `blocker` juzgado solo rechaza si sobrevive al intento de refutarlo y tiene confianza alta; si no, va a la propuesta como punto a revisar por el humano.

## Contrato

- **Entradas:**
  - `specification`: lo que hay que cumplir (los criterios y aclaraciones de la petición y las peticiones de cambio del humano), y si está clara.
  - `change`: el cambio que se verifica, y si tiene `implemented`.
  - `environment`: dónde se ejecuta la verificación, y si está listo.
  - `rationale`: el plan de implementación vigente. Solo explica por qué existe cada cambio; nunca es la referencia de lo que hay que cumplir.
- **Expone:** `planned`, `question_raised` con la duda, `validated` con el acta y `rejected` con el acta y los defectos de cada dimensión.

## Ciclo de vida

| Stage | Guard | Servicio | Milestones |
| --- | --- | --- | --- |
| `VerificationPlanning` | La `specification` está clara y la Verification no tiene `planned` ni `question_raised` | [`plan_verification`](../services/plan_verification.md) | `planned`, `question_raised` |
| `Validation` | Con `planned`, el `change` con `implemented`, el `environment` listo y sin `validated` ni `rejected` | [`validate`](../services/validate.md) | `validated`, `rejected` |

| Milestone | Se alcanza cuando | Se invalida cuando |
| --- | --- | --- |
| `planned` | Hay un plan de verificación que cubre la `specification`, con todas las peticiones de cambio del humano | La `specification` deja de estar clara o llega una petición de cambio del humano. Nunca por un rechazo de la validación: lo que se pidió sigue siendo lo que se pidió |
| `question_raised` | Al planificar aparece una duda funcional que la `specification` no resuelve | La `specification` vuelve a estar clara |
| `validated` | Las tres dimensiones pasan según la regla de decisión, con el acta completa | El `change` pierde `implemented` |
| `rejected` | Las tres dimensiones tienen resultado y al menos una rechaza. También cuando el humano devuelve el turno tras agotarse el tope, con el acta del último intento | El `change` pierde `implemented` |

## Tope de rechazos

El proyecto fija cuántos rechazos seguidos, sin intervención humana, admite el bucle de corrección. Si no lo fija, son 3.

- Cuando un intento rechazado alcanza el tope, `Validation` se cierra con `failed` en lugar de `rejected`. El acta se publica igual, pero el rechazo no llega al plan.
- Cuando el humano devuelve el turno, el contador vuelve a cero, se invalida `failed` y la Verification alcanza `rejected` con el acta de ese último intento: el rechazo entra al plan junto con lo que el humano haya escrito. Si lo que ha escrito pide algo, también llega a la Verification como petición de cambio del humano.
- Con un tope de 1 no hay corrección automática: el primer rechazo cede el turno.

Si un método previsto no puede ejecutarse (no arranca el navegador, no hay base de datos de test), es un fallo del servicio (`failed`), no un rechazo del cambio.
