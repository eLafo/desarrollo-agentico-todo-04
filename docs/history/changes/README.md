# Cambios documentados

Cada caso que cierra el workflow `adw` deja aquí su documento de cierre: qué se construyó, cómo se usa, cómo se probó y qué conviene saber para trabajar sobre ello. Lo escribe la skill `document` en la rama del caso, y viaja en el PR con el código.

Este índice es condicional: cada entrada dice cuándo conviene cargar su documento. Se llega a él desde [`docs/README.md`](../../README.md): quien planifica, como las skills `feature`, `bug` y `chore`, lee solo los documentos cuyas condiciones encajan con el encargo.

Un documento de cierre cuenta qué pasó en su caso, y no se reescribe cuando otro caso cambia el mismo código. Lo vigente está en `docs/knowledge/` y `docs/procedures/`.

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
- [feature-24-el-boton-de-anadir-dice-anadir-tarea.md](feature-24-el-boton-de-anadir-dice-anadir-tarea.md): el botón del formulario de nueva tarea pasa de «Añadir» a «Añadir tarea»
  - Cuando se trabaje con `templates/tasks/index.plush.html` o con el formulario de nueva tarea
  - Cuando se cambie el texto, el ancho o la disposición del botón de añadir, también en pantallas estrechas
  - Cuando se escriban tests de navegador en `e2e/tareas/` que busquen el botón o el formulario por su nombre
  - Cuando se modifique el campo «Nueva tarea», su texto de ayuda o sus mensajes de error
