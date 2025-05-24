package model

import (
	"github.com/limaJavier/timetabling/pkg/sat"

	"github.com/samber/lo"
)

type isolatedRoomTimetabler struct {
	solver                  sat.SATSolver
	hybrid                  bool
	roomSimilarityThreshold float32
}

func NewIsolatedRoomTimetabler(solver sat.SATSolver, hybrid bool, roomSimilarityThreshold float32) Timetabler {
	return &isolatedRoomTimetabler{
		solver:                  solver,
		hybrid:                  hybrid,
		roomSimilarityThreshold: roomSimilarityThreshold,
	}
}

func (timetabler *isolatedRoomTimetabler) Build(modelInput ModelInput) (timetable [][6]uint64, variables uint64, clauses uint64, err error) {
	//** Extract attributes's domains
	totalRooms := uint64(1)
	totalPeriods, totalDays, totalLessons, totalSubjectProfessors, totalGroups, _ := getAttributes(modelInput)

	//** Initialize dependencies
	isolatedEvaluator := newPredicateEvaluatorIsolatedRoom(modelInput, timetabler.roomSimilarityThreshold)
	standardEvaluator := newPredicateEvaluator(modelInput, timetabler.roomSimilarityThreshold)
	indexer := newIndexer(totalPeriods, totalDays, totalLessons, totalSubjectProfessors, totalGroups, totalRooms)
	generator := newPermutationGenerator(totalPeriods, totalDays, totalLessons, totalSubjectProfessors, totalGroups, totalRooms)

	//** Build SAT instance
	variables = totalPeriods * totalDays * totalLessons * totalSubjectProfessors * totalGroups * totalRooms

	// Constraints functions
	constraints := []func(state constraintState) [][]int64{
		professorConstraints,
		studentConstraints,
		subjectPermissibilityConstraints,
		professorAvailabilityConstraints,
		lessonConstraints,
		completenessConstraints,
		negationConstraints,
		uniquenessConstraints,
	}
	if timetabler.hybrid {
		constraints = append(constraints, roomSimilarityConstraints)
	}

	state := constraintState{
		evaluator:         isolatedEvaluator,
		indexer:           indexer,
		generator:         generator,
		periods:           totalPeriods,
		days:              totalDays,
		lessons:           totalLessons,
		subjectProfessors: totalSubjectProfessors,
		groups:            totalGroups,
		rooms:             totalRooms,
	}

	satInstance, explicitVariables := buildSat(variables, constraints, state)

	//** Solve SAT instance
	solution, err := timetabler.solver.Solve(satInstance)
	if err != nil {
		return nil, 0, 0, err
	} else if solution == nil { // Return nil if the SAT instance is not satisfiable
		return nil, variables, uint64(len(satInstance.Clauses)), nil
	}

	// Filter solution by taking only positive and explicit variables
	solution = lo.Filter(solution, func(variable int64, _ int) bool {
		return variable > 0 && explicitVariables[variable]
	})

	timetable, err = roomAssignment(solution, indexer, standardEvaluator, modelInput)
	return timetable, variables, uint64(len(satInstance.Clauses)), err
}

func (timetabler *isolatedRoomTimetabler) Verify(timetable [][6]uint64, modelInput ModelInput) bool {
	return verify(timetable, modelInput)
}
