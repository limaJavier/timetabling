package model

import (
	"fmt"
	"slices"

	"github.com/samber/lo"
)

type Preprocessor interface {
	AddSingletonGroups(classesCurriculum [][]bool, groupsPerSubjectProfessor map[uint64][][]uint64)
	ExtractCurriculumAndGroups(groupsPerSubjectProfessor map[uint64][][]uint64) ([][]bool, map[uint64][]uint64)
	BuildGroupsGraph(groups map[uint64][]uint64) [][]bool
}

func NewPreprocessor() Preprocessor {
	return &preprocessorImplementation{}
}

type preprocessorImplementation struct {
}

func (preprocessor *preprocessorImplementation) AddSingletonGroups(classesCurriculum [][]bool, groupsPerSubjectProfessor map[uint64][][]uint64) {
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
func (preprocessor *preprocessorImplementation) ExtractCurriculumAndGroups(groupsPerSubjectProfessor map[uint64][][]uint64) ([][]bool, map[uint64][]uint64) {
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
func (preprocessor *preprocessorImplementation) BuildGroupsGraph(groups map[uint64][]uint64) [][]bool {
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
