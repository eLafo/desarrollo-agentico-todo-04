---
name: setup_environment
description: Deja la copia de trabajo en la rama del caso, al día, con el entorno de desarrollo en marcha.
skills: [env:start]
---

# setup_environment

## Entradas

- `work_item`: la petición de trabajo (identidad, tipo y título).
- La rama del caso, si ya se asignó.

## Qué hace

1. Si el caso no tiene rama, asígnale una a partir de la petición y regístrala. La rama no cambia durante el caso.
2. Sitúa la copia de trabajo del caso, en la que corre la sesión, en la rama, al día con su versión publicada. Si la rama todavía no existe, créala desde la base.
3. Arranca el entorno con el comando del proyecto `env:start` (`.claude/commands/env/start.md`) y comprueba que responde.

## Resultado

La copia de trabajo en la rama del caso y el entorno respondiendo. Si el entorno no arranca, es un fallo.
