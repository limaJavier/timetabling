package model

import "math"

type constraintState struct {
	evaluator predicateEvaluator
	indexer   indexer
	generator permutationGenerator

	periods,
	days,
	lessons,
	subjectProfessors,
	groups,
	rooms uint64
}

func professorConstraints(state constraintState) [][]int64 {
	permutations := state.generator.ConstrainedPermutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 1
		func(permutation []uint64) bool {
			lesson, subjectProfessor, group := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Teaches(group, subjectProfessor, lesson)
		},
		// Allowed(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor, group := permutation[0], permutation[1], permutation[3], permutation[4]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Allowed(subjectProfessor, group, day, period)
		},
		// ProfessorAvailable(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor := permutation[0], permutation[1], permutation[3]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.ProfessorAvailable(subjectProfessor, day, period)
		},
		// Assigned(r, i) = 1
		func(permutation []uint64) bool {
			subjectProfessor, group, room := permutation[3], permutation[4], permutation[5]

			return subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Assigned(room, subjectProfessor, group)
		},
		// Fits(k, r) = 1
		func(permutation []uint64) bool {
			group, room := permutation[4], permutation[5]

			return group == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Fits(group, room)
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
			if period1 == period2 && day1 == day2 && state.evaluator.SameProfessor(subjectProfessor1, subjectProfessor2) {
				index1 := state.indexer.Index(period1, day1, lesson1, subjectProfessor1, group1, room1)
				index2 := state.indexer.Index(period2, day2, lesson2, subjectProfessor2, group2, room2)
				clauses = append(clauses, []int64{-int64(index1), -int64(index2)})
			}
		}
	}

	return clauses
}

func studentConstraints(state constraintState) [][]int64 {
	permutations := state.generator.ConstrainedPermutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 1
		func(permutation []uint64) bool {
			lesson, subjectProfessor, group := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Teaches(group, subjectProfessor, lesson)
		},
		// Allowed(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor, group := permutation[0], permutation[1], permutation[3], permutation[4]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Allowed(subjectProfessor, group, day, period)
		},
		// ProfessorAvailable(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor := permutation[0], permutation[1], permutation[3]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.ProfessorAvailable(subjectProfessor, day, period)
		},
		// Assigned(r, i) = 1
		func(permutation []uint64) bool {
			subjectProfessor, group, room := permutation[3], permutation[4], permutation[5]

			return subjectProfessor == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Assigned(room, subjectProfessor, group)
		},
		// Fits(k, r) = 1
		func(permutation []uint64) bool {
			group, room := permutation[4], permutation[5]

			return group == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Fits(group, room)
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
			if period1 == period2 && day1 == day2 && !state.evaluator.Disjoint(group1, group2) && !state.evaluator.SameProfessor(subjectProfessor1, subjectProfessor2) {
				index1 := state.indexer.Index(period1, day1, lesson1, subjectProfessor1, group1, room1)
				index2 := state.indexer.Index(period2, day2, lesson2, subjectProfessor2, group2, room2)

				clauses = append(clauses, []int64{-int64(index1), -int64(index2)})
			}
		}
	}

	return clauses
}

func subjectPermissibilityConstraints(state constraintState) [][]int64 {
	permutations := state.generator.ConstrainedPermutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 1
		func(permutation []uint64) bool {
			lesson, subjectProfessor, group := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Teaches(group, subjectProfessor, lesson)
		},
		// Allowed(i, d, t) = 0
		func(permutation []uint64) bool {
			period, day, subjectProfessor, group := permutation[0], permutation[1], permutation[3], permutation[4]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				!state.evaluator.Allowed(subjectProfessor, group, day, period)
		},
		// ProfessorAvailable(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor := permutation[0], permutation[1], permutation[3]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.ProfessorAvailable(subjectProfessor, day, period)
		},
		// Assigned(r, i) = 1
		func(permutation []uint64) bool {
			subjectProfessor, group, room := permutation[3], permutation[4], permutation[5]

			return subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Assigned(room, subjectProfessor, group)
		},
		// Fits(k, r) = 1
		func(permutation []uint64) bool {
			group, room := permutation[4], permutation[5]

			return group == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Fits(group, room)
		},
	})

	clauses := make([][]int64, 0)

	for _, permutation := range permutations {
		period, day, lesson, subjectProfessor, group, room := permutation[0], permutation[1], permutation[2], permutation[3], permutation[4], permutation[5]

		index := state.indexer.Index(period, day, lesson, subjectProfessor, group, room)

		clauses = append(clauses, []int64{-int64(index)})
	}

	return clauses
}

