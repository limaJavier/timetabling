package model

import "math"

type permutationGeneratorImplementation struct {
	periods, days, lessons, subjectProfessors, groups uint64
}

func (generator permutationGeneratorImplementation) ConstrainedPermutations(constraints []func(permutation []uint64) bool) [][]uint64 {
	permutations := make([][]uint64, 0, generator.periods*generator.days*generator.lessons*generator.subjectProfessors*generator.groups)
	generator.constrainedPermutations(
		constraints,
		[]uint64{generator.periods, generator.days, generator.lessons, generator.subjectProfessors, generator.groups},
		0,
		[]uint64{math.MaxUint64, math.MaxUint64, math.MaxUint64, math.MaxUint64, math.MaxUint64},
		&permutations,
	)
	return permutations
}

func (generator permutationGeneratorImplementation) constrainedPermutations(
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

	for i := uint64(0); i < domains[currentDomain]; i++ {
		permutation[currentDomain] = i
		constraintViolated := false
		for _, constraint := range constraints {
			if !constraint(permutation) {
				constraintViolated = true
				break
			}
		}

		if constraintViolated {
			continue
		}

		generator.constrainedPermutations(constraints, domains, currentDomain+1, permutation, permutations)
	}

	permutation[currentDomain] = math.MaxUint64
}
