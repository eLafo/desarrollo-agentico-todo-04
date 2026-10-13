package models

import "errors"

// ErrInvalidDimension indica que un valor de dimensión no es "", "true" ni "false".
var ErrInvalidDimension = errors.New("valor de dimensión no válido")

// ParseDimension convierte un valor de formulario en una dimensión de Eisenhower:
// "" la deja sin estimar (nil), "true" la marca y "false" la desmarca.
func ParseDimension(value string) (*bool, error) {
	switch value {
	case "":
		return nil, nil
	case "true":
		v := true
		return &v, nil
	case "false":
		v := false
		return &v, nil
	}
	return nil, ErrInvalidDimension
}

// DimensionValue es la inversa de ParseDimension: nil → "", true → "true", false → "false".
func DimensionValue(d *bool) string {
	if d == nil {
		return ""
	}
	if *d {
		return "true"
	}
	return "false"
}

// EisenhowerMatrix reparte tareas entre los cuatro cuadrantes de Eisenhower.
type EisenhowerMatrix struct {
	Do          Tasks // urgente e importante
	Schedule    Tasks // importante y no urgente
	Delegate    Tasks // urgente y no importante
	Eliminate   Tasks // ni urgente ni importante
	Unestimated int   // tareas a las que les falta alguna dimensión
}

// NewEisenhowerMatrix clasifica las tareas conservando su orden; las que no
// tienen las dos dimensiones estimadas solo se cuentan en Unestimated.
func NewEisenhowerMatrix(tasks Tasks) EisenhowerMatrix {
	m := EisenhowerMatrix{Do: Tasks{}, Schedule: Tasks{}, Delegate: Tasks{}, Eliminate: Tasks{}}
	for _, t := range tasks {
		if t.Urgent == nil || t.Important == nil {
			m.Unestimated++
			continue
		}
		switch {
		case *t.Urgent && *t.Important:
			m.Do = append(m.Do, t)
		case !*t.Urgent && *t.Important:
			m.Schedule = append(m.Schedule, t)
		case *t.Urgent && !*t.Important:
			m.Delegate = append(m.Delegate, t)
		default:
			m.Eliminate = append(m.Eliminate, t)
		}
	}
	return m
}
