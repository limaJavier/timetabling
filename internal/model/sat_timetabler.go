package model

import (
	"fmt"
	"math"
	"slices"
	"timetabling/internal/sat"

	"github.com/samber/lo"
)

type satTimetabler struct {
	//** Dependencies
	evaluator    PredicateEvaluator
	indexer      Indexer
	permutations func(constraints []func(permutation []uint64) bool) [][]uint64 // TODO: (Refactor) Instead of a function this should be an interface ConstrainedPermutator to ensure the permutation-contract from an interface level
	solver       sat.SATSolver

	groups     map[uint64][][]uint64

	periods           uint64
	days              uint64
	lessons           uint64
	subjectProfessors uint64
	classes           uint64
}

func newSatTimetabler(solver sat.SATSolver) *satTimetabler {
	return &satTimetabler{
		solver: solver,
	}
}

func (timetabler *satTimetabler) Build(
	curriculum [][]uint64,
	groups map[uint64][][]uint64, // TODO: (Optional) Consider lessons as well
	availability map[uint64][][]bool,
	rooms map[uint64]uint64,
	professors map[uint64]uint64,
) ([][5]uint64, error) {
	//** Extract attributes's domains
	timetabler.getAttributes(curriculum, availability)

	//** Add singleton groups for completeness
	timetabler.addSingletonGroups(curriculum, groups)
	timetabler.groups = groups

	//** Initialize dependencies
	timetabler.evaluator = NewPredicateEvaluator(
		curriculum,
		groups,
		availability,
		rooms,
		professors,
		timetabler.lessons,
		timetabler.subjectProfessors,
	)
	timetabler.indexer = NewIndexer(timetabler.periods, timetabler.days, timetabler.lessons, timetabler.subjectProfessors, timetabler.classes)
	timetabler.permutations = MakeConstrainedPermutations(timetabler.periods, timetabler.days, timetabler.lessons, timetabler.subjectProfessors, timetabler.classes)

	//** Build SAT instance
	satInstance := sat.SAT{
		Variables: timetabler.periods * timetabler.days * timetabler.lessons * timetabler.subjectProfessors * timetabler.classes,
		Clauses:   [][]int64{},
	}

	// TODO: To improve performance each set of clauses could be built in a different goroutine.
	satInstance.Clauses = append(satInstance.Clauses, timetabler.professorConstraints()...)
	satInstance.Clauses = append(satInstance.Clauses, timetabler.studentConstraints()...)
	satInstance.Clauses = append(satInstance.Clauses, timetabler.professorAvailabilityConstraints()...)
	satInstance.Clauses = append(satInstance.Clauses, timetabler.roomCompatibilityConstraints()...)
	satInstance.Clauses = append(satInstance.Clauses, timetabler.completenessConstraints()...)
	satInstance.Clauses = append(satInstance.Clauses, timetabler.negationConstraints()...)
	satInstance.Clauses = append(satInstance.Clauses, timetabler.uniquenessConstraints()...)

	//** Solve SAT instance
	solution, err := timetabler.solver.Solve(satInstance)
	if err != nil {
		return nil, err
	} else if solution == nil { // Return nil if the SAT instance is not satisfiable
		return nil, nil
	}

	present := func(variable int64) bool {
		return lo.SomeBy(satInstance.Clauses, func(clause []int64) bool { return slices.Contains(clause, variable) })
	}

	timetable := [][5]uint64{}
	for _, variable := range solution {
		if variable > 0 && present(variable) {
			positive := [5]uint64{}
			positive[0], positive[1], positive[2], positive[3], positive[4] = timetabler.indexer.Attributes(uint64(variable))
			timetable = append(timetable, positive)
		}
	}

	return timetable, nil
}

