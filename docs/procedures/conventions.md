# Convenciones

Cómo se construye esta aplicación y con qué rúbricas se juzga un cambio. Es la referencia de la dimensión estructural de la verificación del workflow `adw` y de los juicios que la acompañan. Las rúbricas están fijadas de antemano: un juez aplica sus criterios y niveles, cita el material y razona cada criterio. No inventa criterios nuevos.

## Capas

La aplicación es Buffalo sobre PostgreSQL, con Pop como ORM y plantillas Plush.

| Capa | Ruta | Responsabilidad |
| --- | --- | --- |
| Rutas y middleware | `actions/app.go` | Declarar rutas y middleware. Nada más |
| Acciones | `actions/` | Coordinar una petición: leer parámetros, normalizar la entrada, llamar al modelo y elegir la respuesta (render, redirección, flash) |
| Modelo | `models/` | Tipos, validaciones (`Validate`), reglas y cálculos del dominio, constantes como `TitleMaxLength` |
| Plantillas | `templates/` | Presentar los valores que pone la acción. Sin cálculos del dominio |
| Esquema y datos | `migrations/` | Fizz para el esquema y SQL para los datos, cada migración con su `down` |
| Recursos | `public/assets/` | CSS y JavaScript sin dependencias, como mejora progresiva sobre formularios que ya funcionan sin él |
| Datos de test | `fixtures/` | Escenarios TOML que cargan los tests |
| Tests de navegador | `e2e/`, `playwright.config.ts` | Recorridos en Chromium contra la aplicación arrancada con `GO_ENV=test`. Prueban lo que depende del navegador |
| Arranque | `cmd/app/` | Migrar y servir |
| Documentación | `docs/` | Lo que hace el sistema y cómo se trabaja en él, y la historia de los casos. Ver `docs/README.md` |
| Trigger | `adws/` | El lanzador del workflow `adw`. No es parte de la aplicación |

`models` no importa `actions` ni `templates`. Las dependencias van de las acciones al modelo.

## Patrones del repositorio

- **Acciones.** Toman la transacción con `tx := c.Value("tx").(*pop.Connection)`. Si no encuentran el registro, responden `c.Error(http.StatusNotFound, err)`. Normalizan la entrada (`strings.TrimSpace`) y validan con `tx.ValidateAndCreate` o `tx.ValidateAndUpdate`. Con errores de validación, vuelven a mostrar el formulario con `422` y `errors`. Si todo va bien, añaden un flash y redirigen con `303` a `/`. Ver `actions/tasks.go`.
- **Modelo.** Las reglas de validación viven en `Validate` y sus mensajes son frases completas en español: «El título es obligatorio.». Ver `models/task.go`.
- **Plantillas.** La salida se escapa siempre con `<%= %>`. Los formularios que cambian estado son `POST` con `authenticity_token` y, para `PUT` o `DELETE`, el campo `_method`. Los controles tienen texto accesible (`aria-label`, `sr-only`) y los errores, `role="alert"` y `aria-invalid`.
- **Textos.** La interfaz está en español. El vocabulario es el de la lista: tarea, completar, reabrir, editar, eliminar.
- **Comentarios.** Los identificadores exportados llevan un comentario en español que empieza por su nombre: `// TasksToggle alterna una tarea entre completada y pendiente.`.
- **Tests.** `ActionSuite` (peticiones y render) y `ModelSuite` (reglas del modelo) sobre `github.com/gobuffalo/suite/v4`. Se nombran `Test_<Recurso><Acción>_<Caso>`. Preparan datos con `as.LoadFixture` o `ValidateAndCreate` y comprueban el código HTTP y el efecto en la base de datos. Ver `actions/tasks_test.go`.
- **Tests de navegador.** Lo que depende del navegador, como el JavaScript de `public/assets/` o un recorrido que solo se ve en la página, se prueba con Playwright en `e2e/<grupo>/<caso>.spec.ts`: un test por fichero, que importa `test` y `expect` de `e2e/fixtures.ts`, prepara las tareas con la fixture `tasks`, usa localizadores por rol y nombre accesible (`getByRole`) y comprueba el efecto visible, y el persistido recargando la página. Ver `e2e/tareas/`.
- **Formato y comprobaciones.** `gofmt`, `go vet ./...`, `go test -p 1 ./...` con la base de datos de test migrada, `go build ./...`, `npx playwright test` y `docker build`, como en `.github/workflows/ci.yaml`.
- **Documentación.** Lo que describe el sistema vive en `docs/knowledge/` (qué hace) y `docs/procedures/` (cómo se trabaja), con los índices que parten de `docs/README.md`. Un cambio que modifica un comportamiento, una regla, un texto o un comando que describen esos documentos los actualiza en el mismo cambio. Se escribe lo vigente, como si siempre hubiera sido así: sin «antes» ni «ahora» y sin el número del caso. Una funcionalidad nueva que no encaja en un documento existente tiene el suyo en `docs/knowledge/features/`, con su entrada en el índice.

