---
name: publish
description: Registra los cambios del caso y publica la propuesta para revisión, o la actualiza si ya existe.
---

# publish

## Entradas

- Los cambios de la copia de trabajo, con el documento de cierre, y el informe de `implement`.
- El plan vigente.
- El acta vigente de la verificación, con el cambio que validó, y el plan de verificación vigente.
- El tipo de trabajo.

## Qué hace

1. Registra todos los cambios con un mensaje convencional según el tipo, documento incluido. El checkout estaba limpio al empezar, así que todos los cambios son del caso.
2. Comprueba que lo registrado es lo que se validó: respecto al cambio validado del acta solo pueden diferir los ficheros que escribe `document`, el documento de cierre y las decisiones con sus índices. Si difiere algo más, no publiques: es un fallo, porque la verificación ya no vale.
3. Publícalos.
4. Si la propuesta no existe, créala con la descripción que se indica abajo. Después avisa en la petición.
5. Si la propuesta ya existe, actualízala con los cambios y renueva su descripción. Si guardaba trabajo sin validar, déjala lista para revisión.

La descripción es donde el humano revisa la validación: la IA ha validado y le entrega la evidencia, y tiene que poder verla sin buscarla ni leer un informe. Va en este orden:

1. El vínculo con la petición y qué cambia, en una o dos frases.
2. La evidencia: en qué commit se validó el cambio y, por cada criterio y cada petición del humano, un bloque con sus palabras como título, su prueba para el revisor a la vista y una frase con lo medido. La prueba es la que fijó el plan de verificación y ya publicó el acta: no se vuelve a subir. Después, en cuatro líneas como mucho, lo demás: tests, gates, seguridad y código. Termina con los enlaces al acta y al plan de verificación, que tienen el detalle.
3. Lo que tiene que decidir el humano: los `should-fix`, los juicios de confianza baja y las verificaciones humanas, cada uno como una pregunta que pueda contestar.
4. Plegado, cómo está hecho: las capas y los ficheros, los tests, el documento de cierre, el enlace al plan vigente, el informe de `implement` con los comandos que fallaron, si los hubo, y las observaciones sin severidad.

## Resultado

La propuesta publicada con todos los cambios del caso.
