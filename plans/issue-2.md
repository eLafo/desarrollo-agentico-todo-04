# Feature: Ordenar las tareas manualmente (arrastrar y soltar con ratón o táctil, y alternativa con teclado)

## Encargo y criterios de aceptación

**Encargo:** issue #2 «Ordenar las tareas», https://github.com/eLafo/desarrollo-agentico-todo-04/issues/2 — «Me gustaría ordenar las tareas utilizando drag and drop».

**Decisiones funcionales ya aclaradas y aceptadas por el humano en la issue** (forman parte del encargo y no se reabren):

1. **Persistencia:** el orden manual se guarda y se mantiene al recargar la página. Hay un único orden global compartido, porque la app no tiene cuentas.
2. **Tareas nuevas:** aparecen al final de la lista.
3. **Completadas:** hay una única lista ordenable que mezcla pendientes y completadas. Completar o reabrir una tarea no cambia su posición.
4. **Interacción:** se reordena arrastrando con ratón **y** en pantallas táctiles, y además hay una **alternativa con teclado** accesible.

### Criterios de aceptación explícitos (salen del encargo y de las decisiones)

- **CA1 — Ratón:** al arrastrar una tarea desde su asa con el ratón y soltarla en otro punto de la lista, la tarea queda en esa posición.
- **CA2 — Táctil:** el mismo arrastre funciona con el dedo en una pantalla táctil. Iniciar el arrastre desde el asa no desplaza la página; fuera del asa la página se desplaza con normalidad.
- **CA3 — Teclado:** sin usar el puntero, una persona puede llegar con `Tab` a los botones «Subir tarea: …» y «Bajar tarea: …» de cada tarea y moverla una posición con `Enter` o `Espacio`. El foco se queda en el botón pulsado, así que puede repetir el movimiento. La nueva posición se anuncia a los lectores de pantalla («Tarea «X» movida a la posición N de M.»).
- **CA4 — Persistencia global:** después de reordenar (de cualquiera de las tres formas), al recargar `/` o abrirlo desde otro navegador se ve el mismo orden.
- **CA5 — Nuevas al final:** una tarea creada desde el formulario aparece la última, sea cual sea el orden manual que haya.
- **CA6 — Lista única:** pendientes y completadas se muestran mezcladas según el orden manual. Completar o reabrir una tarea, en la interfaz o con `PUT /tasks/{id}/toggle`, no cambia su posición ni al momento ni después de recargar.

### Inferencias (no se pidieron literalmente; se adoptan para que lo anterior sea coherente y robusto)

- **IN1:** las tareas que ya existen conservan, al desplegar el cambio, el orden que se ve hoy (`created_at asc, id asc`).
- **IN2:** editar el título o eliminar una tarea no altera el orden relativo de las demás.
- **IN3:** si el servidor rechaza un orden o falla la red, la lista vuelve al último orden guardado y aparece un mensaje visible («No se pudo guardar el nuevo orden. Recarga la página e inténtalo de nuevo.»). Nunca se guarda un orden parcial.
- **IN4:** reordenar solo cambia la posición. No modifica el título, el estado ni `updated_at`.
- **IN5:** sin JavaScript la lista se ve y funciona como hoy. Los controles de reordenación no se muestran para que no haya botones inertes.
- **IN6:** los títulos se siguen escapando en las nuevas etiquetas (`aria-label`), igual que ya ocurre con «Editar tarea: …».

## Comportamiento actual y propuesta

### Actual (observado en el repositorio)

- **Rutas** (`actions/app.go`): `GET /`, `POST /tasks`, `GET /tasks/{task_id}/edit`, `PUT /tasks/{task_id}`, `PUT /tasks/{task_id}/toggle`, `DELETE /tasks/{task_id}` y `GET /healthz`. Las rutas se evalúan **en el orden en que se registran**: `buffalo@v1.1.4/route_mappings.go` no las ordena y gorilla/mux usa la primera que coincide. Middlewares: `csrf.New`, que acepta el token en la cabecera `X-CSRF-Token` o en el campo `authenticity_token` y se desactiva cuando `GO_ENV=test`, y `popmw.Transaction`, que abre una transacción por petición y hace rollback si el estado es ≥ 400 o hay error.
- **Acciones** (`actions/tasks.go`): `TasksIndex` lista con `tx.Order("created_at asc, id asc")` (línea 19). La rama de error de validación de `TasksCreate` repite la misma consulta (línea 46). `TasksToggle` y `TasksUpdate` cargan la tarea y la guardan entera con `tx.Update` / `tx.ValidateAndUpdate`.
- **Modelo** (`models/task.go`): `Task{ID, Title, Completed, CreatedAt, UpdatedAt}`. No tiene campo de orden ni callbacks.
- **Migraciones** (`migrations/`): `20260901000000_create_tasks.up.fizz` (esquema) y `20260901000001_seed_tasks.up.sql` (tres tareas de ejemplo). `migrations/embed.go` embebe `*.fizz *.sql`. `cmd/app/main.go` ejecuta `models.Migrate()` al arrancar, y la CI usa `soda migrate -e test`.
- **Template** (`templates/tasks/index.plush.html`): cada tarea es un `<li class="task">` dentro de `<ul class="task-list">` con el formulario de completar, el título, el estado oculto («Completada») y `.task-actions` (editar y eliminar). En dispositivos con puntero fino, las acciones solo se ven al pasar el ratón por encima o al recibir foco (`public/assets/app.css`).
- **JS** (`public/assets/app.js`): confirma antes de eliminar e intercepta el formulario de completar para actualizar la vista al momento. **Defecto previo, fuera de alcance:** ese código termina con `fetch(form.action)`, que hace un `GET` a `/tasks/{id}/toggle`, una ruta que no existe. Por eso completar desde la interfaz no persiste al recargar. No influye en el orden.
- **Tests:** `actions/tasks_test.go` (`ActionSuite`) y `models/task_test.go` (`ModelSuite`), ambos con PostgreSQL, y la fixture `fixtures/tasks.toml` («tres tareas»).
- **Resultado:** la lista se ordena por fecha de creación y no guarda ningún orden propio.

### Propuesta (cambio vertical mínimo)

