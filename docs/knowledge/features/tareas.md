# Lista de tareas

Lo que ve y puede hacer el usuario con sus tareas. Todo pasa en dos pantallas: la lista (`/`) y la edición de una tarea (`/tasks/{task_id}/edit`). En las dos, la pestaña del navegador dice «Mis tareas». El código está en `actions/tasks.go`, `models/task.go`, `templates/tasks/` y `public/assets/app.js`.

## La lista

- El enlace de la navegación lateral y el título que encabeza la lista dicen «Tareas».
- Debajo del título, el número de tareas: «3 tareas en tu lista», y en singular «1 tarea en tu lista».
- Las tareas salen en el orden que deja el usuario (ver [Ordenar](#ordenar)). Una tarea nueva sale al final.
- Cada tarea muestra:
  - un botón redondo para completarla o reabrirla;
  - su título;
  - la etiqueta «Completada», si lo está;
  - dos botones para subirla o bajarla un puesto, un enlace para editarla y un botón para eliminarla.
- Los botones solo tienen icono, y su nombre accesible incluye el título: «Completar tarea: Comprar pan», «Reabrir tarea: …», «Subir tarea: …», «Bajar tarea: …», «Editar tarea: …», «Eliminar tarea: …».
- Sin tareas, la lista muestra «No hay tareas todavía. Añade la primera con el formulario.», y el número dice «0 tareas en tu lista».

## Reglas del título

Las comprueba `Task.Validate` al crear y al editar:

| Regla | Mensaje si no se cumple |
| --- | --- |
| Es obligatorio. Un título solo con espacios cuenta como vacío | «El título es obligatorio.» |
| Tiene como mucho 200 caracteres (`models.TitleMaxLength`), contados sin los espacios del principio y del final | «El título no puede superar los 200 caracteres.» |

Antes de guardar, la acción quita los espacios del principio y del final. Los campos de título del navegador tienen también `maxlength="200"`.

## Añadir

- El formulario está debajo de la lista: el campo «Nueva tarea», con el texto de ayuda «Añadir una tarea…», y el botón «Añadir».
- Si el título es válido, la tarea se crea pendiente, aparece al final de la lista y sale el aviso «Tarea creada.».
- Si no, la lista se vuelve a mostrar con `422`: el campo conserva lo escrito, queda marcado como inválido y el mensaje sale debajo.

## Editar

- Se llega desde el enlace de editar de cada tarea. La pantalla tiene el enlace «← Volver a la lista», el título «Editar tarea» y el campo «Título» con el título actual. Debajo, la ayuda: «Hasta 200 caracteres. Los demás datos de la tarea no cambian.».
- «Guardar cambios» cambia solo el título: si la tarea estaba completada, lo sigue estando. Después vuelve a la lista con el aviso «Tarea actualizada.».
- Con un título inválido, la pantalla se vuelve a mostrar con `422` y el mensaje. «Cancelar» vuelve a la lista sin cambiar nada.
- Si la tarea no existe, `404`.

## Completar y reabrir

- El botón de cada tarea la alterna entre pendiente y completada. Una tarea completada lleva la etiqueta «Completada», y su botón pasa a reabrirla.
- Sin JavaScript, el formulario envía `PUT /tasks/{task_id}/toggle`, la acción cambia el estado y vuelve a la lista. No hay aviso.
- Con JavaScript, `app.js` cambia la tarea en la página sin recargarla (la etiqueta, el estado del botón y su nombre accesible) y guarda el cambio con la misma petición que el formulario, enviada con `fetch`. Al recargar, la tarea sigue como se dejó.

## Ordenar

- Con el ratón, se arrastra una tarea y se suelta en otro lugar de la lista. Mientras se arrastra, la lista muestra dónde quedará. Si se suelta fuera de la lista o se pulsa Escape, vuelve a su sitio.
- «Subir» y «Bajar» la mueven un puesto. Con el teclado se llega a ellos con el tabulador y se pulsan con Intro o la barra espaciadora; el foco se queda en el botón. En la primera tarea «Subir» no hace nada, y en la última, «Bajar».
- Las completadas se ordenan igual que las pendientes. Completar o reabrir una tarea no la cambia de sitio.
- El orden se guarda: se mantiene al recargar y desde cualquier navegador o dispositivo.
- Sin JavaScript no se puede arrastrar. Los botones envían `PUT /tasks/{task_id}/move` con el lugar de destino, y la lista se vuelve a mostrar. Con JavaScript, `app.js` mueve la tarea en la página sin recargarla y guarda el lugar con la misma petición, enviada con `fetch`.
- No hay aviso.

## Eliminar

- El botón de eliminar pide confirmación: «¿Seguro que quieres eliminar esta tarea?». Si se cancela, no pasa nada. Sin JavaScript, elimina sin preguntar.
- Al eliminar, la tarea desaparece de la lista y sale el aviso «Tarea eliminada.».
- Si la tarea no existe, `404`.

## Avisos

Los avisos salen arriba, en la primera pantalla que se muestra después de la acción, y se anuncian a los lectores de pantalla (`role="status"`):

| Acción | Aviso |
| --- | --- |
| Añadir | «Tarea creada.» |
| Editar | «Tarea actualizada.» |
| Eliminar | «Tarea eliminada.» |
| Completar o reabrir | Ninguno |
| Ordenar | Ninguno |
