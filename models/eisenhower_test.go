package models

func mark(v bool) *bool { return &v }

func (ms *ModelSuite) Test_ParseDimension() {
	v, err := ParseDimension("")
	ms.NoError(err)
	ms.Nil(v)

	v, err = ParseDimension("true")
	ms.NoError(err)
	ms.NotNil(v)
	ms.True(*v)

	v, err = ParseDimension("false")
	ms.NoError(err)
	ms.NotNil(v)
	ms.False(*v)

	for _, invalid := range []string{"sí", "1", "TRUE", " true"} {
		v, err = ParseDimension(invalid)
		ms.ErrorIs(err, ErrInvalidDimension, invalid)
		ms.Nil(v, invalid)
	}
}

func (ms *ModelSuite) Test_DimensionValue() {
	ms.Equal("", DimensionValue(nil))
	ms.Equal("true", DimensionValue(mark(true)))
	ms.Equal("false", DimensionValue(mark(false)))

	for _, value := range []string{"", "true", "false"} {
		d, err := ParseDimension(value)
		ms.NoError(err)
		ms.Equal(value, DimensionValue(d))
	}
}

func (ms *ModelSuite) Test_NewEisenhowerMatrix_Classifies() {
	states := []*bool{nil, mark(true), mark(false)}
	tasks := Tasks{}
	for i, u := range states {
		for j, im := range states {
			title := "t" + string(rune('0'+i)) + string(rune('0'+j))
			tasks = append(tasks, Task{Title: title, Urgent: u, Important: im})
		}
	}

	m := NewEisenhowerMatrix(tasks)
	ms.Len(m.Do, 1)
	ms.Len(m.Schedule, 1)
	ms.Len(m.Delegate, 1)
	ms.Len(m.Eliminate, 1)
	ms.Equal("t11", m.Do[0].Title)       // urgente e importante
	ms.Equal("t21", m.Schedule[0].Title) // no urgente e importante
	ms.Equal("t12", m.Delegate[0].Title) // urgente y no importante
	ms.Equal("t22", m.Eliminate[0].Title)
	ms.Equal(5, m.Unestimated)
}

func (ms *ModelSuite) Test_NewEisenhowerMatrix_KeepsOrder() {
	tasks := Tasks{
		{Title: "A", Urgent: mark(true), Important: mark(true)},
		{Title: "B", Urgent: mark(true), Important: mark(true)},
		{Title: "C", Urgent: mark(true), Important: mark(true)},
	}
	m := NewEisenhowerMatrix(tasks)
	ms.Len(m.Do, 3)
	ms.Equal("A", m.Do[0].Title)
	ms.Equal("B", m.Do[1].Title)
	ms.Equal("C", m.Do[2].Title)
}

func (ms *ModelSuite) Test_NewEisenhowerMatrix_Empty() {
	m := NewEisenhowerMatrix(Tasks{})
	ms.NotNil(m.Do)
	ms.NotNil(m.Schedule)
	ms.NotNil(m.Delegate)
	ms.NotNil(m.Eliminate)
	ms.Len(m.Do, 0)
	ms.Len(m.Schedule, 0)
	ms.Len(m.Delegate, 0)
	ms.Len(m.Eliminate, 0)
	ms.Equal(0, m.Unestimated)
}
