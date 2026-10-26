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

// assertPageTitle comprueba que la página tiene una sola etiqueta <title> y que dice exactamente «Mis tareas».
func (as *ActionSuite) assertPageTitle(body string) {
	as.Equal(1, strings.Count(body, "<title>"))
	as.Contains(body, "<title>Mis tareas</title>")
}

func (as *ActionSuite) Test_TasksIndex_PageTitle() {
	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.assertPageTitle(res.Body.String())
}

func (as *ActionSuite) Test_TasksIndex_Heading() {
	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "<h1>Tareas</h1>")
}

func (as *ActionSuite) Test_TasksCreate_Invalid_PageTitle() {
	res := as.HTML("/tasks").Post(&models.Task{Title: ""})
	as.Equal(422, res.Code)
	as.assertPageTitle(res.Body.String())
}

func (as *ActionSuite) Test_TasksEdit_PageTitle() {
	task := as.taskForEditing(false)
	res := as.HTML("/tasks/%s/edit", task.ID).Get()
	as.Equal(200, res.Code)
	as.assertPageTitle(res.Body.String())
}

func (as *ActionSuite) Test_TasksUpdate_Invalid_PageTitle() {
	task := as.taskForEditing(false)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {""}})
	as.Equal(422, res.Code)
	as.assertPageTitle(res.Body.String())
}

// createTasks crea las tareas con esos títulos, en ese orden.
func (as *ActionSuite) createTasks(titles ...string) []*models.Task {
	tasks := []*models.Task{}
	for _, title := range titles {
		task := &models.Task{Title: title}
		verrs, err := as.DB.ValidateAndCreate(task)
		as.NoError(err)
		as.False(verrs.HasAny())
		tasks = append(tasks, task)
	}
	return tasks
}

// taskTitles devuelve los títulos guardados, en el orden de la lista.
func (as *ActionSuite) taskTitles() []string {
	tasks, err := models.AllTasks(as.DB)
	as.NoError(err)
	titles := []string{}
	for _, task := range tasks {
		titles = append(titles, task.Title)
	}
	return titles
}

// assertListOrder comprueba que la página muestra los títulos en ese orden.
func (as *ActionSuite) assertListOrder(body string, titles ...string) {
	last := -1
	for _, title := range titles {
		index := strings.Index(body, `<span class="task-title">`+title+`</span>`)
		as.NotEqual(-1, index, "falta la tarea %q", title)
		as.Greater(index, last, "la tarea %q sale antes de lo esperado", title)
		last = index
	}
}

func (as *ActionSuite) Test_TasksIndex_ShowsSavedOrder() {
	tasks := as.createTasks("A", "B", "C")
	as.NoError(tasks[2].MoveTo(as.DB, 0))

	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.assertListOrder(res.Body.String(), "C", "A", "B")
}

func (as *ActionSuite) Test_TasksIndex_MoveButtons() {
	tasks := as.createTasks("A", "B", "C")

	res := as.HTML("/").Get()
	body := res.Body.String()
	for _, task := range tasks {
		as.Contains(body, `action="/tasks/`+task.ID.String()+`/move"`)
		as.Contains(body, "Subir tarea: "+task.Title)
		as.Contains(body, "Bajar tarea: "+task.Title)
	}

	start := strings.Index(body, `action="/tasks/`+tasks[1].ID.String()+`/move"`)
	end := start + strings.Index(body[start:], "</form>")
	form := body[start:end]
	as.Contains(form, `name="position" value="0"`)
	as.Contains(form, `name="position" value="2"`)
}

func (as *ActionSuite) Test_TasksMove_Persists() {
	tasks := as.createTasks("A", "B", "C")

	res := as.HTML("/tasks/%s/move", tasks[0].ID).Put(url.Values{"position": {"2"}})
	as.Equal(303, res.Code)
	as.Equal("/", res.Header().Get("Location"))

	as.Equal([]string{"B", "C", "A"}, as.taskTitles())
}

func (as *ActionSuite) Test_TasksMove_CompletedTask() {
	tasks := as.createTasks("A", "B", "C")
	tasks[0].Completed = true
	as.NoError(as.DB.Update(tasks[0]))

	res := as.HTML("/tasks/%s/move", tasks[0].ID).Put(url.Values{"position": {"2"}})
	as.Equal(303, res.Code)

	as.Equal([]string{"B", "C", "A"}, as.taskTitles())
	as.NoError(as.DB.Reload(tasks[0]))
	as.True(tasks[0].Completed)
}

func (as *ActionSuite) Test_TasksMove_OutOfRange() {
	tasks := as.createTasks("A", "B", "C")

	res := as.HTML("/tasks/%s/move", tasks[0].ID).Put(url.Values{"position": {"-1"}})
	as.Equal(303, res.Code)
	as.Equal([]string{"A", "B", "C"}, as.taskTitles())

	res = as.HTML("/tasks/%s/move", tasks[0].ID).Put(url.Values{"position": {"99"}})
	as.Equal(303, res.Code)
	as.Equal([]string{"B", "C", "A"}, as.taskTitles())
}

func (as *ActionSuite) Test_TasksMove_InvalidPosition() {
	tasks := as.createTasks("A", "B", "C")

	res := as.HTML("/tasks/%s/move", tasks[0].ID).Put(url.Values{"position": {"abc"}})
	as.Equal(422, res.Code)

	res = as.HTML("/tasks/%s/move", tasks[0].ID).Put(url.Values{})
	as.Equal(422, res.Code)

	as.Equal([]string{"A", "B", "C"}, as.taskTitles())
}

func (as *ActionSuite) Test_TasksMove_NotFound() {
	res := as.HTML("/tasks/%s/move", uuid.Must(uuid.NewV4())).Put(url.Values{"position": {"0"}})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksMove_InvalidID() {
	res := as.HTML("/tasks/no-es-un-uuid/move").Put(url.Values{"position": {"0"}})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksCreate_AppendsAtEnd() {
	tasks := as.createTasks("A", "B", "C")
	as.NoError(tasks[2].MoveTo(as.DB, 0))

	res := as.HTML("/tasks").Post(&models.Task{Title: "D"})
	as.Equal(303, res.Code)
	as.Equal([]string{"C", "A", "B", "D"}, as.taskTitles())

	res = as.HTML("/").Get()
	as.assertListOrder(res.Body.String(), "C", "A", "B", "D")
}

func (as *ActionSuite) Test_TasksToggle_KeepsPosition() {
	tasks := as.createTasks("A", "B", "C")
	as.NoError(tasks[2].MoveTo(as.DB, 0))

	res := as.HTML("/tasks/%s/toggle", tasks[0].ID).Put(url.Values{})
	as.Equal(303, res.Code)
	as.Equal([]string{"C", "A", "B"}, as.taskTitles())

	res = as.HTML("/tasks/%s/toggle", tasks[0].ID).Put(url.Values{})
	as.Equal(303, res.Code)
	as.Equal([]string{"C", "A", "B"}, as.taskTitles())

	res = as.HTML("/").Get()
	as.assertListOrder(res.Body.String(), "C", "A", "B")
}
