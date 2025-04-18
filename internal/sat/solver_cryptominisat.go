package sat

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

const cryptominisatPath = "cryptominisat"

type cryptominisatSolver struct{}

func NewCryptominisatSolver() SATSolver {
	return &cryptominisatSolver{}
}

func (solver *cryptominisatSolver) Solve(sat *SAT) (SATSolution, error) {
	dimacs := sat.ToDIMACS() // Transform SAT into DIMACS-CNF string format

	cmd := exec.Command(cryptominisatPath, "--verb", "0")
	cmd.Stdin = strings.NewReader(dimacs) // Feed dimacs into cryptominisat's standard input

	var stdOut bytes.Buffer
	cmd.Stdout = &stdOut
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	// Exit-code of 10 stands for satisfiable and exit-code 20 stands for unsatisfiable
	if err != nil && cmd.ProcessState.ExitCode() != 10 && cmd.ProcessState.ExitCode() != 20 {
		return nil, fmt.Errorf("an occurred during cryptominisat execution: %v : %v", err.Error(), stderr.String())
	} else if cmd.ProcessState.ExitCode() == 20 {
		return nil, nil
	}

	out := stdOut.String()

	return parseSolution(out), nil
}
