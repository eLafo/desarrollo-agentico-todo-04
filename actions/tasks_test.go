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
