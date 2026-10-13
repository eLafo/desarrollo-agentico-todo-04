package actions

import (
	"net/http"
	"strings"

	"todo/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
)

// taskForm recoge los campos que la persona usuaria fija al crear o editar una tarea.
type taskForm struct {
	Title     string `form:"title" json:"title"`
	Urgent    string `form:"urgent" json:"urgent"`
	Important string `form:"important" json:"important"`
}

// applyTo copia el formulario en la tarea: recorta el título y convierte las
// dimensiones; si alguna no es válida devuelve el error sin tocar la tarea.
func (f taskForm) applyTo(task *models.Task) error {
	urgent, err := models.ParseDimension(f.Urgent)
	if err != nil {
		return err
	}
	important, err := models.ParseDimension(f.Important)
	if err != nil {
		return err
	}
	task.Title = strings.TrimSpace(f.Title)
	task.Urgent = urgent
	task.Important = important
	return nil
}

// TasksIndex muestra la lista de tareas y el formulario de creación.
func TasksIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	tasks := models.Tasks{}
	if err := tx.Order("created_at asc, id asc").All(&tasks); err != nil {
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

	input := taskForm{}
	if err := c.Bind(&input); err != nil {
		return err
	}
	task := &models.Task{}
	if err := input.applyTo(task); err != nil {
		return c.Error(http.StatusBadRequest, err)
	}

	verrs, err := tx.ValidateAndCreate(task)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		tasks := models.Tasks{}
		if err := tx.Order("created_at asc, id asc").All(&tasks); err != nil {
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

// TasksUpdate actualiza el título y las dimensiones de Eisenhower, conservando los demás datos de la tarea.
func TasksUpdate(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	input := taskForm{}
	if err := c.Bind(&input); err != nil {
		return err
	}
	if err := input.applyTo(task); err != nil {
		return c.Error(http.StatusBadRequest, err)
	}

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

// TasksEisenhower muestra en los cuatro cuadrantes de Eisenhower las tareas con
// las dos dimensiones estimadas y cuenta las que no aparecen por falta de
// alguna. Las completadas solo se incluyen con show_completed=1.
func TasksEisenhower(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	showCompleted := c.Param("show_completed") == "1"

	query := tx.Order("created_at asc, id asc")
	if !showCompleted {
		query = query.Where("completed = ?", false)
	}
	tasks := models.Tasks{}
	if err := query.All(&tasks); err != nil {
		return err
	}

	c.Set("matrix", models.NewEisenhowerMatrix(tasks))
	c.Set("show_completed", showCompleted)
	return c.Render(http.StatusOK, r.HTML("tasks/eisenhower.plush.html"))
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
