# Limitar el título de una tarea a 120 caracteres

**Issue:** #22 · **Tipo:** feature

## Resumen
El título de una tarea podía tener hasta 200 caracteres y en la lista los títulos tan largos se leían mal. El límite baja a 120: el servidor rechaza más de 120 con un mensaje claro al añadir y al editar, los campos de título no dejan escribir más, y la ayuda de la pantalla de edición dice el límite nuevo.

## Qué se construyó
- Al añadir o editar, un título de más de 120 caracteres se rechaza con «El título no puede superar los 120 caracteres.» (C1).
- Un título de 120 caracteres exactos se acepta y se guarda entero (C2).
- Los campos «Nueva tarea» y «Título» no dejan escribir más de 120 caracteres, y la ayuda de la edición dice «Hasta 120 caracteres. Los demás datos de la tarea no cambian.» (C3).
- Tests de Go para el modelo y las acciones, y dos tests de navegador.

## Implementación técnica

| Fichero | Papel |
| --- | --- |
| `models/task.go` | `TitleMaxLength` pasa de 200 a 120; la validación y el mensaje usan esa constante |
| `templates/tasks/index.plush.html` | `maxlength="120"` en el campo «Nueva tarea» |
| `templates/tasks/edit.plush.html` | `maxlength="120"` en el campo «Título» y ayuda con el límite nuevo |
| `models/task_test.go`, `actions/tasks_test.go` | Tests de 120 (válido) y 121 (rechazado) al crear y al editar, y de la ayuda |
| `e2e/tareas/anadir-titulo-hasta-120-caracteres.spec.ts`, `e2e/tareas/editar-titulo-hasta-120-caracteres.spec.ts` | Escribir de más deja 120 caracteres, se guarda y sigue tras recargar |
| `docs/knowledge/features/tareas.md`, `docs/knowledge/overview.md` | Cuentan el límite de 120 |

Cambios que importan:

- **La regla vive en el modelo.** Solo cambia la constante `TitleMaxLength`; las acciones no se tocan. El rechazo usa el camino de `422` que ya existía.
- **Sin migración.** La columna `title` sigue en `varchar(200)`, así que nada válido puede dar un error de base de datos.
- **El 120 está escrito a mano** en las dos plantillas (`maxlength` y la ayuda), además de en `TitleMaxLength`. Si el límite vuelve a cambiar hay que tocar los tres sitios.
- **Los tests de Go ya no usan la constante** en los títulos de prueba: escriben 120 y 121 a mano, para que un cambio accidental de la constante falle.

## Cómo se usa
- **Añadir:** se escribe en «Nueva tarea»; el campo corta en 120 caracteres. Si un envío directo trae más, la página vuelve con `422` y el mensaje bajo el campo.
- **Editar:** igual en «Título», con la ayuda que dice el límite.
- Los espacios del principio y del final no cuentan: se quitan antes de validar.

## Configuración
No requiere configuración.

## Cómo se probó
La verificación quedó validada (intento 1 de 3, árbol `aafba40`). Acta: https://github.com/eLafo/desarrollo-agentico-todo-04/issues/22#issuecomment-6286163453. Planes: [verificación, revisión 1](https://github.com/eLafo/desarrollo-agentico-todo-04/issues/22#issuecomment-6285288643) e [implementación, revisión 1](https://github.com/eLafo/desarrollo-agentico-todo-04/issues/22#issuecomment-6285452480).

| Afirmación | Resultado |
| --- | --- |
| Los tests de 121 al añadir y al editar (y el del modelo) fallan en la base (`expected: 422 actual: 303`) y pasan en el cambio | Cumplida |
| Los tests de 120 pasan; también pasan en la base, como se declaró | Cumplida |
| Los tests de navegador de escribir de más y de la ayuda fallan en la base (130 caracteres en el campo) y pasan en el cambio | Cumplida |
| 121 caracteres (acentuados, emoji, con espacios alrededor) dan 422 y el mensaje exacto al crear y al editar, y no se guarda nada | Cumplida |
| 120 caracteres dan 303 y se guardan enteros (también 240 y 480 bytes) | Cumplida |
| Al teclear 128 caracteres, los dos campos se quedan con los 120 primeros y la ayuda dice 120 | Cumplida |
| Una tarea guardada antes con 200 caracteres se ve entera; guardarla sin acortar da 422 | Cumplida |
| Vacío y solo espacios siguen dando «El título es obligatorio.» | Cumplida |
| Suite de Go, 9 de 9 tests de navegador, `gofmt`, `go vet` y `go build` | En verde |

Para repetir las pruebas:

```bash
GO_ENV=test go test -p 1 -count=1 ./actions ./models
npx playwright test e2e/tareas/anadir-titulo-hasta-120-caracteres.spec.ts e2e/tareas/editar-titulo-hasta-120-caracteres.spec.ts
bash .claude/skills/check/scripts/check.sh
```

## Notas
- **Tareas antiguas.** Las tareas guardadas con 121 a 200 caracteres se conservan y se ven enteras en la lista, pero para guardarlas al editar hay que acortarlas a 120. Ningún test automático lo protege: lo sostiene la columna en 200 y la medida de la verificación. `docs/knowledge/features/tareas.md` no lo dice de forma expresa (se deduce de la regla «al crear y al editar»).
- **Emoji.** El navegador cuenta cada emoji como dos unidades para `maxlength`: en «Nueva tarea» caben 60 emoji aunque el servidor acepta 120. Ya pasaba con 200. Verificación del humano pendiente (A27).
- **Lectura en la lista.** Si un título de 120 caracteres se lee bien es el motivo de la issue; en la verificación cada título de 120 ocupó dos líneas completas, sin cortar. Valoración del humano pendiente (A28).
- **Ayuda en el error.** Con un error en la edición, el campo pasa a `aria-describedby="title-error"` y la ayuda deja de anunciarse como su descripción, aunque sigue visible. Ya era así antes y queda fuera de esta issue.
- **Nombres de tests.** `Test_Task_Title_120Chars_Valid` y `Test_Task_Title_121Chars_Invalid` nombran el caso por la cifra; si el límite cambia, dejan de ser ciertos (`nit` de la verificación).
- Fuera de alcance: recortar títulos en la lista, contador de caracteres, ayuda nueva en «Nueva tarea» y migrar las tareas antiguas.
