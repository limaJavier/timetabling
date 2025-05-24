package sat

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

type slimeSolver struct{}

func NewSlimeSolver() SATSolver {
	return &slimeSolver{}
}

func (solver *slimeSolver) Solve(sat SAT) (SATSolution, error) {
	slimePath := getExecutablePath("slimePath")
	dimacs := sat.ToDIMACS() // Transform SAT into DIMACS-CNF string format

	// Create a temporary file to hold the DIMACS content
	tmpFile, err := os.CreateTemp("./", "dimacs-*.cnf")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name()) // Ensure the file is removed after execution

	// Write the DIMACS content to the temporary file
	if _, err := tmpFile.WriteString(dimacs); err != nil {
		return nil, fmt.Errorf("failed to write DIMACS to temporary file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temporary file: %v", err)
	}

	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			fmt.Printf("warning: failed to remove temporary file %s: %v\n", tmpFile.Name(), err)
		}
	}()

	cmd := exec.Command(slimePath)
	// Set the temporary file as the input for the command
	cmd.Args = append(cmd.Args, tmpFile.Name())

	var stdOut bytes.Buffer
	cmd.Stdout = &stdOut
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Exit-code of 10 stands for satisfiable and exit-code 20 stands for unsatisfiable
	err = cmd.Run()
	if err != nil && cmd.ProcessState.ExitCode() != 10 && cmd.ProcessState.ExitCode() != 20 {
		return nil, fmt.Errorf("an occurred during slime execution: %v : %v", err.Error(), stderr.String())
	} else if cmd.ProcessState.ExitCode() == 20 {
		return nil, nil
	}

	return parseSolution(stdOut.String()), nil
}