func professorAvailabilityConstraints(state constraintState) [][]int64 {
	permutations := state.generator.ConstrainedPermutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 1
		func(permutation []uint64) bool {
			lesson, subjectProfessor, group := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Teaches(group, subjectProfessor, lesson)
		},
		// Allowed(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor, group := permutation[0], permutation[1], permutation[3], permutation[4]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Allowed(subjectProfessor, group, day, period)
		},
		// ProfessorAvailable(i, d, t) = 0
		func(permutation []uint64) bool {
			period, day, subjectProfessor := permutation[0], permutation[1], permutation[3]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||

				// Actual predicate
				!state.evaluator.ProfessorAvailable(subjectProfessor, day, period)
		},
		// Assigned(r, i) = 1
		func(permutation []uint64) bool {
			subjectProfessor, group, room := permutation[3], permutation[4], permutation[5]

			return subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Assigned(room, subjectProfessor, group)
		},
		// Fits(k, r) = 1
		func(permutation []uint64) bool {
			group, room := permutation[4], permutation[5]

			return group == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Fits(group, room)
		},
	})

	clauses := make([][]int64, 0)

	for _, permutation := range permutations {
		period, day, lesson, subjectProfessor, group, room := permutation[0], permutation[1], permutation[2], permutation[3], permutation[4], permutation[5]

		index := state.indexer.Index(period, day, lesson, subjectProfessor, group, room)

		clauses = append(clauses, []int64{-int64(index)})
	}

	return clauses
}

func lessonConstraints(state constraintState) [][]int64 {
	permutations := state.generator.ConstrainedPermutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 1
		func(permutation []uint64) bool {
			lesson, subjectProfessor, group := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Teaches(group, subjectProfessor, lesson)
		},
		// Allowed(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor, group := permutation[0], permutation[1], permutation[3], permutation[4]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Allowed(subjectProfessor, group, day, period)
		},
		// ProfessorAvailable(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor := permutation[0], permutation[1], permutation[3]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.ProfessorAvailable(subjectProfessor, day, period)
		},
		// Assigned(r, i) = 1
		func(permutation []uint64) bool {
			subjectProfessor, group, room := permutation[3], permutation[4], permutation[5]

			return subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Assigned(room, subjectProfessor, group)
		},
		// Fits(k, r) = 1
		func(permutation []uint64) bool {
			group, room := permutation[4], permutation[5]

			return group == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Fits(group, room)
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
				index1 := state.indexer.Index(period1, day1, lesson1, subjectProfessor1, group1, room1)
				index2 := state.indexer.Index(period2, day2, lesson2, subjectProfessor2, group2, room2)

				clauses = append(clauses, []int64{-int64(index1), -int64(index2)})
			}
		}
	}

	return clauses
}

func roomConstraints(state constraintState) [][]int64 {
	permutations := state.generator.ConstrainedPermutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 1
		func(permutation []uint64) bool {
			lesson, subjectProfessor, group := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Teaches(group, subjectProfessor, lesson)
		},
		// Allowed(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor, group := permutation[0], permutation[1], permutation[3], permutation[4]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Allowed(subjectProfessor, group, day, period)
		},
		// ProfessorAvailable(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor := permutation[0], permutation[1], permutation[3]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.ProfessorAvailable(subjectProfessor, day, period)
		},
		// Assigned(r, i) = 1
		func(permutation []uint64) bool {
			subjectProfessor, group, room := permutation[3], permutation[4], permutation[5]

			return subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Assigned(room, subjectProfessor, group)
		},
		// Fits(k, r) = 1
		func(permutation []uint64) bool {
			group, room := permutation[4], permutation[5]

			return group == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Fits(group, room)
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
			if period1 == period2 && day1 == day2 && room1 == room2 && !state.evaluator.SameProfessor(subjectProfessor1, subjectProfessor2) && group1 != group2 {
				index1 := state.indexer.Index(period1, day1, lesson1, subjectProfessor1, group1, room1)
				index2 := state.indexer.Index(period2, day2, lesson2, subjectProfessor2, group2, room2)
				clauses = append(clauses, []int64{-int64(index1), -int64(index2)})
			}
		}
	}

	return clauses
}

