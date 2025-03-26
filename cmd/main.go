package main

import (
	"fmt"
	"log"
	"timetabling/internal/model"
	"timetabling/internal/sat"
)

func main() {
	curriculum := [][]uint64{
		{2, 2},
		{2, 2},
	}

	availability := map[uint64][][]bool{
		0: {
			{true, true},
			{true, true},
		},
		1: {
			{true, true},
			{true, true},
		},
	}

	rooms := map[uint64]uint64{
		0: 0,
		1: 1,
	}

	professors := map[uint64]uint64{
		0: 0,
		1: 1,
	}

	solver := sat.NewKissatSolver()
	timetabler := model.NewTimetabler(solver)

	solution, err := timetabler.Build(curriculum, availability, rooms, professors)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(solution)
}
