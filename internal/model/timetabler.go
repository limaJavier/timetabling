package model

import "timetabling/internal/sat"

type Timetabler interface {
	Build() sat.SATSolution
}

func NewTimetabler() Timetabler {
	panic("not implemented")
}
