package model

import (
	"slices"
	"timetabling/internal/sat"
)

type satTimetabler struct {
	evaluator    PredicateEvaluator
	indexer      Indexer
	permutations func(constraints []func(permutation []uint64) bool) [][]uint64
}

func newSatTimetabler() *satTimetabler {
	panic("not implemented")
}

func (timetabler *satTimetabler) Build() sat.SATSolution {
	panic("not implemented")
}

// TODO: This method should be private
func (timetabler *satTimetabler) ProfessorConstraints() [][]int64 {
	permutations := timetabler.permutations([]func(permutation []uint64) bool{
		func(permutation []uint64) bool {
			lesson, subjectProfessor, class := permutation[2], permutation[3], permutation[4]
			return timetabler.evaluator.Teaches(class, subjectProfessor, lesson)
		},
		func(permutation []uint64) bool {
			period, day, subjectProfessor := permutation[0], permutation[1], permutation[3]
			return timetabler.evaluator.ProfessorAvailable(subjectProfessor, day, period)
		},
	})

	clauses := make([][]int64, 0, len(permutations)*len(permutations))

	for i := range len(permutations) - 1 {
		for j := i + 1; j < len(permutations); j++ {
			permutation1, permutation2 := permutations[i], permutations[j]
			subjectProfessor1, subjectProfessor2 := permutation1[3], permutation2[3]

			if !slices.Equal(permutation1, permutation2) && timetabler.evaluator.SameProfessor(subjectProfessor1, subjectProfessor2) {
				period1, day1, lesson1, class1 := permutation1[0], permutation1[1], permutation1[2], permutation1[4]
				period2, day2, lesson2, class2 := permutation2[0], permutation2[1], permutation2[2], permutation2[4]

				index1 := timetabler.indexer.Index(period1, day1, lesson1, subjectProfessor1, class1)
				index2 := timetabler.indexer.Index(period2, day2, lesson2, subjectProfessor2, class2)
				clauses = append(clauses, []int64{-int64(index1), -int64(index2)})
			}
		}
	}

	return clauses
}
