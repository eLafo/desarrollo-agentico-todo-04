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

func (as *ActionSuite) createTask(title string, position int, completed bool) *models.Task {
	task := &models.Task{Title: title, Position: position, Completed: completed}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())
	return task
}

func (as *ActionSuite) positions(tasks ...*models.Task) []int {
	result := make([]int, len(tasks))
	for i, task := range tasks {
		as.NoError(as.DB.Reload(task))
		result[i] = task.Position
	}
	return result
}

func (as *ActionSuite) Test_TasksIndex_OrdersByPosition() {
	as.createTask("Tercera", 3, false)
	as.createTask("Primera", 1, false)
	as.createTask("Segunda", 2, false)

	body := as.HTML("/").Get().Body.String()
	first, second, third := strings.Index(body, "Primera"), strings.Index(body, "Segunda"), strings.Index(body, "Tercera")
	as.True(first >= 0 && first < second && second < third)
}

func (as *ActionSuite) Test_TasksIndex_TieBreaksByCreatedAt() {
	as.createTask("Antigua", 1, false)
	as.createTask("Reciente", 1, false)

	body := as.HTML("/").Get().Body.String()
	as.True(strings.Index(body, "Antigua") < strings.Index(body, "Reciente"))
}

func (as *ActionSuite) Test_TasksIndex_RendersDragAttributes() {
	task := as.createTask("Arrastrable", 1, false)

	body := as.HTML("/").Get().Body.String()
	as.Contains(body, `draggable="true"`)
	as.Contains(body, `data-task-id="`+task.ID.String()+`"`)
	as.Contains(body, `data-reorder-url="/tasks/order"`)
}

func (as *ActionSuite) Test_TasksCreate_AppendsAtEnd() {
	res := as.HTML("/tasks").Post(&models.Task{Title: "Primera de la lista"})
	as.Equal(303, res.Code)
	first := &models.Task{}
	as.NoError(as.DB.Where("title = ?", "Primera de la lista").First(first))
	as.Equal(1, first.Position)

	as.createTask("Colocada en la 5", 5, false)
	res = as.HTML("/tasks").Post(&models.Task{Title: "Al final"})
	as.Equal(303, res.Code)
	last := &models.Task{}
	as.NoError(as.DB.Where("title = ?", "Al final").First(last))
	as.Equal(6, last.Position)
}

func (as *ActionSuite) Test_TasksReorder_Persists() {
	a := as.createTask("A", 1, false)
	b := as.createTask("B", 2, false)
	c := as.createTask("C", 3, false)

	res := as.HTML("/tasks/order").Put(url.Values{"task_ids": {c.ID.String(), a.ID.String(), b.ID.String()}})
	as.Equal(204, res.Code)

	as.Equal([]int{2, 3, 1}, as.positions(a, b, c))

	body := as.HTML("/").Get().Body.String()
	as.True(strings.Index(body, ">C<") < strings.Index(body, ">A<"))
	as.True(strings.Index(body, ">A<") < strings.Index(body, ">B<"))
}

func (as *ActionSuite) Test_TasksReorder_DoesNotChangeCompleted() {
	a := as.createTask("A", 1, false)
	b := as.createTask("B", 2, true)
	as.NoError(as.DB.Reload(b))
	title, createdAt := b.Title, b.CreatedAt

	res := as.HTML("/tasks/order").Put(url.Values{"task_ids": {b.ID.String(), a.ID.String()}})
	as.Equal(204, res.Code)

	as.NoError(as.DB.Reload(b))
	as.True(b.Completed)
	as.Equal(title, b.Title)
	as.True(b.CreatedAt.Equal(createdAt))
	as.Equal(1, b.Position)
}

func (as *ActionSuite) Test_TasksReorder_InvalidID() {
	a := as.createTask("A", 1, false)
	b := as.createTask("B", 2, false)

	res := as.HTML("/tasks/order").Put(url.Values{"task_ids": {b.ID.String(), "no-es-un-uuid"}})
	as.Equal(422, res.Code)
	as.Equal([]int{1, 2}, as.positions(a, b))
}

func (as *ActionSuite) Test_TasksReorder_UnknownID() {
	a := as.createTask("A", 1, false)
	b := as.createTask("B", 2, false)

	res := as.HTML("/tasks/order").Put(url.Values{"task_ids": {b.ID.String(), uuid.Must(uuid.NewV4()).String(), a.ID.String()}})
	as.Equal(422, res.Code)
	as.Equal([]int{1, 2}, as.positions(a, b))
}

func (as *ActionSuite) Test_TasksReorder_DuplicateIDs() {
	a := as.createTask("A", 1, false)
	b := as.createTask("B", 2, false)

	res := as.HTML("/tasks/order").Put(url.Values{"task_ids": {b.ID.String(), b.ID.String(), a.ID.String()}})
	as.Equal(422, res.Code)
	as.Equal([]int{1, 2}, as.positions(a, b))
}

func (as *ActionSuite) Test_TasksReorder_Empty() {
	a := as.createTask("A", 1, false)

	res := as.HTML("/tasks/order").Put(url.Values{})
	as.Equal(422, res.Code)
	as.Equal([]int{1}, as.positions(a))
}
