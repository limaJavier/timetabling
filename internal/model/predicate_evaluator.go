package model

type PredicateEvaluator interface {
	// Checks whether the subjectProfessor1 and subjectProfessor2 share the same professor
	SameProfessor(subjectProfessor1, subjectProfessor2 uint64) bool

	// Checks whether the subjectProfessor is available to teach at the given day and period
	ProfessorAvailable(subjectProfessor, day, period uint64) bool

	// Checks whether the subjectProfessor1 and subjectProfessor2 have the same assigned room
	SameRoom(subjectProfessor1, subjectProfessor2 uint64) bool

	// Checks whether the subjectProfessor teaches the lesson to the class
	Teaches(class, subjectProfessor, lesson uint64) bool
}

func NewPredicateEvaluator() PredicateEvaluator {
	// return newMatrixPredicateEvaluator()
	panic("not implemented")
}