## Zonas fuera del alcance de un caso

Un caso no toca estas rutas salvo que su issue lo pida expresamente: `.github/`, `.claude/`, `.devcontainer/`, `adws/`, `config/`, `go.mod` y `go.sum`, `package.json`, `package-lock.json` y `playwright.config.ts`, las migraciones ya existentes, estas convenciones y los documentos de otros casos en `docs/history/`. La fidelidad lo comprueba sobre `git diff --name-only` respecto a la base.

Los documentos de `docs/history/` del propio caso (su documento de cierre, sus decisiones y sus entradas en los índices) no forman parte del cambio que se juzga: los escribe `document` al cerrar el caso, y la contención no los evalúa. Los de `docs/knowledge/` y `docs/procedures/` sí: los actualiza el cambio, y los juzgan D1 y D3.

## Severidad

| Severidad | Significado |
| --- | --- |
| `blocker` | El cambio no debe integrarse así: rompe una capa o un patrón que el siguiente cambio tendría que deshacer, abre un riesgo (seguridad, pérdida de datos) o deja un comportamiento sin probar |
| `should-fix` | Deuda que conviene pagar pronto, pero que no impide integrar |
| `nit` | Preferencia menor |

## Rúbricas

Cada rúbrica dice qué material juzga y qué distingue cada nivel. Lo que no aparece en un nivel `blocker` no puede ser `blocker`.

### E1 · Cada cosa en su capa

Material: el diff de `actions/`, `models/`, `templates/` y `public/assets/`.

- **Pasa:** las reglas y los cálculos del dominio están en `models/`; la acción solo coordina; la plantilla presenta valores que recibe.
- **`should-fix`:** lógica de presentación enrevesada en la plantilla (condiciones anidadas que podrían ser un valor preparado por la acción); la misma consulta repetida en dos acciones.
- **`blocker`:** un cálculo del dominio en la plantilla o en JavaScript (por ejemplo, contar las tareas pendientes recorriendo `tasks` en Plush); una validación del dominio solo en el cliente; SQL a mano donde Pop lo cubre; `models` importando `actions` o `templates`.

### E2 · Diseño simple

Material: el diff de código Go.

- **Pasa:** cada función tiene una responsabilidad; no hay abstracciones de un solo uso; se reutiliza lo que ya existe (`TitleMaxLength`, `Task.Validate`).
- **`should-fix`:** una función que mezcla dos responsabilidades; un parámetro o una rama que nadie usa todavía.
- **`blocker`:** una interfaz, un paquete, una capa o una dependencia nueva sin un segundo uso ni un criterio que la exija; una regla duplicada que puede divergir de la existente (un segundo límite de longitud del título, por ejemplo).

### E3 · Nombres y legibilidad

Material: los identificadores y comentarios nuevos o cambiados.

- **Pasa:** nombres idiomáticos de Go (MixedCaps, receptores cortos, sin repetir el paquete); acciones `<Recurso><Acción>`; tests `Test_<Recurso><Acción>_<Caso>`; comentarios de documentación en español que empiezan por el nombre.
- **`nit`:** un nombre mejorable que no confunde.
- **`should-fix`:** un identificador exportado sin comentario; un nombre que no dice lo que hace.

### E4 · Errores

Material: el diff de código Go.

- **Pasa:** todo error se devuelve o se trata; un registro que no existe da `404`; una entrada inválida da `422` con el formulario y sus errores.
- **`should-fix`:** un error que pierde un contexto útil al devolverse.
- **`blocker`:** un error ignorado en una operación de base de datos; un `500` donde el patrón del repositorio da `404` o `422`; un `panic` alcanzable desde la entrada del usuario.

### E5 · Seguridad web

Material: el diff de `actions/`, `templates/` y `public/assets/`.

- **Pasa:** la salida se escapa; los formularios que cambian estado llevan `authenticity_token`; la validación del servidor no depende del cliente.
- **`blocker`:** datos del usuario en la salida sin escapar; un formulario o una petición que cambia estado sin token CSRF o por `GET`; un límite que solo impone el navegador (`maxlength`) sin validación en el modelo.

