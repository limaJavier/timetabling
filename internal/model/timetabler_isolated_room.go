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
	totalPeriods, totalDays, totalLessons, totalSubjectProfessors, totalGroups, _ := getAttributes(modelInput, curriculum)
	totalRooms := uint64(1)

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
	satInstance := sat.SAT{
		Variables: totalPeriods * totalDays * totalLessons * totalSubjectProfessors * totalGroups * totalRooms,
		Clauses:   [][]int64{},
	}

	explicitVariables := make(map[int64]bool)  // Variables that are explicitly stated in the clauses
	constraintsChannel := make(chan [][]int64) // Channel to collect constraints

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

	// Execute constraints functions on different goroutines to improve performance
	for _, constraint := range constraints {
		go func(constraint func(state ConstraintState) [][]int64) {
			constraintsChannel <- constraint(state)
		}(constraint)
	}

	// Collect generated constraints
	collectedConstraints := 0
	for clauses := range constraintsChannel {
		for _, clause := range clauses {
			for _, variable := range clause {
				// Check whether the variable is positive, since required explicit variables ought to be positive
				if variable > 0 {
					explicitVariables[variable] = true
				}
			}
		}
		// Append clauses to the SAT instance
		satInstance.Clauses = append(satInstance.Clauses, clauses...)

		// Check whether all constraints have been collected to properly close the channel
		if collectedConstraints++; collectedConstraints == len(constraints) {
			close(constraintsChannel)
		}
	}

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
