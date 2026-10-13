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

// TaskListOrder es el orden de la lista: posición manual y, en caso de empate,
// fecha de creación e identificador.
const TaskListOrder = "position asc, created_at asc, id asc"

// ErrInvalidTaskOrder indica que el orden recibido no contiene exactamente
// una vez cada tarea existente.
var ErrInvalidTaskOrder = errors.New("el orden no coincide con las tareas existentes")

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

// BeforeCreate coloca cada tarea nueva al final del orden manual.
func (t *Task) BeforeCreate(tx *pop.Connection) error {
	last := &Task{}
	err := tx.Order("position desc").First(last)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	t.Position = last.Position + 1
	return nil
}

// ReorderTasks guarda el orden manual de la lista. ids debe contener exactamente
// una vez cada tarea existente; cada una recibe su posición empezando en 1.
func ReorderTasks(tx *pop.Connection, ids []uuid.UUID) error {
	tasks := Tasks{}
	if err := tx.All(&tasks); err != nil {
		return err
	}
	if len(ids) != len(tasks) {
		return ErrInvalidTaskOrder
	}

	current := make(map[uuid.UUID]int, len(tasks))
	for _, task := range tasks {
		current[task.ID] = task.Position
	}

	seen := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		if _, ok := current[id]; !ok || seen[id] {
			return ErrInvalidTaskOrder
		}
		seen[id] = true
	}

	for i, id := range ids {
		position := i + 1
		if current[id] == position {
			continue
		}
		if err := tx.RawQuery("UPDATE tasks SET position = ? WHERE id = ?", position, id).Exec(); err != nil {
			return err
		}
	}
	return nil
}