func (timetabler *satTimetabler) Verify(
	timetable [][5]uint64,
	curriculum [][]uint64,
	availability map[uint64][][]bool,
	rooms map[uint64]uint64,
	professors map[uint64]uint64,
) bool {

	//** Extract attributes's domains
	timetabler.getAttributes(curriculum, availability)

	//** Initialize derived-curriculum
	derivedCurriculum := make([][]uint64, timetabler.classes)
	for i := range derivedCurriculum {
		derivedCurriculum[i] = make([]uint64, timetabler.subjectProfessors)
	}

	//** Initialize professor-assistance
	professorAssistance := make(map[uint64][][]bool, 0)
	for professor := range len(professors) {
		professorAssistance[uint64(professor)] = make([][]bool, timetabler.periods)
		for i := range professorAssistance[uint64(professor)] {
			professorAssistance[uint64(professor)][i] = make([]bool, timetabler.days)
		}
	}

	//** Initialize class-assistance
	classAssistance := make(map[uint64][][]bool, 0)
	for class := range timetabler.classes {
		classAssistance[class] = make([][]bool, timetabler.periods)
		for i := range classAssistance[class] {
			classAssistance[class][i] = make([]bool, timetabler.days)
		}
	}

	for _, positive := range timetable {
		period, day, subjectProfessor, class := positive[0], positive[1], positive[3], positive[4]
		professor := professors[subjectProfessor]

		// Check professor is actually available for that period and day, and that he/she have not assisted already. Check as well for previous class assistance
		if !availability[professor][period][day] || professorAssistance[professor][period][day] || classAssistance[class][period][day] {
			return false
		}

		professorAssistance[professor][period][day] = true // Store professor assistance
		classAssistance[class][period][day] = true         // Store class assistance
		derivedCurriculum[class][subjectProfessor]++       // Store lesson taught
	}

	// Check that curriculum and derivedCurriculum are the same
	return !lo.SomeBy(
		lo.Zip2(curriculum, derivedCurriculum),
		func(rows lo.Tuple2[[]uint64, []uint64]) bool {
			return !slices.Equal(rows.A, rows.B)
		},
	)
}

func (timetabler *satTimetabler) professorConstraints() [][]int64 {
	permutations := timetabler.permutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 1
		func(permutation []uint64) bool {
			lesson, subjectProfessor, class := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				class == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Teaches(class, subjectProfessor, lesson)
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
	})

	clauses := make([][]int64, 0, len(permutations)*len(permutations))

	// Due to the nature of the iteration process we're are certain that we won't find the case where: k = k', i = i', j = j', d = d', t = t'
	for i := range len(permutations) - 1 {
		for j := i + 1; j < len(permutations); j++ {
			permutation1, permutation2 := permutations[i], permutations[j]
			period1, period2 := permutation1[0], permutation2[0]
			day1, day2 := permutation1[1], permutation2[1]
			subjectProfessor1, subjectProfessor2 := permutation1[3], permutation2[3]
			class1, class2 := permutation1[4], permutation2[4]

			// d == d', t == t', SameProfessor(i, i') = 1, !Together(k1, k2, i) (where i == i')
			if period1 == period2 && day1 == day2 && timetabler.evaluator.SameProfessor(subjectProfessor1, subjectProfessor2) && (subjectProfessor1 != subjectProfessor2 || !timetabler.evaluator.Together(subjectProfessor1, class1, class2)) {
				lesson1 := permutation1[2]
				lesson2 := permutation2[2]

				index1 := timetabler.indexer.Index(period1, day1, lesson1, subjectProfessor1, class1)
				index2 := timetabler.indexer.Index(period2, day2, lesson2, subjectProfessor2, class2)
				clauses = append(clauses, []int64{-int64(index1), -int64(index2)})
			}
		}
	}

	return clauses
}

func (timetabler *satTimetabler) studentConstraints() [][]int64 {
	permutations := timetabler.permutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 1
		func(permutation []uint64) bool {
			lesson, subjectProfessor, class := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				class == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Teaches(class, subjectProfessor, lesson)
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
	})

	clauses := make([][]int64, 0, len(permutations)*len(permutations))

	for i := range len(permutations) - 1 {
		for j := i + 1; j < len(permutations); j++ {
			permutation1, permutation2 := permutations[i], permutations[j]
			period1, period2 := permutation1[0], permutation2[0]
			day1, day2 := permutation1[1], permutation2[1]
			subjectProfessor1, subjectProfessor2 := permutation1[3], permutation2[3]
			class1, class2 := permutation1[4], permutation2[4]

			// k = k', d = d', t = t', SameProfessor(i, i') = 0
			if class1 == class2 && period1 == period2 && day1 == day2 && !timetabler.evaluator.SameProfessor(subjectProfessor1, subjectProfessor2) {
				lesson1 := permutation1[2]
				lesson2 := permutation2[2]

				index1 := timetabler.indexer.Index(period1, day1, lesson1, subjectProfessor1, class1)
				index2 := timetabler.indexer.Index(period2, day2, lesson2, subjectProfessor2, class2)
				clauses = append(clauses, []int64{-int64(index1), -int64(index2)})
			}
		}
	}

	return clauses
}

