package main

import (
	"fmt"
	"log"
	"timetabling/internal/model"
	"timetabling/internal/sat"
)

func main() {
	// subjectTeachersStr := []string{
	// 	"algebraConf",
	// 	"logicaConf",
	// 	"analisisConf",
	// 	"programacionConf",
	// 	"algebraCp",
	// 	"algebraCp",
	// 	"logicaCp",
	// 	"logicaCp",
	// 	"analisisCp",
	// 	"analisisCp",
	// 	"programacionCp",
	// 	"programacionCp",
	// 	"discretaConf",
	// 	"edaConf",
	// 	"arquitecturaConf",
	// 	"edoConf",
	// 	"discretaCp",
	// 	"discretaCp",
	// 	"edaCp",
	// 	"edaCp",
	// 	"arquitecturaCp",
	// 	"arquitecturaCp",
	// 	"edoCp",
	// 	"edoCp",
	// }

	// professorsStr := []string{
	// 	"pancho",
	// 	"luisa",
	// 	"juan",
	// 	"manuel",
	// 	"pedro",
	// 	"rosa",
	// 	"miguel",
	// 	"beatriz",
	// 	"julia",
	// 	"juana",
	// 	"julian",
	// 	"sonia",
	// 	"esteban",
	// 	"rocio",
	// 	"leonardo",
	// 	"minerva",
	// 	"penelope",
	// 	"alejandro",
	// 	"carlos",
	// 	"helen",
	// 	"rachel",
	// 	"ross",
	// 	"chandler",
	// 	"phoebe",
	// 	"joseph",
	// }

	curriculum := [][]uint64{
		{1, 1},
		{1, 1},
		// {1, 1, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0},
		// {1, 1, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0},
		// {0, 0, 0, 0, 0, 0, 1, 1, 1, 0, 1, 0},
		// {0, 0, 0, 0, 0, 0, 1, 1, 0, 1, 0, 1},
	}

	groups := map[uint64][][]uint64{
		0: {{0, 1}},
		1: {{0, 1}},
		// 6: {{2, 3}},
		// 7: {{2, 3}},
	}

	professors := map[uint64]uint64{}

	availability := map[uint64][][]bool{}

	for i := range 50 {
		professors[uint64(i)] = uint64(i)
	}

	for i := range 50 {
		availability[uint64(i)] = [][]bool{
			{true, true},
			{true, true},
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

	timetable, err := timetabler.Build(curriculum, groups, availability, rooms, professors)
	if err != nil {
		log.Fatal(err)
	} else if timetable == nil {
		fmt.Println("Not satisfiable")
		return
	}

	for _, positive := range timetable {
		// str := fmt.Sprintf("%v~%v", subjectTeachersStr[positive[3]], professorsStr[professors[positive[3]]])
		str := positive[3]
		fmt.Printf("Period: %v, Day: %v, Lesson: %v, SubjectProfessor: %v, Class: %v \n", positive[0], positive[1], positive[2], str, positive[4])
	}

	if !timetabler.Verify(timetable, curriculum, availability, rooms, professors) {
		log.Fatal("Verification failed")
	}
}
