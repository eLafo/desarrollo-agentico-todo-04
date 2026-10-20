# Al completar una tarea, al recargar vuelve a estar pendiente

**Issue:** #15 · **Tipo:** bug

## Resumen
Completar o reabrir una tarea con el círculo de la izquierda cambiaba la página al momento, pero el cambio no se guardaba: al recargar, la tarea volvía a su estado anterior. El JavaScript enviaba la petición de guardado con un método que el servidor no acepta; ahora envía la misma petición que el formulario.

## Qué se construyó
- Completar una tarea con el círculo se guarda y sigue completada tras recargar (C1).
- Reabrir una tarea con «Reabrir» se guarda y sigue pendiente tras recargar (C2).
- Dos tests de navegador que recargan la página y comprueban el estado guardado.

## Implementación técnica

| Fichero | Papel |
| --- | --- |
| `public/assets/app.js` | Al enviar el formulario de una tarea, además de actualizar la página, guarda el cambio |
| `templates/application.plush.html` | Sube la versión del script de `app.js?v=2` a `app.js?v=3` |
| `e2e/tareas/completar-una-tarea.spec.ts` | Test nuevo de C1 |
| `e2e/tareas/reabrir-una-tarea.spec.ts` | Test nuevo de C2 |

Cambios que importan:

- **Causa:** el `fetch(form.action)` era un `GET` sin cuerpo, y la ruta de alternar solo responde a `PUT` (a través del formulario). El servidor no cambiaba nada.
- **Arreglo:** `fetch(form.action, { method: "POST", body: new URLSearchParams(new FormData(form)) })`. Envía los campos del formulario, incluidos `authenticity_token` y `_method=PUT`, que Buffalo convierte en `PUT /tasks/{id}/toggle`. No hay cambios en Go.
- **Caché:** el script se sirve con caché de un año, por eso se sube el `?v=` de la plantilla. Si se vuelve a cambiar `app.js`, hay que subirlo otra vez.
- **Sin JavaScript** el formulario sigue funcionando como antes (`POST` con redirección 303).

## Cómo se usa
Sin cambios para el usuario: pulsa el círculo de una tarea para completarla o «Reabrir» para volver a dejarla pendiente. La página cambia al momento y el estado queda guardado.

## Configuración
No requiere configuración.

## Cómo se probó
La verificación quedó validada (intento 1 de 3, árbol `1c43448`). Acta: https://github.com/eLafo/desarrollo-agentico-todo-04/issues/15#issuecomment-6270266316

| Afirmación | Resultado |
| --- | --- |
| Los dos tests nuevos fallan en la base (justo tras recargar) y pasan en el cambio | Cumplida |
| C1: completar y recargar deja la tarea completada, en la página y en la base de datos | Cumplida |
| C2: reabrir y recargar deja la tarea pendiente, en la página y en la base de datos | Cumplida |
| Varios cambios seguidos quedan guardados; el resto de la lista no cambia | Cumplida |
| Sin JavaScript, completar y reabrir siguen guardándose | Cumplida |
| Un `GET` a la ruta de alternar no cambia nada (404) y un `POST` sin token CSRF se rechaza (403) | Cumplida |
| Suite de Go y de navegador (7 de 7), `gofmt`, `go vet` y `go build` | En verde |

Para repetir las pruebas:

```bash
npx playwright test e2e/tareas/completar-una-tarea.spec.ts e2e/tareas/reabrir-una-tarea.spec.ts
bash .claude/skills/check/scripts/check.sh
```

## Notas
- Si el guardado falla (sin red, token caducado, error del servidor), la página no avisa: muestra el cambio y al recargar vuelve el estado anterior. Ya ocurría antes y queda fuera de esta issue.
- Con `GO_ENV=test` el middleware CSRF no comprueba nada, así que los tests de navegador no detectarían que el JavaScript dejara de enviar el token. La verificación lo cubrió contra la aplicación en desarrollo.
- Los tests esperan la respuesta del servidor antes de comprobar la página, así que no distinguen si el cambio se ve al pulsar o al llegar la respuesta.
