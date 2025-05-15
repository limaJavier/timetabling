package model

type predicateEvaluator interface {
	// Checks whether the subjectProfessor1 and subjectProfessor2 share the same professor
	SameProfessor(subjectProfessor1, subjectProfessor2 uint64) bool

	// Checks whether the subjectProfessor is available to teach at the given day and period
	ProfessorAvailable(subjectProfessor, day, period uint64) bool

	// Checks whether the subjectProfessor teaches the lesson to the group
	Teaches(group, subjectProfessor, lesson uint64) bool

	// Checks whether group1 and group2 do not share any common class (they're disjoint)
	Disjoint(group1, group2 uint64) bool

	// Checks whether subjectProfessor is allowed to teach (or be taught) on that given period and day
	Allowed(subjectProfessor, group, day, period uint64) bool

	// Checks whether the room is assigned to the subjectProfessor
	Assigned(room, subjectProfessor, group uint64) bool

	// Checks whether the group's size is smaller than or equal to the room's capacity (i.e. the group fits in the room)
	Fits(group, room uint64) bool

	// Checks whether two subject-professor are similar according to their assigned-rooms
	RoomSimilar(subjectProfessor1, subjectProfessor2, group1, group2 uint64) bool
}
