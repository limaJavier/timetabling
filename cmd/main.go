package main

import (
	"fmt"
	"log"
	"slices"
	"timetabling/internal/model"
	"timetabling/internal/sat"

	"github.com/samber/lo"
)

func main() {
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
		// 0:  {{0, 1}},
		// 1:  {{0, 1}},
		// 2:  {{0, 1}},
		// 3:  {{0, 1}},
		// 12: {{2, 3}},
		// 13: {{2, 3}},
		// 14: {{2, 3}},
		// 15: {{2, 3}},
	}

	addSingletonGroups(classesCurriculum, groupsPerSubjectProfessor)
	curriculum, groups := extractCurriculumAndGroups(groupsPerSubjectProfessor)

	groupsGraph := buildGroupsGraph(groups)

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
		// str := fmt.Sprintf("%v~%v", subjectTeachersStr[positive[3]], professorsStr[professors[positive[3]]])
		str := positive[3]
		fmt.Printf("Period: %v, Day: %v, Lesson: %v, SubjectProfessor: %v, Group: %v \n", positive[0], positive[1], positive[2], str, positive[4])
	}

	if !timetabler.Verify(timetable, curriculum, groupsGraph, lessons, availability, rooms, professors, groupsPerSubjectProfessor) {
		log.Fatal("Verification failed")
	}

	fmt.Println("Well done!")
}

// TODO: Test this
func addSingletonGroups(classesCurriculum [][]bool, groupsPerSubjectProfessor map[uint64][][]uint64) {
	classes := uint64(len(classesCurriculum))
	subjectProfessors := uint64(len(classesCurriculum[0]))

	for class := range classes {
		for subjectProfessor := range subjectProfessors {
			if classesCurriculum[class][subjectProfessor] {
				contained := false

				for _, group := range groupsPerSubjectProfessor[subjectProfessor] {
					if slices.Contains(group, class) {
						contained = true
						break
					}
				}

				if !contained {
					groupsPerSubjectProfessor[subjectProfessor] = append(groupsPerSubjectProfessor[subjectProfessor], []uint64{class})
				}
			}
		}
	}
}

// TODO: Test this
func extractCurriculumAndGroups(groupsPerSubjectProfessor map[uint64][][]uint64) ([][]bool, map[uint64][]uint64) {
	subjectProfessors := len(groupsPerSubjectProfessor)
	curriculum := make([][]bool, 0)
	groups := make(map[uint64][]uint64)

	currentId := uint64(0)
	for subjectProfessor, associatedGroups := range groupsPerSubjectProfessor {
		associatedClasses := make(map[uint64]bool)

		for _, group := range associatedGroups {
			// Verify associated groups are disjoint
			lo.ForEach(group, func(class uint64, _ int) {
				if _, ok := associatedClasses[class]; ok {
					panic(fmt.Sprintf("groups associated to the same subjectProfessor \"%v\" must be disjoint sets: class \"%v\" is present in more than one group or group \"%v\" is not a set", subjectProfessor, class, group))
				}
				associatedClasses[class] = true
			})

			groupCopy := make([]uint64, len(group))
			copy(groupCopy, group)
			slices.Sort(groupCopy)

			exists := false
			for groupId, group := range groups {
				if slices.Equal(group, groupCopy) {
					curriculum[groupId][subjectProfessor] = true
					exists = true
					break
				}
			}

			if !exists {
				groups[currentId] = groupCopy
				currentId++

				row := make([]bool, subjectProfessors)
				row[subjectProfessor] = true
				curriculum = append(curriculum, row)
			}
		}
	}

	return curriculum, groups
}

// TODO: Test this
func buildGroupsGraph(groups map[uint64][]uint64) [][]bool {
	groupsGraph := make([][]bool, len(groups))

	groupsIds := make([]uint64, 0, len(groups))
	for id := range groups {
		groupsIds = append(groupsIds, id)
		groupsGraph[id] = make([]bool, len(groups)) // Initialize each row
	}

	for i := range len(groupsIds) - 1 {
		groupsGraph[i][i] = true // For completeness we assume that groups[i][i] = true for all i
		for j := i + 1; j < len(groupsIds); j++ {
			id1, id2 := groupsIds[i], groupsIds[j]
			group1, group2 := groups[id1], groups[id2]

			// Verify group1 and group2 have a class in common
			if lo.SomeBy(group1, func(class uint64) bool {
				return slices.Contains(group2, class)
			}) {
				groupsGraph[id1][id2] = true
				groupsGraph[id2][id1] = true
			}
		}
	}
	groupsGraph[len(groups)-1][len(groups)-1] = true // Set last index from diagonal to true since the previous iteration does not account for it

	return groupsGraph
}
