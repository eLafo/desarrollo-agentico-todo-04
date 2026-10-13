package actions

import (
	"fmt"
	"net/http"
	"strings"

	"todo/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// orderedTasks devuelve todas las tareas en el orden definido por el usuario.
// created_at e id desempatan de forma determinista.
func orderedTasks(tx *pop.Connection) (models.Tasks, error) {
	tasks := models.Tasks{}
	if err := tx.Order("position asc, created_at asc, id asc").All(&tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

// TasksIndex muestra la lista de tareas y el formulario de creación.
func TasksIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	tasks, err := orderedTasks(tx)
	if err != nil {
		return err
	}

	c.Set("tasks", tasks)
	c.Set("task", &models.Task{})
	c.Set("errors", validate.NewErrors())
	return c.Render(http.StatusOK, r.HTML("tasks/index.plush.html"))
}

// TasksCreate crea una tarea nueva a partir del formulario.
func TasksCreate(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := c.Bind(task); err != nil {
		return err
	}
	task.Title = strings.TrimSpace(task.Title)

	var maxPosition int
	if err := tx.RawQuery("SELECT COALESCE(MAX(position), 0) FROM tasks").First(&maxPosition); err != nil {
		return err
	}
	task.Position = maxPosition + 1

	verrs, err := tx.ValidateAndCreate(task)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		tasks, err := orderedTasks(tx)
		if err != nil {
			return err
		}

		c.Set("tasks", tasks)
		c.Set("task", task)
		c.Set("errors", verrs)
		return c.Render(http.StatusUnprocessableEntity, r.HTML("tasks/index.plush.html"))
	}

	c.Flash().Add("success", "Tarea creada.")
	return c.Redirect(http.StatusSeeOther, "/")
}

// TasksEdit muestra el formulario de edición de una tarea existente.
func TasksEdit(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	c.Set("task", task)
	c.Set("errors", validate.NewErrors())
	return c.Render(http.StatusOK, r.HTML("tasks/edit.plush.html"))
}

// TasksUpdate actualiza solo el título, conservando los demás datos de la tarea.
func TasksUpdate(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	input := struct {
		Title string `form:"title" json:"title"`
	}{}
	if err := c.Bind(&input); err != nil {
		return err
	}
	task.Title = strings.TrimSpace(input.Title)

	verrs, err := tx.ValidateAndUpdate(task)
	if err != nil {
		return err
	}
	if verrs.HasAny() {
		c.Set("task", task)
		c.Set("errors", verrs)
		return c.Render(http.StatusUnprocessableEntity, r.HTML("tasks/edit.plush.html"))
	}

	c.Flash().Add("success", "Tarea actualizada.")
	return c.Redirect(http.StatusSeeOther, "/")
}

// TasksToggle alterna una tarea entre completada y pendiente.
func TasksToggle(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	task.Completed = !task.Completed
	if err := tx.Update(task); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/")
}

// TasksDestroy elimina una tarea.
func TasksDestroy(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := tx.Destroy(task); err != nil {
		return err
	}

	c.Flash().Add("success", "Tarea eliminada.")
	return c.Redirect(http.StatusSeeOther, "/")
}

// TasksReorder guarda el orden completo de las tareas recibido como una lista
// repetida de task_ids. No toca updated_at: reordenar no edita el contenido.
func TasksReorder(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	if err := c.Request().ParseForm(); err != nil {
		return c.Error(http.StatusUnprocessableEntity, err)
	}
	rawIDs := c.Request().Form["task_ids"]
	if len(rawIDs) == 0 {
		return c.Error(http.StatusUnprocessableEntity, fmt.Errorf("task_ids es obligatorio"))
	}

	ids := make([]uuid.UUID, 0, len(rawIDs))
	seen := make(map[uuid.UUID]bool, len(rawIDs))
	for _, raw := range rawIDs {
		id, err := uuid.FromString(raw)
		if err != nil {
			return c.Error(http.StatusUnprocessableEntity, fmt.Errorf("identificador no válido: %q", raw))
		}
		if seen[id] {
			return c.Error(http.StatusUnprocessableEntity, fmt.Errorf("identificador duplicado: %s", id))
		}
		seen[id] = true
		ids = append(ids, id)
	}

	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	count, err := tx.Where("id in (?)", args...).Count(&models.Task{})
	if err != nil {
		return err
	}
	if count != len(ids) {
		return c.Error(http.StatusUnprocessableEntity, fmt.Errorf("alguna tarea no existe"))
	}

	for i, id := range ids {
		if err := tx.RawQuery("UPDATE tasks SET position = ? WHERE id = ?", i+1, id).Exec(); err != nil {
			return err
		}
	}

	c.Response().WriteHeader(http.StatusNoContent)
	return nil
}
