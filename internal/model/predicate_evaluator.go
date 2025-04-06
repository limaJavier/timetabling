package model

type PredicateEvaluator interface {
	// Checks whether the subjectProfessor1 and subjectProfessor2 share the same professor
	SameProfessor(subjectProfessor1, subjectProfessor2 uint64) bool

	// Checks whether the subjectProfessor is available to teach at the given day and period
	ProfessorAvailable(subjectProfessor, day, period uint64) bool

	// Checks whether the subjectProfessor1 and subjectProfessor2 have the same assigned room
	SameRoom(subjectProfessor1, subjectProfessor2 uint64) bool

	// Checks whether the subjectProfessor teaches the lesson to the group
	Teaches(group, subjectProfessor, lesson uint64) bool

	// Checks whether group1 and group2 do not share any common class (they're disjoint)
	Disjoint(group1, group2 uint64) bool

	// Check whether subjectProfessor is allowed to teach (or be taught) on that given period and day
	Allowed(subjectProfessor, day, period uint64) bool
}

func NewPredicateEvaluator(
	curriculum [][]bool,
	groupsGraph [][]bool,
	lessons map[uint64]uint64,
	permissibility map[uint64][][]bool,
	availability map[uint64][][]bool,
	rooms map[uint64]uint64,
	professors map[uint64]uint64,
) PredicateEvaluator {

	return newMatrixPredicateEvaluator(
		curriculum,
		groupsGraph,
		lessons,
		permissibility,
		availability,
		rooms,
		professors,
	)
}
