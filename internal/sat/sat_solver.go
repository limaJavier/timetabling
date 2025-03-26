package sat

import (
	"log"
	"strconv"
	"strings"

	"github.com/samber/lo"
)

type SATSolver interface {
	Solve(SAT) (SATSolution, error)
}

func NewKissatSolver() SATSolver {
	return &kissatSolver{}
}

func ParseSolution(solverOutput string) SATSolution {
	resultLine, ok := lo.Find(strings.Split(solverOutput, "\n"), func(line string) bool { return len(line) > 0 && line[0] == 'v' })

	if !ok {
		return nil
	} else if len(resultLine) == 3 {
		return SATSolution{}
	}

	splits := strings.Split(resultLine[2:len(resultLine)-2], " ")

	var solution SATSolution = make(SATSolution, 0, len(splits))
	lo.ForEach(splits, func(item string, _ int) {
		value, err := strconv.ParseInt(item, 10, 64)
		if item != "" && err != nil {
			log.Panicf("invalid literal in kissat output: %v", err)
		}
		solution = append(solution, value)
	})

	return solution
}