func (timetabler *satTimetabler) professorAvailabilityConstraints() [][]int64 {
	permutations := timetabler.permutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 1
		func(permutation []uint64) bool {
			lesson, subjectProfessor, class := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				class == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Teaches(class, subjectProfessor, lesson)
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

	clauses := make([][]int64, 0, len(permutations)*len(permutations))

	for _, permutation := range permutations {
		period, day, lesson, subjectProfessor, class := permutation[0], permutation[1], permutation[2], permutation[3], permutation[4]

		index := timetabler.indexer.Index(period, day, lesson, subjectProfessor, class)

		clauses = append(clauses, []int64{-int64(index)})
	}

	return clauses
}

func (timetabler *satTimetabler) roomCompatibilityConstraints() [][]int64 {
	permutations := timetabler.permutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 1
		func(permutation []uint64) bool {
			lesson, subjectProfessor, class := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				class == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Teaches(class, subjectProfessor, lesson)
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
	})

	clauses := make([][]int64, 0, len(permutations)*len(permutations))

	// Due to the nature of the iteration we're are certain that we won't find the case where: k = k', i = i', j = j', d = d', t = t'
	for i := range len(permutations) - 1 {
		for j := i + 1; j < len(permutations); j++ {
			permutation1, permutation2 := permutations[i], permutations[j]
			period1, period2 := permutation1[0], permutation2[0]
			day1, day2 := permutation1[1], permutation2[1]
			subjectProfessor1, subjectProfessor2 := permutation1[3], permutation2[3]
			class1, class2 := permutation1[4], permutation2[4]

			// d == d', t == t', SameRoom(i, i') = 1, !Together(k1, k2, i) (where i == i')
			if period1 == period2 && day1 == day2 && timetabler.evaluator.SameRoom(subjectProfessor1, subjectProfessor2) && (subjectProfessor1 != subjectProfessor2 || !timetabler.evaluator.Together(subjectProfessor1, class1, class2)) {
				lesson1 := permutation1[2]
				lesson2 := permutation2[2]

				index1 := timetabler.indexer.Index(period1, day1, lesson1, subjectProfessor1, class1)
				index2 := timetabler.indexer.Index(period2, day2, lesson2, subjectProfessor2, class2)
				clauses = append(clauses, []int64{-int64(index1), -int64(index2)})
			}
		}
	}

	return clauses
}

func (timetabler *satTimetabler) completenessConstraints() [][]int64 {
	// <Lesson, SubjectProfessor, Class> triplets
	triplets := make([][3]uint64, 0)
	_ = timetabler.permutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 1
		func(permutation []uint64) bool {
			lesson, subjectProfessor, class := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				class == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Teaches(class, subjectProfessor, lesson)
		},
		// According to how ConstrainedPermutations works this predicate will be evaluated only if the previous one evaluates to true, therefore any triplet that reaches it will be a valid one
		func(permutation []uint64) bool {
			lesson, subjectProfessor, class := permutation[2], permutation[3], permutation[4]
			triplet := [3]uint64{lesson, subjectProfessor, class}

			if lesson != math.MaxUint64 && subjectProfessor != math.MaxUint64 && class != math.MaxUint64 {
				triplets = append(triplets, triplet)
			}
			return true // Always return true since class will be the last attribute to fill during backtracking, so there will be no further ado
		},
	})

	clauses := make([][]int64, 0, len(triplets)*int(timetabler.periods)*int(timetabler.days))

	for _, triplet := range triplets {
		lesson, subjectProfessor, class := triplet[0], triplet[1], triplet[2]
		clause := []int64{}
					for period := range timetabler.periods {
						for day := range timetabler.days {
				// ProfessorAvailable(i, d, t) = 1
							if timetabler.evaluator.ProfessorAvailable(subjectProfessor, day, period) {
								index := timetabler.indexer.Index(period, day, lesson, subjectProfessor, class)
					clause = append(clause, int64(index))
				}
			}
		}
		clauses = append(clauses, clause)
	}

	return clauses
}

