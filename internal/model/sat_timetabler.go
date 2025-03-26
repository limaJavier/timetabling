package model

import (
	"slices"
	"timetabling/internal/sat"

	"github.com/samber/lo"
)

type satTimetabler struct {
	evaluator    PredicateEvaluator
	indexer      Indexer
	permutations func(constraints []func(permutation []uint64) bool) [][]uint64
	solver       sat.SATSolver
}

func newSatTimetabler(solver sat.SATSolver) *satTimetabler {
	return &satTimetabler{
		solver: solver,
	}
}

func (timetabler *satTimetabler) Build(
	curriculum [][]uint64,
	availability map[uint64][][]bool,
	rooms map[uint64]uint64,
	professors map[uint64]uint64,
) (sat.SATSolution, error) {

	periods := uint64(len(availability[0]))
	days := uint64(len(availability[0][0]))
	lessons := lo.Reduce(curriculum, func(max uint64, row []uint64, _ int) uint64 {
		current := lo.Max(row)
		if current > max {
			return current
		}
		return max
	}, 0)
	subjectTeachers := uint64(len(curriculum[0]))
	classes := uint64(len(curriculum))

	timetabler.evaluator = NewPredicateEvaluator(
		availability,
		rooms,
		professors,
		curriculum,
		lessons,
		subjectTeachers,
	)
	timetabler.indexer = NewIndexer(periods, days, lessons, subjectTeachers, classes)
	timetabler.permutations = MakeConstrainedPermutations(periods, days, lessons, subjectTeachers, classes)

	satInstance := sat.SAT{
		Variables: periods * days * lessons * subjectTeachers * classes,
		Clauses:   [][]int64{},
	}

	satInstance.Clauses = append(satInstance.Clauses, timetabler.professorConstraints()...)

	solution, err := timetabler.solver.Solve(satInstance)
	if err != nil {
		return nil, err
	}
	return solution, nil
}

func (timetabler *satTimetabler) professorConstraints() [][]int64 {
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
