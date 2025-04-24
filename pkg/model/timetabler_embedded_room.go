package model

import "timetabling/pkg/sat"

type embeddedRoomTimetabler struct {
	solver sat.SATSolver
}

func NewEmbeddedRoomTimetabler(solver sat.SATSolver) Timetabler {
	return &embeddedRoomTimetabler{
		solver: solver,
	}
}

func (timetabler *embeddedRoomTimetabler) Build(modelInput ModelInput) ([][6]uint64, error) {
	//** Preprocess input
	curriculum, groups, groupsGraph := preprocessInput(modelInput)

	//** Extract attributes's domains
	totalPeriods, totalDays, totalLessons, totalSubjectProfessors, totalGroups, totalRooms := getAttributes(modelInput, curriculum)

	//** Initialize dependencies
	evaluator := newPredicateEvaluator(
		modelInput,
		curriculum,
		groups,
		groupsGraph,
		0,
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
		roomConstraints,
		roomNegationConstraints,
		completenessConstraints,
		negationConstraints,
		uniquenessConstraints,
	}

	state := constraintState{
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
	solution, err := timetabler.solver.Solve(satInstance)
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

func (timetabler *embeddedRoomTimetabler) Verify(timetable [][6]uint64, modelInput ModelInput) bool {
	return verify(timetable, modelInput)
}
