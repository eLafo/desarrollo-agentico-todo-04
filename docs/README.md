# Documentación

La documentación de la aplicación tiene tres partes:

- **[knowledge/](knowledge/README.md):** qué es el sistema y qué hace hoy.
- **[procedures/](procedures/README.md):** cómo se trabaja en él: convenciones, arranque, tests y CI.
- **[history/](history/README.md):** qué pasó: el documento de cierre de cada caso y las decisiones del humano.

`knowledge/` y `procedures/` describen el sistema tal como es. Un cambio que deja desfasado uno de sus documentos lo actualiza en el mismo cambio, y la verificación lo comprueba. `history/` no se reescribe: cada caso añade lo suyo al cerrarse.

## Cómo se lee

Lee lo de «Siempre». Después, entra en «Según la tarea» solo en lo que encaje con lo que vas a hacer, y recorre el índice de cada carpeta de la misma manera. Así lo hacen `/prime` y las skills de planificación (`feature`, `bug` y `chore`).

## Siempre

- [knowledge/overview.md](knowledge/overview.md): propósito, stack, estructura y recorrido de una petición.
- [procedures/conventions.md](procedures/conventions.md): capas, patrones del repositorio y rúbricas con las que se juzga un cambio.

## Según la tarea

- [knowledge/README.md](knowledge/README.md): qué hace la aplicación, funcionalidad por funcionalidad.
  - Cuando se vaya a cambiar o investigar el comportamiento de la aplicación
- [procedures/README.md](procedures/README.md): cómo se arranca, se prueba y se integra.
  - Cuando haya que arrancar la aplicación, ejecutar tests o tocar la CI
- [history/changes/README.md](history/changes/README.md): los documentos de cierre de casos anteriores, con sus condiciones de carga.
  - Cuando se trabaje sobre código que ya cambió un caso anterior
  - Cuando se investigue por qué algo es como es
- [history/decisions/README.md](history/decisions/README.md): lo que decidió el humano al responder las preguntas de un caso.
  - Cuando el encargo toque algo que el humano ya decidió
  - Antes de preguntarle algo que quizá ya respondió