### E6 · Tests

Material: el diff de los ficheros `_test.go`, `fixtures/` y `e2e/`.

- **Pasa:** cada comportamiento nuevo tiene test en la suite que le toca, con su caso negativo cuando lo hay, y comprueba el efecto, no solo el código HTTP. Lo que depende del navegador tiene test en `e2e/`.
- **`should-fix`:** un test que solo comprueba el código HTTP de algo que cambia datos.
- **`blocker`:** un comportamiento nuevo sin test; un cambio en el JavaScript de `public/assets/` probado solo con tests de Go, que no lo ejecutan; un test existente desactivado, saltado (`t.Skip`, `test.skip`, `test.fixme`) o debilitado.

### F1 · Mensajes y textos

Material: las cadenas nuevas o cambiadas que ve el usuario (plantillas, flashes, mensajes de validación, `aria-label`), extraídas del diff.

- **Pasa:** español correcto (ortografía, tildes, signos de apertura); frases completas con punto en los mensajes; el mismo vocabulario y registro que la interfaz; plurales correctos («1 tarea», «2 tareas»); el texto accesible nombra la acción y el objeto.
- **`nit`:** una preferencia de estilo.
- **`should-fix`:** vocabulario o registro distinto del resto de la interfaz.
- **`blocker`:** una falta de ortografía o de gramática visible, un texto en otro idioma, un plural incorrecto o un mensaje que no dice qué ha pasado.

### D1 · Contención

Material: cada fragmento del diff, salvo los documentos de `docs/history/` del caso, el plan de implementación vigente y la especificación (criterios, aclaraciones y peticiones de cambio).

- **Pasa:** el fragmento responde a un criterio o una petición a través de un paso del plan, o es un cambio mecánico que alguno exige: el formato del código tocado, el test de un criterio, la migración o el fixture que necesita, la actualización de un documento de `docs/knowledge/` o `docs/procedures/` que el cambio deja desfasado.
- **`nit`:** un ajuste de formato en una línea contigua al cambio.
- **`blocker`:** una funcionalidad, una opción o un texto que nadie ha pedido; un refactor, un renombre o una reordenación que ningún criterio necesita; un cambio en una zona fuera del alcance del caso; un comportamiento existente cambiado o eliminado sin que se pida; documentación reescrita sobre algo que el cambio no toca. Una mejora razonable sigue siendo un cambio no pedido.

El plan de implementación sirve para entender por qué existe un fragmento. No justifica por sí solo un fragmento que ningún criterio pide.

### D2 · Test asociado a cada criterio

Material: cada criterio, la afirmación que lo cubre y el test que se le asocia.

- **Pasa:** el test, de Go o de navegador, ejercita el caso del criterio y comprueba su resultado esperado.
- **`blocker`:** el test asociado prueba otra cosa o un caso más débil que el criterio; ningún test cubre el criterio.

Solo se aplica a criterios con comportamiento observable. Uno sin él (documentación, una tarea sin efecto visible) no necesita test: basta con la declaración de su motivo.

### D3 · Documentación al día

Material: el diff, y los documentos de `docs/knowledge/` y `docs/procedures/` que describen lo que el diff toca, localizados recorriendo los índices desde `docs/README.md`.

- **Pasa:** cada documento que describía un comportamiento, una regla, un texto o un comando que el cambio modifica lo describe ya como queda; lo nuevo que ve el usuario está descrito en `docs/knowledge/`; los documentos cuentan lo vigente, sin la historia del caso. Si el cambio no toca nada descrito, pasa con la lista de los documentos revisados.
- **`nit`:** redacción mejorable; un detalle secundario que se podría añadir.
- **`should-fix`:** un documento que cuenta la historia del cambio («antes…, ahora…», el número del caso) en vez de lo vigente; lo nuevo descrito en un documento que no le toca, o un documento nuevo sin entrada en el índice de su carpeta; un detalle para desarrolladores que falta en `docs/procedures/`.
- **`blocker`:** un documento que sigue describiendo el comportamiento, la regla, el texto o el comando de antes del cambio; un documento que describe algo que el diff no hace; una funcionalidad o una regla nueva que ve el usuario sin describir en ningún documento de `docs/knowledge/`.

Un documento está desfasado por lo que afirma: lo que sigue siendo cierto después del cambio no lo está, aunque mencione lo mismo (el tamaño de una columna que el cambio no toca, por ejemplo). Varios fragmentos desfasados por la misma omisión son un solo hallazgo, con todas sus citas.
