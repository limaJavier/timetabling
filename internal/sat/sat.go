package sat

import "fmt"

type SATSolution []int64

type SAT struct {
	Variables uint64
	Clauses   [][]int64
}

func (s *SAT) ToDIMACS() string {
	var dimacs string
	dimacs += fmt.Sprintf("p cnf %d %d\n", s.Variables, len(s.Clauses))
	for _, clause := range s.Clauses {
		for _, literal := range clause {
			dimacs += fmt.Sprintf("%d ", literal)
		}
		dimacs += "0\n"
	}
	return dimacs
}