func roomSimilarityConstraints(state constraintState) [][]int64 {
	permutations := state.generator.ConstrainedPermutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 1
		func(permutation []uint64) bool {
			lesson, subjectProfessor, group := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Teaches(group, subjectProfessor, lesson)
		},
		// Allowed(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor, group := permutation[0], permutation[1], permutation[3], permutation[4]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Allowed(subjectProfessor, group, day, period)
		},
		// ProfessorAvailable(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor := permutation[0], permutation[1], permutation[3]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.ProfessorAvailable(subjectProfessor, day, period)
		},
		// Assigned(r, i) = 1
		func(permutation []uint64) bool {
			subjectProfessor, group, room := permutation[3], permutation[4], permutation[5]

			return subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Assigned(room, subjectProfessor, group)
		},
		// Fits(k, r) = 1
		func(permutation []uint64) bool {
			group, room := permutation[4], permutation[5]

			return group == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Fits(group, room)
		},
	})

	clauses := make([][]int64, 0)

	// Due to the nature of the iteration process we're are certain that we won't find the case where: k = k', i = i', j = j', d = d', t = t'
	for i := range len(permutations) - 1 {
		for j := i + 1; j < len(permutations); j++ {
			permutation1, permutation2 := permutations[i], permutations[j]
			period1, day1, lesson1, subjectProfessor1, group1, room1 := permutation1[0], permutation1[1], permutation1[2], permutation1[3], permutation1[4], permutation1[5]
			period2, day2, lesson2, subjectProfessor2, group2, room2 := permutation2[0], permutation2[1], permutation2[2], permutation2[3], permutation2[4], permutation2[5]

			// d == d', t == t', RoomSimilar(i, i') = 1 , SameProfessor(i, i') = 0, k != k'
			if period1 == period2 && day1 == day2 && state.evaluator.RoomSimilar(subjectProfessor1, subjectProfessor2, group1, group2) && !state.evaluator.SameProfessor(subjectProfessor1, subjectProfessor2) && group1 != group2 {
				index1 := state.indexer.Index(period1, day1, lesson1, subjectProfessor1, group1, room1)
				index2 := state.indexer.Index(period2, day2, lesson2, subjectProfessor2, group2, room2)
				clauses = append(clauses, []int64{-int64(index1), -int64(index2)})
			}
		}
	}

	return clauses
}

func roomNegationConstraints(state constraintState) [][]int64 {
	permutations := state.generator.ConstrainedPermutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 1
		func(permutation []uint64) bool {
			lesson, subjectProfessor, group := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Teaches(group, subjectProfessor, lesson)
		},
		// Allowed(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor, group := permutation[0], permutation[1], permutation[3], permutation[4]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Allowed(subjectProfessor, group, day, period)
		},
		// ProfessorAvailable(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor := permutation[0], permutation[1], permutation[3]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.ProfessorAvailable(subjectProfessor, day, period)
		},
		// Assigned(r, i) = 0 or Fits(k, r) = 0
		func(permutation []uint64) bool {
			subjectProfessor, group, room := permutation[3], permutation[4], permutation[5]

			return subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				!state.evaluator.Assigned(room, subjectProfessor, group) ||
				!state.evaluator.Fits(group, room)
		},
	})

	clauses := make([][]int64, 0)

	for _, permutation := range permutations {
		period, day, lesson, subjectProfessor, group, room := permutation[0], permutation[1], permutation[2], permutation[3], permutation[4], permutation[5]

		index := state.indexer.Index(period, day, lesson, subjectProfessor, group, room)

		clauses = append(clauses, []int64{-int64(index)})
	}

	return clauses
}