- Columna `position integer NOT NULL DEFAULT 0` en `tasks`. Una migración la rellena (`1..N`) con el orden actual.
- La lista se ordena por `position asc, created_at asc, id asc`. Los dos últimos criterios desempatan de forma determinista.
- Cada tarea nueva recibe `MAX(position) + 1` mediante un callback `BeforeCreate` del modelo (CA5).
- Un único endpoint `PUT /tasks/order` recibe en JSON el **orden completo** (`{"task_ids": [...]}`), lo valida (tienen que aparecer exactamente las tareas existentes, cada una una vez) y asigna las posiciones `1..N`. Responde `204` si va bien, `422` con JSON `{"error": …}` si la lista no es válida y `400` si el cuerpo está mal formado.
- En el template, cada tarea lleva su `data-task-id`, un asa de arrastre (solo para puntero) y dos botones «Subir»/«Bajar» (teclado y lectores de pantalla). La lista expone la URL del endpoint y el token CSRF, y hay una región `role="status"` para los anuncios.
- En `app.js`, sin dependencias: arrastre con **Pointer Events** (ratón, lápiz y táctil con el mismo código) desde el asa, botones Subir/Bajar y guardado con `fetch` (`PUT` con la cabecera `X-CSRF-Token`). Las peticiones se envían de una en una y, si una falla, se vuelve al último orden guardado.
- Completar o reabrir no toca `position`: `TasksToggle` solo invierte `Completed` y conserva la posición cargada, y el JS no mueve el `<li>` (CA6).

## Alcance

### Incluido

- Migración fizz (up/down) que añade `position` y rellena las tareas existentes.
- Modelo: campo `Position`, constante de orden de la lista, error `ErrInvalidTaskOrder`, callback `BeforeCreate` y función `ReorderTasks`.
- Acción `TasksReorder` y ruta `PUT /tasks/order`, registrada **antes** de `PUT /tasks/{task_id}`.
- Cambio de orden en `TasksIndex` y en la rama de error de `TasksCreate`.
- Template de la lista (atributos de datos, asa, botones Subir/Bajar y región de anuncios).
- JavaScript propio (arrastre con Pointer Events, botones y guardado) y CSS (columnas, asa, estado de arrastre, ocultación sin JS, móvil).
- Nuevas versiones de `app.css` y `app.js` en el layout para invalidar la caché del navegador.
- Tests de modelo y de acción descritos en «Tests y validación».

### Excluido

- Corregir el defecto previo del JS de completar (`fetch(form.action)` hace un `GET` y no persiste). Es un problema independiente del orden: merece su propia issue o bug.
- Orden por persona o por sesión (la decisión 1 fija un orden global), varias listas, secciones o separar las completadas.
- Ordenación automática (por fecha, alfabética…) y botones «mover al principio/al final».
- Reordenar sin JavaScript. Los controles se ocultan en ese caso (IN5).
- Librerías de terceros (JS o Go) y cambios en `go.mod`/`go.sum`. Tampoco se añaden índices de base de datos ni restricciones de unicidad sobre `position`.
- Tests automáticos de JavaScript o de navegador. El repositorio no tiene esa infraestructura y se cubre con la verificación manual descrita más abajo.
- Cambios en `fixtures/tasks.toml`: las filas toman `position = 0` por defecto, que es válido.
- Arrastre con desplazamiento automático avanzado. Solo se incluye un desplazamiento básico cerca de los bordes de la ventana.

## Diseño por capas

### Persistencia y migración

- Nueva migración `20261013000000_add_position_to_tasks` en **fizz**, el patrón que ya usa el esquema del repositorio. `add_column` genera `ALTER TABLE "tasks" ADD COLUMN "position" integer NOT NULL DEFAULT '0';` (fizz@v1.14.4 crea las columnas `NOT NULL` salvo que se indique `null: true`). Después, un `sql(...)` rellena las posiciones `1..N` con `ROW_NUMBER() OVER (ORDER BY created_at ASC, id ASC)`, de modo que se conserva el orden visible actual (IN1).
- Pop ejecuta cada migración dentro de una transacción (`pop/v6@v6.4.1/migrator.go`). Las varias sentencias del fizz se lanzan con un único `Exec` sin argumentos, y pgx v5 usa entonces el protocolo simple, que admite varias sentencias. El `down` elimina la columna.
- `position` es una palabra clave no reservada de PostgreSQL y vale como nombre de columna en `SELECT`, `UPDATE … SET` y `ORDER BY`.
- La migración de semillas (`…000001_seed_tasks`) es anterior y queda cubierta por el relleno.

### Modelo (`models/task.go`)

- `Position int` con las etiquetas `json:"position" db:"position" form:"-"`. Con `form:"-"` no se puede asignar desde formularios.
- `TaskListOrder = "position asc, created_at asc, id asc"`: un único punto de verdad para las dos consultas de la lista.
- `BeforeCreate(tx)`: pop lo invoca dentro de `Create`/`ValidateAndCreate`, después de validar y antes del `INSERT` (`pop/v6@v6.4.1/executors.go`). Consulta la tarea con mayor `position` (`tx.Order("position desc").First(last)`, que devuelve `sql.ErrNoRows` si no hay ninguna) y asigna `last.Position + 1`. Siempre sobrescribe el valor, así que toda creación va al final.
- `ReorderTasks(tx, ids []uuid.UUID) error`: carga todas las tareas y exige que `ids` contenga **exactamente** las tareas existentes, cada una una vez. Si no, devuelve `ErrInvalidTaskOrder` **antes de escribir nada**. Después actualiza con `UPDATE tasks SET position = ? WHERE id = ?` solo las filas cuya posición cambia. No toca `updated_at` (IN4). Toda la operación va dentro de la transacción de la petición.
- Las posiciones pueden tener huecos tras eliminar tareas. No es un problema: el orden relativo se mantiene, la siguiente tarea nueva sigue siendo `MAX + 1` y la próxima reordenación vuelve a numerar `1..N`.

### Acción y ruta (`actions/tasks.go`, `actions/app.go`)

