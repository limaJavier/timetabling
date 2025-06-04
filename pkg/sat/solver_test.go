package sat

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
)

const testDirectory = "../../test/cnfs/"

func TestKissat(t *testing.T) {
	solver := NewKissatSolver()
	t.Run("Satisfiable instances", func(t *testing.T) {
		satisfiableExecution(t, solver)
	})
}

func TestCadical(t *testing.T) {
	solver := NewCadicalSolver()
	t.Run("Satisfiable instances", func(t *testing.T) {
		satisfiableExecution(t, solver)
	})
}

func TestCryptominisat(t *testing.T) {
	solver := NewCryptominisatSolver()
	t.Run("Satisfiable instances", func(t *testing.T) {
		satisfiableExecution(t, solver)
	})
}

func TestMinisat(t *testing.T) {
	solver := NewMinisatSolver()
	t.Run("Satisfiable instances", func(t *testing.T) {
		satisfiableExecution(t, solver)
	})
}

func TestSlime(t *testing.T) {
	solver := NewSlimeSolver()
	t.Run("Satisfiable instances", func(t *testing.T) {
		satisfiableExecution(t, solver)
	})
}

func TestOrtoolsat(t *testing.T) {
	solver := NewOrtoolsatSolver()
	t.Run("Satisfiable instances", func(t *testing.T) {
		satisfiableExecution(t, solver)
	})
}

func TestGlucoseSimp(t *testing.T) {
	solver := NewGlucoseSimpSolver()
	t.Run("Satisfiable instances", func(t *testing.T) {
		satisfiableExecution(t, solver)
	})
}

func TestGlucoseSyrup(t *testing.T) {
	solver := NewGlucoseSyrupSolver()
	t.Run("Satisfiable instances", func(t *testing.T) {
		satisfiableExecution(t, solver)
	})
}

func satisfiableExecution(t *testing.T, solver SATSolver) {
	testFiles, err := os.ReadDir(testDirectory)
	if err != nil {
		log.Fatalf("cannot read directory: %v", err)
	}

	for _, file := range testFiles {
		//** Arrange
		filename := testDirectory + file.Name()
		sat, err := parseDIMACSFile(filename)
		if err != nil {
			log.Fatalf("cannot parse file: %v", err)
		}

		//** Act
		solution, err := solver.Solve(sat)
		if err != nil {
			log.Fatalf("cannot solve SAT: %v", err)
		}

		//** Assert
		assertSATSolution(sat, solution)
	}
}

func parseDIMACSFile(fileName string) (SAT, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return SAT{}, fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	var sat SAT
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		// Skip comments
		if strings.HasPrefix(line, "c") {
			continue
		}
		// Problem line
		if strings.HasPrefix(line, "p cnf") {
			parts := strings.Fields(line)
			if len(parts) != 4 {
				return SAT{}, fmt.Errorf("invalid problem line: %s", line)
			}
			vars, err1 := strconv.ParseUint(parts[2], 10, 64)
			if err1 != nil {
				return SAT{}, fmt.Errorf("invalid variable count: %w", err1)
			}
			sat.Variables = vars
			continue
		}
		// Clause line
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		var clause []int64
		for _, litStr := range fields {
			lit, err := strconv.ParseInt(litStr, 10, 64)
			if err != nil {
				return SAT{}, fmt.Errorf("invalid literal '%s': %w", litStr, err)
			}
			if lit == 0 {
				break
			}
			clause = append(clause, lit)
		}
		if len(clause) > 0 {
			sat.Clauses = append(sat.Clauses, clause)
		}
	}

	if err := scanner.Err(); err != nil {
		return SAT{}, fmt.Errorf("error reading file: %w", err)
	}

	return sat, nil
}

func assertSATSolution(satInstance SAT, satSolution SATSolution) bool {
	// Make sure there are no duplicates nor contradictions
	literals := make(map[int64]bool)
	for _, literal := range satSolution {
		if literals[literal] || literals[-literal] {
			return false
		}
		literals[literal] = true
	}

	// Check that all clauses are satisfied
	for _, clause := range satInstance.Clauses {
		satisfied := false
		for _, literal := range clause {
			if literals[literal] {
				satisfied = true
				break
			}
		}
		if !satisfied {
			return false
		}
	}

	return true
}
