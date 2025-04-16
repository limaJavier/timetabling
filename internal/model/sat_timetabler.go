package model

import (
	"log"
	"math"
	"timetabling/internal/sat"

	"github.com/samber/lo"
)

type satTimetabler struct {
	evaluator PredicateEvaluator
	indexer   Indexer
	generator PermutationGenerator
	solver    sat.SATSolver

	periods,
	days,
	lessons,
	subjectProfessors,
	groups,
	rooms uint64
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
	timetabler.getAttributes(modelInput, curriculum)

	//** Initialize dependencies
	timetabler.evaluator = NewPredicateEvaluator(
		modelInput,
		curriculum,
		groups,
		groupsGraph,
	)

	timetabler.indexer = NewIndexer(timetabler.periods, timetabler.days, timetabler.lessons, timetabler.subjectProfessors, timetabler.groups, timetabler.rooms)
	timetabler.generator = NewPermutationGenerator(timetabler.periods, timetabler.days, timetabler.lessons, timetabler.subjectProfessors, timetabler.groups, timetabler.rooms)

	//** Build SAT instance
	satInstance := sat.SAT{
		Variables: timetabler.periods * timetabler.days * timetabler.lessons * timetabler.subjectProfessors * timetabler.groups * timetabler.rooms,
		Clauses:   [][]int64{},
	}

	explicitVariables := make(map[int64]bool)  // Variables that are explicitly stated in the clauses
	constraintsChannel := make(chan [][]int64) // Channel to collect constraints

	// Constraints functions
	constraints := []func() [][]int64{
		timetabler.professorConstraints,
		timetabler.studentConstraints,
		timetabler.subjectPermissibilityConstraints,
		timetabler.professorAvailabilityConstraints,
		timetabler.lessonConstraints,
		timetabler.roomConstraints,
		timetabler.roomNegationConstraints,
		timetabler.completenessConstraints,
		timetabler.negationConstraints,
		timetabler.uniquenessConstraints,
	}

	// Execute constraints functions on different goroutines to improve performance
	for _, constraint := range constraints {
		go func(constraint func() [][]int64) {
			constraintsChannel <- constraint()
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
			positive[0], positive[1], positive[2], positive[3], positive[4], positive[5] = timetabler.indexer.Attributes(uint64(variable))
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

	//** Extract attributes's domains
	timetabler.getAttributes(modelInput, curriculum)

	//** Initialize derived-lessons
	derivedLessons := make(map[uint64]uint64)

	//** Initialize professor-assistance
	professorAssistance := make(map[uint64][][]bool)
	for professor := range len(modelInput.Professors) {
		professorAssistance[uint64(professor)] = make([][]bool, timetabler.periods)
		for i := range professorAssistance[uint64(professor)] {
			professorAssistance[uint64(professor)][i] = make([]bool, timetabler.days)
		}
	}

	//** Initialize group-assistance
	groupAssistance := make(map[uint64][][]bool)
	for group := range timetabler.groups {
		groupAssistance[group] = make([][]bool, timetabler.periods)
		for i := range groupAssistance[group] {
			groupAssistance[group][i] = make([]bool, timetabler.days)
		}
	}

	lessonTaught := make(map[[3]uint64]bool)

	for _, positive := range timetable {
		period, day, subjectProfessorId, group := positive[0], positive[1], positive[3], positive[4]
		professor := modelInput.SubjectProfessors[subjectProfessorId].Professor
		_, alreadyTaught := lessonTaught[[3]uint64{group, subjectProfessorId, day}]

		// Check that:
		// - SubjectProfessor is allowed to teach (or to be taught) in the period and day
		// - Professor is available in the period and day
		// - Professor is not already assisting in the period and day
		// - A group with a common class is not already scheduled in the period and day (no collision)
		// - A subjectProfessor can only teach (or be taught) a group once a day
		if !modelInput.SubjectProfessors[subjectProfessorId].Permissibility[period][day] || !modelInput.Professors[professor].Availability[period][day] || professorAssistance[professor][period][day] || collide(groupsGraph, groupAssistance, group, period, day) || alreadyTaught {
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

func (timetabler *satTimetabler) professorConstraints() [][]int64 {
	permutations := timetabler.generator.ConstrainedPermutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 1
		func(permutation []uint64) bool {
			lesson, subjectProfessor, group := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Teaches(group, subjectProfessor, lesson)
		},
		// Allowed(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor := permutation[0], permutation[1], permutation[3]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Allowed(subjectProfessor, day, period)
		},
		// ProfessorAvailable(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor := permutation[0], permutation[1], permutation[3]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.ProfessorAvailable(subjectProfessor, day, period)
		},
		// Assigned(r, i) = 1
		func(permutation []uint64) bool {
			subjectProfessor, room := permutation[3], permutation[5]

			return subjectProfessor == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Assigned(room, subjectProfessor)
		},
		// Fits(k, r) = 1
		func(permutation []uint64) bool {
			group, room := permutation[4], permutation[5]

			return group == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Fits(group, room)
		},
	})

	clauses := make([][]int64, 0)

	// Due to the nature of the iteration process we're are certain that we won't find the case where: k = k', i = i', j = j', d = d', t = t'
	for i := range len(permutations) - 1 {
		for j := i + 1; j < len(permutations); j++ {
			permutation1, permutation2 := permutations[i], permutations[j]
			period1, day1, lesson1, subjectProfessor1, group1, room1 := permutation1[0], permutation1[1], permutation1[2], permutation1[3], permutation1[4], permutation1[5]
			period2, day2, lesson2, subjectProfessor2, group2, room2 := permutation2[0], permutation2[1], permutation2[2], permutation2[3], permutation2[4], permutation2[5]

			// d == d', t == t', SameProfessor(i, i') = 1
			if period1 == period2 && day1 == day2 && timetabler.evaluator.SameProfessor(subjectProfessor1, subjectProfessor2) {
				index1 := timetabler.indexer.Index(period1, day1, lesson1, subjectProfessor1, group1, room1)
				index2 := timetabler.indexer.Index(period2, day2, lesson2, subjectProfessor2, group2, room2)
				clauses = append(clauses, []int64{-int64(index1), -int64(index2)})
			}
		}
	}

	return clauses
}

func (timetabler *satTimetabler) studentConstraints() [][]int64 {
	permutations := timetabler.generator.ConstrainedPermutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 1
		func(permutation []uint64) bool {
			lesson, subjectProfessor, group := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Teaches(group, subjectProfessor, lesson)
		},
		// Allowed(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor := permutation[0], permutation[1], permutation[3]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Allowed(subjectProfessor, day, period)
		},
		// ProfessorAvailable(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor := permutation[0], permutation[1], permutation[3]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.ProfessorAvailable(subjectProfessor, day, period)
		},
		// Assigned(r, i) = 1
		func(permutation []uint64) bool {
			subjectProfessor, room := permutation[3], permutation[5]

			return subjectProfessor == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Assigned(room, subjectProfessor)
		},
		// Fits(k, r) = 1
		func(permutation []uint64) bool {
			group, room := permutation[4], permutation[5]

			return group == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Fits(group, room)
		},
	})

	clauses := make([][]int64, 0)

	// Due to the nature of the iteration process we're are certain that we won't find the case where: k = k', i = i', j = j', d = d', t = t'
	for i := range len(permutations) - 1 {
		for j := i + 1; j < len(permutations); j++ {
			permutation1, permutation2 := permutations[i], permutations[j]
			period1, day1, lesson1, subjectProfessor1, group1, room1 := permutation1[0], permutation1[1], permutation1[2], permutation1[3], permutation1[4], permutation1[5]
			period2, day2, lesson2, subjectProfessor2, group2, room2 := permutation2[0], permutation2[1], permutation2[2], permutation2[3], permutation2[4], permutation2[5]

			// Disjoint(k, k') = 0, d = d', t = t', SameProfessor(i, i') = 0
			if period1 == period2 && day1 == day2 && !timetabler.evaluator.Disjoint(group1, group2) && !timetabler.evaluator.SameProfessor(subjectProfessor1, subjectProfessor2) {
				index1 := timetabler.indexer.Index(period1, day1, lesson1, subjectProfessor1, group1, room1)
				index2 := timetabler.indexer.Index(period2, day2, lesson2, subjectProfessor2, group2, room2)

				clauses = append(clauses, []int64{-int64(index1), -int64(index2)})
			}
		}
	}

	return clauses
}

