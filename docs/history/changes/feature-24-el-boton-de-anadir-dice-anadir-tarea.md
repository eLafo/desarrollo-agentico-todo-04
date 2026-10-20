# El botón de añadir dice «Añadir tarea»

**Issue:** #24 · **Tipo:** feature

## Resumen

El botón del formulario de nueva tarea pasa de decir «Añadir» a decir «Añadir tarea», para que diga qué añade. El resto del formulario no cambia: el campo, su texto de ayuda y los mensajes de error.

## Qué se construyó

- El botón de envío del formulario de la lista de tareas dice exactamente «Añadir tarea».
- Sin cambios: el campo «Nueva tarea», su texto de ayuda «Añadir una tarea…» y los mensajes de error.

## Implementación técnica

| Fichero | Papel |
| --- | --- |
| `templates/tasks/index.plush.html` | Cambia el texto del `<button type="submit">` del formulario de nueva tarea. |
| `docs/knowledge/features/tareas.md` | La sección «Añadir» describe el botón con su texto nuevo. |
| `e2e/tareas/boton-anadir-tarea.spec.ts` | Test de navegador nuevo: texto y nombre accesible del botón, y nombre y texto de ayuda del campo. |

- Solo cambia el texto del botón; no hay cambios de rutas, acciones, modelos ni estilos.
- El test busca el formulario por su nombre accesible «Añadir tarea» y comprueba el botón con `toHaveText` y `toHaveAccessibleName`.

## Cómo se usa

En la pantalla de la lista, debajo de las tareas, escribe el título en «Nueva tarea» y pulsa «Añadir tarea».

## Configuración

No requiere configuración.

## Cómo se probó

Acta de la verificación, validada en el intento 1 de 3: [comentario en la issue](https://github.com/eLafo/desarrollo-agentico-todo-04/issues/24#issuecomment-6294937865). Planes: [implementación](https://github.com/eLafo/desarrollo-agentico-todo-04/issues/24#issuecomment-6294454552) y [verificación](https://github.com/eLafo/desarrollo-agentico-todo-04/issues/24#issuecomment-6294687428).

| Afirmación | Resultado |
| --- | --- |
| El test nuevo falla en `todo-06` (recibe «Añadir») y pasa con el cambio | Verificado |
| El botón dice «Añadir tarea», único y de envío, con la lista vacía, con tareas y tras un error | Verificado |
| El campo, su texto de ayuda y el mensaje de error son iguales que en `todo-06` | Verificado |
| Las suites de Go y de navegador completas: 8 tests de navegador pasan | Verificado |

Para repetirlo:

```bash
bash .claude/skills/verify/scripts/red-green.sh --base de62988 e2e/tareas/boton-anadir-tarea.spec.ts
go test -p 1 ./...
npx playwright test
```

## Notas

- Sin preguntas con opciones respondidas por el humano: no hay registro de decisiones.
- El botón es unos 38 px más ancho y el campo pierde otro tanto (de 653 a 615 px en el ordenador, de 213 a 176 px a 390 px). El texto cabe entero y en una línea; a 320 px el botón baja a su propia línea, como ya hacía en `todo-06`.
- Quedó como duda para el humano si le gusta el formulario con el botón más ancho.
- Sugerencias menores del acta, sin aplicar: el título del test no nombra el campo, y un comentario del test dice «sigue igual» aunque compara con valores fijos.
- Sin test: el límite de 200 caracteres y el autocompletado desactivado del campo en la pantalla de añadir; no cambian aquí y solo los cubrió la comparación de medidas con `todo-06`.
