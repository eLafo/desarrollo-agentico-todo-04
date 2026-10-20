---
name: document
description: Escribe o actualiza el documento de cierre del caso y su entrada en el índice condicional, a partir del diff real y del acta de la verificación, y registra las decisiones del humano.
skills: [document]
---

# document

## Entradas

- `work_item`: identidad, tipo y título de la petición, y las preguntas con opciones que respondió el humano, con sus respuestas.
- `change`: el diff de la rama respecto a su base, incluidos los cambios sin registrar.
- El acta vigente de la verificación, para la sección de pruebas.
- Las rutas del documento del caso y de sus decisiones, que da el adaptador.

## Qué hace

1. Ejecuta la skill `document` (`.claude/skills/document/SKILL.md`) con la petición, el diff, el acta, las preguntas respondidas y las rutas.
2. Si el caso ya tiene documento, actualízalo para que describa el diff vigente. No crees otro.
3. No registres ni publiques los cambios: es trabajo de `publish`.
4. Comprueba que el documento y su entrada en el índice existen, también las decisiones si hubo preguntas, y que no has tocado ningún otro fichero. La documentación viva no se toca aquí: la actualiza el cambio y la juzga la verificación.

## Resultado

El documento y su entrada en el índice, y las decisiones si las hubo, escritos en la copia de trabajo.
