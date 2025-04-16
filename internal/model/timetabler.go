package model

import "timetabling/internal/sat"

type Timetabler interface {
	Build(
		modelInput ModelInput,
		curriculum [][]bool,
		groups map[uint64][]uint64,
		groupsGraph [][]bool,
	) ([][6]uint64, error)

	Verify(
		timetable [][6]uint64,
		modelInput ModelInput,
		curriculum [][]bool,
		groups map[uint64][]uint64,
		groupsGraph [][]bool,
	) bool
}

func NewTimetabler(solver sat.SATSolver) Timetabler {
	return newSatTimetabler(solver)
}
