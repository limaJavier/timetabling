package model

import "timetabling/internal/sat"

type Timetabler interface {
	Build(
		modelInput ModelInput,
		curriculum [][]bool,
		groups map[uint64][]uint64,
		groupsGraph [][]bool,
	) ([][5]uint64, error)

	Verify(
		timetable [][5]uint64,
		modelInput ModelInput,
		curriculum [][]bool,
		groups map[uint64][]uint64,
		groupsGraph [][]bool,
	) bool
}

func NewTimetabler(solver sat.SATSolver) Timetabler {
	return newSatTimetabler(solver)
}
