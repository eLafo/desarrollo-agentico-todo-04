package models

import (
	"strings"

	"github.com/gofrs/uuid"
)

func (ms *ModelSuite) Test_Task_Create_Valid() {
	task := &Task{Title: "Comprar café"}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	ms.NotZero(task.ID)
	ms.False(task.Completed)

	count, err := ms.DB.Count(&Task{})
	ms.NoError(err)
	ms.Equal(1, count)
}

func (ms *ModelSuite) Test_Task_Title_Required() {
	task := &Task{Title: ""}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.True(verrs.HasAny())
	ms.Contains(verrs.Get("title"), "El título es obligatorio.")

	count, err := ms.DB.Count(&Task{})
	ms.NoError(err)
	ms.Equal(0, count)
}

func (ms *ModelSuite) Test_Task_Title_OnlySpaces_Invalid() {
	task := &Task{Title: "   \t  "}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.True(verrs.HasAny())
	ms.Contains(verrs.Get("title"), "El título es obligatorio.")
}

func (ms *ModelSuite) Test_Task_Title_MaxLength() {
	task := &Task{Title: strings.Repeat("a", TitleMaxLength)}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.False(verrs.HasAny())
}

func (ms *ModelSuite) Test_Task_Title_TooLong_Invalid() {
	task := &Task{Title: strings.Repeat("a", TitleMaxLength+1)}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.True(verrs.HasAny())
	ms.NotEmpty(verrs.Get("title"))
}

func (ms *ModelSuite) Test_Task_Toggle_Persists() {
	task := &Task{Title: "Revisar el correo"}
	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	task.Completed = true
	ms.NoError(ms.DB.Update(task))

	reloaded := &Task{}
	ms.NoError(ms.DB.Find(reloaded, task.ID))
	ms.True(reloaded.Completed)
}

func (ms *ModelSuite) createTask(title string, completed bool) *Task {
	task := &Task{Title: title, Completed: completed}
	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.False(verrs.HasAny())
	ms.NoError(ms.DB.Reload(task))
	return task
}

func (ms *ModelSuite) positionOf(task *Task) int {
	reloaded := &Task{}
	ms.NoError(ms.DB.Find(reloaded, task.ID))
	return reloaded.Position
}

func (ms *ModelSuite) Test_Task_Create_AppendsAtEnd() {
	first := ms.createTask("Primera", false)
	second := ms.createTask("Segunda", false)
	third := ms.createTask("Tercera", false)

	ms.Equal(1, first.Position)
	ms.Equal(2, second.Position)
	ms.Equal(3, third.Position)

	ms.NoError(ms.DB.Destroy(second))
	fourth := ms.createTask("Cuarta", false)
	ms.Equal(4, fourth.Position)
}

func (ms *ModelSuite) Test_Task_Create_KeepsExplicitPosition() {
	task := &Task{Title: "Con posición", Position: 7}
	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	ms.Equal(7, ms.positionOf(task))
}

func (ms *ModelSuite) Test_ReorderTasks_AssignsPositionsInGivenOrder() {
	a := ms.createTask("A", true)
	b := ms.createTask("B", false)
	c := ms.createTask("C", false)

	ms.NoError(ReorderTasks(ms.DB, []uuid.UUID{c.ID, a.ID, b.ID}))

	ms.Equal(1, ms.positionOf(c))
	ms.Equal(2, ms.positionOf(a))
	ms.Equal(3, ms.positionOf(b))

	reloaded := &Task{}
	ms.NoError(ms.DB.Find(reloaded, a.ID))
	ms.Equal(a.Title, reloaded.Title)
	ms.True(reloaded.Completed)
	ms.True(reloaded.UpdatedAt.Equal(a.UpdatedAt))

	tasks := Tasks{}
	ms.NoError(ms.DB.Order(TaskListOrder).All(&tasks))
	ms.Len(tasks, 3)
	ms.Equal(c.ID, tasks[0].ID)
	ms.Equal(a.ID, tasks[1].ID)
	ms.Equal(b.ID, tasks[2].ID)
}

func (ms *ModelSuite) Test_ReorderTasks_RejectsMismatch() {
	a := ms.createTask("A", false)
	b := ms.createTask("B", false)
	c := ms.createTask("C", false)

	cases := []struct {
		name string
		ids  []uuid.UUID
	}{
		{"lista vacía", []uuid.UUID{}},
		{"falta una tarea", []uuid.UUID{a.ID, b.ID}},
		{"ID duplicado", []uuid.UUID{a.ID, a.ID, b.ID}},
		{"UUID desconocido", []uuid.UUID{a.ID, b.ID, uuid.Must(uuid.NewV4())}},
	}

	for _, tc := range cases {
		err := ReorderTasks(ms.DB, tc.ids)
		ms.ErrorIs(err, ErrTaskOrderMismatch, tc.name)

		ms.Equal(1, ms.positionOf(a), tc.name)
		ms.Equal(2, ms.positionOf(b), tc.name)
		ms.Equal(3, ms.positionOf(c), tc.name)
	}
}
