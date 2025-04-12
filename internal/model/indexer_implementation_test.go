package model

import (
	"math/rand"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexAndAttributesDeterministic(t *testing.T) {
	// Arrange
	scenarios := [][]uint64{
		{3, 3, 3, 3, 3, 3},
		{20, 5, 10, 5, 20, 25},
		{15, 7, 7, 10, 5, 8},
		{10, 6, 8, 35, 20, 35},
		{5, 7, 5, 20, 5, 7},
		{1, 4, 5, 45, 20, 15},
	}

	for _, scenario := range scenarios {
		var Periods uint64 = scenario[0]
		var Days uint64 = scenario[1]
		var Lessons uint64 = scenario[2]
		var SubjectProfessors = scenario[3]
		var Groups uint64 = scenario[4]
		var Rooms uint64 = scenario[5]

		// Act
		indexer := NewIndexer(Periods, Days, Lessons, SubjectProfessors, Groups, Rooms)

		indices := make([]uint64, 0, Periods*Days*Lessons*SubjectProfessors*Groups)

		for period := uint64(0); period < Periods; period++ {
			for day := uint64(0); day < Days; day++ {
				for lesson := uint64(0); lesson < Lessons; lesson++ {
					for subjectProfessor := uint64(0); subjectProfessor < SubjectProfessors; subjectProfessor++ {
						for group := uint64(0); group < Groups; group++ {
							for room := uint64(0); room < Rooms; room++ {
								indices = append(indices, indexer.Index(period, day, lesson, subjectProfessor, group, room))
							}
						}
					}
				}
			}
		}

		// Assert
		for _, index := range indices {
			period, day, lesson, subjectProfessor, group, room := indexer.Attributes(index)
			assert.Equal(t, index, indexer.Index(period, day, lesson, subjectProfessor, group, room))
		}
	}
}

func TestIndexAndAttributesNonDeterministic(t *testing.T) {
	for range 10 {
		// Arrange
		var Periods uint64 = uint64(rand.Intn(20) + 1)
		var Days uint64 = uint64(rand.Intn(7) + 1)
		var Lessons uint64 = uint64(rand.Intn(10) + 1)
		var SubjectProfessors uint64 = uint64(rand.Intn(50) + 1)
		var Groups uint64 = uint64(rand.Intn(20) + 1)
		var Rooms uint64 = uint64(rand.Intn(50) + 1)

		// Act
		indexer := NewIndexer(Periods, Days, Lessons, SubjectProfessors, Groups, Rooms)

		indices := make([]uint64, 0, Periods*Days*Lessons*SubjectProfessors*Groups)

		for period := uint64(0); period < Periods; period++ {
			for day := uint64(0); day < Days; day++ {
				for lesson := uint64(0); lesson < Lessons; lesson++ {
					for subjectProfessor := uint64(1); subjectProfessor < SubjectProfessors; subjectProfessor++ {
						for group := uint64(0); group < Groups; group++ {
							for room := uint64(0); room < Rooms; room++ {
								indices = append(indices, indexer.Index(period, day, lesson, subjectProfessor, group, room))
							}
						}
					}
				}
			}
		}

		// Assert
		for _, index := range indices {
			period, day, lesson, subjectProfessor, group, room := indexer.Attributes(index)
			assert.Equal(t, index, indexer.Index(period, day, lesson, subjectProfessor, group, room))
		}
	}
}

func TestIntegerConstraints(t *testing.T) {
	for range 10 {
		// Arrange
		var Periods uint64 = uint64(rand.Intn(20) + 1)
		var Days uint64 = uint64(rand.Intn(7) + 1)
		var Lessons uint64 = uint64(rand.Intn(10) + 1)
		var SubjectProfessors uint64 = uint64(rand.Intn(50) + 1)
		var Groups uint64 = uint64(rand.Intn(20) + 1)
		var Rooms uint64 = uint64(rand.Intn(50) + 1)

		// Act
		indexer := NewIndexer(Periods, Days, Lessons, SubjectProfessors, Groups, Rooms)

		indices := make([]uint64, 0, Periods*Days*Lessons*SubjectProfessors*Groups)

		for period := uint64(0); period < Periods; period++ {
			for day := uint64(0); day < Days; day++ {
				for lesson := uint64(0); lesson < Lessons; lesson++ {
					for subjectProfessor := uint64(0); subjectProfessor < SubjectProfessors; subjectProfessor++ {
						for group := uint64(0); group < Groups; group++ {
							for room := uint64(0); room < Rooms; room++ {
								indices = append(indices, indexer.Index(period, day, lesson, subjectProfessor, group, room))
							}
						}
					}
				}
			}
		}

		slices.Sort(indices)

		// Assert
		for i, index := range indices {
			if i == 0 {
				// First index should be 1
				assert.Equal(t, uint64(1), index)
				continue
			}

			// Each index should be one more than the previous index
			assert.Equal(t, indices[i-1]+1, index)
		}
	}
}
