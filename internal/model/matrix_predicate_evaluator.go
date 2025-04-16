package model

import (
	"slices"

	"github.com/samber/lo"
)

// Curriculum matrix
// Professor available matrix
// Allocation matrix can be built from the curriculum matrix

type matrixPredicateEvaluator struct {
	modelInput  ModelInput
	allocations map[uint64][][]bool // Allocation matrix per group
	groupsGraph [][]bool            // Groups matrix' coordinate (i, j) = true if and only if group_i and group_j have at least one class in common (i.e. it represents an undirected graph where an edge indicate that two groups share a common class). For completeness we assume that groups[i][i] = true for all i
}

func newMatrixPredicateEvaluator(
	modelInput ModelInput,
	curriculum [][]bool,
	groupsGraph [][]bool,
) *matrixPredicateEvaluator {
	subjectProfessors := uint64(len(modelInput.SubjectProfessors))
	maxLessons := lo.Max(lo.Map(modelInput.SubjectProfessors, func(subjectProfessor SubjectProfessor, _ int) uint64 {
		return subjectProfessor.Lessons
	}))

	evaluator := matrixPredicateEvaluator{
		modelInput:  modelInput,
		groupsGraph: groupsGraph,
	}

	evaluator.allocations = make(map[uint64][][]bool) // Initialize dictionary
	for group := range curriculum {                   // For each group
		evaluator.allocations[uint64(group)] = make([][]bool, subjectProfessors) // Initialize allocation per group

		for subjectProfessor := range curriculum[group] { // For each subjectProfessor
			evaluator.allocations[uint64(group)][subjectProfessor] = make([]bool, maxLessons) // Initialize subjectProfessor row
			for i := range modelInput.SubjectProfessors[subjectProfessor].Lessons {
				if curriculum[group][subjectProfessor] {
					evaluator.allocations[uint64(group)][subjectProfessor][i] = true // Set to true the first j lessons where j is the number of lessons assigned for "subjectProfessor" to teach to "group" (i.e. curriculum[group][subjectProfessor])
				}
			}
		}
	}

	return &evaluator
}

func (evaluator *matrixPredicateEvaluator) SameProfessor(subjectProfessor1, subjectProfessor2 uint64) bool {
	professor1 := evaluator.modelInput.SubjectProfessors[subjectProfessor1].Professor
	professor2 := evaluator.modelInput.SubjectProfessors[subjectProfessor2].Professor
	return professor1 == professor2
}

func (evaluator *matrixPredicateEvaluator) ProfessorAvailable(subjectProfessor, day, period uint64) bool {
	professorId := evaluator.modelInput.SubjectProfessors[subjectProfessor].Professor
	distribution := evaluator.modelInput.Professors[professorId].Availability
	return distribution[period][day]
}

func (evaluator *matrixPredicateEvaluator) SameRoom(subjectProfessor1, subjectProfessor2 uint64) bool {
	rooms1 := evaluator.modelInput.SubjectProfessors[subjectProfessor1].Rooms
	rooms2 := evaluator.modelInput.SubjectProfessors[subjectProfessor2].Rooms
	return slices.Equal(rooms1, rooms2)
}

func (evaluator *matrixPredicateEvaluator) Teaches(group, subjectProfessor, lesson uint64) bool {
	allocation, ok := evaluator.allocations[group]
	if !ok {
		panic("group not found")
	}
	return allocation[subjectProfessor][lesson]
}

func (evaluator *matrixPredicateEvaluator) Disjoint(group1, group2 uint64) bool {
	return !evaluator.groupsGraph[group1][group2]
}

func (evaluator *matrixPredicateEvaluator) Allowed(subjectProfessor, day, period uint64) bool {
	distribution := evaluator.modelInput.SubjectProfessors[subjectProfessor].Permissibility
	return distribution[period][day]
}

func (evaluator *matrixPredicateEvaluator) Assigned(room, subjectProfessor uint64) bool {
	return true
}

func (evaluator *matrixPredicateEvaluator) Fits(group, room uint64) bool {
	return true
}