func (timetabler *satTimetabler) subjectPermissibilityConstraints() [][]int64 {
	permutations := timetabler.generator.ConstrainedPermutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 1
		func(permutation []uint64) bool {
			lesson, subjectProfessor, group := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Teaches(group, subjectProfessor, lesson)
		},
		// Allowed(i, d, t) = 0
		func(permutation []uint64) bool {
			period, day, subjectProfessor := permutation[0], permutation[1], permutation[3]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||

				// Actual predicate
				!timetabler.evaluator.Allowed(subjectProfessor, day, period)
		},
	})

	clauses := make([][]int64, 0)

	for _, permutation := range permutations {
		period, day, lesson, subjectProfessor, group, room := permutation[0], permutation[1], permutation[2], permutation[3], permutation[4], permutation[5]

		index := timetabler.indexer.Index(period, day, lesson, subjectProfessor, group, room)

		clauses = append(clauses, []int64{-int64(index)})
	}

	return clauses
}

func (timetabler *satTimetabler) professorAvailabilityConstraints() [][]int64 {
	permutations := timetabler.generator.ConstrainedPermutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 1
		func(permutation []uint64) bool {
			lesson, subjectProfessor, group := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Teaches(group, subjectProfessor, lesson)
		},
		// ProfessorAvailable(i, d, t) = 0
		func(permutation []uint64) bool {
			period, day, subjectProfessor := permutation[0], permutation[1], permutation[3]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||

				// Actual predicate
				!timetabler.evaluator.ProfessorAvailable(subjectProfessor, day, period)
		},
	})

	clauses := make([][]int64, 0)

	for _, permutation := range permutations {
		period, day, lesson, subjectProfessor, group, room := permutation[0], permutation[1], permutation[2], permutation[3], permutation[4], permutation[5]

		index := timetabler.indexer.Index(period, day, lesson, subjectProfessor, group, room)

		clauses = append(clauses, []int64{-int64(index)})
	}

	return clauses
}

