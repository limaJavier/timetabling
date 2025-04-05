package main

import (
	"fmt"
	"log"
	"timetabling/internal/model"
	"timetabling/internal/sat"
)

func main() {
	preprocessor := model.NewPreprocessor()

	subjectTeachersStr := []string{
		"algebraConf",
		"logicaConf",
		"analisisConf",
		"programacionConf",
		"algebraCp",
		"algebraCp",
		"logicaCp",
		"logicaCp",
		"analisisCp",
		"analisisCp",
		"programacionCp",
		"programacionCp",
		"discretaConf",
		"edaConf",
		"arquitecturaConf",
		"edoConf",
		"discretaCp",
		"discretaCp",
		"edaCp",
		"edaCp",
		"arquitecturaCp",
		"arquitecturaCp",
		"edoCp",
		"edoCp",
	}

	professorsStr := []string{
		"pancho",
		"luisa",
		"juan",
		"manuel",
		"pedro",
		"rosa",
		"miguel",
		"beatriz",
		"julia",
		"juana",
		"julian",
		"sonia",
		"esteban",
		"rocio",
		"leonardo",
		"minerva",
		"penelope",
		"alejandro",
		"carlos",
		"helen",
		"rachel",
		"ross",
		"chandler",
		"phoebe",
		"joseph",
	}

	classesCurriculumInt := [][]uint64{
		{1, 1, 1, 1, 1, 0, 1, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{1, 1, 1, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 0, 1, 0, 1, 0, 1, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 1, 0, 1, 0, 1, 0, 1},
	}

	classesCurriculum := [][]bool{}

	for i, row := range classesCurriculumInt {
		classesCurriculum = append(classesCurriculum, make([]bool, len(row)))
		for j, bit := range row {
			if bit == 1 {
				classesCurriculum[i][j] = true
			}
		}
	}

	groupsPerSubjectProfessor := map[uint64][][]uint64{
		0:  {{0, 1}},
		1:  {{0, 1}},
		2:  {{0, 1}},
		3:  {{0, 1}},
		12: {{2, 3}},
		13: {{2, 3}},
		14: {{2, 3}},
		15: {{2, 3}},
	}

	preprocessor.AddSingletonGroups(classesCurriculum, groupsPerSubjectProfessor)
	curriculum, groups := preprocessor.ExtractCurriculumAndGroups(groupsPerSubjectProfessor)

	groupsGraph := preprocessor.BuildGroupsGraph(groups)

	lessons := map[uint64]uint64{}

	for i := range len(curriculum[0]) {
		lessons[uint64(i)] = 1
	}

	professors := map[uint64]uint64{}

	availability := map[uint64][][]bool{}

	for i := range 50 {
		professors[uint64(i)] = uint64(i)
	}

	for i := range 50 {
		availability[uint64(i)] = [][]bool{
			{true, true, true, true, true},
			{true, true, true, true, true},
			{true, true, true, true, true},
			{true, true, true, true, true},
			{true, true, true, true, true},
			{true, true, true, true, true},
		}
	}

	rooms := map[uint64]uint64{
		0:  0,
		1:  0,
		2:  0,
		3:  0,
		4:  0,
		5:  1,
		6:  0,
		7:  1,
		8:  0,
		9:  1,
		10: 0,
		11: 1,
		12: 2,
		13: 2,
		14: 2,
		15: 2,
		16: 2,
		17: 3,
		18: 2,
		19: 3,
		20: 2,
		21: 3,
		22: 2,
		23: 3,
	}

	solver := sat.NewKissatSolver()
	timetabler := model.NewTimetabler(solver)

	timetable, err := timetabler.Build(curriculum, groupsGraph, lessons, availability, rooms, professors)
	if err != nil {
		log.Fatal(err)
	} else if timetable == nil {
		fmt.Println("Not satisfiable")
		return
	}

	for _, positive := range timetable {
		subjectProfessor := fmt.Sprintf("%v~%v", subjectTeachersStr[positive[3]], professorsStr[professors[positive[3]]])
		// subjectProfessor := positive[3]
		for _, class := range groups[positive[4]] {
			fmt.Printf("Period: %v, Day: %v, Lesson: %v, SubjectProfessor: %v, Class: %v \n", positive[0], positive[1], positive[2], subjectProfessor, class)
		}
	}

	if !timetabler.Verify(timetable, curriculum, groupsGraph, lessons, availability, rooms, professors, groupsPerSubjectProfessor) {
		log.Fatal("Verification failed")
	}

	fmt.Println("Well done!")
}

// TODO: Test this