func completenessConstraints(state constraintState) [][]int64 {
	// <Lesson, SubjectProfessor, Group> triplets
	triplets := make([][3]uint64, 0)
	_ = state.generator.ConstrainedPermutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 1
		func(permutation []uint64) bool {
			lesson, subjectProfessor, group := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Teaches(group, subjectProfessor, lesson)
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
		for period := range state.periods {
			for day := range state.days {
				for room := range state.rooms {
					// Allowed(i, d, t) = 1, ProfessorAvailable(i, d, t) = 1, Assigned(r, i) = 1, Fits(k, r) = 1
					if state.evaluator.Allowed(subjectProfessor, group, day, period) &&
						state.evaluator.ProfessorAvailable(subjectProfessor, day, period) &&
						state.evaluator.Assigned(room, subjectProfessor, group) &&
						state.evaluator.Fits(group, room) {
						index := state.indexer.Index(period, day, lesson, subjectProfessor, group, room)
						clause = append(clause, int64(index))
					}
				}
			}
		}
		clauses = append(clauses, clause)
	}

	return clauses
}

func negationConstraints(state constraintState) [][]int64 {
	permutations := state.generator.ConstrainedPermutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 0
		func(permutation []uint64) bool {
			lesson, subjectProfessor, group := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				!state.evaluator.Teaches(group, subjectProfessor, lesson)
		},
		// Allowed(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor, group := permutation[0], permutation[1], permutation[3], permutation[4]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Allowed(subjectProfessor, group, day, period)
		},
		// ProfessorAvailable(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor := permutation[0], permutation[1], permutation[3]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.ProfessorAvailable(subjectProfessor, day, period)
		},
		// Assigned(r, i) = 1
		func(permutation []uint64) bool {
			subjectProfessor, group, room := permutation[3], permutation[4], permutation[5]

			return subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Assigned(room, subjectProfessor, group)
		},
		// Fits(k, r) = 1
		func(permutation []uint64) bool {
			group, room := permutation[4], permutation[5]

			return group == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Fits(group, room)
		},
	})

	clauses := make([][]int64, 0)

	for _, permutation := range permutations {
		period, day, lesson, subjectProfessor, group, room := permutation[0], permutation[1], permutation[2], permutation[3], permutation[4], permutation[5]

		index := state.indexer.Index(period, day, lesson, subjectProfessor, group, room)

		clauses = append(clauses, []int64{-int64(index)})
	}

	return clauses
}

// TODO: (Optional) This method can be performance-optimized by a triple for loop instead of going through all permutations
func uniquenessConstraints(state constraintState) [][]int64 {
	permutations := state.generator.ConstrainedPermutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 1
		func(permutation []uint64) bool {
			lesson, subjectProfessor, group := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Teaches(group, subjectProfessor, lesson)
		},
		// Allowed(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor, group := permutation[0], permutation[1], permutation[3], permutation[4]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Allowed(subjectProfessor, group, day, period)
		},
		// ProfessorAvailable(i, d, t) = 1
		func(permutation []uint64) bool {
			period, day, subjectProfessor := permutation[0], permutation[1], permutation[3]

			return period == math.MaxUint64 ||
				day == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.ProfessorAvailable(subjectProfessor, day, period)
		},
		// Assigned(r, i) = 1
		func(permutation []uint64) bool {
			subjectProfessor, group, room := permutation[3], permutation[4], permutation[5]

			return subjectProfessor == math.MaxUint64 ||
				group == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Assigned(room, subjectProfessor, group)
		},
		// Fits(k, r) = 1
		func(permutation []uint64) bool {
			group, room := permutation[4], permutation[5]

			return group == math.MaxUint64 ||
				room == math.MaxUint64 ||

				// Actual predicate
				state.evaluator.Fits(group, room)
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
				index1 := state.indexer.Index(period1, day1, lesson1, subjectProfessor1, group1, room1)
				index2 := state.indexer.Index(period2, day2, lesson2, subjectProfessor2, group2, room2)
				clauses = append(clauses, []int64{-int64(index1), -int64(index2)})
			}
		}
	}

	return clauses
}
