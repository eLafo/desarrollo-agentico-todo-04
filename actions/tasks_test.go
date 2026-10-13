package actions

import (
	"net/url"
	"strings"

	"todo/models"

	"github.com/gofrs/uuid"
)

func (as *ActionSuite) Test_TasksIndex_EmptyState() {
	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "No hay tareas todavía")
}

func (as *ActionSuite) Test_TasksIndex_ListsTasks() {
	as.LoadFixture("tres tareas")

	res := as.HTML("/").Get()
	as.Equal(200, res.Code)

	body := res.Body.String()
	as.Contains(body, "Comprar material de oficina")
	as.Contains(body, "Llamar al gestor")
	as.Contains(body, "Publicar la oferta de empleo")
	as.Contains(body, "Completada")
}

func (as *ActionSuite) Test_TasksCreate_Valid() {
	res := as.HTML("/tasks").Post(&models.Task{Title: "  Nueva tarea de prueba  "})
	as.Equal(303, res.Code)
	as.Equal("/", res.Header().Get("Location"))

	task := &models.Task{}
	as.NoError(as.DB.First(task))
	as.Equal("Nueva tarea de prueba", task.Title)
	as.False(task.Completed)
}

func (as *ActionSuite) Test_TasksCreate_EmptyTitle_Invalid() {
	res := as.HTML("/tasks").Post(&models.Task{Title: ""})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksCreate_OnlySpaces_Invalid() {
	res := as.HTML("/tasks").Post(&models.Task{Title: "     "})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksToggle_CompletesAndReopens() {
	task := &models.Task{Title: "Pendiente de completar"}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/tasks/%s/toggle", task.ID).Put(url.Values{})
	as.Equal(303, res.Code)

	as.NoError(as.DB.Reload(task))
	as.True(task.Completed)

	res = as.HTML("/tasks/%s/toggle", task.ID).Put(url.Values{})
	as.Equal(303, res.Code)

	as.NoError(as.DB.Reload(task))
	as.False(task.Completed)
}

func (as *ActionSuite) Test_TasksToggle_NotFound() {
	id := uuid.Must(uuid.NewV4())
	res := as.HTML("/tasks/%s/toggle", id).Put(url.Values{})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksToggle_InvalidID() {
	res := as.HTML("/tasks/no-es-un-uuid/toggle").Put(url.Values{})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksDestroy() {
	task := &models.Task{Title: "Tarea que se elimina"}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/tasks/%s", task.ID).Delete()
	as.Equal(303, res.Code)

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksDestroy_NotFound() {
	id := uuid.Must(uuid.NewV4())
	res := as.HTML("/tasks/%s", id).Delete()
	as.Equal(404, res.Code)
}

func (as *ActionSuite) taskForEditing(completed bool) *models.Task {
	task := &models.Task{Title: "Título original", Completed: completed}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())
	as.NoError(as.DB.Reload(task))
	return task
}

func (as *ActionSuite) Test_TasksIndex_EditLink() {
	task := as.taskForEditing(false)
	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "/tasks/"+task.ID.String()+"/edit")
	as.Contains(res.Body.String(), "Editar tarea: Título original")
}

func (as *ActionSuite) Test_TasksEdit_PrefillsTitle() {
	task := as.taskForEditing(true)
	res := as.HTML("/tasks/%s/edit", task.ID).Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), `value="Título original"`)
	as.Contains(res.Body.String(), `value="PUT"`)
	as.Contains(res.Body.String(), "Guardar cambios")
	as.Contains(res.Body.String(), "Cancelar")
	as.NoError(as.DB.Reload(task))
	as.True(task.Completed)
}

func (as *ActionSuite) Test_TasksEdit_NotFound() {
	res := as.HTML("/tasks/%s/edit", uuid.Must(uuid.NewV4())).Get()
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksEdit_InvalidID() {
	res := as.HTML("/tasks/no-es-un-uuid/edit").Get()
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksEdit_EscapesTitle() {
	task := &models.Task{Title: `<img src=x onerror="alert(1)">`}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())
	res := as.HTML("/tasks/%s/edit", task.ID).Get()
	as.Equal(200, res.Code)
	as.NotContains(res.Body.String(), `<img src=x`)
	as.Contains(res.Body.String(), "&lt;img")
}

func (as *ActionSuite) Test_TasksUpdate_PersistsTitleOnly() {
	task := as.taskForEditing(true)
	id, createdAt, updatedAt := task.ID, task.CreatedAt, task.UpdatedAt
	position := task.Position
	res := as.HTML("/tasks/%s", id).Put(url.Values{
		"title": {"  Título actualizado  "}, "completed": {"false"},
		"id": {uuid.Must(uuid.NewV4()).String()}, "created_at": {"2000-01-01T00:00:00Z"},
	})
	as.Equal(303, res.Code)
	as.Equal("/", res.Header().Get("Location"))
	as.NoError(as.DB.Reload(task))
	as.Equal("Título actualizado", task.Title)
	as.Equal(id, task.ID)
	as.Equal(position, task.Position)
	as.True(task.Completed)
	as.True(task.CreatedAt.Equal(createdAt))
	as.False(task.UpdatedAt.Before(updatedAt))
	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(1, count)
}

func (as *ActionSuite) Test_TasksUpdate_PendingTaskRemainsPending() {
	task := as.taskForEditing(false)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {"Nuevo título"}, "completed": {"true"}})
	as.Equal(303, res.Code)
	as.NoError(as.DB.Reload(task))
	as.Equal("Nuevo título", task.Title)
	as.False(task.Completed)
}

