package model

import "timetabling/internal/sat"

type Timetabler interface {
	Build(curriculum [][]uint64, availability map[uint64][][]bool, rooms map[uint64]uint64, professors map[uint64]uint64) ([][5]uint64, error)
	Verify(timetabler [][5]uint64, curriculum [][]uint64, availability map[uint64][][]bool, rooms map[uint64]uint64, professors map[uint64]uint64) bool
}

func NewTimetabler(solver sat.SATSolver) Timetabler {
	return newSatTimetabler(solver)
}
