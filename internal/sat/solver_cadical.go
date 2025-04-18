package sat

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

const cadicalPath = "cadical"

type cadicalSolver struct{}

func NewCadicalSolver() SATSolver {
	return &cadicalSolver{}
}

func (solver *cadicalSolver) Solve(sat *SAT) (SATSolution, error) {
	dimacs := sat.ToDIMACS() // Transform SAT into DIMACS-CNF string format

	cmd := exec.Command(cadicalPath, "-q")
	cmd.Stdin = strings.NewReader(dimacs) // Feed dimacs into cadical's standard input

	var stdOut bytes.Buffer
	cmd.Stdout = &stdOut
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	// Exit-code of 10 stands for satisfiable and exit-code 20 stands for unsatisfiable
	if err != nil && cmd.ProcessState.ExitCode() != 10 && cmd.ProcessState.ExitCode() != 20 {
		return nil, fmt.Errorf("an occurred during cadical execution: %v : %v", err.Error(), stderr.String())
	} else if cmd.ProcessState.ExitCode() == 20 {
		return nil, nil
	}

	return parseSolution(stdOut.String()), nil
}
