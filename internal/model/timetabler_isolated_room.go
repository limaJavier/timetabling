package model

import (
	"log"
	"timetabling/internal/sat"
)

type isolatedRoomTimetabler struct {
	solver sat.SATSolver
}

func NewIsolatedRoomTimetabler(solver sat.SATSolver) Timetabler {
	return &isolatedRoomTimetabler{
		solver: solver,
	}
}

func (timetabler *isolatedRoomTimetabler) Build(
	modelInput ModelInput,
	curriculum [][]bool,
	groups map[uint64][]uint64,
	groupsGraph [][]bool,
) ([][6]uint64, error) {

	//** Extract attributes's domains
	totalRooms := uint64(1)
	totalPeriods, totalDays, totalLessons, totalSubjectProfessors, totalGroups, _ := getAttributes(modelInput, curriculum)

	//** Initialize dependencies
	evaluator := NewPredicateEvaluatorIsolatedRoom(
		modelInput,
		curriculum,
		groups,
		groupsGraph,
	)
	indexer := NewIndexer(totalPeriods, totalDays, totalLessons, totalSubjectProfessors, totalGroups, totalRooms)
	generator := NewPermutationGenerator(totalPeriods, totalDays, totalLessons, totalSubjectProfessors, totalGroups, totalRooms)

	//** Build SAT instance
	variables := totalPeriods * totalDays * totalLessons * totalSubjectProfessors * totalGroups * totalRooms

	// Constraints functions
	constraints := []func(state ConstraintState) [][]int64{
		professorConstraints,
		studentConstraints,
		subjectPermissibilityConstraints,
		professorAvailabilityConstraints,
		lessonConstraints,
		completenessConstraints,
		negationConstraints,
		uniquenessConstraints,
	}

	state := ConstraintState{
		evaluator:         evaluator,
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
	log.Println("Start solver")
	solution, err := timetabler.solver.Solve(satInstance)
	log.Println("Solver done")
	if err != nil {
		return nil, err
	} else if solution == nil { // Return nil if the SAT instance is not satisfiable
		return nil, nil
	}

	timetable := [][6]uint64{}
	for _, variable := range solution {
		// Acknowledge only positive variables that are explicitly stated in the clauses
		if variable > 0 && explicitVariables[variable] {
			positive := [6]uint64{}
			positive[0], positive[1], positive[2], positive[3], positive[4], positive[5] = indexer.Attributes(uint64(variable))
			timetable = append(timetable, positive)
		}
	}

	return timetable, nil
}

func (timetabler *isolatedRoomTimetabler) Verify(
	timetable [][6]uint64,
	modelInput ModelInput,
	curriculum [][]bool,
	groups map[uint64][]uint64,
	groupsGraph [][]bool,
) bool {
	return verify(timetable, modelInput, curriculum, groups, groupsGraph)
}
