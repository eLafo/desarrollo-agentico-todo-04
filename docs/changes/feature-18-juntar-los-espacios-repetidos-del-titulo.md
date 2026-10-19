# Juntar los espacios repetidos del título

**Issue:** #18 · **Tipo:** feature

## Resumen
Al añadir o editar una tarea, cada grupo de espacios seguidos del título (tabuladores incluidos) se guarda como un solo espacio, y los de los extremos se siguen quitando. El límite de 200 caracteres se mide con los espacios ya juntos, y para eso los formularios dejan de cortar el título al escribir.

## Qué se construyó
- Añadir: «Comprar    pan» se guarda como «Comprar pan».
- Editar: el título editado pasa por la misma normalización.
- Tabuladores: «Comprar⇥pan» se guarda como «Comprar pan».
- Límite: un título que solo supera 200 caracteres por los espacios repetidos se acepta; si con los espacios juntos sigue pasando de 200, se rechaza con «El título no puede superar los 200 caracteres.».
- Las tareas ya guardadas no cambian hasta que alguien las edita.

## Implementación técnica

| Fichero | Papel |
| --- | --- |
| `actions/tasks.go` | Nueva `normalizeTitle`, usada en `TasksCreate` y `TasksUpdate` |
| `templates/tasks/index.plush.html` | Campo «Nueva tarea» sin `maxlength` |
| `templates/tasks/edit.plush.html` | Campo de edición sin `maxlength` |
| `actions/tasks_test.go` | Tests de Go de juntar, tabuladores, límite tras juntar y tareas antiguas |
| `e2e/tareas/anadir-titulo-con-espacios-repetidos.spec.ts`, `editar-titulo-con-espacios-repetidos.spec.ts`, `anadir-titulo-que-sigue-largo-tras-juntar-espacios.spec.ts` | Tests de navegador nuevos |

Cambios que importan:

- **`normalizeTitle`:** `strings.Join(strings.Fields(title), " ")`. Sustituye a `strings.TrimSpace` y quita también los extremos.
- **Límite:** sigue solo en el modelo (`models/task.go`); la acción normaliza antes de validar, así que el modelo mide el título ya junto.
- **Sin `maxlength`:** con `maxlength="200"` el navegador cortaba lo escrito antes de que el servidor pudiera juntar los espacios, y un título válido tras juntar se perdía.
- **Sin migración:** no se reescribe ningún título guardado.

## Cómo se usa
Sin cambios visibles: se escribe el título como siempre. Al guardar, los espacios repetidos desaparecen. Si el título sigue pasando de 200 tras juntar, aparece el error de siempre y el campo muestra el título ya junto.

## Configuración
No requiere configuración.

## Cómo se probó
Verificación validada (intento 2 de 3, árbol `f102e78`). Acta vigente: https://github.com/eLafo/desarrollo-agentico-todo-04/issues/18#issuecomment-6274563546 · Plan de verificación: https://github.com/eLafo/desarrollo-agentico-todo-04/issues/18#issuecomment-6273992621

| Afirmación | Resultado |
| --- | --- |
| 33 afirmaciones del plan de verificación, todas las bloqueantes | Cumplidas |
| Tests nuevos de Go y de navegador: en rojo sobre la base y en verde con el cambio | Cumplida |
| Añadir, editar, tabuladores y límite tras juntar, en el navegador y en la base | Cumplida |
| Tareas guardadas intactas, sin migraciones, extremos quitados | Cumplida |
| Suite completa de Go y 10 tests de navegador | En verde |
| `gofmt`, `go vet`, `go build` y Docker | En verde |

Para repetir las pruebas:

```bash
npx playwright test e2e/tareas/anadir-titulo-con-espacios-repetidos.spec.ts e2e/tareas/editar-titulo-con-espacios-repetidos.spec.ts e2e/tareas/anadir-titulo-que-sigue-largo-tras-juntar-espacios.spec.ts
bash .claude/skills/check/scripts/check.sh
```

## Notas
- **Capa de acciones:** la normalización vive en `normalizeTitle`, no en el modelo. Cualquier otra vía que escriba el título (otra acción, importación, script) debe pasar por ella, o el límite se medirá sin juntar.
- **Qué cuenta como espacio:** `strings.Fields` usa `unicode.IsSpace`: también el espacio duro (U+00A0), otros espacios Unicode y los saltos de línea. Es el mismo conjunto que usaba `strings.TrimSpace` en los extremos. Pendiente de confirmar por el humano.
- **Tras un rechazo** el campo muestra el título ya junto, no lo escrito. Pendiente de confirmar por el humano.
- **Sin `maxlength`:** un título demasiado largo ya no se corta al escribir; se rechaza al pulsar «Añadir» o «Guardar cambios». Pendiente de confirmar por el humano.
- **Tareas antiguas:** no se migran; conservan sus espacios repetidos y los pierden al editarlas y guardar.
- **Ayuda del campo:** «Hasta 200 caracteres.» no dice que se cuenta con los espacios ya juntos.
- **Nombres de tests:** `Test_Tasks(Create|Update)_TitleOverLimitOnlyBySpaces` suenan a rechazo, pero cubren el caso que se acepta. No hay test de Go que edite con 201 caracteres tras juntar; lo cubren el test previo y la verificación.
