---
name: plan_verification
description: Convierte la especificación en un plan de verificación con afirmaciones refutables antes de probar nada, o lo amplía con lo que pide una revisión del humano.
---

# plan_verification

## Entradas

- `specification`: título, descripción, criterios, aclaraciones y tipo de la petición, y las peticiones de cambio del humano con los comentarios que las motivan, si las hay.
- El plan de verificación vigente, si lo hay.
- Las convenciones del proyecto (`docs/procedures/conventions.md`), con las rúbricas de los juicios.
- La documentación viva de la base, desde `docs/README.md`, para saber qué documentos describen lo que pide la especificación.
- La ruta donde escribir el plan, que da el adaptador.

## Qué hace

1. Parte de la especificación, no del plan de implementación: no lo uses. Así la validación puede detectar un cambio que cumple el plan pero no el criterio. Puedes leer el código de la rama base para concretar pantallas, rutas o datos, pero no ejecutes ni modifiques nada.
2. Convierte cada criterio, aclaración y petición en afirmaciones refutables. Añade las que la especificación implica: casos límite y negativos, regresiones del comportamiento existente y lo que pide cada dimensión.
3. Para cada afirmación fija, antes de ejecutar nada:

   | Campo | Contenido |
   | --- | --- |
   | Afirmación | Qué se dice que es verdad y de qué criterio o petición sale |
   | Dimensión | Estructural, funcional o fidelidad |
   | Método | Automatizado, juicio del LLM o humano. Si es juicio, la rúbrica que aplica |
   | Procedimiento | Cómo se comprueba, descrito por lo observable (pantalla, acción, dato), no por detalles de implementación que todavía no existen. En el navegador, con los comandos sueltos de `playwright-cli` que describe la skill `verify` (abrir, escribir, pulsar, medir con `eval`, capturar y grabar): `run-code` está denegado |
   | Resultado esperado | El valor concreto que debe observarse |
   | Condición de refutación | Qué resultado demostraría que está mal |
   | Evidencia prevista | Qué fichero quedará y cómo se reproduce |
   | Bloqueante | Si su fallo rechaza el cambio o solo se informa |

4. Elige el método con estas reglas: si se puede automatizar, se automatiza, y el LLM no juzga lo que un test puede decidir; lo automatizado prepara el juicio, extrayendo el material que el LLM juzga; lo que no puede decidir ni el código ni el LLM (gusto, decisiones de producto, interpretaciones de algo ambiguo) es verificación humana y va a la lista de revisión de la propuesta.
5. Incluye siempre:
   - **Estructural:** los gates de formato, análisis y compilación, bloqueantes, y la imagen de Docker, informativa; el juicio del diff con las rúbricas de estructura de las convenciones.
   - **Funcional:** la suite completa sin regresiones, con los tests de navegador, bloqueante, y, para cada test nuevo que demuestra un criterio, su ejecución en rojo sobre la base y en verde sobre el cambio. En un bug, el test de regresión es obligatorio. Si el criterio depende del navegador, como el JavaScript de `public/assets/`, el test que lo demuestra es de navegador: uno de Go no ejecuta ese código y pasaría también en la base.
   - **Fidelidad:** la cobertura (cada criterio con sus afirmaciones y, si tiene comportamiento observable, un test asociado), la contención (cada cambio del diff trazado a un criterio) y la documentación al día (cada documento de `docs/knowledge/` y `docs/procedures/` que describe lo que cambia, actualizado; nombra los que ya sabes que lo describen), las tres bloqueantes.

   Las verificaciones humanas y las declaraciones de que algo no tiene comportamiento observable nunca son bloqueantes.
6. Fija la evidencia según el método:

   | Método | Evidencia |
   | --- | --- |
   | Gates y linters | Registro de la ejecución: comando, código de salida y salida recortada |
   | Test nuevo | Ejecución en rojo sobre la base y en verde sobre el cambio |
   | Comportamiento HTTP | Transcripción: petición, estado y extracto de la respuesta |
   | Comportamiento visible | Valor medido en el DOM, más una captura (estado) o un vídeo con capítulos (recorrido) |
   | Persistencia | Consulta antes y después |
   | Casos negativos y límite | Cualquiera de las anteriores, con la entrada inválida o el caso límite |
   | Sin comportamiento observable | Declaración con el motivo |
   | Juicio del LLM | Acta de juicio: rúbrica, cita literal, razonamiento por criterio, refutación, confianza y procedencia |

7. Para cada criterio y cada petición, fija su **prueba para el revisor**: la evidencia de una de sus afirmaciones bloqueantes que lo demuestra a la vista, la que verá el humano en la propuesta. Un vídeo si es un recorrido, una captura si es un estado, dos capturas si se comparan dos estados (con y sin, antes y después) o un extracto corto de salida si no tiene interfaz. Es una evidencia más, no un sustituto: la decisión sigue saliendo del valor medido.
8. Si es una revisión, conserva las afirmaciones anteriores y añade las de las peticiones nuevas. Si una petición o una aclaración del humano cambia lo que pedía un criterio, sustituye sus afirmaciones: la anterior se queda en el plan marcada «sustituida por Ak». Nada más justifica quitar o suavizar una afirmación: ni un rechazo ni lo que se descubra al validar.
9. Si aparece un caso que la especificación no resuelve (por ejemplo, qué mostrar cuando no hay tareas), no lo decidas ni preguntes: devuélvelo como duda funcional.
10. Comprueba que cada criterio y cada petición tienen al menos una afirmación y su prueba para el revisor, y que cada afirmación tiene resultado esperado, condición de refutación y evidencia prevista.
11. Publica el plan para informar, no para que lo aprueben. A la vista, lo que recoge, qué cambia respecto a la revisión anterior, cómo se va a probar cada criterio y petición y lo que le tocará verificar al humano; plegado, el detalle de cada afirmación.

## Resultado

El plan de verificación publicado, o una duda funcional que la especificación no resuelve.
