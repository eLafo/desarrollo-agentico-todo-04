# Documentation

El documento de cierre del caso: qué se construyó, cómo se usa, qué configuración necesita, cómo se probó y qué conviene saber para trabajar sobre ello. Se escribe en la rama y viaja dentro de la propuesta: se revisa y se integra con el código, no después.

Es la historia del caso. La documentación que describe el sistema tal como es no pertenece a este artefacto: forma parte del Change, porque la actualiza el propio cambio y la juzga la verificación.

## Información

- El documento del caso. Hay uno por caso: si el cambio vuelve a pasar por la verificación, se actualiza el mismo.
- Su entrada en el índice condicional, con las condiciones en que conviene cargarlo al planificar otro trabajo.
- Las decisiones del humano: las preguntas con opciones que respondió en el caso, con su contexto y su elección, y su entrada en su índice. Solo si las hubo.

## Contrato

- **Entradas:**
  - `work_item`: la petición de trabajo (identidad, tipo, título y criterios) y las preguntas con opciones que respondió el humano.
  - `change`: el cambio que se documenta y el acta de su verificación, y si está listo.
  - `environment`: dónde se escribe el documento, y si la copia de trabajo está en la rama del caso.
- **Expone:** `documented`.

## Ciclo de vida

| Stage | Guard | Servicio | Milestones |
| --- | --- | --- | --- |
| `Documenting` | El `change` está listo, la copia de trabajo está en la rama del caso y la Documentation no tiene `documented` | [`document`](../services/document.md) | `documented` |

| Milestone | Se alcanza cuando | Se invalida cuando |
| --- | --- | --- |
| `documented` | El documento, que cita el acta vigente, y su entrada en el índice existen y describen el diff vigente; y, si el humano respondió preguntas con opciones, sus decisiones están registradas | El `change` deja de estar listo |
