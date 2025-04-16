package model

import (
	"log"
	"timetabling/internal/sat"

	"github.com/samber/lo"
)

type satTimetabler struct {
	solver sat.SATSolver
}

func newSatTimetabler(solver sat.SATSolver) *satTimetabler {
	return &satTimetabler{
		solver: solver,
	}
}

func (timetabler *satTimetabler) Build(
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

func (timetabler *satTimetabler) Verify(
	timetable [][6]uint64,
	modelInput ModelInput,
	curriculum [][]bool,
	groups map[uint64][]uint64,
	groupsGraph [][]bool,
) bool {
	//** Initialize dependencies
	evaluator := NewPredicateEvaluator(
		modelInput,
		curriculum,
		groups,
		groupsGraph,
	)

	//** Extract attributes's domains
	totalPeriods, totalDays, _, _, totalGroups, _ := getAttributes(modelInput, curriculum)

	//** Initialize derived-lessons
	derivedLessons := make(map[uint64]uint64)

	//** Initialize professor-assistance
	professorAssistance := make(map[uint64][][]bool)
	for professor := range len(modelInput.Professors) {
		professorAssistance[uint64(professor)] = make([][]bool, totalPeriods)
		for i := range professorAssistance[uint64(professor)] {
			professorAssistance[uint64(professor)][i] = make([]bool, totalDays)
		}
	}

	//** Initialize group-assistance
	groupAssistance := make(map[uint64][][]bool)
	for group := range totalGroups {
		groupAssistance[group] = make([][]bool, totalPeriods)
		for i := range groupAssistance[group] {
			groupAssistance[group][i] = make([]bool, totalDays)
		}
	}

	lessonTaught := make(map[[3]uint64]bool)

	for _, positive := range timetable {
		period, day, subjectProfessorId, group, room := positive[0], positive[1], positive[3], positive[4], positive[5]
		professor := modelInput.SubjectProfessors[subjectProfessorId].Professor

		_, alreadyTaught := lessonTaught[[3]uint64{group, subjectProfessorId, day}]
		// Check that:
		// - SubjectProfessor is allowed to teach (or to be taught) in the period and day
		// - Professor is available in the period and day
		// - Professor is not already assisting in the period and day
		// - A group with a common class is not already scheduled in the period and day (no collision)
		// - A subjectProfessor can only teach (or be taught) a group once a day
		// - Room is not assigned to subjectProfessor
		// - Group does not fit in room
		if !modelInput.SubjectProfessors[subjectProfessorId].Permissibility[period][day] ||
			!evaluator.ProfessorAvailable(subjectProfessorId, day, period) ||
			professorAssistance[professor][period][day] ||
			collide(groupsGraph, groupAssistance, group, period, day) ||
			alreadyTaught ||
			!evaluator.Assigned(room, subjectProfessorId) ||
			!evaluator.Fits(group, room) {
			return false
		}

		professorAssistance[professor][period][day] = true             // Store professor assistance
		groupAssistance[group][period][day] = true                     // Store group assistance
		derivedLessons[subjectProfessorId]++                           // Store lesson taught
		lessonTaught[[3]uint64{group, subjectProfessorId, day}] = true // Store lesson taught
	}

	for _, subjectProfessor := range modelInput.SubjectProfessors {
		subjectProfessorId, groups := subjectProfessor.Id, subjectProfessor.Groups
		derivedLessons[subjectProfessorId] /= uint64(len(groups))
	}

	// Check whether the number of lessons taught for each subjectProfessor is equal to the number of lessons assigned in the curriculum
	return !lo.SomeBy(modelInput.SubjectProfessors, func(subjectProfessor SubjectProfessor) bool {
		return derivedLessons[subjectProfessor.Id] != subjectProfessor.Lessons
	})
}

func getAttributes(modelInput ModelInput, curriculum [][]bool) (periods, days, lessons, subjectProfessors, groups, rooms uint64) {
	periods = uint64(len(modelInput.Professors[0].Availability))
	days = uint64(len(modelInput.Professors[0].Availability[0]))
	subjectProfessors = uint64(len(modelInput.Professors))
	groups = uint64(len(curriculum))
	rooms = uint64(len(modelInput.Rooms))

	lessons = lo.Max(lo.Map(modelInput.SubjectProfessors, func(subjectProfessor SubjectProfessor, _ int) uint64 {
		return subjectProfessor.Lessons
	}))

	return periods, days, lessons, subjectProfessors, groups, rooms
}

// Checks whether two groups sharing a common class (or classes) are scheduled in the same period and day
func collide(groupsGraph [][]bool, groupAssistance map[uint64][][]bool, group, period, day uint64) bool {
	for neighborGroup, notDisjoint := range groupsGraph[group] {
		if notDisjoint && groupAssistance[uint64(neighborGroup)][period][day] {
			return true
		}
	}
	return false
}