- `TasksReorder`: `c.Bind` a `struct{ TaskIDs []string \`json:"task_ids"\` }` (el binder JSON de Buffalo se elige por `Content-Type`/`Accept`). Si el binding falla, devuelve `c.Error(400, err)`. Convierte cada id con `uuid.FromString`; si alguno es inválido, responde `422`. Llama a `models.ReorderTasks` y, si devuelve `ErrInvalidTaskOrder`, responde `422` con `r.JSON(map[string]string{"error": "…"})`. Si todo va bien, responde `c.Render(http.StatusNoContent, nil)`: `DefaultContext.Render` con renderer `nil` solo escribe el estado, y `popmw` confirma la transacción porque el estado es < 400.
- **Ruta:** `app.PUT("/tasks/order", TasksReorder)` se registra **antes** de `app.PUT("/tasks/{task_id}", TasksUpdate)`. Si no, `{task_id}` capturaría `"order"` y la petición acabaría en un `404` de `TasksUpdate`. Un test cubre este punto.
- **CSRF:** en desarrollo y producción el JS envía el token enmascarado de `authenticity_token` en la cabecera `X-CSRF-Token`. El middleware la lee antes que el campo del formulario. El token enmascarado se puede reutilizar mientras dure la sesión.
- `TasksIndex` y la rama de error de `TasksCreate` pasan a `tx.Order(models.TaskListOrder)`.
- **Por qué el orden completo y no «mover la tarea X a la posición N»:** es idempotente, sirve igual para arrastre y teclado, y detecta listas desactualizadas (otra pestaña creó o eliminó tareas). En ese caso responde `422`, el cliente restaura el orden y pide recargar. El coste de varios `UPDATE` es despreciable para una lista de tareas.

### Renderizado (`templates/tasks/index.plush.html`, `templates/application.plush.html`)

- `<ul class="task-list" data-task-list data-reorder-url="/tasks/order" data-csrf-token="<%= authenticity_token %>">`. Las rutas se escriben a mano, como en el resto del template.
- `<li … data-task-id="<%= t.ID %>">`.
- Dentro de `.task-actions`, **antes** de «Editar», un grupo `<span class="task-move-group">` con dos `<button type="button" class="icon-action task-move" data-move="up|down">`. Sus `aria-label` son «Subir tarea: <título>» y «Bajar tarea: <título>», con `title` e iconos de flecha `aria-hidden`. Son botones nativos: se activan con `Enter`/`Espacio` y cualquier tecnología de apoyo los anuncia.
- Al final del `<li>`, el asa `<span class="task-handle" data-drag-handle aria-hidden="true" title="Arrastrar para reordenar">` con un icono de seis puntos. Es solo para puntero: no es enfocable y queda oculta a los lectores de pantalla, que usan Subir/Bajar. Va en una cuarta columna a la derecha para no romper la alineación entre los botones de completar y el «+» del formulario.
- Tras `</ul>`, dentro de la rama con tareas: `<p class="sr-only" role="status" data-reorder-feedback></p>` (región de anuncios; se hace visible si hay errores).
- Layout: `app.css?v=3` pasa a `?v=4` y `app.js?v=2` pasa a `?v=3`, como ya se hizo en d8dc18d.

### JavaScript (`public/assets/app.js`)

- Se añade un bloque IIFE en el mismo estilo ES5 del archivo (`var`, `function`), sin dependencias, que solo actúa si existe `[data-task-list]`.
- Al iniciarse, añade a la lista la clase `task-list-reorderable` (el CSS solo muestra los controles con ella, IN5) y guarda el orden inicial como último orden confirmado.
- **Pointer Events** en lugar de HTML5 Drag and Drop: unifican ratón, lápiz y táctil, mientras que el DnD nativo no funciona de forma fiable en móviles y obligaría a escribir otra implementación táctil. El arrastre empieza con `pointerdown` sobre el asa (botón principal, puntero primario) y usa `setPointerCapture`. En cada `pointermove` coloca la tarea antes de la primera tarea cuyo punto medio vertical quede por debajo del puntero y desplaza la ventana si el puntero se acerca a los bordes. Con `pointerup` guarda si el orden cambió, y con `pointercancel` restaura el orden.
- **Clave de robustez:** para recolocar una tarea se mueven sus **vecinas** (`placeBefore`) y nunca el propio `<li>` movido. Así no sale del DOM y no pierde ni la captura del puntero ni el foco del teclado (al quitar un nodo del DOM, el navegador suelta la captura y el foco).
- Para el teclado, un `click` delegado en `[data-move]` intercambia la tarea con su vecina mediante `placeBefore`. En los extremos no mueve nada y anuncia «La tarea ya es la primera/última de la lista.».
- **Guardado:** `fetch(PUT, JSON, X-CSRF-Token)`. Solo hay una petición en curso; si llegan más cambios mientras tanto, al terminar se envía el último orden. Si falla, se restaura el último orden confirmado, se conserva el foco y se muestra un error visible (la región de anuncios pierde `sr-only` y toma `field-error`).

### CSS (`public/assets/app.css`)

- Sin JS, `.task-list:not(.task-list-reorderable)` oculta `.task-handle` y `.task-move-group`.
- Con JS, la rejilla pasa a `44px minmax(0, 1fr) auto 32px` (en móvil, `36px minmax(0, 1fr) auto 32px`).
- El asa tiene `touch-action: none` (el arrastre táctil no desplaza la página), `user-select: none`, `-webkit-touch-callout: none` y `cursor: grab`. Durante el arrastre se aplican `.task-dragging` (fondo y sombra) y `.is-reordering` en `body` (`cursor: grabbing` y sin selección de texto).
- Con puntero fino, el asa aparece al pasar por encima o con el foco dentro, igual que `.task-actions`.
- En `max-width: 760px`, `.task-move-group:not(:focus-within)` queda visualmente oculto con la técnica `sr-only`: no llena las filas estrechas, los lectores de pantalla lo siguen alcanzando y vuelve a verse en cuanto recibe el foco del teclado. Las personas que usan pantallas táctiles arrastran desde el asa.
- Se reutiliza `.field-error` para el mensaje de error. `prefers-reduced-motion` ya anula las transiciones en todo el documento.

## Archivos relevantes