// TODO: Generate similar constraints of that of professorAvailabilityConstraints but for rooms

func (timetabler *satTimetabler) lessonConstraints() [][]int64 {
	permutations := timetabler.generator.ConstrainedPermutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 1
		func(permutation []uint64) bool {
			lesson, subjectProfessor, group := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Teaches(group, subjectProfessor, lesson)
		},
		// Allowed(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor := permutation[0], permutation[1], permutation[3]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Allowed(subjectProfessor, day, period)
		},
		// ProfessorAvailable(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor := permutation[0], permutation[1], permutation[3]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.ProfessorAvailable(subjectProfessor, day, period)
		},
		// Assigned(r, i) = 1
		func(permutation []uint64) bool {
			subjectProfessor, room := permutation[3], permutation[5]

			return subjectProfessor == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Assigned(room, subjectProfessor)
		},
		// Fits(k, r) = 1
		func(permutation []uint64) bool {
			group, room := permutation[4], permutation[5]

			return group == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Fits(group, room)
		},
	})

	clauses := make([][]int64, 0)

	// Due to the nature of the iteration process we're are certain that we won't find the case where: k = k', i = i', j = j', d = d', t = t'
	for i := range len(permutations) - 1 {
		for j := i + 1; j < len(permutations); j++ {
			permutation1, permutation2 := permutations[i], permutations[j]
			period1, day1, lesson1, subjectProfessor1, group1, room1 := permutation1[0], permutation1[1], permutation1[2], permutation1[3], permutation1[4], permutation1[5]
			period2, day2, lesson2, subjectProfessor2, group2, room2 := permutation2[0], permutation2[1], permutation2[2], permutation2[3], permutation2[4], permutation2[5]

			// k = k', i = i', d = d', j != j'
			if group1 == group2 && subjectProfessor1 == subjectProfessor2 && day1 == day2 && lesson1 != lesson2 {
				index1 := timetabler.indexer.Index(period1, day1, lesson1, subjectProfessor1, group1, room1)
				index2 := timetabler.indexer.Index(period2, day2, lesson2, subjectProfessor2, group2, room2)

				clauses = append(clauses, []int64{-int64(index1), -int64(index2)})
			}
		}
	}

	return clauses
}

