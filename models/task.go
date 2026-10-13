package models

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// TitleMaxLength es el número máximo de caracteres permitido en el título.
const TitleMaxLength = 200

// TaskListOrder es el orden en que se muestra la lista: la posición que fija la
// persona y, a igualdad de posición, la fecha de creación y el identificador.
const TaskListOrder = "position asc, created_at asc, id asc"

// Task representa una tarea de la lista.
type Task struct {
	ID        uuid.UUID `json:"id" db:"id" form:"-"`
	Title     string    `json:"title" db:"title" form:"title"`
	Completed bool      `json:"completed" db:"completed" form:"-"`
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

// BeforeCreate coloca cada tarea nueva al final de la lista.
func (t *Task) BeforeCreate(tx *pop.Connection) error {
	last := &Task{}
	err := tx.Order("position desc").First(last)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	t.Position = last.Position + 1
	return nil
}

// MoveTask lleva la tarea id a la posición indicada de la lista (1 es la
// primera) y numera de nuevo la lista completa de 1 a N. Una posición fuera de
// rango se ajusta a la primera o a la última. Devuelve sql.ErrNoRows si la
// tarea no existe.
func MoveTask(tx *pop.Connection, id uuid.UUID, position int) error {
	tasks := Tasks{}
	if err := tx.Order(TaskListOrder).All(&tasks); err != nil {
		return err
	}

	from := slices.IndexFunc(tasks, func(t Task) bool { return t.ID == id })
	if from == -1 {
		return sql.ErrNoRows
	}

	moved := tasks[from]
	tasks = slices.Delete(tasks, from, from+1)
	to := min(max(position-1, 0), len(tasks))
	tasks = slices.Insert(tasks, to, moved)

	for i := range tasks {
		if tasks[i].Position == i+1 {
			continue
		}
		tasks[i].Position = i + 1
		if err := tx.UpdateColumns(&tasks[i], "position"); err != nil {
			return err
		}
	}
	return nil
}