func (as *ActionSuite) Test_TasksUpdate_EmptyTitle() {
	task := as.taskForEditing(false)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {""}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")
	as.Contains(res.Body.String(), `aria-describedby="title-error"`)
	as.NoError(as.DB.Reload(task))
	as.Equal("Título original", task.Title)
}

func (as *ActionSuite) Test_TasksUpdate_OnlySpaces() {
	task := as.taskForEditing(true)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {"   "}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")
	as.NoError(as.DB.Reload(task))
	as.Equal("Título original", task.Title)
	as.True(task.Completed)
}

func (as *ActionSuite) Test_TasksUpdate_TitleTooLong() {
	task := as.taskForEditing(false)
	title := strings.Repeat("á", models.TitleMaxLength+1)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {title}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título no puede superar los 200 caracteres.")
	as.Contains(res.Body.String(), title)
	as.NoError(as.DB.Reload(task))
	as.Equal("Título original", task.Title)
}

func (as *ActionSuite) Test_TasksUpdate_TitleAtLimit() {
	task := as.taskForEditing(false)
	title := strings.Repeat("á", models.TitleMaxLength)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {title}})
	as.Equal(303, res.Code)
	as.NoError(as.DB.Reload(task))
	as.Equal(title, task.Title)
}

func (as *ActionSuite) Test_TasksUpdate_NotFound() {
	res := as.HTML("/tasks/%s", uuid.Must(uuid.NewV4())).Put(url.Values{"title": {"Título nuevo"}})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksUpdate_InvalidID() {
	res := as.HTML("/tasks/no-es-un-uuid").Put(url.Values{"title": {"Título nuevo"}})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) createTasks(titles ...string) []*models.Task {
	tasks := make([]*models.Task, 0, len(titles))
	for _, title := range titles {
		task := &models.Task{Title: title}
		verrs, err := as.DB.ValidateAndCreate(task)
		as.NoError(err)
		as.False(verrs.HasAny())
		as.NoError(as.DB.Reload(task))
		tasks = append(tasks, task)
	}
	return tasks
}

// assertOrder comprueba que los textos aparecen en body en el orden dado.
func (as *ActionSuite) assertOrder(body string, texts ...string) {
	last := -1
	for _, text := range texts {
		index := strings.Index(body, text)
		as.GreaterOrEqual(index, 0, text)
		as.Greater(index, last, text)
		last = index
	}
}

// assertPositions comprueba que las tareas tienen las posiciones 1..n en el orden dado.
func (as *ActionSuite) assertPositions(tasks ...*models.Task) {
	for i, task := range tasks {
		reloaded := &models.Task{}
		as.NoError(as.DB.Find(reloaded, task.ID))
		as.Equal(i+1, reloaded.Position, task.Title)
	}
}

func (as *ActionSuite) reorder(tasks ...*models.Task) int {
	ids := make([]string, 0, len(tasks))
	for _, task := range tasks {
		ids = append(ids, task.ID.String())
	}
	return as.HTML("/tasks/order").Put(url.Values{"task_ids": ids}).Code
}

func (as *ActionSuite) Test_TasksIndex_OrdersByPosition() {
	for i, title := range []string{"Tercera tarea", "Segunda tarea", "Primera tarea"} {
		task := &models.Task{Title: title, Position: 3 - i}
		verrs, err := as.DB.ValidateAndCreate(task)
		as.NoError(err)
		as.False(verrs.HasAny())
	}

	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.assertOrder(res.Body.String(), "Primera tarea", "Segunda tarea", "Tercera tarea")
}

func (as *ActionSuite) Test_TasksIndex_RendersReorderControls() {
	task := as.taskForEditing(false)

	res := as.HTML("/").Get()
	as.Equal(200, res.Code)

	body := res.Body.String()
	as.Contains(body, `data-task-id="`+task.ID.String()+`"`)
	as.Contains(body, `aria-label="Mover tarea: Título original"`)
	as.Contains(body, `aria-describedby="task-order-help"`)
	as.Contains(body, `action="/tasks/order"`)
	as.Contains(body, "data-task-order")
}

func (as *ActionSuite) Test_TasksIndex_EmptyState_NoReorderForm() {
	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.NotContains(res.Body.String(), "/tasks/order")
}

func (as *ActionSuite) Test_TasksReorder_PersistsOrder() {
	tasks := as.createTasks("Primera tarea", "Segunda tarea", "Tercera tarea")
	a, b, c := tasks[0], tasks[1], tasks[2]

	as.Equal(204, as.reorder(c, a, b))
	as.assertPositions(c, a, b)

	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.assertOrder(res.Body.String(), "Tercera tarea", "Primera tarea", "Segunda tarea")
}

func (as *ActionSuite) Test_TasksReorder_IncompleteList() {
	tasks := as.createTasks("Primera tarea", "Segunda tarea", "Tercera tarea")
	a, b, c := tasks[0], tasks[1], tasks[2]

	as.Equal(422, as.reorder(c, a))
	as.assertPositions(a, b, c)
}

func (as *ActionSuite) Test_TasksReorder_DuplicateID() {
	tasks := as.createTasks("Primera tarea", "Segunda tarea", "Tercera tarea")
	a, b, c := tasks[0], tasks[1], tasks[2]

	as.Equal(422, as.reorder(c, c, a))
	as.assertPositions(a, b, c)
}

func (as *ActionSuite) Test_TasksReorder_UnknownID() {
	tasks := as.createTasks("Primera tarea", "Segunda tarea", "Tercera tarea")
	a, b, c := tasks[0], tasks[1], tasks[2]

	unknown := &models.Task{ID: uuid.Must(uuid.NewV4())}
	as.Equal(422, as.reorder(c, a, unknown))
	as.assertPositions(a, b, c)
}

func (as *ActionSuite) Test_TasksReorder_InvalidID() {
	tasks := as.createTasks("Primera tarea", "Segunda tarea", "Tercera tarea")
	a, b, c := tasks[0], tasks[1], tasks[2]

	res := as.HTML("/tasks/order").Put(url.Values{
		"task_ids": {c.ID.String(), "no-es-un-uuid", b.ID.String()},
	})
	as.Equal(422, res.Code)
	as.assertPositions(a, b, c)
}

func (as *ActionSuite) Test_TasksCreate_AppendsAtEnd() {
	tasks := as.createTasks("Primera tarea", "Segunda tarea", "Tercera tarea")
	a, b, c := tasks[0], tasks[1], tasks[2]
	as.Equal(204, as.reorder(c, a, b))

	res := as.HTML("/tasks").Post(url.Values{"title": {"Nueva tarea"}, "position": {"1"}})
	as.Equal(303, res.Code)

	created := &models.Task{}
	as.NoError(as.DB.Where("title = ?", "Nueva tarea").First(created))
	as.Equal(4, created.Position)

	res = as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.assertOrder(res.Body.String(), "Tercera tarea", "Primera tarea", "Segunda tarea", "Nueva tarea")
}

func (as *ActionSuite) Test_TasksToggle_KeepsPosition() {
	tasks := as.createTasks("Primera tarea", "Segunda tarea", "Tercera tarea")
	a, b, c := tasks[0], tasks[1], tasks[2]
	as.Equal(204, as.reorder(c, a, b))

	res := as.HTML("/tasks/%s/toggle", a.ID).Put(url.Values{})
	as.Equal(303, res.Code)

	as.NoError(as.DB.Reload(a))
	as.True(a.Completed)
	as.assertPositions(c, a, b)
}
