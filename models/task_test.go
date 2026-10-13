package models

import (
	"database/sql"
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
		tasks = append(tasks, task)
	}
	return tasks
}

func (ms *ModelSuite) listedTitles() []string {
	tasks := Tasks{}
	ms.NoError(ms.DB.Order(TaskListOrder).All(&tasks))
	titles := make([]string, 0, len(tasks))
	for _, task := range tasks {
		titles = append(titles, task.Title)
	}
	return titles
}

func (ms *ModelSuite) Test_Task_Create_AppendsAtEnd() {
	tasks := ms.createTasks("Primera", "Segunda", "Tercera")
	for i, task := range tasks {
		ms.NoError(ms.DB.Reload(task))
		ms.Equal(i+1, task.Position)
	}
}

func (ms *ModelSuite) Test_MoveTask_Down() {
	tasks := ms.createTasks("Primera", "Segunda", "Tercera")
	ms.NoError(MoveTask(ms.DB, tasks[0].ID, 3))
	ms.Equal([]string{"Segunda", "Tercera", "Primera"}, ms.listedTitles())
}

func (ms *ModelSuite) Test_MoveTask_Up() {
	tasks := ms.createTasks("Primera", "Segunda", "Tercera")
	ms.NoError(MoveTask(ms.DB, tasks[2].ID, 1))
	ms.Equal([]string{"Tercera", "Primera", "Segunda"}, ms.listedTitles())

	positions := []int{}
	for _, task := range tasks {
		ms.NoError(ms.DB.Reload(task))
		positions = append(positions, task.Position)
	}
	ms.Equal([]int{2, 3, 1}, positions)
}

func (ms *ModelSuite) Test_MoveTask_ClampsOutOfRange() {
	tasks := ms.createTasks("Primera", "Segunda", "Tercera")
	ms.NoError(MoveTask(ms.DB, tasks[1].ID, 0))
	ms.Equal([]string{"Segunda", "Primera", "Tercera"}, ms.listedTitles())
	ms.NoError(MoveTask(ms.DB, tasks[1].ID, 99))
	ms.Equal([]string{"Primera", "Tercera", "Segunda"}, ms.listedTitles())
}

func (ms *ModelSuite) Test_MoveTask_KeepsOtherFields() {
	tasks := ms.createTasks("Primera", "Segunda")
	task := tasks[1]
	task.Completed = true
	ms.NoError(ms.DB.Update(task))
	ms.NoError(ms.DB.Reload(task))
	updatedAt := task.UpdatedAt

	ms.NoError(MoveTask(ms.DB, task.ID, 1))
	ms.NoError(ms.DB.Reload(task))
	ms.Equal(1, task.Position)
	ms.Equal("Segunda", task.Title)
	ms.True(task.Completed)
	ms.True(task.UpdatedAt.Equal(updatedAt))
}

func (ms *ModelSuite) Test_MoveTask_NotFound() {
	ms.createTasks("Primera")
	ms.ErrorIs(MoveTask(ms.DB, uuid.Must(uuid.NewV4()), 1), sql.ErrNoRows)
}

func (ms *ModelSuite) Test_Task_Create_AfterMove_AppendsAtEnd() {
	tasks := ms.createTasks("Primera", "Segunda")
	ms.NoError(MoveTask(ms.DB, tasks[1].ID, 1))
	ms.createTasks("Tercera")
	ms.Equal([]string{"Segunda", "Primera", "Tercera"}, ms.listedTitles())
}

func (ms *ModelSuite) Test_Task_Toggle_KeepsPosition() {
	tasks := ms.createTasks("Primera", "Segunda", "Tercera")
	ms.NoError(MoveTask(ms.DB, tasks[2].ID, 1))
	task := tasks[0]
	ms.NoError(ms.DB.Reload(task))
	task.Completed = true
	ms.NoError(ms.DB.Update(task))
	ms.Equal([]string{"Tercera", "Primera", "Segunda"}, ms.listedTitles())
}
