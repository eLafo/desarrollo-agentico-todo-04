# El botón de añadir dice «Añadir tarea»

**Issue:** #24 · **Tipo:** feature

## Resumen

El botón del formulario de nueva tarea pasa de decir «Añadir» a decir «Añadir tarea», para que diga qué añade. Al revisar la propuesta, el humano pidió además que el botón lleve un `title` con ese mismo texto, para que se vea al pasar el ratón. El resto del formulario no cambia: el campo, su texto de ayuda y los mensajes de error.

## Qué se construyó

- El botón de envío del formulario de la lista de tareas dice exactamente «Añadir tarea».
- El botón lleva `title="Añadir tarea"`, que el navegador muestra al pasar el ratón. Salió de la petición del humano en la [revisión del PR #28](https://github.com/eLafo/desarrollo-agentico-todo-04/pull/28#issuecomment-6295007056): «Añade también al botón un `title` con el texto «Añadir tarea», para que se vea al pasar el ratón.»
- Sin cambios: el campo «Nueva tarea», su texto de ayuda «Añadir una tarea…» y los mensajes de error.

## Implementación técnica

| Fichero | Papel |
| --- | --- |
| `templates/tasks/index.plush.html` | Cambia el texto del `<button type="submit">` del formulario de nueva tarea y le añade `title="Añadir tarea"`. |
| `docs/knowledge/features/tareas.md` | La sección «Añadir» describe el botón con su texto nuevo y con el mismo texto al pasar el ratón. |
| `e2e/tareas/boton-anadir-tarea.spec.ts` | Test de navegador nuevo: texto y nombre accesible del botón, y nombre y texto de ayuda del campo. |
| `e2e/tareas/boton-anadir-tarea-al-pasar-el-raton.spec.ts` | Test de navegador nuevo: el botón lleva `title="Añadir tarea"`. |

- Solo cambia una línea de la plantilla: el texto fijo del botón y su atributo `title`; no hay cambios de rutas, acciones, modelos, estilos ni lógica.
- El `title` no altera el nombre accesible del botón, que sigue siendo su texto.
- Los tests buscan el formulario por su nombre accesible «Añadir tarea» y comprueban el botón con `toHaveText`, `toHaveAccessibleName` y `toHaveAttribute('title', …)`.

## Cómo se usa

En la pantalla de la lista, debajo de las tareas, escribe el título en «Nueva tarea» y pulsa «Añadir tarea» (con un clic o con Intro). Al pasar el ratón sobre el botón, el navegador muestra «Añadir tarea».

## Configuración

No requiere configuración.

## Cómo se probó

Acta de la verificación, validada en el intento 1 de 3 sobre el commit `5f0aca9`: [comentario en la issue](https://github.com/eLafo/desarrollo-agentico-todo-04/issues/24#issuecomment-6296470253).

| Afirmación | Resultado |
| --- | --- |
| `boton-anadir-tarea.spec.ts` falla en `todo-06` («Añadir») y pasa con el cambio | Verificado |
| El botón dice «Añadir tarea» (texto y nombre para lectores de pantalla) con la lista vacía, con tareas y tras un error; se lee entero a 1280, 390 y 320 px | Verificado |
| El campo, su texto de ayuda y los mensajes («El título es obligatorio.», «El título no puede superar los 200 caracteres.», con `422`) son los de `todo-06`; en el formulario solo cambian el texto del botón y su `title` | Verificado |
| `boton-anadir-tarea-al-pasar-el-raton.spec.ts` falla en `todo-06` (sin `title`) y pasa con el cambio | Verificado |
| El botón lleva `title="Añadir tarea"` en los tres estados y en el HTML que envía el servidor; su nombre para lectores de pantalla sigue siendo su texto | Verificado |
| Sin regresiones: `go test -p 1 ./...` y `npx playwright test` (9 tests) en verde; añadir con un clic y con Intro funciona; la lista y la pantalla de editar son iguales que en `todo-06` salvo el botón, y los demás `title` no cambian | Verificado |
| `gofmt`, `go vet`, `go build` y `docker build` en verde; cada cambio del diff responde a un criterio o a la petición, y la documentación de la lista está al día | Verificado |

Para repetirlo:

```bash
npx playwright test e2e/tareas/boton-anadir-tarea.spec.ts e2e/tareas/boton-anadir-tarea-al-pasar-el-raton.spec.ts
GO_ENV=test go test -p 1 ./...
bash .claude/skills/verify/scripts/red-green.sh --base de62988 <test>
```

## Notas

- Sin preguntas con opciones respondidas por el humano: no hay registro de decisiones. La petición del `title` fue un comentario de revisión, no una respuesta a una pregunta con opciones.
- El botón es más ancho y el campo pierde otro tanto. El texto cabe entero y en una línea; a 320 px el botón baja a su propia línea, como ya hacía en `todo-06`.
- Puntos para el humano que no bloquean: si le gusta cómo queda el botón más ancho; una sugerencia menor sobre el título del test del texto, sin aplicar; y que el mensaje de más de 200 caracteres solo lo fija un test de la pantalla de editar, no uno de la pantalla de añadir.
