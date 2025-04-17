package sat

import (
	"log"
	"strconv"
	"strings"

	"github.com/samber/lo"
)

type SATSolver interface {
	Solve(SAT) (SATSolution, error) // Returns a solution of the SAT instance if satisfiable, else returns nil (these are valid outputs where error shall be nil)
}

func ParseSolution(solverOutput string) SATSolution {
	values := lo.Map(
		lo.Reduce(
			lo.Filter(strings.Split(solverOutput, "\n"), func(line string, _ int) bool {
				return len(line) > 0 && line[0] == 'v'
			}),
			func(values []string, line string, _ int) []string {
				return append(values, strings.Split(line[2:], " ")...)
			},
			[]string{},
		),
		func(valueStr string, _ int) int64 {
			value, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				log.Panicf("invalid literal in kissat output: %v", err)
			}
			return value
		},
	)
	return values[:len(values)-1]
}