func (timetabler *satTimetabler) roomConstraints() [][]int64 {
	permutations := timetabler.generator.ConstrainedPermutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 1
		func(permutation []uint64) bool {
			lesson, subjectProfessor, group := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Teaches(group, subjectProfessor, lesson)
		},
		// Allowed(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor := permutation[0], permutation[1], permutation[3]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Allowed(subjectProfessor, day, period)
		},
		// ProfessorAvailable(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor := permutation[0], permutation[1], permutation[3]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.ProfessorAvailable(subjectProfessor, day, period)
		},
		// Assigned(r, i) = 1
		func(permutation []uint64) bool {
			subjectProfessor, room := permutation[3], permutation[5]

			return subjectProfessor == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Assigned(room, subjectProfessor)
		},
		// Fits(k, r) = 1
		func(permutation []uint64) bool {
			group, room := permutation[4], permutation[5]

			return group == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Fits(group, room)
		},
	})

	clauses := make([][]int64, 0)

	// Due to the nature of the iteration process we're are certain that we won't find the case where: k = k', i = i', j = j', d = d', t = t'
	for i := range len(permutations) - 1 {
		for j := i + 1; j < len(permutations); j++ {
			permutation1, permutation2 := permutations[i], permutations[j]
			period1, day1, lesson1, subjectProfessor1, group1, room1 := permutation1[0], permutation1[1], permutation1[2], permutation1[3], permutation1[4], permutation1[5]
			period2, day2, lesson2, subjectProfessor2, group2, room2 := permutation2[0], permutation2[1], permutation2[2], permutation2[3], permutation2[4], permutation2[5]

			// d == d', t == t', r == r', SameProfessor(i, i') = 0, k != k'
			if period1 == period2 && day1 == day2 && room1 == room2 && !timetabler.evaluator.SameProfessor(subjectProfessor1, subjectProfessor2) && group1 != group2 {
				index1 := timetabler.indexer.Index(period1, day1, lesson1, subjectProfessor1, group1, room1)
				index2 := timetabler.indexer.Index(period2, day2, lesson2, subjectProfessor2, group2, room2)
				clauses = append(clauses, []int64{-int64(index1), -int64(index2)})
			}
		}
	}

	return clauses
}

func (timetabler *satTimetabler) roomNegationConstraints() [][]int64 {
	permutations := timetabler.generator.ConstrainedPermutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 1
		func(permutation []uint64) bool {
			lesson, subjectProfessor, group := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Teaches(group, subjectProfessor, lesson)
		},
		// Assigned(r, i) = 0 or Fits(k, r) = 0
		func(permutation []uint64) bool {
			subjectProfessor, group, room := permutation[3], permutation[4], permutation[5]

			return subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				!timetabler.evaluator.Assigned(room, subjectProfessor) ||
				!timetabler.evaluator.Fits(group, room)
		},
	})

	clauses := make([][]int64, 0)

	for _, permutation := range permutations {
		period, day, lesson, subjectProfessor, group, room := permutation[0], permutation[1], permutation[2], permutation[3], permutation[4], permutation[5]

		index := timetabler.indexer.Index(period, day, lesson, subjectProfessor, group, room)

		clauses = append(clauses, []int64{-int64(index)})
	}

	return clauses
}

func (timetabler *satTimetabler) completenessConstraints() [][]int64 {
	// <Lesson, SubjectProfessor, Group> triplets
	triplets := make([][3]uint64, 0)
	_ = timetabler.generator.ConstrainedPermutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 1
		func(permutation []uint64) bool {
			lesson, subjectProfessor, group := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Teaches(group, subjectProfessor, lesson)
		},
		// According to how ConstrainedPermutations works this predicate will be evaluated only if the previous one evaluates to true, therefore any triplet that reaches it will be a valid one
		func(permutation []uint64) bool {
			lesson, subjectProfessor, group := permutation[2], permutation[3], permutation[4]
			triplet := [3]uint64{lesson, subjectProfessor, group}

			if lesson != math.MaxUint64 && subjectProfessor != math.MaxUint64 && group != math.MaxUint64 {
				triplets = append(triplets, triplet)
			}
			return true // Always return true since group will be the last attribute to fill during backtracking, so there will be no further ado
		},
	})

	clauses := make([][]int64, 0)

	for _, triplet := range triplets {
		lesson, subjectProfessor, group := triplet[0], triplet[1], triplet[2]
		clause := []int64{}
		for period := range timetabler.periods {
			for day := range timetabler.days {
				for room := range timetabler.rooms {
					// Allowed(i, d, t) = 1, ProfessorAvailable(i, d, t) = 1, Assigned(r, i) = 1, Fits(k, r) = 1
					if timetabler.evaluator.Allowed(subjectProfessor, day, period) &&
						timetabler.evaluator.ProfessorAvailable(subjectProfessor, day, period) &&
						timetabler.evaluator.Assigned(room, subjectProfessor) &&
						timetabler.evaluator.Fits(group, room) {
						index := timetabler.indexer.Index(period, day, lesson, subjectProfessor, group, room)
						clause = append(clause, int64(index))
					}
				}
			}
		}
		clauses = append(clauses, clause)
	}

	return clauses
}

