package main

import (
	"fmt"
	"log"
	"slices"
	"timetabling/internal/model"
	"timetabling/internal/sat"
)

var Days = map[uint64]string{
	0: "Monday",
	1: "Tuesday",
	2: "Wednesday",
	3: "Thursday",
	4: "Friday",
	5: "Saturday",
	6: "Sunday",
}

func main() {
	const File string = "../test/out/1.json"
	preprocessor := model.NewPreprocessor()

	input, err := model.InputFromJson(File)
	if err != nil {
		log.Fatalf("cannot parse input file: %v", err)
	}

	groupsPerSubjectProfessor, lessons, rooms, professors, permissibility, availability := input.GetGroupsPerSubjectProfessors(), input.GetLessons(), input.GetRooms(), input.GetProfessors(), input.GetPermissibility(), input.GetAvailability()

	curriculum, groups := preprocessor.ExtractCurriculumAndGroups(groupsPerSubjectProfessor)

	groupsGraph := preprocessor.BuildGroupsGraph(groups)

	solver := sat.NewKissatSolver()
	timetabler := model.NewTimetabler(solver)

	timetable, err := timetabler.Build(curriculum, groupsGraph, lessons, permissibility, availability, rooms, professors)
	if err != nil {
		log.Fatal(err)
	} else if timetable == nil {
		fmt.Println("Not satisfiable")
		return
	}

	slices.SortFunc(timetable, func(a, b [5]uint64) int {
		if a[1] < b[1] {
			return -1
		} else if a[1] > b[1] {
			return 1
		}
		return 0
	})

	for _, positive := range timetable {
		day := Days[positive[1]]
		// day := positive[1]

		subjectProfessor := fmt.Sprintf("%v~%v", input.Metadata.SubjectProfessors[positive[3]], input.Metadata.Professors[professors[positive[3]]])
		// subjectProfessor := positive[3]

		for _, class := range groups[positive[4]] {
			fmt.Printf("Period: %v, Day: %v, Lesson: %v, SubjectProfessor: %v, Class: %v \n", positive[0], day, positive[2], subjectProfessor, class)
		}
	}

	if !timetabler.Verify(timetable, curriculum, groupsGraph, lessons, permissibility, availability, rooms, professors, groupsPerSubjectProfessor) {
		log.Fatal("Verification failed")
	}

	fmt.Println("Well done!")
}
