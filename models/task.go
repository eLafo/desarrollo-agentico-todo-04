package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// TitleMaxLength es el número máximo de caracteres permitido en el título.
const TitleMaxLength = 200

// Task representa una tarea de la lista.
type Task struct {
	ID        uuid.UUID `json:"id" db:"id" form:"-"`
	Title     string    `json:"title" db:"title" form:"title"`
	Completed bool      `json:"completed" db:"completed" form:"-"`
	// Position es el lugar de la tarea en la lista: sale antes la de menor posición.
	Position  int       `json:"position" db:"position" form:"-"`
	CreatedAt time.Time `json:"created_at" db:"created_at" form:"-"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" form:"-"`
}

// Tasks es una colección de Task.
type Tasks []Task

// Validate se ejecuta en cada llamada a los métodos "pop.Validate*".
func (t *Task) Validate(tx *pop.Connection) (*validate.Errors, error) {
	verrs := validate.NewErrors()

	title := strings.TrimSpace(t.Title)
	if title == "" {
		verrs.Add("title", "El título es obligatorio.")
	} else if len([]rune(title)) > TitleMaxLength {
		verrs.Add("title", fmt.Sprintf("El título no puede superar los %d caracteres.", TitleMaxLength))
	}

	return verrs, nil
}

// AllTasks devuelve todas las tareas en el orden de la lista: por posición y, a igual
// posición, de la más antigua a la más reciente.
func AllTasks(tx *pop.Connection) (Tasks, error) {
	tasks := Tasks{}
	if err := tx.Order("position asc, created_at asc, id asc").All(&tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

// BeforeCreate coloca la tarea nueva al final de la lista.
func (t *Task) BeforeCreate(tx *pop.Connection) error {
	last := Tasks{}
	if err := tx.Order("position desc").Limit(1).All(&last); err != nil {
		return err
	}
	t.Position = 0
	if len(last) > 0 {
		t.Position = last[0].Position + 1
	}
	return nil
}

// MoveTo lleva la tarea al lugar indicado de la lista (0 es el primero) y renumera todas sin
// cambiar el orden de las demás. Un lugar fuera de la lista se ajusta al primero o al último.
func (t *Task) MoveTo(tx *pop.Connection, position int) error {
	tasks, err := AllTasks(tx)
	if err != nil {
		return err
	}

	others := Tasks{}
	for _, task := range tasks {
		if task.ID != t.ID {
			others = append(others, task)
		}
	}
	position = max(0, min(position, len(others)))

	ordered := Tasks{}
	ordered = append(ordered, others[:position]...)
	ordered = append(ordered, *t)
	ordered = append(ordered, others[position:]...)

	for i := range ordered {
		if ordered[i].Position == i {
			continue
		}
		ordered[i].Position = i
		if err := tx.UpdateColumns(&ordered[i], "position"); err != nil {
			return err
		}
	}

	t.Position = position
	return nil
}
