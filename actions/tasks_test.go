package actions

import (
	"net/url"
	"strings"

	"todo/models"

	"github.com/gofrs/uuid"
)

func mark(v bool) *bool { return &v }

// quadrantHTML devuelve el fragmento que empieza en el cuadrante slug y acaba
// en el primer </section> siguiente, o "" si no encuentra el cuadrante.
func quadrantHTML(body, slug string) string {
	start := strings.Index(body, `id="quadrant-`+slug+`"`)
	if start < 0 {
		return ""
	}
	rest := body[start:]
	end := strings.Index(rest, "</section>")
	if end < 0 {
		return rest
	}
	return rest[:end]
}

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
	as.Nil(task.Urgent)
	as.Nil(task.Important)
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

func (as *ActionSuite) Test_TasksIndex_DimensionFields() {
	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	body := res.Body.String()
	as.Contains(body, `name="urgent"`)
	as.Contains(body, `name="important"`)
	as.Equal(2, strings.Count(body, `<option value="" selected>Sin estimar</option>`))
}

func (as *ActionSuite) Test_TasksCreate_WithDimensions() {
	res := as.HTML("/tasks").Post(url.Values{"title": {"Con dimensiones"}, "urgent": {"true"}, "important": {"false"}})
	as.Equal(303, res.Code)

	task := &models.Task{}
	as.NoError(as.DB.First(task))
	as.NotNil(task.Urgent)
	as.NotNil(task.Important)
	as.True(*task.Urgent)
	as.False(*task.Important)
}

func (as *ActionSuite) Test_TasksCreate_Invalid_KeepsDimensions() {
	res := as.HTML("/tasks").Post(url.Values{"title": {""}, "urgent": {"true"}, "important": {"false"}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), `<option value="true" selected>Urgente</option>`)
	as.Contains(res.Body.String(), `<option value="false" selected>No importante</option>`)

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksCreate_InvalidDimension() {
	res := as.HTML("/tasks").Post(url.Values{"title": {"Título válido"}, "urgent": {"quizá"}})
	as.Equal(400, res.Code)

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksEdit_PrefillsDimensions() {
	task := &models.Task{Title: "Con dimensiones", Urgent: mark(false), Important: mark(true)}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/tasks/%s/edit", task.ID).Get()
	as.Equal(200, res.Code)
	body := res.Body.String()
	as.Contains(body, `<option value="false" selected>No urgente</option>`)
	as.Contains(body, `<option value="true" selected>Importante</option>`)
	as.NotContains(body, `<option value="" selected>`)
}

func (as *ActionSuite) Test_TasksUpdate_PersistsAndClearsDimensions() {
	task := as.taskForEditing(true)

	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {"Título original"}, "urgent": {"true"}, "important": {"true"}})
	as.Equal(303, res.Code)
	as.NoError(as.DB.Reload(task))
	as.NotNil(task.Urgent)
	as.NotNil(task.Important)
	as.True(*task.Urgent)
	as.True(*task.Important)
	as.True(task.Completed)

	res = as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {"Título original"}, "urgent": {""}, "important": {""}})
	as.Equal(303, res.Code)
	as.NoError(as.DB.Reload(task))
	as.Nil(task.Urgent)
	as.Nil(task.Important)
	as.Equal("Título original", task.Title)
	as.True(task.Completed)
}

func (as *ActionSuite) Test_TasksUpdate_InvalidDimension() {
	task := as.taskForEditing(false)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {"Otro título"}, "important": {"2"}})
	as.Equal(400, res.Code)
	as.NoError(as.DB.Reload(task))
	as.Equal("Título original", task.Title)
	as.Nil(task.Urgent)
	as.Nil(task.Important)
}

func (as *ActionSuite) Test_TasksUpdate_Invalid_KeepsDimensions() {
	task := as.taskForEditing(false)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {""}, "urgent": {"false"}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), `<option value="false" selected>No urgente</option>`)
}