| Ruta | Responsabilidad | Cambio |
| --- | --- | --- |
| `migrations/20261013000000_add_position_to_tasks.up.fizz` | Añadir `position` y rellenar `1..N` con el orden actual | Nuevo |
| `migrations/20261013000000_add_position_to_tasks.down.fizz` | Eliminar `position` | Nuevo |
| `migrations/embed.go` | Embebe `*.fizz *.sql` (recoge la nueva migración sin cambios) | Sin cambios |
| `migrations/20260901000001_seed_tasks.up.sql` | Semillas anteriores (las cubre el relleno) | Sin cambios |
| `models/task.go` | Campo `Position`, `TaskListOrder`, `ErrInvalidTaskOrder`, `BeforeCreate`, `ReorderTasks` | Modificar |
| `models/task_test.go` | Tests de posición al crear y de reordenación | Modificar |
| `actions/tasks.go` | Orden de la lista en `TasksIndex`/`TasksCreate`; nueva acción `TasksReorder` | Modificar |
| `actions/app.go` | Ruta `PUT /tasks/order` antes de `PUT /tasks/{task_id}` | Modificar |
| `actions/tasks_test.go` | Tests del endpoint, del orden en la lista, de creación al final y de completar sin mover | Modificar |
| `templates/tasks/index.plush.html` | `data-*`, asa, botones Subir/Bajar y región de anuncios | Modificar |
| `templates/application.plush.html` | Nuevas versiones de `app.css`/`app.js` | Modificar |
| `public/assets/app.js` | Arrastre con Pointer Events, botones, guardado y restauración | Modificar (añadir bloque) |
| `public/assets/app.css` | Columnas, asa, estado de arrastre, ocultación sin JS y adaptación a móvil | Modificar |
| `fixtures/tasks.toml`, `go.mod`, `go.sum`, `.github/workflows/ci.yaml` | Referencia; no se tocan | Sin cambios |

## Implementación paso a paso

> El código Go usa **tabuladores** (gofmt). Los imports van en los mismos grupos que ya existen (biblioteca estándar · `todo/...` · terceros) y en orden alfabético dentro de cada grupo. No hay Go en este checkout, así que hay que escribirlo con cuidado para que compile y salga ya formateado.

### 0. Comprobación inicial

- `git status --short --branch` debe mostrar la rama `adw/issue-2` sin más cambios que este plan.
- Antes de editar, confirma que las líneas citadas siguen coincidiendo: `actions/tasks.go:19` y `:46` con `"created_at asc, id asc"`, y el orden de las rutas en `actions/app.go:51-57`.

### 1. Migración

`migrations/20261013000000_add_position_to_tasks.up.fizz`:

```
add_column("tasks", "position", "integer", {"default": 0})
sql("UPDATE tasks SET position = ordered.new_position FROM (SELECT id, ROW_NUMBER() OVER (ORDER BY created_at ASC, id ASC) AS new_position FROM tasks) AS ordered WHERE tasks.id = ordered.id")
```

`migrations/20261013000000_add_position_to_tasks.down.fizz`:

```
drop_column("tasks", "position")
```

(El `sql(...)` va en una sola línea y sin comillas dobles dentro.)

### 2. Modelo (`models/task.go`)

Imports y declaraciones nuevas (el resto del archivo no cambia):

```go
import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// TitleMaxLength es el número máximo de caracteres permitido en el título.
const TitleMaxLength = 200

// TaskListOrder es el orden de la lista: posición manual y, en caso de empate,
// fecha de creación e identificador.
const TaskListOrder = "position asc, created_at asc, id asc"

// ErrInvalidTaskOrder indica que el orden recibido no contiene exactamente
// una vez cada tarea existente.
var ErrInvalidTaskOrder = errors.New("el orden no coincide con las tareas existentes")

// Task representa una tarea de la lista.
type Task struct {
	ID        uuid.UUID `json:"id" db:"id" form:"-"`
	Title     string    `json:"title" db:"title" form:"title"`
	Completed bool      `json:"completed" db:"completed" form:"-"`
	Position  int       `json:"position" db:"position" form:"-"`
	CreatedAt time.Time `json:"created_at" db:"created_at" form:"-"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" form:"-"`
}
```

Después de `Validate`, añadir:

```go
// BeforeCreate coloca cada tarea nueva al final del orden manual.
func (t *Task) BeforeCreate(tx *pop.Connection) error {
	last := &Task{}
	err := tx.Order("position desc").First(last)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	t.Position = last.Position + 1
	return nil
}

// ReorderTasks guarda el orden manual de la lista. ids debe contener exactamente
// una vez cada tarea existente; cada una recibe su posición empezando en 1.
func ReorderTasks(tx *pop.Connection, ids []uuid.UUID) error {
	tasks := Tasks{}
	if err := tx.All(&tasks); err != nil {
		return err
	}
	if len(ids) != len(tasks) {
		return ErrInvalidTaskOrder
	}

	current := make(map[uuid.UUID]int, len(tasks))
	for _, task := range tasks {
		current[task.ID] = task.Position
	}

	seen := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		if _, ok := current[id]; !ok || seen[id] {
			return ErrInvalidTaskOrder
		}
		seen[id] = true
	}

	for i, id := range ids {
		position := i + 1
		if current[id] == position {
			continue
		}
		if err := tx.RawQuery("UPDATE tasks SET position = ? WHERE id = ?", position, id).Exec(); err != nil {
			return err
		}
	}
	return nil
}
```

### 3. Tests de modelo (`models/task_test.go`)

Imports: `"strings"` y, en un grupo aparte, `"github.com/gofrs/uuid"`. Añadir:

```go
func (ms *ModelSuite) createTasks(titles ...string) []*Task {
	tasks := make([]*Task, 0, len(titles))
	for _, title := range titles {
		task := &Task{Title: title}
		verrs, err := ms.DB.ValidateAndCreate(task)
		ms.NoError(err)
		ms.False(verrs.HasAny())
		ms.NoError(ms.DB.Reload(task))
		tasks = append(tasks, task)
	}
	return tasks
}

func (ms *ModelSuite) positions(tasks ...*Task) []int {
	positions := make([]int, 0, len(tasks))
	for _, task := range tasks {
		ms.NoError(ms.DB.Reload(task))
		positions = append(positions, task.Position)
	}
	return positions
}

func (ms *ModelSuite) Test_Task_Create_AppendsAtEnd() {
	tasks := ms.createTasks("Primera", "Segunda", "Tercera")
	ms.Equal([]int{1, 2, 3}, ms.positions(tasks...))
}

func (ms *ModelSuite) Test_ReorderTasks_AssignsConsecutivePositions() {
	tasks := ms.createTasks("Primera", "Segunda", "Tercera")
	a, b, c := tasks[0], tasks[1], tasks[2]
	updatedAt := a.UpdatedAt

	ms.NoError(ReorderTasks(ms.DB, []uuid.UUID{c.ID, a.ID, b.ID}))
	ms.Equal([]int{2, 3, 1}, ms.positions(a, b, c))
	ms.True(a.UpdatedAt.Equal(updatedAt))

	next := ms.createTasks("Cuarta")[0]
	ms.Equal(4, next.Position)
}

