package model

// Returns a stateful function (that's bound to periods, days, lessons, subjectProfessors and classes) that builds all the variables's attributes permutations that hold the constraints.
// Attributes' order in the permutation parameter is the following: Period, Day, Lesson, SubjectProfessor, Class.
// All the constraints must take into account that if the value of permutation[i] (for all feasible i's) is zero then the permutation is not ready to be evaluated if this evaluation involves permutation[i]
//
// Example:
//
//	constrainedPermutations := model.MakeConstrainedPermutations(Periods, Days, Lessons, SubjectProfessors, Classes)
//
//	permutations := constrainedPermutations([]func(permutation []uint64) bool{
//				func(permutation []uint64) bool {
//	       		// Verify "permutation[1] == 0", since the predicate "permutation[1] == 1" relies in this index
//					return permutation[1] == 0 || permutation[1] == 1
//				},
//			})
func MakeConstrainedPermutations(periods, days, lessons, subjectProfessors, classes uint64) func(constraints []func(permutation []uint64) bool) [][]uint64 {
	return func(constraints []func(permutation []uint64) bool) [][]uint64 {
		permutations := make([][]uint64, 0, periods*days*lessons*subjectProfessors*classes)
		constrainedPermutations(
			constraints,
			[]uint64{periods, days, lessons, subjectProfessors, classes},
			0,
			make([]uint64, 5),
			&permutations,
		)
		return permutations
	}
}

func constrainedPermutations(
	constraints []func(permutation []uint64) bool,
	domains []uint64,
	currentDomain uint64,
	permutation []uint64,
	permutations *[][]uint64) {

	if currentDomain >= uint64(len(domains)) {
		permutationCopy := make([]uint64, len(permutation))
		copy(permutationCopy, permutation)
		*permutations = append(*permutations, permutationCopy)
		return
	}

	for i := uint64(1); i <= domains[currentDomain]; i++ {
		permutation[currentDomain] = i
		constraintViolated := false
		for _, constraint := range constraints {
			constraintViolated = !constraint(permutation)
		}

		if constraintViolated {
			continue
		}

		constrainedPermutations(constraints, domains, currentDomain+1, permutation, permutations)
	}

	permutation[currentDomain] = 0
}
