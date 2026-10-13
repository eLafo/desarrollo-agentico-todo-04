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

func (ms *ModelSuite) createTasks(titles ...string) []*Task {
	tasks := make([]*Task, 0, len(titles))
	for _, title := range titles {
		task := &Task{Title: title}
		verrs, err := ms.DB.ValidateAndCreate(task)
		ms.NoError(err)
		ms.False(verrs.HasAny())
		ms.NoError(ms.DB.Reload(task))
		tasks = append(tasks, task)
	}
	return tasks
}

func (ms *ModelSuite) positions(tasks ...*Task) []int {
	positions := make([]int, 0, len(tasks))
	for _, task := range tasks {
		ms.NoError(ms.DB.Reload(task))
		positions = append(positions, task.Position)
	}
	return positions
}

func (ms *ModelSuite) Test_Task_Create_AppendsAtEnd() {
	tasks := ms.createTasks("Primera", "Segunda", "Tercera")
	ms.Equal([]int{1, 2, 3}, ms.positions(tasks...))
}

func (ms *ModelSuite) Test_ReorderTasks_AssignsConsecutivePositions() {
	tasks := ms.createTasks("Primera", "Segunda", "Tercera")
	a, b, c := tasks[0], tasks[1], tasks[2]
	updatedAt := a.UpdatedAt

	ms.NoError(ReorderTasks(ms.DB, []uuid.UUID{c.ID, a.ID, b.ID}))
	ms.Equal([]int{2, 3, 1}, ms.positions(a, b, c))
	ms.True(a.UpdatedAt.Equal(updatedAt))

	next := ms.createTasks("Cuarta")[0]
	ms.Equal(4, next.Position)
}

func (ms *ModelSuite) Test_ReorderTasks_RejectsInvalidLists() {
	tasks := ms.createTasks("Primera", "Segunda")
	a, b := tasks[0], tasks[1]
	unknown := uuid.Must(uuid.NewV4())

	for _, ids := range [][]uuid.UUID{
		{b.ID},
		{b.ID, b.ID},
		{b.ID, unknown},
		{b.ID, a.ID, unknown},
	} {
		ms.ErrorIs(ReorderTasks(ms.DB, ids), ErrInvalidTaskOrder)
	}
	ms.Equal([]int{1, 2}, ms.positions(a, b))
}