func (ms *ModelSuite) Test_ReorderTasks_RejectsInvalidLists() {
	tasks := ms.createTasks("Primera", "Segunda")
	a, b := tasks[0], tasks[1]
	unknown := uuid.Must(uuid.NewV4())

	for _, ids := range [][]uuid.UUID{
		{b.ID},
		{b.ID, b.ID},
		{b.ID, unknown},
		{b.ID, a.ID, unknown},
	} {
		ms.ErrorIs(ReorderTasks(ms.DB, ids), ErrInvalidTaskOrder)
	}
	ms.Equal([]int{1, 2}, ms.positions(a, b))
}
```

### 4. Acciones (`actions/tasks.go`)

- Imports: añadir `"errors"` al grupo estándar (antes de `"net/http"`) y `"github.com/gofrs/uuid"` al grupo de terceros (después de `"github.com/gobuffalo/validate/v3"`).
- Línea 19 (`TasksIndex`) y línea 46 (`TasksCreate`): `tx.Order("created_at asc, id asc")` pasa a ser `tx.Order(models.TaskListOrder)`.
- Añadir después de `TasksToggle`:

```go
// TasksReorder guarda el orden manual de la lista completa de tareas.
func TasksReorder(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	input := struct {
		TaskIDs []string `json:"task_ids"`
	}{}
	if err := c.Bind(&input); err != nil {
		return c.Error(http.StatusBadRequest, err)
	}

	invalidOrder := func() error {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(map[string]string{
			"error": "El orden no coincide con las tareas actuales. Recarga la página e inténtalo de nuevo.",
		}))
	}

	ids := make([]uuid.UUID, 0, len(input.TaskIDs))
	for _, value := range input.TaskIDs {
		id, err := uuid.FromString(value)
		if err != nil {
			return invalidOrder()
		}
		ids = append(ids, id)
	}

	if err := models.ReorderTasks(tx, ids); err != nil {
		if errors.Is(err, models.ErrInvalidTaskOrder) {
			return invalidOrder()
		}
		return err
	}

	return c.Render(http.StatusNoContent, nil)
}
```

### 5. Ruta (`actions/app.go`)

Entre `app.POST("/tasks", TasksCreate)` y `app.GET("/tasks/{task_id}/edit", TasksEdit)`:

```go
		// Antes de "/tasks/{task_id}": las rutas se evalúan en orden de registro.
		app.PUT("/tasks/order", TasksReorder)
```

### 6. Tests de acción (`actions/tasks_test.go`)

No hacen falta imports nuevos (ya están `net/url`, `strings`, `todo/models` y `github.com/gofrs/uuid`). Añadir:

```go
func (as *ActionSuite) createTasks(titles ...string) []*models.Task {
	tasks := make([]*models.Task, 0, len(titles))
	for _, title := range titles {
		task := &models.Task{Title: title}
		verrs, err := as.DB.ValidateAndCreate(task)
		as.NoError(err)
		as.False(verrs.HasAny())
		tasks = append(tasks, task)
	}
	return tasks
}

func (as *ActionSuite) assertListedInOrder(body string, titles ...string) {
	last := -1
	for _, title := range titles {
		index := strings.Index(body, `class="task-title">`+title+`</span>`)
		as.Greater(index, last, title)
		last = index
	}
}

func (as *ActionSuite) Test_TasksIndex_OrdersByPosition() {
	tasks := as.createTasks("Tarea alfa", "Tarea beta", "Tarea gamma")
	as.NoError(models.ReorderTasks(as.DB, []uuid.UUID{tasks[2].ID, tasks[0].ID, tasks[1].ID}))

	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.assertListedInOrder(res.Body.String(), "Tarea gamma", "Tarea alfa", "Tarea beta")
}

func (as *ActionSuite) Test_TasksIndex_RendersReorderControls() {
	task := as.createTasks("Tarea alfa")[0]

	res := as.HTML("/").Get()
	as.Equal(200, res.Code)

	body := res.Body.String()
	as.Contains(body, `data-reorder-url="/tasks/order"`)
	as.Contains(body, `data-task-id="`+task.ID.String()+`"`)
	as.Contains(body, "data-drag-handle")
	as.Contains(body, `aria-label="Subir tarea: Tarea alfa"`)
	as.Contains(body, `aria-label="Bajar tarea: Tarea alfa"`)
	as.Contains(body, "data-reorder-feedback")
}

func (as *ActionSuite) Test_TasksReorder_PersistsOrder() {
	tasks := as.createTasks("Tarea alfa", "Tarea beta", "Tarea gamma")
	a, b, c := tasks[0], tasks[1], tasks[2]

	res := as.JSON("/tasks/order").Put(map[string][]string{
		"task_ids": {c.ID.String(), a.ID.String(), b.ID.String()},
	})
	as.Equal(204, res.Code)

	as.NoError(as.DB.Reload(a))
	as.NoError(as.DB.Reload(b))
	as.NoError(as.DB.Reload(c))
	as.Equal(2, a.Position)
	as.Equal(3, b.Position)
	as.Equal(1, c.Position)

	page := as.HTML("/").Get()
	as.assertListedInOrder(page.Body.String(), "Tarea gamma", "Tarea alfa", "Tarea beta")
}

func (as *ActionSuite) Test_TasksReorder_InvalidList() {
	tasks := as.createTasks("Tarea alfa", "Tarea beta")
	a, b := tasks[0], tasks[1]

	for _, ids := range [][]string{
		{b.ID.String()},
		{b.ID.String(), b.ID.String()},
		{b.ID.String(), a.ID.String(), uuid.Must(uuid.NewV4()).String()},
		{b.ID.String(), "no-es-un-uuid"},
	} {
		res := as.JSON("/tasks/order").Put(map[string][]string{"task_ids": ids})
		as.Equal(422, res.Code)
		as.Contains(res.Body.String(), "Recarga la página")
	}

	as.NoError(as.DB.Reload(a))
	as.NoError(as.DB.Reload(b))
	as.Equal(1, a.Position)
	as.Equal(2, b.Position)
}

func (as *ActionSuite) Test_TasksCreate_AppendsAtEnd() {
	tasks := as.createTasks("Tarea alfa", "Tarea beta")
	as.NoError(models.ReorderTasks(as.DB, []uuid.UUID{tasks[1].ID, tasks[0].ID}))

	res := as.HTML("/tasks").Post(&models.Task{Title: "Tarea gamma"})
	as.Equal(303, res.Code)

	created := &models.Task{}
	as.NoError(as.DB.Where("title = ?", "Tarea gamma").First(created))
	as.Equal(3, created.Position)

	page := as.HTML("/").Get()
	as.assertListedInOrder(page.Body.String(), "Tarea beta", "Tarea alfa", "Tarea gamma")
}

