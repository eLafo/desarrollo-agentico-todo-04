# La pestaña del navegador dice «Mis tareas»

**Issue:** #25 · **Tipo:** feature

## Resumen
La etiqueta `<title>` de todas las páginas que usan la plantilla común pasa de «Tareas» a «Mis tareas», de modo que la pestaña del navegador nombra la aplicación. El encabezado visible de la lista sigue diciendo «Tareas».

## Qué se construyó
- El título del documento es exactamente «Mis tareas» en la lista y en la edición de una tarea, también con error de validación (criterio 1).
- El encabezado `<h1>` de la lista sigue diciendo «Tareas» (criterio 2). El enlace de la navegación lateral y el encabezado «Editar tarea» tampoco cambian.

## Implementación técnica
| Fichero | Papel |
| --- | --- |
| `templates/application.plush.html` | Plantilla común: `<title>Tareas</title>` pasa a `<title>Mis tareas</title>`. Es el único cambio de código. |
| `actions/tasks_test.go` | Cinco tests nuevos de acción y un ayudante, `assertPageTitle`. |
| `docs/knowledge/features/tareas.md` | La documentación viva recoge que la pestaña dice «Mis tareas» y que «Tareas» es el enlace lateral y el encabezado de la lista. |

- `assertPageTitle` exige una sola etiqueta `<title>` y que su texto sea «Mis tareas».
- Cuatro tests comprueban el título: la lista, la lista tras un POST inválido (422), la edición y la edición tras un PUT inválido (422). El quinto comprueba que la lista sigue con `<h1>Tareas</h1>`.
- Al vivir el título en la plantilla común, no hay un valor por pantalla: las dos pantallas comparten el mismo.

## Cómo se usa
Abrir la lista (`/`) o la edición de una tarea (`/tasks/{id}/edit`): la pestaña dice «Mis tareas». No hay nada que activar.

## Configuración
No requiere configuración.

## Cómo se probó
Acta de la verificación, validada en el intento 1 de 3 sobre el commit `de62988` con cambios sin registrar: https://github.com/eLafo/desarrollo-agentico-todo-04/issues/25#issuecomment-6294900078

| Afirmación | Resultado |
| --- | --- |
| Los cuatro tests del título fallan sobre la base y pasan con el cambio | Cumplido |
| El test del encabezado `<h1>Tareas</h1>` pasa en la base y en el cambio | Cumplido |
| El servidor y el navegador dan «Mis tareas», con una sola etiqueta `<title>`, en cinco pantallas (también la lista vacía) | Cumplido |
| El enlace lateral sigue diciendo «Tareas»; una tarea inexistente sigue dando 404 con la página de error de Buffalo | Cumplido |
| Suite de Go, 7 tests de navegador y gates (gofmt, vet, build, docker) | En verde |
| Juicios de estructura, textos y fidelidad | Sin hallazgos que bloqueen |

Para repetir las pruebas, con la base de test migrada y el `.env` del worktree cargado:

```bash
GO_ENV=test go test -p 1 ./actions/ -run Test_ActionSuite -testify.m 'PageTitle|Heading' -v
bash .claude/skills/check/scripts/check.sh
```

## Notas
- Las páginas de error del framework (404, 500) no usan la plantilla común y quedan fuera del cambio.
- La marca, el enlace lateral y el encabezado «Editar tarea» no cambian.
- Un navegador sin ventana no dibuja la pestaña: la verificación midió el título del documento, y ver la pestaña real queda para quien revise.
- Los tests nuevos no llevan comentario propio, como el resto de tests del repositorio; el juicio lo dio por bueno con confianza baja.
- No hubo decisiones del humano en el caso.
