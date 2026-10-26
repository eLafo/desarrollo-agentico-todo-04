package models

import "strings"

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

// createTasks crea las tareas con esos títulos, en ese orden.
func (ms *ModelSuite) createTasks(titles ...string) Tasks {
	tasks := Tasks{}
	for _, title := range titles {
		task := Task{Title: title}
		verrs, err := ms.DB.ValidateAndCreate(&task)
		ms.NoError(err)
		ms.False(verrs.HasAny())
		tasks = append(tasks, task)
	}
	return tasks
}

// titles devuelve los títulos de las tareas en el orden de la lista.
func (ms *ModelSuite) titles() []string {
	tasks, err := AllTasks(ms.DB)
	ms.NoError(err)
	titles := []string{}
	for _, task := range tasks {
		titles = append(titles, task.Title)
	}
	return titles
}

// positions devuelve las posiciones guardadas, en el orden de la lista.
func (ms *ModelSuite) positions() []int {
	tasks, err := AllTasks(ms.DB)
	ms.NoError(err)
	positions := []int{}
	for _, task := range tasks {
		positions = append(positions, task.Position)
	}
	return positions
}

func (ms *ModelSuite) Test_Task_Create_AppendsAtEnd() {
	ms.createTasks("A", "B", "C")

	ms.Equal([]string{"A", "B", "C"}, ms.titles())
	ms.Equal([]int{0, 1, 2}, ms.positions())
}

func (ms *ModelSuite) Test_Task_Create_AfterMove_AppendsAtEnd() {
	tasks := ms.createTasks("A", "B", "C")
	ms.NoError(tasks[2].MoveTo(ms.DB, 0))

	created := ms.createTasks("D")

	ms.Equal([]string{"C", "A", "B", "D"}, ms.titles())
	ms.Equal(3, created[0].Position)
}

func (ms *ModelSuite) Test_Task_MoveTo_First() {
	tasks := ms.createTasks("A", "B", "C")

	ms.NoError(tasks[2].MoveTo(ms.DB, 0))

	ms.Equal([]string{"C", "A", "B"}, ms.titles())
	ms.Equal([]int{0, 1, 2}, ms.positions())
}

func (ms *ModelSuite) Test_Task_MoveTo_Last() {
	tasks := ms.createTasks("A", "B", "C")

	ms.NoError(tasks[0].MoveTo(ms.DB, 2))

	ms.Equal([]string{"B", "C", "A"}, ms.titles())
	ms.Equal([]int{0, 1, 2}, ms.positions())
}

func (ms *ModelSuite) Test_Task_MoveTo_OutOfRange() {
	tasks := ms.createTasks("A", "B", "C")

	ms.NoError(tasks[0].MoveTo(ms.DB, -1))
	ms.Equal([]string{"A", "B", "C"}, ms.titles())

	ms.NoError(tasks[0].MoveTo(ms.DB, 10))
	ms.Equal([]string{"B", "C", "A"}, ms.titles())
	ms.Equal(2, tasks[0].Position)
}

func (ms *ModelSuite) Test_Task_MoveTo_CompletedTask() {
	tasks := ms.createTasks("A", "B", "C")
	tasks[2].Completed = true
	ms.NoError(ms.DB.Update(&tasks[2]))

	ms.NoError(tasks[2].MoveTo(ms.DB, 0))

	ms.Equal([]string{"C", "A", "B"}, ms.titles())
	reloaded := &Task{}
	ms.NoError(ms.DB.Find(reloaded, tasks[2].ID))
	ms.True(reloaded.Completed)
}

func (ms *ModelSuite) Test_AllTasks_SamePosition_ByCreation() {
	tasks := ms.createTasks("A", "B", "C")
	for i := range tasks {
		tasks[i].Position = 0
		ms.NoError(ms.DB.UpdateColumns(&tasks[i], "position"))
	}

	ms.Equal([]string{"A", "B", "C"}, ms.titles())
}