func (timetabler *satTimetabler) negationConstraints() [][]int64 {
	permutations := timetabler.permutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 0
		func(permutation []uint64) bool {
			lesson, subjectProfessor, class := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				class == math.MaxUint64 ||

				// Actual predicate
				!timetabler.evaluator.Teaches(class, subjectProfessor, lesson)
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
	})

	clauses := make([][]int64, 0, len(permutations)*len(permutations))

	for _, permutation := range permutations {
		period, day, lesson, subjectProfessor, class := permutation[0], permutation[1], permutation[2], permutation[3], permutation[4]

		index := timetabler.indexer.Index(period, day, lesson, subjectProfessor, class)

		clauses = append(clauses, []int64{-int64(index)})
	}

	return clauses
}

// TODO: This method can be performance-optimized by a triple for loop instead of going through all permutations
func (timetabler *satTimetabler) uniquenessConstraints() [][]int64 {
	permutations := timetabler.permutations([]func(permutation []uint64) bool{
		// A_k(i,j) = 1
		func(permutation []uint64) bool {
			lesson, subjectProfessor, class := permutation[2], permutation[3], permutation[4]

			return lesson == math.MaxUint64 ||
				subjectProfessor == math.MaxUint64 ||
				class == math.MaxUint64 ||

				// Actual predicate
				timetabler.evaluator.Teaches(class, subjectProfessor, lesson)
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
	})

	clauses := make([][]int64, 0, len(permutations)*len(permutations))

	// Due to the nature of the iteration we're are certain that we won't find the case where: k = k', i = i', j = j', d = d', t = t'
	for i := range len(permutations) - 1 {
		for j := i + 1; j < len(permutations); j++ {
			permutation1, permutation2 := permutations[i], permutations[j]
			period1, day1, lesson1, subjectProfessor1, class1 := permutation1[0], permutation1[1], permutation1[2], permutation1[3], permutation1[4]
			period2, day2, lesson2, subjectProfessor2, class2 := permutation2[0], permutation2[1], permutation2[2], permutation2[3], permutation2[4]

			// k == k', i == i', j == j'
			if class1 == class2 && subjectProfessor1 == subjectProfessor2 && lesson1 == lesson2 {
				index1 := timetabler.indexer.Index(period1, day1, lesson1, subjectProfessor1, class1)
				index2 := timetabler.indexer.Index(period2, day2, lesson2, subjectProfessor2, class2)
				clauses = append(clauses, []int64{-int64(index1), -int64(index2)})
			}
		}
	}

	return clauses
}

func (timetabler *satTimetabler) getAttributes(curriculum [][]uint64, availability map[uint64][][]bool) {
	timetabler.periods = uint64(len(availability[0]))
	timetabler.days = uint64(len(availability[0][0]))
	timetabler.lessons = lo.Reduce(curriculum, func(max uint64, row []uint64, _ int) uint64 {
		current := lo.Max(row)
		if current > max {
			return current
		}
		return max
	}, 0)
	timetabler.subjectProfessors = uint64(len(curriculum[0]))
	timetabler.classes = uint64(len(curriculum))
}

func (timetabler *satTimetabler) addSingletonGroups(curriculum [][]uint64, groups map[uint64][][]uint64) {
	for class := range timetabler.classes {
		for subjectProfessor := range timetabler.subjectProfessors {
			if curriculum[class][subjectProfessor] > 0 {
				contained := false

				for _, group := range groups[subjectProfessor] {
					if slices.Contains(group, class) {
						contained = true
						break
					}
				}

				if !contained {
					groups[subjectProfessor] = append(groups[subjectProfessor], []uint64{class})
				}
			}
		}
	}
}

