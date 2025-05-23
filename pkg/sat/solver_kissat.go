package sat

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

var kissatPath = Config["kissatPath"]

type kissatSolver struct{}

func NewKissatSolver() SATSolver {
	return &kissatSolver{}
}

func (solver *kissatSolver) Solve(sat SAT) (SATSolution, error) {
	dimacs := sat.ToDIMACS() // Transform SAT into DIMACS-CNF string format

	cmd := exec.Command(kissatPath, "-q", "--relaxed")
	cmd.Stdin = strings.NewReader(dimacs) // Feed dimacs into kissat's standard input

	var stdOut bytes.Buffer
	cmd.Stdout = &stdOut
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	// Exit-code of 10 stands for satisfiable and exit-code 20 stands for unsatisfiable
	if err != nil && cmd.ProcessState.ExitCode() != 10 && cmd.ProcessState.ExitCode() != 20 {
		return nil, fmt.Errorf("an occurred during kissat execution: %v : %v", err.Error(), stderr.String())
	} else if cmd.ProcessState.ExitCode() == 20 {
		return nil, nil
	}

	return parseSolution(stdOut.String()), nil
}