func (timetabler *satTimetabler) negationConstraints() [][]int64 {
	permutations := timetabler.generator.ConstrainedPermutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 0
		func(permutation []uint64) bool {
			lesson, subjectProfessor, group := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				!timetabler.evaluator.Teaches(group, subjectProfessor, lesson)
		},
		// Allowed(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor := permutation[0], permutation[1], permutation[3]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Allowed(subjectProfessor, day, period)
		},
		// ProfessorAvailable(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor := permutation[0], permutation[1], permutation[3]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.ProfessorAvailable(subjectProfessor, day, period)
		},
		// Assigned(r, i) = 1
		func(permutation []uint64) bool {
			subjectProfessor, room := permutation[3], permutation[5]

			return subjectProfessor == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Assigned(room, subjectProfessor)
		},
		// Fits(k, r) = 1
		func(permutation []uint64) bool {
			group, room := permutation[4], permutation[5]

			return group == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Fits(group, room)
		},
	})

	clauses := make([][]int64, 0)

	for _, permutation := range permutations {
		period, day, lesson, subjectProfessor, group, room := permutation[0], permutation[1], permutation[2], permutation[3], permutation[4], permutation[5]

		index := timetabler.indexer.Index(period, day, lesson, subjectProfessor, group, room)

		clauses = append(clauses, []int64{-int64(index)})
	}

	return clauses
}

// TODO: (Optional) This method can be performance-optimized by a triple for loop instead of going through all permutations
func (timetabler *satTimetabler) uniquenessConstraints() [][]int64 {
	permutations := timetabler.generator.ConstrainedPermutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 1
		func(permutation []uint64) bool {
			lesson, subjectProfessor, group := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Teaches(group, subjectProfessor, lesson)
		},
		// Allowed(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor := permutation[0], permutation[1], permutation[3]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Allowed(subjectProfessor, day, period)
		},
		// ProfessorAvailable(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor := permutation[0], permutation[1], permutation[3]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.ProfessorAvailable(subjectProfessor, day, period)
		},
		// Assigned(r, i) = 1
		func(permutation []uint64) bool {
			subjectProfessor, room := permutation[3], permutation[5]

			return subjectProfessor == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Assigned(room, subjectProfessor)
		},
		// Fits(k, r) = 1
		func(permutation []uint64) bool {
			group, room := permutation[4], permutation[5]

			return group == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Fits(group, room)
		},
	})

	clauses := make([][]int64, 0)

	// Due to the nature of the iteration we're are certain that we won't find the case where: k = k', i = i', j = j', d = d', t = t'
	for i := range len(permutations) - 1 {
		for j := i + 1; j < len(permutations); j++ {
			permutation1, permutation2 := permutations[i], permutations[j]
			period1, day1, lesson1, subjectProfessor1, group1, room1 := permutation1[0], permutation1[1], permutation1[2], permutation1[3], permutation1[4], permutation1[5]
			period2, day2, lesson2, subjectProfessor2, group2, room2 := permutation2[0], permutation2[1], permutation2[2], permutation2[3], permutation2[4], permutation2[5]

			// k == k', i == i', j == j'
			if group1 == group2 && subjectProfessor1 == subjectProfessor2 && lesson1 == lesson2 {
				index1 := timetabler.indexer.Index(period1, day1, lesson1, subjectProfessor1, group1, room1)
				index2 := timetabler.indexer.Index(period2, day2, lesson2, subjectProfessor2, group2, room2)
				clauses = append(clauses, []int64{-int64(index1), -int64(index2)})
			}
		}
	}

	return clauses
}

func (timetabler *satTimetabler) getAttributes(modelInput ModelInput, curriculum [][]bool) {
	timetabler.periods = uint64(len(modelInput.Professors[0].Availability))
	timetabler.days = uint64(len(modelInput.Professors[0].Availability[0]))
	timetabler.subjectProfessors = uint64(len(modelInput.Professors))
	timetabler.groups = uint64(len(curriculum))
	timetabler.rooms = uint64(len(modelInput.Rooms))

	timetabler.lessons = lo.Max(lo.Map(modelInput.SubjectProfessors, func(subjectProfessor SubjectProfessor, _ int) uint64 {
		return subjectProfessor.Lessons
	}))
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
