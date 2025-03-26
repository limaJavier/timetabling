package model

import (
	"math/rand"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexAndAttributesDeterministic(t *testing.T) {
	for range 10 {
		// Arranges
		var Periods uint64 = uint64(rand.Intn(20) + 1)
		var Days uint64 = uint64(rand.Intn(7) + 1)
		var Lessons uint64 = uint64(rand.Intn(10) + 1)
		var SubjectTeachers uint64 = uint64(rand.Intn(50) + 1)
		var Classes uint64 = uint64(rand.Intn(20) + 1)

		// Act
		indexer := NewIndexer(Periods, Days, Lessons, SubjectTeachers, Classes)

		indices := make([]uint64, 0, Periods*Days*Lessons*SubjectTeachers*Classes)

		for period := uint64(1); period <= Periods; period++ {
			for day := uint64(1); day <= Days; day++ {
				for lesson := uint64(1); lesson <= Lessons; lesson++ {
					for subjectTeacher := uint64(1); subjectTeacher <= SubjectTeachers; subjectTeacher++ {
						for class := uint64(1); class <= Classes; class++ {
							indices = append(indices, indexer.Index(period, day, lesson, subjectTeacher, class))
						}
					}
				}
			}
		}

		// Act
		for _, index := range indices {
			period, day, lesson, subjectTeacher, class := indexer.Attributes(index)
			assert.Equal(t, index, indexer.Index(period, day, lesson, subjectTeacher, class))
		}
	}
}

func TestIndexAndAttributesNonDeterministic(t *testing.T) {
	// Arrange
	scenarios := [][]uint64{
		{20, 5, 10, 5, 20},
		{15, 7, 7, 10, 5},
		{10, 6, 8, 35, 20},
		{5, 7, 5, 20, 5},
		{1, 4, 5, 45, 20},
	}

	for _, scenario := range scenarios {
		var Periods uint64 = scenario[0]
		var Days uint64 = scenario[1]
		var Lessons uint64 = scenario[2]
		var SubjectTeachers = scenario[3]
		var Classes uint64 = scenario[4]

		// Act
		indexer := NewIndexer(Periods, Days, Lessons, SubjectTeachers, Classes)

		indices := make([]uint64, 0, Periods*Days*Lessons*SubjectTeachers*Classes)

		for period := uint64(1); period <= Periods; period++ {
			for day := uint64(1); day <= Days; day++ {
				for lesson := uint64(1); lesson <= Lessons; lesson++ {
					for subjectTeacher := uint64(1); subjectTeacher <= SubjectTeachers; subjectTeacher++ {
						for class := uint64(1); class <= Classes; class++ {
							indices = append(indices, indexer.Index(period, day, lesson, subjectTeacher, class))
						}
					}
				}
			}
		}

		// Assert
		for _, index := range indices {
			period, day, lesson, subjectTeacher, class := indexer.Attributes(index)
			assert.Equal(t, index, indexer.Index(period, day, lesson, subjectTeacher, class))
		}
	}
}

func TestIntegerConstraints(t *testing.T) {
	for range 10 {
		// Arrange
		var Periods uint64 = uint64(rand.Intn(20) + 1)
		var Days uint64 = uint64(rand.Intn(7) + 1)
		var Lessons uint64 = uint64(rand.Intn(10) + 1)
		var SubjectTeachers uint64 = uint64(rand.Intn(50) + 1)
		var Classes uint64 = uint64(rand.Intn(20) + 1)

		// Act
		indexer := NewIndexer(Periods, Days, Lessons, SubjectTeachers, Classes)

		indices := make([]uint64, 0, Periods*Days*Lessons*SubjectTeachers*Classes)

		for period := uint64(1); period <= Periods; period++ {
			for day := uint64(1); day <= Days; day++ {
				for lesson := uint64(1); lesson <= Lessons; lesson++ {
					for subjectTeacher := uint64(1); subjectTeacher <= SubjectTeachers; subjectTeacher++ {
						for class := uint64(1); class <= Classes; class++ {
							indices = append(indices, indexer.Index(period, day, lesson, subjectTeacher, class))
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
