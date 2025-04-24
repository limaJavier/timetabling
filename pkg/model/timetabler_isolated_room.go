package model

import (
	"timetabling/pkg/sat"

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

func (timetabler *isolatedRoomTimetabler) Build(modelInput ModelInput) ([][6]uint64, error) {
	//** Preprocess input
	curriculum, groups, groupsGraph := preprocessInput(modelInput)

	//** Extract attributes's domains
	totalRooms := uint64(1)
	totalPeriods, totalDays, totalLessons, totalSubjectProfessors, totalGroups, _ := getAttributes(modelInput, curriculum)

	//** Initialize dependencies
	isolatedEvaluator := newPredicateEvaluatorIsolatedRoom(
		modelInput,
		curriculum,
		groups,
		groupsGraph,
		timetabler.roomSimilarityThreshold,
	)
	standardEvaluator := newPredicateEvaluator(
		modelInput,
		curriculum,
		groups,
		groupsGraph,
		timetabler.roomSimilarityThreshold,
	)
	indexer := newIndexer(totalPeriods, totalDays, totalLessons, totalSubjectProfessors, totalGroups, totalRooms)
	generator := newPermutationGenerator(totalPeriods, totalDays, totalLessons, totalSubjectProfessors, totalGroups, totalRooms)

	//** Build SAT instance
	variables := totalPeriods * totalDays * totalLessons * totalSubjectProfessors * totalGroups * totalRooms

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
		return nil, err
	} else if solution == nil { // Return nil if the SAT instance is not satisfiable
		return nil, nil
	}

	// Filter solution by taking only positive and explicit variables
	solution = lo.Filter(solution, func(variable int64, _ int) bool {
		return variable > 0 && explicitVariables[variable]
	})

	return roomAssignment(solution, indexer, standardEvaluator, modelInput)
}

func (timetabler *isolatedRoomTimetabler) Verify(timetable [][6]uint64, modelInput ModelInput) bool {
	return verify(timetable, modelInput)
}
