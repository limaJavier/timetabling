package model

import (
	"log"
	"os"
	"testing"
	"timetabling/pkg/sat"

	"github.com/stretchr/testify/assert"
)

func TestKissatBasedEmbeddedRoomTimetabler(t *testing.T) {
	solver := sat.NewKissatSolver()
	timetabler := NewEmbeddedRoomTimetabler(solver)

	t.Run("Satisfiable instances", func(t *testing.T) {
		satisfiableExecution(t, timetabler)
	})
}

func TestCadicalBasedEmbeddedRoomTimetabler(t *testing.T) {
	solver := sat.NewCadicalSolver()
	timetabler := NewEmbeddedRoomTimetabler(solver)

	t.Run("Satisfiable instances", func(t *testing.T) {
		satisfiableExecution(t, timetabler)
	})
}

func TestMinisatBasedEmbeddedRoomTimetabler(t *testing.T) {
	solver := sat.NewCryptominisatSolver()
	timetabler := NewEmbeddedRoomTimetabler(solver)

	t.Run("Satisfiable instances", func(t *testing.T) {
		satisfiableExecution(t, timetabler)
	})
}

func TestCryptominisatBasedEmbeddedRoomTimetabler(t *testing.T) {
	solver := sat.NewCryptominisatSolver()
	timetabler := NewEmbeddedRoomTimetabler(solver)

	t.Run("Satisfiable instances", func(t *testing.T) {
		satisfiableExecution(t, timetabler)
	})
}

func TestSlimeBasedEmbeddedRoomTimetabler(t *testing.T) {
	solver := sat.NewSlimeSolver()
	timetabler := NewEmbeddedRoomTimetabler(solver)

	t.Run("Satisfiable instances", func(t *testing.T) {
		satisfiableExecution(t, timetabler)
	})
}

func satisfiableExecution(t *testing.T, timetabler Timetabler) {
	testFiles, err := os.ReadDir(satisfiableTestDirectory)
	if err != nil {
		log.Fatalf("cannot read directory: %v", err)
	}

	for _, file := range testFiles {
		//** Arrange
		filename := satisfiableTestDirectory + file.Name()
		input, err := InputFromJson(filename)
		if err != nil {
			log.Fatalf("cannot parse input file: %v", err)
		}

		//** Act
		timetable, err := timetabler.Build(input)

		//** Assert
		assert.Nil(t, err)
		assert.NotNil(t, timetable)
		assert.True(t, timetabler.Verify(timetable, input))
	}
}
