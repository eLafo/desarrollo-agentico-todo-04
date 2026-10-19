# Decir qué tarea se ha eliminado

**Issue:** #20 · **Tipo:** feature

## Resumen
Al eliminar una tarea, el aviso decía solo «Tarea eliminada.», y al borrar varias seguidas no se sabía cuál se había quitado. Ahora el aviso nombra la tarea: «Tarea eliminada: Comprar pan.». Si el título ya acaba en punto no se añade otro («Tarea eliminada: Llamar a Juan.»), por petición del humano en la issue.

## Qué se construyó
- El aviso de eliminar incluye el título de la tarea, entero y tal cual se guardó; `<`, `&` y similares se ven como texto (C1, C2).
- Si el título acaba en «.», el aviso no añade otro punto (petición del humano).
- Los avisos de crear y actualizar no cambian (C3).

## Implementación técnica

| Fichero | Papel |
| --- | --- |
| `actions/tasks.go` | `TasksDestroy` compone el aviso con el título de la tarea que ya había leído |
| `actions/tasks_test.go` | Tres tests nuevos: título normal, título de 200 caracteres con `<`, `&` y `á`, y título que acaba en punto |
| `e2e/tareas/eliminar-dice-que-tarea.spec.ts` | Test de navegador nuevo: tres eliminaciones seguidas con comparación exacta del aviso |
| `e2e/tareas/eliminar-pide-confirmacion.spec.ts` | El texto esperado pasa a «Tarea eliminada: Enviar la factura.» |

Cambios que importan:

- **Dónde se compone:** en `TasksDestroy`, con la `task` ya cargada; no hay segunda consulta. La plantilla `templates/tasks/index.plush.html` no cambia respecto a la base.
- **Escapado:** el aviso se imprime con `<%= %>` en la plantilla del flash, que escapa el HTML. No se usa `raw` ni `template.HTML`.
- **Regla del punto:** `strings.HasSuffix(task.Title, ".")`; si es falso, se añade «.».
- **Sin recortes:** el título sale completo, sea cual sea su longitud (máximo `models.TitleMaxLength`, 200).

## Cómo se usa
Sin pasos nuevos: se elimina una tarea con su botón (tras confirmar) y el aviso verde dice cuál.

## Configuración
No requiere configuración.

## Cómo se probó
La verificación quedó validada (intento 1 de 1, árbol `44c1483`). Acta: https://github.com/eLafo/desarrollo-agentico-todo-04/issues/20#issuecomment-6277156868

| Afirmación | Resultado |
| --- | --- |
| Tras eliminar «Comprar pan», el aviso dice «Tarea eliminada: Comprar pan.»; dos seguidas, cada aviso nombra la suya y no reaparece al recargar | Cumplida |
| Un título de 200 caracteres sale entero; `<`, `&` y `%` se ven como texto | Cumplida |
| Crear y actualizar siguen con «Tarea creada.» y «Tarea actualizada.» | Cumplida |
| «Llamar a Juan.» da «Tarea eliminada: Llamar a Juan.», sin doble punto | Cumplida |
| Los tests nuevos fallan en la base y pasan en el cambio; el de la regla del punto falla también en el intento anterior (`9ff6482`) | Cumplida |
| Suites de Go y de navegador (8 de 8), `gofmt`, `go vet`, `go build`; la confirmación y cancelación al eliminar siguen probadas | En verde |

Para repetir las pruebas:

```bash
bash .claude/skills/check/scripts/check.sh
GO_ENV=test go test -p 1 ./actions -run Test_ActionSuite -testify.m 'Test_TasksDestroy'
npx playwright test e2e/tareas/eliminar-dice-que-tarea.spec.ts
```

## Notas
- **La regla del punto solo exime el «.».** Tras «?», «!» o «…» (U+2026) se añade el punto: «Tarea eliminada: ¿Llamo al fontanero?.». La petición solo hablaba del punto; ampliarlo es decisión del humano.
- **Aviso en `TasksDestroy`, escapado en la plantilla.** Quien cambie el texto del aviso lo hace en la acción; quien toque la plantilla del flash debe conservar `<%= %>` para que el título no se interprete como HTML.
- **Título largo sin espacios (pendiente de decisión del humano).** Un título de 200 caracteres sin espacios se sale de la caja del aviso y de la ventana de 1280 px, porque `.flash` no tiene `overflow-wrap`. Con espacios se reparte en líneas y cabe. No se ha cambiado el CSS.
- Los tests de Go buscan el aviso con `Contains` en toda la página; la comparación exacta la hacen los de navegador. El título largo del test de navegador está escrito a mano, sin relación con `TitleMaxLength`.
