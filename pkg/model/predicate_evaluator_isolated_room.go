package model

type predicateEvaluatorIsolatedRoom struct {
	e predicateEvaluator
}

func newPredicateEvaluatorIsolatedRoom(modelInput ModelInput, roomSimilarityThreshold float32) predicateEvaluator {
	evaluator := predicateEvaluatorIsolatedRoom{
		e: newPredicateEvaluator(modelInput, roomSimilarityThreshold),
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

func (evaluator *predicateEvaluatorIsolatedRoom) Allowed(subjectProfessor, group, day, period uint64) bool {
	return evaluator.e.Allowed(subjectProfessor, group, day, period)
}

func (evaluator *predicateEvaluatorIsolatedRoom) Assigned(room, subjectProfessor, group uint64) bool {
	return true
}

func (evaluator *predicateEvaluatorIsolatedRoom) Fits(group, room uint64) bool {
	return true
}

func (evaluator *predicateEvaluatorIsolatedRoom) RoomSimilar(subjectProfessor1, subjectProfessor2, group1, group2 uint64) bool {
	return evaluator.e.RoomSimilar(subjectProfessor1, subjectProfessor2, group1, group2)
}
