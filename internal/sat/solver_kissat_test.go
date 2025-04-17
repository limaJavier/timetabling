package sat

import (
	"log"
	"math/rand/v2"
	"testing"
)

func TestKissatSatisfiable(t *testing.T) {
	solver := NewKissatSolver()
	unsatisfiableCount := 0

	for range 10 {
		literals := uint64(rand.IntN(100) + 1)
		clauses := rand.IntN(200) + 1
		instance := GenerateSATInstance(literals, clauses)

		solution, err := solver.Solve(instance)
		if err != nil {
			t.Errorf("an error occurred while solving a SAT instance: %v", err)
		}

		if len(solution) == 0 {
			unsatisfiableCount++
			continue
		}

		if !AssertSATSolution(instance, solution) {
			t.Error("Wrong answer")
		}
	}

	log.Printf("Unsatisfiable instances: %v", unsatisfiableCount)
}
