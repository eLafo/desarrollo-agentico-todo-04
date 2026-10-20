---
name: validate
description: Verifica el cambio contra el plan de verificación vigente en tres dimensiones independientes, recoge las evidencias y publica el acta con el veredicto.
skills: [check, verify, judge]
---

# validate

## Entradas

- El plan de verificación vigente.
- `specification`: la petición con sus criterios y aclaraciones, y las peticiones de cambio del humano.
- `change`: el diff de la rama respecto a su base, incluidos los cambios sin registrar.
- `rationale`: el plan de implementación vigente, solo para la fidelidad y solo para entender por qué existe un cambio.
- El número del intento en la racha de rechazos y el tope.
- Un `environment` listo: la copia de trabajo en la rama del caso, con la aplicación en marcha.
- El directorio de evidencias del intento, que da el adaptador.

El informe del implementador no es una entrada: nadie que verifique lo recibe.

## Qué hace

1. Anota el cambio que se valida: la rama, el commit de partida y la huella de la copia de trabajo, cambios sin registrar incluidos. El acta queda ligada a él. Comprueba que la aplicación en marcha sirve ese código y no una versión anterior.
2. Verifica cada dimensión por separado, con sus afirmaciones del plan y sin contexto previo. Ninguna verificación recibe el razonamiento de otra.
   - **Estructural:** ejecuta los gates de formato, análisis, compilación e imagen con la skill `check` y juzga el diff con la skill `judge` y las rúbricas de estructura de `docs/procedures/conventions.md`.
   - **Funcional:** ejecuta la suite y los tests de navegador con `check`, y las afirmaciones automatizadas contra el sistema real con la skill `verify`: tests nuevos, de Go o de navegador, en rojo sobre la base y en verde sobre el cambio, HTTP, valores del DOM con capturas y vídeos, consultas a la base de datos. Juzga con `judge` los mensajes y textos que muestra el cambio.
   - **Fidelidad:** construye la matriz de trazabilidad: cada criterio con las afirmaciones que lo cubren y su test asociado, y cada fragmento del diff con el criterio que lo justifica a través de un paso del plan. Juzga con `judge` la contención y la documentación al día: que cada documento de `docs/knowledge/` y `docs/procedures/` que describe lo que el diff toca lo cuente ya como queda. No vuelve a probar: los resultados de las afirmaciones los ponen las otras dimensiones.
3. Cada verificación puede añadir afirmaciones, como regresiones cerca del código tocado o reglas para las zonas que tocó el diff, pero nunca quitarlas ni suavizarlas. Solo escribe en el directorio de evidencias.
4. Somete cada hallazgo `blocker` de un juicio a un intento de refutación por un juez distinto, sin contexto previo: recibe el hallazgo, la rúbrica y el material citado, y trata de tumbarlo. El hallazgo cuenta si el juez lo emitió con confianza alta y el refutador lo sostiene con confianza alta; si no, pasa a la lista para el humano.
5. Con las tres dimensiones terminadas, completa la matriz de fidelidad con los resultados de las otras dos: un criterio está cubierto si se cumplen todas sus afirmaciones bloqueantes y las demás están declaradas con su motivo. Aplica la regla de decisión de cada dimensión (ver [Verification](../artifacts/verification.md)).
6. Comprueba que el acta es completa: cada afirmación tiene veredicto y evidencia, o el motivo por el que no es verificable, y cada fichero citado existe.
7. Escribe el acta y publícala con sus capturas y vídeos. A la vista: el veredicto; si rechaza, lo que falla, con lo esperado y lo observado, y lo que pasa después; la prueba para el revisor de cada criterio y petición, y lo que decide el humano, cada punto como una pregunta. Plegado, el árbol de comprobaciones: cada dimensión con el recuento de sus resultados y, dentro, cada comprobación con su resultado, quién la decide (una herramienta, un agente o el humano) y su detalle; cada juicio, con un plegado por criterio de su rúbrica. Si el intento rechazado alcanza el tope, el acta lo dice a la vista.

Si un método previsto no puede ejecutarse (no arranca el navegador, no hay base de datos de test), no rechaces el cambio: es un fallo del servicio.

## Qué hace buena a una evidencia

- **Trazable:** se asocia a una afirmación, y esta a un criterio.
- **Reproducible:** incluye el comando o los pasos exactos.
- **Discriminante:** demuestra que la prueba podía fallar: rojo en la base y verde en el cambio, el estado antes y después, un caso negativo.
- **Auténtica:** es la salida real guardada en un fichero, ligada al cambio validado. Una captura vista en sesión o una medición contada en prosa no cuentan.
- **Completa:** cubre todas las afirmaciones; las no verificables se declaran con su motivo.
- **Proporcionada:** recortada a lo relevante y sin secretos.

La captura es para el humano; la decisión sale del valor medido.

## Resultado

El acta publicada y el veredicto: validado, o rechazado con los defectos de cada dimensión para el plan.
