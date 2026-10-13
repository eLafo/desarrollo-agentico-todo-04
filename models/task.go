package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// TitleMaxLength es el número máximo de caracteres permitido en el título.
const TitleMaxLength = 200

// TaskListOrder es el orden de la lista: la posición elegida por la persona
// usuaria y, si hay empate, la antigüedad.
const TaskListOrder = "position asc, created_at asc, id asc"

// ErrTaskOrderMismatch indica que un nuevo orden no incluye cada tarea exactamente una vez.
var ErrTaskOrderMismatch = errors.New("el nuevo orden debe incluir cada tarea exactamente una vez")

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
	if t.Position > 0 {
		return nil
	}
	last := &Task{}
	err := tx.Order("position desc").First(last)
	if errors.Is(err, sql.ErrNoRows) {
		t.Position = 1
		return nil
	}
	if err != nil {
		return err
	}
	t.Position = last.Position + 1
	return nil
}

// ReorderTasks guarda un nuevo orden: ids[i] pasa a la posición i+1; si ids no
// contiene exactamente una vez cada tarea existente, devuelve ErrTaskOrderMismatch
// sin cambiar nada.
func ReorderTasks(tx *pop.Connection, ids []uuid.UUID) error {
	tasks := Tasks{}
	if err := tx.All(&tasks); err != nil {
		return err
	}
	if len(ids) != len(tasks) {
		return ErrTaskOrderMismatch
	}
	existing := make(map[uuid.UUID]bool, len(tasks))
	for _, task := range tasks {
		existing[task.ID] = true
	}
	seen := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		if !existing[id] || seen[id] {
			return ErrTaskOrderMismatch
		}
		seen[id] = true
	}
	for i, id := range ids {
		if err := tx.RawQuery("UPDATE tasks SET position = ? WHERE id = ?", i+1, id).Exec(); err != nil {
			return err
		}
	}
	return nil
}
