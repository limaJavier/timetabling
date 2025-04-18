package model

import (
	"log"
	"os"
	"testing"
	"timetabling/internal/sat"

	"github.com/stretchr/testify/assert"
)

func TestKissatBasedEmbeddedRoomTimetabler(t *testing.T) {
	preprocessor := NewPreprocessor()
	solver := sat.NewKissatSolver()
	timetabler := NewEmbeddedRoomTimetabler(solver)

	t.Run("Satisfiable instances", func(t *testing.T) {
		satisfiableExecution(t, preprocessor, timetabler)
	})
}

func TestCadicalBasedEmbeddedRoomTimetabler(t *testing.T) {
	preprocessor := NewPreprocessor()
	solver := sat.NewCadicalSolver()
	timetabler := NewEmbeddedRoomTimetabler(solver)

	t.Run("Satisfiable instances", func(t *testing.T) {
		satisfiableExecution(t, preprocessor, timetabler)
	})
}

func TestMinisatBasedEmbeddedRoomTimetabler(t *testing.T) {
	preprocessor := NewPreprocessor()
	solver := sat.NewCryptominisatSolver()
	timetabler := NewEmbeddedRoomTimetabler(solver)

	t.Run("Satisfiable instances", func(t *testing.T) {
		satisfiableExecution(t, preprocessor, timetabler)
	})
}

func TestCryptominisatBasedEmbeddedRoomTimetabler(t *testing.T) {
	preprocessor := NewPreprocessor()
	solver := sat.NewCryptominisatSolver()
	timetabler := NewEmbeddedRoomTimetabler(solver)

	t.Run("Satisfiable instances", func(t *testing.T) {
		satisfiableExecution(t, preprocessor, timetabler)
	})
}

func TestSlimeBasedEmbeddedRoomTimetabler(t *testing.T) {
	preprocessor := NewPreprocessor()
	solver := sat.NewSlimeSolver()
	timetabler := NewEmbeddedRoomTimetabler(solver)

	t.Run("Satisfiable instances", func(t *testing.T) {
		satisfiableExecution(t, preprocessor, timetabler)
	})
}

var satisfiableExecution = func(t *testing.T, preprocessor Preprocessor, timetabler Timetabler) {
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
		curriculum, groups := preprocessor.ExtractCurriculumAndGroups(input)
		groupsGraph := preprocessor.BuildGroupsGraph(groups)

		//** Act
		timetable, err := timetabler.Build(input, curriculum, groups, groupsGraph)

		//** Assert
		assert.Nil(t, err)
		assert.NotNil(t, timetable)
		assert.True(t, timetabler.Verify(timetable, input, curriculum, groups, groupsGraph))
	}
}
