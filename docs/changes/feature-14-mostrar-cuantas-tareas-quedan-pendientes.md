# Mostrar cuántas tareas quedan pendientes

**Issue:** #14 · **Tipo:** feature

## Resumen
La cabecera de la lista decía «N tareas en tu lista». Ahora dice «X pendientes de N tareas», donde X son las tareas sin completar, y el texto cambia en el momento al completar o reabrir una tarea, sin recargar la página.

## Qué se construyó
- Debajo del título, «X pendientes de N tareas», en el mismo sitio que el texto antiguo, que desaparece (C1).
- Singular y plural por separado: «1 pendiente de 1 tarea», «1 pendiente de 3 tareas», «0 pendientes de 1 tarea» (C2).
- Con todas las tareas completadas: «0 pendientes de 3 tareas».
- Sin tareas, el texto no aparece: basta con el mensaje de lista vacía (C3).
- Añadir una tarea la cuenta como pendiente; eliminar la descuenta, y al eliminar la última el texto desaparece (C4).
- Completar o reabrir con el círculo actualiza el texto sin recargar (aclaración, opción a).

## Implementación técnica

| Fichero | Papel |
| --- | --- |
| `models/task.go` | `Tasks.Pending()` cuenta las tareas sin completar |
| `actions/tasks.go` | `listCount(tasks)` compone el texto; `TasksIndex` y la respuesta 422 de `TasksCreate` lo ponen como `listCount` |
| `templates/tasks/index.plush.html` | Pinta `<p class="list-count"><%= listCount %></p>` solo si hay tareas |
| `public/assets/app.js` | Tras guardar el cambio, copia el texto de `.list-count` de la página que devuelve el servidor |
| `templates/application.plush.html` | Sube el script de `app.js?v=3` a `app.js?v=4` |
| `actions/tasks_test.go`, `models/task_test.go` | Tests de Go del texto, de crear, alternar y eliminar, y de `Pending()` |
| `e2e/tareas/pendientes-al-completar-y-reabrir.spec.ts` | Test de navegador nuevo, con marca `data-sin-recargar` |
| `e2e/tareas/anadir-una-tarea.spec.ts` | Solo cambia el texto esperado a «1 pendiente de 1 tarea» |

Cambios que importan:

- **El servidor calcula todo:** el modelo cuenta, la acción elige singular y plural (`pendiente`/`pendientes`, `tarea`/`tareas`) y la plantilla solo pinta. El JavaScript no cuenta ni elige plural.
- **JavaScript:** el `fetch` de #15 ya hacía `POST` con `_method=PUT` a `/tasks/{id}/toggle`. Esa acción redirige (303) a `/` y `fetch` sigue la redirección; con la respuesta, un `DOMParser` lee `.list-count` y se copia su `textContent` a la página. El texto mostrado es siempre el calculado con lo guardado.
- **Sin JavaScript** el formulario sigue igual (`POST` y redirección) y el texto sale bien tras la recarga.
- **Seguridad:** se escribe con `textContent`, sin `innerHTML`; el `POST` mantiene su token CSRF.
- **Caché:** `app.js` se sirve con caché de un año, por eso se sube `?v=`.

## Cómo se usa
Sin acciones nuevas. En la lista de tareas, bajo «Tareas», se lee cuántas quedan por hacer. Al pulsar el círculo para completar o «Reabrir», el texto baja o sube al momento. Al añadir o eliminar una tarea, la página se recarga y el texto cuadra.

## Configuración
No requiere configuración.

## Cómo se probó
La verificación quedó validada (intento 1 de 3, árbol `ebf237b`). Acta: https://github.com/eLafo/desarrollo-agentico-todo-04/issues/14#issuecomment-6271687025. Plan de verificación: https://github.com/eLafo/desarrollo-agentico-todo-04/issues/14#issuecomment-6271428454

Se verificaron 34 afirmaciones del plan (31 bloqueantes) más 4 añadidas, todas sin fallos.

| Afirmación | Resultado |
| --- | --- |
| C1 a C3 y plurales (2 de 3, 1 de 1, 1 de 3, 0 de 1, 0 de 3, lista vacía) al cargar, medidos en el DOM | Cumplida |
| C4: añadir cuenta la tarea como pendiente (3 de 4); la primera tarea hace aparecer el texto | Cumplida |
| Completar y reabrir cambia el texto sin recargar y coincide con la base de datos tras recargar; varios cambios seguidos | Cumplida |
| Eliminar descuenta y eliminar la última quita el texto | Cumplida |
| Regresiones: 422 con título vacío, sin JavaScript, sin errores de consola ni peticiones fallidas | Cumplida |
| Los tests nuevos de Go y de navegador fallan sobre la base y pasan sobre el cambio | Cumplida |
| Mutante (el cambio con el `app.js` de la base): el test de navegador falla al esperar el texto sin recargar | Detecta el «sin recargar» |
| Juicios de estructura (E1 a E6), textos (F1) y fidelidad (D1, D2) | Sin `blocker`; 2 `nit` de nombres |
| Suite de Go (34 tests) y de navegador (8 de 8), `gofmt`, `go vet`, `go build` y Docker | En verde |

Para repetir las pruebas:

```bash
npx playwright test e2e/tareas/pendientes-al-completar-y-reabrir.spec.ts e2e/tareas/anadir-una-tarea.spec.ts
bash .claude/skills/check/scripts/check.sh
```

## Notas
- **Accesibilidad:** `.list-count` cambia sin recargar pero no lleva `role="status"` ni `aria-live`; un lector de pantalla no anuncia el cambio.
- **Guardado fallido o respuestas desordenadas:** el marcado optimista viene de #15. Si el guardado falla, la tarea se ve marcada pero el texto no cambia hasta recargar. Si se pulsan dos círculos muy seguidos, el texto queda con la última respuesta que llega. Ningún test lo cubre.
- **Coste:** cada completar o reabrir descarga la página entera de la lista para leer un párrafo.
- **Caché del script:** si se vuelve a tocar `public/assets/app.js`, subir `?v=` en `templates/application.plush.html`.
- Nombres con `nit`: `Tasks.Pending` devuelve un número (`PendingCount` sería más explícito) y `Test_TasksDestroy_DiscountsTask` no dice que comprueba la cabecera.
- Ningún test automatizado elimina una tarea completada ni comprueba la posición bajo el título; ambos se comprobaron a mano en la interfaz.
