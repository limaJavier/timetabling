package model

import (
	"log"
	"timetabling/internal/sat"
)

type embeddedRoomTimetabler struct {
	solver sat.SATSolver
}

func NewEmbeddedRoomTimetabler(solver sat.SATSolver) Timetabler {
	return &embeddedRoomTimetabler{
		solver: solver,
	}
}

func (timetabler *embeddedRoomTimetabler) Build(
	modelInput ModelInput,
	curriculum [][]bool,
	groups map[uint64][]uint64,
	groupsGraph [][]bool,
) ([][6]uint64, error) {

	//** Extract attributes's domains
	totalPeriods, totalDays, totalLessons, totalSubjectProfessors, totalGroups, totalRooms := getAttributes(modelInput, curriculum)

	//** Initialize dependencies
	evaluator := NewPredicateEvaluator(
		modelInput,
		curriculum,
		groups,
		groupsGraph,
		0,
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
		roomConstraints,
		roomNegationConstraints,
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

func (timetabler *embeddedRoomTimetabler) Verify(
	timetable [][6]uint64,
	modelInput ModelInput,
	curriculum [][]bool,
	groups map[uint64][]uint64,
	groupsGraph [][]bool,
) bool {
	return verify(timetable, modelInput, curriculum, groups, groupsGraph)
}
