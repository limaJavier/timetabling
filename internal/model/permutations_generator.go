package model

type permutationGenerator interface {
	// Attributes' order in the permutation parameter is the following: Period, Day, Lesson, SubjectProfessor, Group, Room.
	// All the constraints must take into account that if the value of permutation[i] (for all feasible i's) is math.MaxUint64 then the permutation is not ready to be evaluated if this evaluation involves permutation[i]
	//
	// Example:
	//
	//	generator := model.newPermutationGenerator(Periods, Days, Lessons, SubjectProfessors, Groups, Rooms)
	//
	//	permutations := generator.ConstrainedPermutations([]func(permutation []uint64) bool{
	//				func(permutation []uint64) bool {
	//	       		// Verify "permutation[1] == math.MaxUint64", since the predicate "permutation[1] == 1" relies in this index
	//					return permutation[1] == math.MaxUint64 || permutation[1] == 1
	//				},
	//			})
	ConstrainedPermutations(constraints []func(permutation []uint64) bool) [][]uint64
}

func newPermutationGenerator(periods, days, lessons, subjectProfessors, groups, rooms uint64) permutationGenerator {
	return &permutationGeneratorImplementation{periods, days, lessons, subjectProfessors, groups, rooms}
}
