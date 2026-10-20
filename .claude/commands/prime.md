---
description: Inspect the Go/Buffalo TODO application and establish working context
allowed-tools: Bash, Read, Glob, Grep
---

# Comprende el proyecto

Construye el contexto necesario para trabajar en este repositorio y, si `$ARGUMENTS` trae un encargo, para ese encargo. No modifiques ningún archivo.

## Exploración

1. Ejecuta `git status --short --branch`.
2. Lee `docs/README.md` y los documentos que indica en «Siempre».
3. Recorre «Según la tarea»: entra solo en los índices y documentos cuyas condiciones encajen con el encargo, y sigue el índice de cada carpeta de la misma manera. Sin encargo, entra en `docs/knowledge/README.md` y `docs/procedures/README.md` y lee lo que indican en «Siempre». `docs/history/` solo si una condición lo pide.
4. Contrasta con el código lo que vayas a necesitar: las rutas, acciones, modelos, plantillas y tests que nombran los documentos leídos. Si un documento no coincide con el código, dilo: la documentación tiene que estar al día.
5. Para un foco específico usa los subcomandos `/prime:codebase`, `/prime:security` o `/prime:product`.

## Resultado

Resume de forma concisa:

- propósito y stack;
- estructura y responsabilidades principales;
- flujo de una petición desde la ruta hasta la persistencia y la vista;
- cómo arrancar, comprobar y probar la aplicación;
- convenciones que debería respetar un cambio;
- qué documentos has leído y por qué, y cualquier desajuste entre ellos y el código;
- dudas o riesgos que todavía no puedan resolverse con el contexto leído.

Distingue hechos observados de inferencias. No propongas una implementación hasta recibir un encargo.
