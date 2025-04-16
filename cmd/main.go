package main

import (
	"fmt"
	"log"
	"slices"
	"strings"
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

	curriculum, groups := preprocessor.ExtractCurriculumAndGroups(input)
	groupsGraph := preprocessor.BuildGroupsGraph(groups)

	solver := sat.NewKissatSolver()
	timetabler := model.NewTimetabler(solver)

	timetable, err := timetabler.Build(input, curriculum, groupsGraph)
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
		dayId := positive[1]
		subjectProfessorId := positive[3]
		groupId := positive[4]

		dayName := Days[dayId]

		subjectProfessorName := fmt.Sprintf("%v~%v",
			input.Subjects[input.SubjectProfessors[subjectProfessorId].Subject].Name,
			input.Professors[input.SubjectProfessors[subjectProfessorId].Professor].Name,
		)

		if !strings.Contains(subjectProfessorName, "_cc_") || !strings.Contains(subjectProfessorName, "_1_") {
			continue
		}

		for _, class := range groups[groupId] {
			className := input.Classes[class].Name

			fmt.Printf("Period: %v, Day: %v, Lesson: %v, SubjectProfessor: %v, Class: %v \n", positive[0], dayName, positive[2], subjectProfessorName, className)
		}
	}

	if !timetabler.Verify(timetable, input, curriculum, groupsGraph) {
		log.Fatal("Verification failed")
	}

	fmt.Println("Well done!")
}
