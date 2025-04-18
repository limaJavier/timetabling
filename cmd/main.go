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

const RoomSimilarityThreshold = 0.50

func main() {
	const File string = "../test/out/satisfiable/1.json"
	preprocessor := model.NewPreprocessor()

	input, err := model.InputFromJson(File)
	if err != nil {
		log.Fatalf("cannot parse input file: %v", err)
	}

	curriculum, groups := preprocessor.ExtractCurriculumAndGroups(input)
	groupsGraph := preprocessor.BuildGroupsGraph(groups)

	// solver := sat.NewKissatSolver()
	// solver := sat.NewCadicalSolver()
	// solver := sat.NewCryptominisatSolver()
	// solver := sat.NewMinisatSolver()
	solver := sat.NewSlimeSolver()
	// timetabler := model.NewEmbeddedRoomTimetabler(solver)
	// timetabler := model.NewIsolatedRoomTimetabler(solver, false, 0)
	timetabler := model.NewIsolatedRoomTimetabler(solver, true, RoomSimilarityThreshold)

	timetable, err := timetabler.Build(input, curriculum, groups, groupsGraph)
	if err != nil {
		log.Fatal(err)
	} else if timetable == nil {
		fmt.Println("Not satisfiable")
		return
	}

	compare := func(a, b [6]uint64, i int) int {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return 1
		}
		return 0
	}

	slices.SortFunc(timetable, func(a, b [6]uint64) int {
		dayComparison := compare(a, b, 1)
		if dayComparison != 0 {
			return dayComparison
		}
		return compare(a, b, 0)
	})

	for _, positive := range timetable {
		dayId := positive[1]
		subjectProfessorId := positive[3]
		groupId := positive[4]
		roomId := positive[5]

		dayName := Days[dayId]
		subjectProfessorName := fmt.Sprintf("%v~%v",
			input.Subjects[input.SubjectProfessors[subjectProfessorId].Subject].Name,
			input.Professors[input.SubjectProfessors[subjectProfessorId].Professor].Name,
		)
		roomName := input.Rooms[roomId].Name

		if !strings.Contains(subjectProfessorName, "_cc_") || !strings.Contains(subjectProfessorName, "_3_") {
			continue
		}

		for _, class := range groups[groupId] {
			className := input.Classes[class].Name

			fmt.Printf("Period: %v, Day: %v, Lesson: %v, SubjectProfessor: %v, Class: %v, Room: %v \n", positive[0], dayName, positive[2], subjectProfessorName, className, roomName)
		}
	}

	if !timetabler.Verify(timetable, input, curriculum, groups, groupsGraph) {
		log.Fatal("Verification failed")
	}

	fmt.Println("Well done!")
}
