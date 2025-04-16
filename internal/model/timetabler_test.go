package model

import (
	"log"
	"os"
	"testing"
	"timetabling/internal/sat"

	"github.com/stretchr/testify/assert"
)

const TestDirectory = "../../test/out/"

func TestBuild(t *testing.T) {
	testFiles, err := os.ReadDir(TestDirectory)
	if err != nil {
		log.Fatalf("cannot read directory: %v", err)
	}

	for _, file := range testFiles {
		//** Arrange
		filename := TestDirectory + file.Name()
		preprocessor := NewPreprocessor()
		input, err := InputFromJson(filename)
		if err != nil {
			log.Fatalf("cannot parse input file: %v", err)
		}

		groupsPerSubjectProfessor, lessons, rooms, professors, permissibility, availability := input.GetGroupsPerSubjectProfessors(), input.GetLessons(), input.GetRooms(), input.GetProfessors(), input.GetPermissibility(), input.GetAvailability()

		curriculum, groups := preprocessor.ExtractCurriculumAndGroups(groupsPerSubjectProfessor)

		groupsGraph := preprocessor.BuildGroupsGraph(groups)

		solver := sat.NewKissatSolver()
		timetabler := NewTimetabler(solver)

		//** Act
		timetable, err := timetabler.Build(curriculum, groupsGraph, lessons, permissibility, availability, rooms, professors)

		//** Assert
		assert.Nil(t, err)
		assert.NotNil(t, timetable)
		assert.True(t, timetabler.Verify(timetable, curriculum, groupsGraph, lessons, permissibility, availability, rooms, professors, groupsPerSubjectProfessor))
	}
}
