package model

type predicateEvaluatorIsolatedRoom struct {
	e PredicateEvaluator
}

func NewPredicateEvaluatorIsolatedRoom(
	modelInput ModelInput,
	curriculum [][]bool,
	groups map[uint64][]uint64,
	groupsGraph [][]bool,
	roomSimilarityThreshold float32,
) PredicateEvaluator {
	evaluator := predicateEvaluatorIsolatedRoom{
		e: NewPredicateEvaluator(modelInput, curriculum, groups, groupsGraph, roomSimilarityThreshold),
	}
	return &evaluator
}

func (evaluator *predicateEvaluatorIsolatedRoom) SameProfessor(subjectProfessor1, subjectProfessor2 uint64) bool {
	return evaluator.e.SameProfessor(subjectProfessor1, subjectProfessor2)
}

func (evaluator *predicateEvaluatorIsolatedRoom) ProfessorAvailable(subjectProfessor, day, period uint64) bool {
	return evaluator.e.ProfessorAvailable(subjectProfessor, day, period)
}

func (evaluator *predicateEvaluatorIsolatedRoom) Teaches(group, subjectProfessor, lesson uint64) bool {
	return evaluator.e.Teaches(group, subjectProfessor, lesson)
}

func (evaluator *predicateEvaluatorIsolatedRoom) Disjoint(group1, group2 uint64) bool {
	return evaluator.e.Disjoint(group1, group2)
}

func (evaluator *predicateEvaluatorIsolatedRoom) Allowed(subjectProfessor, day, period uint64) bool {
	return evaluator.e.Allowed(subjectProfessor, day, period)
}

func (evaluator *predicateEvaluatorIsolatedRoom) Assigned(room, subjectProfessor uint64) bool {
	return true
}

func (evaluator *predicateEvaluatorIsolatedRoom) Fits(group, room uint64) bool {
	return true
}

func (evaluator *predicateEvaluatorIsolatedRoom) RoomSimilar(subjectProfessor1, subjectProfessor2, group1, group2 uint64) bool {
	return evaluator.e.RoomSimilar(subjectProfessor1, subjectProfessor2, group1, group2)
}
