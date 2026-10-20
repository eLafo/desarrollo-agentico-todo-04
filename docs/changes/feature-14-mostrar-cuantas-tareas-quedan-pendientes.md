# Mostrar cuántas tareas quedan pendientes

**Issue:** #14 · **Tipo:** feature

## Resumen
La cabecera de la lista decía «N tareas en tu lista». Ahora dice «**X** pendientes de N tareas», con el número de pendientes en negrita, y el texto cambia en el momento al completar o reabrir una tarea, sin recargar la página. El título de la lista pasa a ser «Mis tareas». Las dos últimas cosas vienen de la revisión del humano en el PR #17.

## Qué se construyó
- Debajo del título, «**X** pendientes de N tareas», en el mismo sitio que el texto antiguo, que desaparece (C1). Solo el número va en negrita.
- Singular y plural por separado: «1 pendiente de 1 tarea», «1 pendiente de 3 tareas», «0 pendientes de 1 tarea» (C2).
- Con todas las tareas completadas: «0 pendientes de 3 tareas».
- Sin tareas, el texto no aparece: basta con el mensaje de lista vacía (C3).
- Añadir una tarea la cuenta como pendiente; eliminar la descuenta, y al eliminar la última el texto desaparece (C4).
- Completar o reabrir con el círculo actualiza el texto sin recargar (aclaración aceptada).
- El `<h1>` de la lista dice «Mis tareas» (commit del humano `5dd1eac`). La pestaña y el enlace de navegación siguen diciendo «Tareas».

## Implementación técnica

| Fichero | Papel |
| --- | --- |
| `models/task.go` | `Tasks.Pending()` cuenta las tareas sin completar |
| `actions/tasks.go` | `TasksIndex` y la respuesta 422 de `TasksCreate` ponen `pendingCount` (`tasks.Pending()`) y `listCountSuffix` (`listCountSuffix(tasks)`, p. ej. «pendientes de 3 tareas») |
| `templates/tasks/index.plush.html` | `<h1>Mis tareas</h1>` y, solo si hay tareas, `<p class="list-count"><strong><%= pendingCount %></strong> <%= listCountSuffix %></p>` |
| `public/assets/app.js` | Tras guardar el cambio, sustituye el párrafo `.list-count` de la página por el que devuelve el servidor |
| `templates/application.plush.html` | Sube el script de `app.js?v=3` a `app.js?v=5` |
| `actions/tasks_test.go`, `models/task_test.go` | Tests de Go del texto y de la negrita, de crear, alternar y eliminar, del título y de `Pending()` |
| `e2e/tareas/pendientes-al-completar-y-reabrir.spec.ts` | Test de navegador, con marca `data-sin-recargar` |
| `e2e/tareas/pendientes-en-negrita.spec.ts` | Test de navegador: peso 700 en el `<strong>` y 400 en el párrafo, al cargar, al completar y al reabrir |
| `e2e/tareas/anadir-una-tarea.spec.ts` | Solo cambia el texto esperado a «1 pendiente de 1 tarea» |

Cambios que importan:

- **El servidor calcula todo:** el modelo cuenta, la acción elige singular y plural (`pendiente`/`pendientes`, `tarea`/`tareas`) en `listCountSuffix`, y la plantilla pinta el número dentro de `<strong>` y el resto fuera. El JavaScript no cuenta ni elige plural.
- **JavaScript:** el `fetch` de #15 hace `POST` con `_method=PUT` a `/tasks/{id}/toggle`. La acción redirige (303) a `/` y `fetch` sigue la redirección; con la respuesta, un `DOMParser` busca `.list-count` y, si existe en la página y en la respuesta, `current.replaceWith(saved)`. Se sustituye el elemento entero, no se copia su texto, por eso la negrita viaja con él. Si la respuesta no es correcta, la página no cambia.
- **Sin JavaScript** el formulario sigue igual (`POST` y redirección) y el texto sale bien tras la recarga.
- **Seguridad:** `<%= %>` escapa el contenido en la plantilla; el HTML que se inserta es el de nuestro propio servidor, y el `POST` mantiene su token CSRF.
- **Caché:** `app.js` se sirve con caché de un año, por eso se sube `?v=` (a 5, porque la primera versión de este cambio ya publicó la 4).

