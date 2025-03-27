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

func NewPredicateEvaluator(
	availability map[uint64][][]bool,
	rooms map[uint64]uint64,
	professors map[uint64]uint64,
	curriculum [][]uint64,
	lessons uint64,
	subjectProfessors uint64,
) PredicateEvaluator {

	return newMatrixPredicateEvaluator(
		availability,
		rooms,
		professors,
		curriculum,
		lessons,
		subjectProfessors,
	)
}
