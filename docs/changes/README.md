# Cambios documentados

Cada caso que cierra el workflow `adw` deja aquí su documento de cierre: qué se construyó, cómo se usa, cómo se probó y qué conviene saber para trabajar sobre ello. Lo escribe la skill `document` en la rama del caso, y viaja en el PR con el código.

Este índice es condicional: cada entrada dice cuándo conviene cargar su documento. Las skills `feature`, `bug` y `chore` lo consultan al planificar y leen solo los documentos cuyas condiciones encajan con el encargo.

## Formato de las entradas

```markdown
- [<tipo>-<N>-<slug>.md](<tipo>-<N>-<slug>.md): <qué cambió, en una línea>
  - Cuando se trabaje con <rutas, acciones o plantillas concretas>
  - Cuando se modifique <modelo, campo o regla relacionados>
  - Cuando se investigue <un problema concreto de esa área>
```

Entre tres y cinco condiciones por documento, y las entradas en orden por número de issue. Si un documento ya tiene entrada, se actualizan sus condiciones en lugar de añadir otra.

## Índice

- [bug-15-al-completar-una-tarea-al-recargar-vuelve-a-estar-pendiente.md](bug-15-al-completar-una-tarea-al-recargar-vuelve-a-estar-pendiente.md): el JavaScript guarda con un POST el completar y reabrir una tarea, que antes se perdía al recargar
  - Cuando se trabaje con `public/assets/app.js` o con la versión del script (`?v=`) en `templates/application.plush.html`
  - Cuando se modifique la ruta `/tasks/{id}/toggle`, su método HTTP o el formulario de completar y reabrir
  - Cuando se cambie la protección CSRF o los campos que envía el formulario de una tarea
  - Cuando se investigue una tarea que cambia en pantalla pero vuelve a su estado al recargar
  - Cuando se escriban tests de navegador en `e2e/tareas/` que dependan de recargar la página
