package model

import (
	"slices"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestExtractCurriculumAndGroups(t *testing.T) {
	preprocessor := preprocessorImplementation{}
	t.Run("Correct flow", func(t *testing.T) {
		//**Arrange
		groupsPerSubjectProfessors := []map[uint64][][]uint64{
			{
				0: {{0, 1}, {2, 3}},
				1: {{0, 1}},
				2: {{2, 3}},
				3: {{1, 3}},
			},
			{
				0: {{0, 1, 2, 3}},
				1: {{0}, {1}, {2}, {3}},
				2: {{0}},
				3: {{1}},
				4: {{2}},
			},
		}

		for i := range len(groupsPerSubjectProfessors) {
			groupsPerSubjectProfessor := groupsPerSubjectProfessors[i]

			//**Act
			curriculum, groups := preprocessor.ExtractCurriculumAndGroups(groupsPerSubjectProfessor)

			//**Assert
			for subjectProfessor, associatedGroups := range groupsPerSubjectProfessor {
				for _, associatedGroup := range associatedGroups {
					slices.Sort(associatedGroup)

					exists := false
					for groupId, group := range groups {
						if slices.Equal(group, associatedGroup) {
							if !curriculum[groupId][subjectProfessor] {
								t.Errorf("subjectProfessor %v is expected to teach group %v", subjectProfessor, groupId)
							}

							exists = true
							break
						}
					}

					if !exists {
						t.Errorf("group %v was not found", associatedGroup)
					}
				}
			}
		}
	})

	t.Run("Panic flow", func(t *testing.T) {
		//**Arrange
		groupsPerSubjectProfessors := []map[uint64][][]uint64{
			{
				0: {{0, 1}},
				1: {{1, 2, 3}, {3}},
				2: {{0, 2, 3}, {2}},
				3: {{0, 1}},
			},
			{
				0: {{0, 1, 2, 3}, {0, 1}},
				1: {{1, 2}},
				2: {{0, 3}},
				3: {{0, 1}},
				4: {{2, 3}},
			},
		}

		for i := range len(groupsPerSubjectProfessors) {
			groupsPerSubjectProfessor := groupsPerSubjectProfessors[i]

			//**Act and assert
			assert.Panics(t, func() {
				preprocessor.ExtractCurriculumAndGroups(groupsPerSubjectProfessor)
			})
		}
	})
}

func TestBuildGroupsGraph(t *testing.T) {
	//**Arrange
	preprocessor := preprocessorImplementation{}
	groupsPerSubjectProfessors := []map[uint64][][]uint64{
		{
			0: {{0, 1}, {2, 3}},
			1: {{0, 1}},
			2: {{2, 3}},
			3: {{1, 3}},
		},
		{
			0: {{0, 1, 2, 3}},
			1: {{0}, {1}, {2}, {3}},
			2: {{0}},
			3: {{1}},
			4: {{2}},
		},
	}

	for i := range len(groupsPerSubjectProfessors) {
		groupsPerSubjectProfessor := groupsPerSubjectProfessors[i]
		_, groups := preprocessor.ExtractCurriculumAndGroups(groupsPerSubjectProfessor)

		//**Act
		groupsGraph := preprocessor.BuildGroupsGraph(groups)

		//**Assert
		for i, row := range groupsGraph {
			for j := range row {
				if i == j && !groupsGraph[i][j] {
					t.Errorf("groupsGraph[%v][%v] must be true", i, i)
				} else if groupsGraph[i][j] {
					if !groupsGraph[j][i] {
						t.Errorf("groupsGraph[%v][%v] == true implies that groupsGraph[%v][%v == true]", i, j, j, i)
					}

					group1, group2 := groups[uint64(i)], groups[uint64(j)]
					if !lo.SomeBy(group1, func(class uint64) bool {
						return slices.Contains(group2, class)
					}) {
						t.Errorf("groups %v and %v are disjoint share an edge", group1, group2)
					}
				}
			}
		}

		for i := range len(groups) - 1 {
			for j := i + 1; j < len(groups); j++ {
				group1, group2 := groups[uint64(i)], groups[uint64(j)]

				if lo.SomeBy(group1, func(class uint64) bool {
					return slices.Contains(group2, class)
				}) && !groupsGraph[i][j] {
					t.Errorf("groups %v and %v are not disjoint and don't share an edge", group1, group2)
				}
			}
		}
	}
}