func (as *ActionSuite) Test_TasksToggle_KeepsPosition() {
	tasks := as.createTasks("Tarea alfa", "Tarea beta", "Tarea gamma")
	a, b, c := tasks[0], tasks[1], tasks[2]
	as.NoError(models.ReorderTasks(as.DB, []uuid.UUID{c.ID, a.ID, b.ID}))

	for i := 0; i < 2; i++ {
		res := as.HTML("/tasks/%s/toggle", a.ID).Put(url.Values{})
		as.Equal(303, res.Code)
		as.NoError(as.DB.Reload(a))
		as.Equal(2, a.Position)
	}

	page := as.HTML("/").Get()
	as.assertListedInOrder(page.Body.String(), "Tarea gamma", "Tarea alfa", "Tarea beta")
}
```

Notas:
- El helper JSON de los tests envía `Accept: application/json` sin `Content-Type`, y Buffalo recurre a `Accept` para elegir el binder JSON. Con `GO_ENV=test` el CSRF está desactivado.
- `Test_TasksReorder_PersistsOrder` también demuestra que la ruta no queda tapada por `/tasks/{task_id}`: en ese caso devolvería `404`.

### 7. Template de la lista (`templates/tasks/index.plush.html`)

Sustituir el bloque `<ul class="task-list"> … </ul>` (rama `else`) por el siguiente. Los formularios de completar y eliminar y el enlace de editar se quedan **exactamente** como están; abajo se abrevian con comentarios:

```html
    <ul class="task-list" data-task-list data-reorder-url="/tasks/order" data-csrf-token="<%= authenticity_token %>">
      <%= for (t) in tasks { %>
        <li class="task<%= if (t.Completed) { %> task-completed<% } %>" data-task-id="<%= t.ID %>">
          <!-- formulario .task-toggle sin cambios -->
          <span class="task-title"><%= t.Title %></span>
          <%= if (t.Completed) { %>
            <span class="task-status">Completada</span>
          <% } %>
          <div class="task-actions">
            <span class="task-move-group">
              <button type="button" class="icon-action task-move" data-move="up" aria-label="Subir tarea: <%= t.Title %>" title="Subir tarea">
                <svg aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M12 19V5M6 11l6-6 6 6" />
                </svg>
              </button>
              <button type="button" class="icon-action task-move" data-move="down" aria-label="Bajar tarea: <%= t.Title %>" title="Bajar tarea">
                <svg aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M12 5v14M6 13l6 6 6-6" />
                </svg>
              </button>
            </span>
            <!-- enlace Editar y formulario Eliminar sin cambios -->
          </div>
          <span class="task-handle" data-drag-handle aria-hidden="true" title="Arrastrar para reordenar">
            <svg viewBox="0 0 24 24" fill="currentColor">
              <circle cx="9" cy="6" r="1.5" />
              <circle cx="15" cy="6" r="1.5" />
              <circle cx="9" cy="12" r="1.5" />
              <circle cx="15" cy="12" r="1.5" />
              <circle cx="9" cy="18" r="1.5" />
              <circle cx="15" cy="18" r="1.5" />
            </svg>
          </span>
        </li>
      <% } %>
    </ul>
    <p class="sr-only" role="status" data-reorder-feedback></p>