func (as *ActionSuite) Test_TasksEisenhower_HidesCompletedByDefault() {
	as.LoadFixture("matriz de eisenhower")

	res := as.HTML("/eisenhower").Get()
	as.Equal(200, res.Code)
	body := res.Body.String()

	as.Contains(quadrantHTML(body, "do"), "Resolver la incidencia del servidor")
	as.Contains(quadrantHTML(body, "schedule"), "Preparar el plan de formación")
	as.Contains(quadrantHTML(body, "delegate"), "Confirmar la reserva de la sala")
	as.Contains(quadrantHTML(body, "eliminate"), "Ordenar el archivo antiguo")

	as.NotContains(body, "Presentar el impuesto trimestral")
	as.NotContains(body, "Revisar los contratos")
	as.NotContains(body, "Actualizar el inventario")
	as.NotContains(body, "Archivar las facturas antiguas")

	as.Contains(body, "2 tareas no aparecen porque les falta estimar la urgencia o la importancia.")
	as.Contains(body, "Las tareas completadas están ocultas.")
	as.Contains(body, `href="/eisenhower?show_completed=1"`)
	for _, name := range []string{"Hacer", "Planificar", "Delegar", "Eliminar"} {
		as.Contains(body, name)
	}
}

func (as *ActionSuite) Test_TasksEisenhower_ShowCompleted() {
	as.LoadFixture("matriz de eisenhower")

	res := as.HTML("/eisenhower?show_completed=1").Get()
	as.Equal(200, res.Code)
	body := res.Body.String()

	do := quadrantHTML(body, "do")
	as.Contains(do, "Presentar el impuesto trimestral")
	as.Contains(do, "Resolver la incidencia del servidor")
	as.Contains(do, "Completada")

	as.Contains(body, "3 tareas no aparecen porque les falta estimar la urgencia o la importancia.")
	as.Contains(body, "Se muestran las tareas completadas.")
	as.Contains(body, `href="/eisenhower" class="button"`)
	as.NotContains(body, "Revisar los contratos")
	as.NotContains(body, "Actualizar el inventario")
	as.NotContains(body, "Archivar las facturas antiguas")
}

func (as *ActionSuite) Test_TasksEisenhower_SingularNotice() {
	task := &models.Task{Title: "Solo urgente", Urgent: mark(true)}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/eisenhower").Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "1 tarea no aparece porque le falta estimar la urgencia o la importancia.")
}

func (as *ActionSuite) Test_TasksEisenhower_EmptyMatrix() {
	res := as.HTML("/eisenhower").Get()
	as.Equal(200, res.Code)
	body := res.Body.String()
	as.NotContains(body, "no aparece")
	as.Equal(4, strings.Count(body, "Sin tareas en este cuadrante."))
}

func (as *ActionSuite) Test_TasksEisenhower_EscapesTitle() {
	task := &models.Task{Title: `<img src=x onerror="alert(1)">`, Urgent: mark(true), Important: mark(true)}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/eisenhower").Get()
	as.Equal(200, res.Code)
	as.NotContains(res.Body.String(), `<img src=x`)
	as.Contains(res.Body.String(), "&lt;img")
}

func (as *ActionSuite) Test_Navigation_MarksCurrentSection() {
	task := as.taskForEditing(false)

	body := as.HTML("/").Get().Body.String()
	as.Contains(body, `<a href="/" class="nav-link" aria-current="page">`)
	as.Contains(body, `<a href="/eisenhower" class="nav-link">`)

	body = as.HTML("/eisenhower").Get().Body.String()
	as.Contains(body, `<a href="/eisenhower" class="nav-link" aria-current="page">`)
	as.Contains(body, `<a href="/" class="nav-link">`)

	body = as.HTML("/tasks/%s/edit", task.ID).Get().Body.String()
	as.Contains(body, `<a href="/" class="nav-link" aria-current="page">`)
}
