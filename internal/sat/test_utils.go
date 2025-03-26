package sat

import "math/rand/v2"

func GenerateSATInstance(literals uint64, clauses int) SAT {
	satInstance := SAT{
		Variables: literals,
		Clauses:   make([][]int64, clauses),
	}

	for i := range clauses {
		satInstance.Clauses[i] = make([]int64, 0, literals)
		for j := range literals {
			if rand.Float32() < 0.5 {
				var sign int64 = 1
				if rand.Float32() < 0.5 {
					sign = -1
				}
				satInstance.Clauses[i] = append(satInstance.Clauses[i], sign*(1+int64(j)))
			}
		}

		if len(satInstance.Clauses[i]) == 0 {
			var sign int64 = 1
			if rand.Float32() < 0.5 {
				sign = -1
			}
			satInstance.Clauses[i] = append(satInstance.Clauses[i], sign*(1+rand.Int64N(int64(literals))))
		}
	}

	return satInstance
}

func AssertSATSolution(satInstance SAT, satSolution SATSolution) bool {
	// Make sure there are no duplicates nor contradictions
	literals := make(map[int64]bool)
	for _, literal := range satSolution {
		if literals[literal] || literals[-literal] {
			return false
		}
		literals[literal] = true
	}

	// Check that all clauses are satisfied
	for _, clause := range satInstance.Clauses {
		satisfied := false
		for _, literal := range clause {
			if literals[literal] {
				satisfied = true
				break
			}
		}
		if !satisfied {
			return false
		}
	}

	return true
}