```

(Los comentarios HTML de arriba solo abrevian el plan; no se copian.)

### 8. Layout (`templates/application.plush.html`)

`/assets/app.css?v=3` pasa a `?v=4` y `/assets/app.js?v=2` pasa a `?v=3`.

### 9. JavaScript (`public/assets/app.js`)

El listener `submit` existente no se toca. Al final del archivo se añade el siguiente bloque (su sintaxis se ha comprobado con Node):

```js
// Reordenación manual: arrastre con ratón o táctil desde el asa y botones Subir/Bajar.
(function () {
  var list = document.querySelector("[data-task-list]");
  if (!list) {
    return;
  }

  var feedback = document.querySelector("[data-reorder-feedback]");
  var savedOrder = taskIds();
  var saving = false;
  var pending = false;
  var drag = null;

  list.classList.add("task-list-reorderable");

  function tasks() {
    return Array.prototype.filter.call(list.children, function (element) {
      return element.classList.contains("task");
    });
  }

  function taskIds() {
    return tasks().map(function (task) {
      return task.dataset.taskId;
    });
  }

  function sameOrder(a, b) {
    return a.length === b.length && a.every(function (id, index) {
      return id === b[index];
    });
  }

  function announce(message, isError) {
    feedback.textContent = message;
    feedback.classList.toggle("sr-only", !isError);
    feedback.classList.toggle("field-error", Boolean(isError));
  }

  function announcePosition(task) {
    var all = tasks();
    var title = task.querySelector(".task-title").textContent;
    announce("Tarea «" + title + "» movida a la posición " + (all.indexOf(task) + 1) + " de " + all.length + ".");
  }

  // Coloca task justo antes de target (null = al final) moviendo solo sus vecinas,
  // para que la tarea movida no salga del DOM y conserve el foco y la captura del puntero.
  function placeBefore(task, target) {
    if (target === task || task.nextElementSibling === target) {
      return;
    }
    var movingDown = target === null || Boolean(task.compareDocumentPosition(target) & Node.DOCUMENT_POSITION_FOLLOWING);
    while (task.nextElementSibling !== target) {
      if (movingDown) {
        list.insertBefore(task.nextElementSibling, task);
      } else {
        list.insertBefore(task.previousElementSibling, task.nextElementSibling);
      }
    }
  }

  function restore(ids) {
    var active = document.activeElement;
    var byId = {};
    tasks().forEach(function (task) {
      byId[task.dataset.taskId] = task;
    });
    ids.forEach(function (id) {
      if (byId[id]) {
        list.appendChild(byId[id]);
      }
    });
    if (active && list.contains(active)) {
      active.focus();
    }
  }

  function save() {
    if (saving) {
      pending = true;
      return;
    }
    var ids = taskIds();
    if (sameOrder(ids, savedOrder)) {
      return;
    }
    saving = true;
    fetch(list.dataset.reorderUrl, {
      method: "PUT",
      credentials: "same-origin",
      headers: {
        "Accept": "application/json",
        "Content-Type": "application/json",
        "X-CSRF-Token": list.dataset.csrfToken
      },
      body: JSON.stringify({ task_ids: ids })
    })
      .then(function (response) {
        if (!response.ok) {
          throw new Error("HTTP " + response.status);
        }
        savedOrder = ids;
      })
      .catch(function () {
        pending = false;
        restore(savedOrder);
        announce("No se pudo guardar el nuevo orden. Recarga la página e inténtalo de nuevo.", true);
      })
      .finally(function () {
        saving = false;
        if (pending) {
          pending = false;
          save();
        }
      });
  }

  // Alternativa con teclado (y lectores de pantalla): botones Subir/Bajar.
  list.addEventListener("click", function (event) {
    var button = event.target.closest("[data-move]");
    if (!button) {
      return;
    }
    var task = button.closest(".task");
    var up = button.dataset.move === "up";
    var neighbour = up ? task.previousElementSibling : task.nextElementSibling;
    if (!neighbour) {
      announce(up ? "La tarea ya es la primera de la lista." : "La tarea ya es la última de la lista.");
      return;
    }
    placeBefore(task, up ? neighbour : neighbour.nextElementSibling);
    announcePosition(task);
    save();
  });

  // Arrastre con ratón, lápiz o táctil (Pointer Events) desde el asa.
  list.addEventListener("pointerdown", function (event) {
    var handle = event.target.closest("[data-drag-handle]");
    if (!handle || drag || !event.isPrimary || event.button !== 0) {
      return;
    }
    event.preventDefault();
    var task = handle.closest(".task");
    handle.setPointerCapture(event.pointerId);
    drag = { task: task, pointerId: event.pointerId, startOrder: taskIds() };
    task.classList.add("task-dragging");
    document.body.classList.add("is-reordering");
  });

  list.addEventListener("pointermove", function (event) {
    if (!drag || event.pointerId !== drag.pointerId) {
      return;
    }
    event.preventDefault();
    var target = null;
    var others = tasks().filter(function (task) {
      return task !== drag.task;
    });
    for (var i = 0; i < others.length; i++) {
      var rect = others[i].getBoundingClientRect();
      if (event.clientY < rect.top + rect.height / 2) {
        target = others[i];
        break;
      }
    }
    placeBefore(drag.task, target);

    var edge = 48;
    if (event.clientY < edge) {
      window.scrollBy(0, -12);
    } else if (event.clientY > window.innerHeight - edge) {
      window.scrollBy(0, 12);
    }
  });

  function endDrag(event, commit) {
    if (!drag || event.pointerId !== drag.pointerId) {
      return;
    }
    var current = drag;
    drag = null;
    current.task.classList.remove("task-dragging");
    document.body.classList.remove("is-reordering");
    if (!commit) {
      restore(current.startOrder);
      return;
    }
    if (!sameOrder(taskIds(), current.startOrder)) {
      announcePosition(current.task);
      save();
    }
  }

  list.addEventListener("pointerup", function (event) {
    endDrag(event, true);
  });

  list.addEventListener("pointercancel", function (event) {
    endDrag(event, false);
  });
})();
```

### 10. CSS (`public/assets/app.css`)

1. Después de `.icon-danger:hover { … }`, añadir:

```css
.task-list:not(.task-list-reorderable) .task-handle,
.task-list:not(.task-list-reorderable) .task-move-group {
  display: none;
}

.task-list-reorderable .task {
  grid-template-columns: 44px minmax(0, 1fr) auto 32px;
}

.task-move-group {
  display: flex;
}

.task-handle {
  display: grid;
  place-items: center;
  min-height: 44px;
  color: var(--color-muted);
  cursor: grab;
  touch-action: none;
  user-select: none;
  -webkit-user-select: none;
  -webkit-touch-callout: none;
}

.task-handle svg {
  width: 18px;
  height: 18px;
}

.task-dragging {
  position: relative;
  z-index: 1;
  background: var(--color-sidebar);
  box-shadow: 0 6px 18px rgba(40, 40, 38, 0.12);
}

.is-reordering,
.is-reordering * {
  cursor: grabbing;
  user-select: none;
  -webkit-user-select: none;
}
```

2. Dentro de `@media (hover: hover) and (pointer: fine)`, el asa debe comportarse como `.task-actions`. Añadir `.task-handle` al selector con `opacity: 0; transition: opacity 120ms ease;` y añadir `.task:hover .task-handle`, `.task:focus-within .task-handle` y `.task-dragging .task-handle` a la regla `opacity: 1`.

3. Dentro de `@media (max-width: 760px)`, **después** de la regla `.task { grid-template-columns: 36px minmax(0, 1fr) auto; }`, añadir:

```css
  .task-list-reorderable .task {
    grid-template-columns: 36px minmax(0, 1fr) auto 32px;
  }

  .task-move-group:not(:focus-within) {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }
