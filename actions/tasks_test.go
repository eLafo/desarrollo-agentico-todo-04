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
	res := as.HTML("/tasks/%s", id).Put(url.Values{
		"title": {"  Título actualizado  "}, "completed": {"false"},
		"id": {uuid.Must(uuid.NewV4()).String()}, "created_at": {"2000-01-01T00:00:00Z"},
	})
	as.Equal(303, res.Code)
	as.Equal("/", res.Header().Get("Location"))
	as.NoError(as.DB.Reload(task))
	as.Equal("Título actualizado", task.Title)
	as.Equal(id, task.ID)
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
		tasks = append(tasks, task)
	}
	return tasks
}

// assertListedInOrder comprueba que los títulos aparecen en la lista en ese orden.
func (as *ActionSuite) assertListedInOrder(titles ...string) {
	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	body := res.Body.String()
	last := -1
	for _, title := range titles {
		index := strings.Index(body, `<span class="task-title">`+title+`</span>`)
		as.Greater(index, last, title)
		last = index
	}
}

// moveTask envía el formulario de mover una tarea y devuelve el código de estado.
func (as *ActionSuite) moveTask(task *models.Task, position string) int {
	return as.HTML("/tasks/%s/move", task.ID).Put(url.Values{"position": {position}}).Code
}

func (as *ActionSuite) Test_TasksIndex_OrdersByPosition() {
	tasks := as.createTasks("Tarea alfa", "Tarea beta", "Tarea gamma")
	as.NoError(models.MoveTask(as.DB, tasks[2].ID, 1))
	as.assertListedInOrder("Tarea gamma", "Tarea alfa", "Tarea beta")
}

func (as *ActionSuite) Test_TasksIndex_RendersReorderControls() {
	tasks := as.createTasks("Tarea alfa", "Tarea beta")

	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	body := res.Body.String()

	as.Contains(body, "data-task-list")
	as.Contains(body, `data-task-id="`+tasks[0].ID.String()+`"`)
	as.Contains(body, `action="/tasks/`+tasks[1].ID.String()+`/move"`)
	as.Contains(body, `aria-label="Subir tarea: Tarea beta"`)
	as.Contains(body, `aria-label="Bajar tarea: Tarea alfa"`)
	as.Contains(body, `name="position" value="1"`)
	as.Contains(body, `name="position" value="3"`)
	as.Contains(body, "data-drag-handle")
	as.Contains(body, `role="status" data-reorder-status`)
}

func (as *ActionSuite) Test_TasksMove_PersistsOrder() {
	tasks := as.createTasks("Tarea alfa", "Tarea beta", "Tarea gamma")

	res := as.HTML("/tasks/%s/move", tasks[0].ID).Put(url.Values{"position": {"3"}})
	as.Equal(303, res.Code)
	as.Equal("/", res.Header().Get("Location"))
	as.assertListedInOrder("Tarea beta", "Tarea gamma", "Tarea alfa")

	as.Equal(303, as.moveTask(tasks[2], "1"))
	as.assertListedInOrder("Tarea gamma", "Tarea beta", "Tarea alfa")
}

func (as *ActionSuite) Test_TasksMove_OutOfRangeIsClamped() {
	tasks := as.createTasks("Tarea alfa", "Tarea beta")

	as.Equal(303, as.moveTask(tasks[0], "0"))
	as.assertListedInOrder("Tarea alfa", "Tarea beta")

	as.Equal(303, as.moveTask(tasks[1], "3"))
	as.assertListedInOrder("Tarea alfa", "Tarea beta")
}

func (as *ActionSuite) Test_TasksMove_InvalidPosition() {
	tasks := as.createTasks("Tarea alfa", "Tarea beta")

	for _, position := range []string{"", "segunda", "1.5"} {
		as.Equal(400, as.moveTask(tasks[1], position), position)
	}
	as.assertListedInOrder("Tarea alfa", "Tarea beta")
}

func (as *ActionSuite) Test_TasksMove_NotFound() {
	res := as.HTML("/tasks/%s/move", uuid.Must(uuid.NewV4())).Put(url.Values{"position": {"1"}})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksMove_InvalidID() {
	res := as.HTML("/tasks/no-es-un-uuid/move").Put(url.Values{"position": {"1"}})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksCreate_AppendsAtEndAfterMove() {
	tasks := as.createTasks("Tarea alfa", "Tarea beta")
	as.NoError(models.MoveTask(as.DB, tasks[1].ID, 1))

	res := as.HTML("/tasks").Post(&models.Task{Title: "Tarea gamma"})
	as.Equal(303, res.Code)
	as.assertListedInOrder("Tarea beta", "Tarea alfa", "Tarea gamma")
}

func (as *ActionSuite) Test_TasksToggle_KeepsPosition() {
	tasks := as.createTasks("Tarea alfa", "Tarea beta", "Tarea gamma")
	as.NoError(models.MoveTask(as.DB, tasks[2].ID, 1))

	for i := 0; i < 2; i++ {
		res := as.HTML("/tasks/%s/toggle", tasks[0].ID).Put(url.Values{})
		as.Equal(303, res.Code)
		as.assertListedInOrder("Tarea gamma", "Tarea alfa", "Tarea beta")
	}
}

func (as *ActionSuite) Test_TasksUpdate_KeepsPosition() {
	tasks := as.createTasks("Tarea alfa", "Tarea beta")
	as.NoError(models.MoveTask(as.DB, tasks[1].ID, 1))

	res := as.HTML("/tasks/%s", tasks[1].ID).Put(url.Values{"title": {"Tarea beta editada"}})
	as.Equal(303, res.Code)
	as.assertListedInOrder("Tarea beta editada", "Tarea alfa")
}
