package sat

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/samber/lo"
)

const glucoseSimpPath = "glucose-simp"

type glucoseSimpSolver struct{}

func NewGlucoseSimpSolver() SATSolver {
	return &glucoseSimpSolver{}
}

func (solver *glucoseSimpSolver) Solve(sat SAT) (SATSolution, error) {
	dimacs := sat.ToDIMACS() // Transform SAT into DIMACS-CNF string format

	// Create a temporary file to hold the DIMACS content
	inputTempFile, err := os.CreateTemp("./", "dimacs-*.cnf")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(inputTempFile.Name()) // Ensure the file is removed after execution

	outputTempFile, err := os.CreateTemp("./", "glucose_simp_output-*.cnf")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(outputTempFile.Name()) // Ensure the file is removed after execution

	// Write the DIMACS content to the temporary file
	if _, err := inputTempFile.WriteString(dimacs); err != nil {
		return nil, fmt.Errorf("failed to write DIMACS to temporary file: %v", err)
	}
	if err := inputTempFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temporary file: %v", err)
	}

	cmd := exec.Command(glucoseSimpPath, "-verb=0")
	// Set the temporary file as the input for the command
	cmd.Args = append(cmd.Args, inputTempFile.Name(), outputTempFile.Name())
	cmd.Stdin = strings.NewReader(dimacs) // Feed dimacs into minisat's standard input

	var stdOut bytes.Buffer
	cmd.Stdout = &stdOut
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Exit-code of 10 stands for satisfiable and exit-code 20 stands for unsatisfiable
	err = cmd.Run()
	if err != nil && cmd.ProcessState.ExitCode() != 10 && cmd.ProcessState.ExitCode() != 20 {
		return nil, fmt.Errorf("an occurred during glucose-simp execution: %v : %v", err.Error(), stderr.String())
	} else if cmd.ProcessState.ExitCode() == 20 {
		return nil, nil
	}

	output, err := io.ReadAll(outputTempFile) // Read the output file
	if err != nil {
		return nil, fmt.Errorf("failed to read output file: %v", err)
	}
	return solver.parseSolution(string(output)), nil
}

func (solver *glucoseSimpSolver) parseSolution(solverOutput string) SATSolution {
	solution := lo.Map(strings.Fields(solverOutput), func(valueStr string, _ int) int64 {
		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil && valueStr != "" {
			log.Panicf("invalid literal in solver output: %v", err)
		}
		return value
	})
	return solution
}