```

(Hace falta `.task-list-reorderable .task` dentro del media query porque la regla general de cuatro columnas es más específica que `.task`.)

### 11. Validación local y revisión final

Ejecuta las comprobaciones del apartado (a) de «Tests y validación» y revisa `git diff` y `git status --short`.

## Tests y validación

### Tests que demuestran el comportamiento

| Test | Demuestra |
| --- | --- |
| `models.Test_Task_Create_AppendsAtEnd` | Las tareas nuevas reciben `1, 2, 3…` (CA5) |
| `models.Test_ReorderTasks_AssignsConsecutivePositions` | Posiciones `1..N` según la lista, `updated_at` intacto y una tarea nueva tras reordenar va al final (CA4, CA5, IN4) |
| `models.Test_ReorderTasks_RejectsInvalidLists` | Listas incompletas, con duplicados o con ids desconocidos se rechazan sin escribir nada (IN3) |
| `actions.Test_TasksIndex_OrdersByPosition` | La lista se ordena por `position`, no por fecha de creación (CA4) |
| `actions.Test_TasksIndex_RendersReorderControls` | El HTML contiene el endpoint, `data-task-id`, el asa, los botones con `aria-label` y la región de anuncios (CA1–CA3) |
| `actions.Test_TasksReorder_PersistsOrder` | `PUT /tasks/order` responde `204`, persiste y la lista recargada respeta el orden; la ruta no queda tapada (CA4) |
| `actions.Test_TasksReorder_InvalidList` | `422` con mensaje para listas inválidas o ids mal formados, sin cambios en la BD (IN3) |
| `actions.Test_TasksCreate_AppendsAtEnd` | Una tarea creada desde el formulario queda la última con un orden manual previo (CA5) |
| `actions.Test_TasksToggle_KeepsPosition` | Completar y reabrir no cambia la posición (CA6) |

Los tests existentes deben seguir en verde sin cambios. `Test_TasksIndex_ListsTasks` usa la fixture con `position = 0` por defecto y solo comprueba que aparecen las tareas.

### (a) Comprobaciones ejecutables en este checkout (sin Go ni PostgreSQL)

No hay toolchain de Go ni PostgreSQL y **no se prepara ningún entorno**: no se instala Go ni se arranca Docker, compose ni la BD.

```bash
git diff --check                                        # espacios finales y conflictos en archivos versionados
node --check public/assets/app.js                       # sintaxis del JS
grep -nE '[[:space:]]+$' migrations/20261013000000_add_position_to_tasks.*.fizz   # archivos nuevos (no aparecen en git diff); no debe imprimir nada
git diff -U0 -- '*.go' | grep -nE '^\+ ' || echo "ok: Go indentado con tabuladores"
git status --short                                      # solo los archivos de «Archivos relevantes» y este plan
```

Revisión manual del formato gofmt, porque no se puede ejecutar:
- indentación con tabuladores;
- imports ordenados alfabéticamente dentro de cada grupo (`database/sql` < `errors` < `fmt`…, y `gobuffalo/validate/v3` < `gofrs/uuid`);
- columnas de `Task` alineadas como en el bloque del paso 2 (nombres a 10 columnas y tipos a 10 columnas);
- sin comentarios al final de las líneas de literales compuestos;
- una línea en blanco entre funciones.

Revisión manual de Plush: cada `<%= if … { %>` / `<%= for … { %>` se cierra con su `<% } %>` y el `<p … data-reorder-feedback>` queda dentro de la rama `else`.

### (b) Comprobaciones que quedan para la CI del PR (`.github/workflows/ci.yaml`)

- `gofmt -l .` (no debe listar archivos).
- `go vet ./...`.
- `soda migrate -e test`: aplica la migración fizz nueva sobre PostgreSQL 17 y valida la sintaxis del `ALTER`/`UPDATE`.
- `go test -p 1 ./...`: tests nuevos y existentes de `actions`, `models` y `adws`.
- `go build ./...` y `docker build -t todo .`.

Si la CI falla en compilación o formato, se corrige en la misma rama sin cambiar el diseño.

### (c) Verificación manual en navegador (devcontainer o entorno con la app arrancada; no en este checkout)

1. Con la BD existente, al arrancar (`models.Migrate()` en `cmd/app/main.go`) las tareas conservan el orden anterior.
2. **Ratón:** arrastrar una tarea desde el asa (aparece al pasar por encima) a otra posición y recargar. El orden se mantiene.
3. **Táctil:** en un dispositivo real o con la emulación táctil de DevTools, arrastrar desde el asa sin que la página se desplace; fuera del asa la página se desplaza. Recargar: el orden se mantiene.
4. **Teclado:** recorrer con `Tab` hasta «Subir tarea: …»/«Bajar tarea: …», pulsar `Enter`/`Espacio` varias veces. El foco sigue en el botón y un lector de pantalla anuncia la posición. En los extremos se anuncia que ya es la primera o la última.
5. Crear una tarea: aparece la última. Completar o reabrir una tarea: no se mueve. Por el defecto previo del JS de completar, el **estado** completado puede no persistir al recargar, pero la **posición** sí.
6. Abrir dos pestañas, crear una tarea en una y reordenar en la otra sin recargar: aparece el mensaje de error visible, la lista vuelve al último orden guardado y al recargar todo es coherente.
7. Con JavaScript desactivado no se muestran el asa ni los botones Subir/Bajar, y la lista funciona como antes.

## Riesgos y recuperación

| Riesgo | Mitigación / recuperación |
| --- | --- |
| No se puede compilar ni ejecutar tests en este checkout | El código del plan se ha contrastado con las fuentes de buffalo@v1.1.4, pop/v6@v6.4.1, fizz@v1.14.4, middleware@v1.0.0, suite/v4, httptest@v1.5.2 y testify@v1.12.1 de la caché de módulos. La CI del PR es la puerta de validación (gofmt, vet, migración, tests, build). |
| `/tasks/{task_id}` captura `/tasks/order` | La ruta nueva se registra antes, con un comentario. Si alguien reordena las rutas, `Test_TasksReorder_PersistsOrder` falla con `404`. |
| Migración sobre datos existentes | Va dentro de una transacción. El relleno conserva el orden actual. El `down` (`soda migrate down -e <entorno>`, un paso) elimina la columna y con ella el orden manual. |
| Concurrencia: dos altas simultáneas obtienen la misma posición; dos reordenaciones simultáneas se entremezclan | No hay restricción de unicidad. Los empates se resuelven siempre igual (`created_at`, `id`) y la siguiente reordenación renumera `1..N`. Es asumible en una app sin cuentas y de uso individual. |
| Lista desactualizada en el navegador (otra pestaña creó o borró tareas) | El servidor responde `422` sin escribir nada, el JS restaura el último orden guardado y pide recargar. |
| El JS no tiene tests automáticos | La lógica se concentra en un bloque pequeño, `node --check` lo valida localmente y hay una lista de verificación manual (c). Mover las vecinas en lugar del `<li>` evita perder el foco y la captura del puntero. |
| Filas estrechas en móvil | Los botones Subir/Bajar quedan visualmente ocultos hasta recibir foco, pero siguen accesibles con teclado y lector de pantalla. El arrastre táctil va desde el asa. |
| Caché de recursos estáticos | Se suben las versiones `?v=` de `app.css` y `app.js`. |
| Pequeño salto de maquetación al cargar (el JS añade la cuarta columna) | Es cosmético y deliberado: así no se muestran controles inertes sin JS (IN5). |
| Defecto previo del JS de completar (`GET` a `/tasks/{id}/toggle`) | Fuera de alcance. Se documenta para que la verificación manual no lo confunda con un fallo del orden. Conviene tratarlo en una issue aparte. |
| Revertir | Revertir el commit basta: el código anterior ignora la columna extra porque pop selecciona columnas explícitas y vuelve a ordenar por fecha. Opcionalmente, se ejecuta el `down` de la migración. Si más adelante se reaplica la feature, las tareas creadas mientras estuvo revertida tendrán `position = 0` y aparecerán primero hasta la siguiente reordenación. |