## Cómo se usa
Sin acciones nuevas. En la lista de tareas, bajo «Mis tareas», se lee cuántas quedan por hacer, con el número en negrita. Al pulsar el círculo para completar o «Reabrir», el texto baja o sube al momento. Al añadir o eliminar una tarea, la página se recarga y el texto cuadra.

## Configuración
No requiere configuración.

## Cómo se probó
La verificación quedó validada (intento 1 de 3, acta 2, árbol `8c08e4c`). Acta: https://github.com/eLafo/desarrollo-agentico-todo-04/issues/14#issuecomment-6272427479

Resultado: las 28 afirmaciones funcionales bloqueantes cumplidas, los 4 gates (`gofmt`, `go vet`, `go build`, Docker) en verde, y los juicios estructurales, de textos y de fidelidad sin `blocker` (2 `nit` de nombres).

| Afirmación | Resultado |
| --- | --- |
| Texto al cargar (2 de 3, 1 de 1, 1 de 3, 0 de 1, 0 de 3, lista vacía), medido en el DOM, debajo del título y sin el texto antiguo | Cumplida |
| Solo el número va en negrita (peso 700, el resto 400): al cargar, al añadir y al completar o reabrir sin recargar; 29 medidas | Cumplida |
| El único `<h1>` dice «Mis tareas» en todos los estados (vacía, 422, sin JavaScript); la pestaña y la navegación siguen en «Tareas» | Cumplida |
| Añadir cuenta la tarea como pendiente; la primera tarea hace aparecer el texto | Cumplida |
| Completar y reabrir cambia el texto sin recargar, coincide con la base de datos tras recargar, y aguanta varios cambios seguidos | Cumplida |
| Eliminar descuenta y eliminar la última quita el texto | Cumplida |
| Regresiones: 422 con título vacío, sin JavaScript, sin errores de consola ni peticiones fallidas (`POST` 303 + `GET /` 200) | Cumplida |
| Los tests nuevos de Go y de navegador (texto, añadir, eliminar, sin recargar, negrita, título) fallan sobre la base y pasan sobre el cambio | Cumplida |
| Suite de Go (36 tests en `actions`) y de navegador (9 de 9) | En verde |
| Fidelidad: los 11 criterios y peticiones cubiertos, sin cambios no pedidos | Cumplida |

Quedaron para revisión humana (A34, A45): que el texto y el cambio se vean bien en las capturas y los vídeos del acta, y que la negrita y «Mis tareas» se lean bien.

Para repetir las pruebas:

```bash
bash .claude/skills/check/scripts/check.sh        # todos los gates
GO_ENV=test go test -p 1 ./...
npx playwright test
```

## Notas
- **Pestaña y título:** la pestaña (`<title>`) y el enlace de navegación dicen «Tareas» y el `<h1>`, «Mis tareas». Es lo pedido; si se quiere que coincidan, hay que tocar `templates/application.plush.html`.
- **Accesibilidad:** `.list-count` cambia sin recargar pero no lleva `role="status"` ni `aria-live`; un lector de pantalla no anuncia el cambio. Al sustituir el elemento, un `aria-live` en el párrafo no serviría: tendría que ir en un contenedor estable.
- **Guardado fallido o respuestas desordenadas:** el marcado optimista viene de #15 y no hay `.catch`. Si el guardado falla, la tarea se ve marcada pero el texto no cambia ni hay aviso hasta recargar. Si se pulsan dos círculos muy seguidos, el texto queda con la última respuesta que llega. Ningún test lo cubre.
- **Coste:** cada completar o reabrir descarga la página entera de la lista para leer un párrafo.
- **Caché del script:** si se vuelve a tocar `public/assets/app.js`, subir `?v=` en `templates/application.plush.html`.
- **Tests:** `pendientes-en-negrita.spec.ts` comprueba el peso del párrafo y del `<strong>`, no el de cada trozo; no detectaría una negrita en un elemento interior que envolviera el resto. Ningún test fija «debajo del título» (se midió en el DOM).
- **Nombres con `nit`:** `Tasks.Pending` devuelve un número (`PendingCount` sería más explícito) y `Test_TasksDestroy_DiscountsTask` no dice que comprueba la cabecera.
