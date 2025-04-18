package sat

import (
	"fmt"
	"strings"
)

type SATSolution []int64

type SAT struct {
	Variables uint64
	Clauses   [][]int64
}

func (s SAT) ToDIMACS() string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "p cnf %d %d\n", s.Variables, len(s.Clauses))
	for _, clause := range s.Clauses {
		for _, literal := range clause {
			fmt.Fprintf(&builder, "%d ", literal)
		}
		builder.WriteString("0\n")
	}
	return builder.String()
}
