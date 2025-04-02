package model

import (
	"testing"
	"timetabling/internal/sat"

	"github.com/stretchr/testify/assert"
)

func TestBuild(t *testing.T) {
	t.Run("Test I", func(t *testing.T) {
		// Arrange
		curriculum := [][]uint64{
			{1, 1, 1, 0, 0, 0, 2, 0, 2, 0, 2, 0, 0, 0, 0, 0, 0, 0},
			{1, 1, 1, 0, 0, 0, 0, 2, 0, 2, 0, 2, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 1, 1, 1, 0, 0, 0, 0, 0, 0, 2, 0, 2, 0, 2, 0},
			{0, 0, 0, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 2, 0, 2, 0, 2},
		}

		professors := map[uint64]uint64{
			0:  0,
			1:  1,
			2:  2,
			3:  3,
			4:  4,
			5:  5,
			6:  6,
			7:  7,
			8:  8,
			9:  9,
			10: 10,
			11: 11,
			12: 12,
			13: 13,
			14: 14,
			15: 15,
			16: 16,
			17: 17,
		}

		availability := map[uint64][][]bool{}

		for i := range 18 {
			availability[uint64(i)] = [][]bool{
				{true, true, true, true, true},
				{true, true, true, true, true},
				{true, true, true, true, true},
			}
		}

		rooms := map[uint64]uint64{
			0:  0,
			1:  1,
			2:  2,
			3:  3,
			4:  4,
			5:  5,
			6:  6,
			7:  7,
			8:  8,
			9:  9,
			10: 10,
			11: 11,
			12: 12,
			13: 13,
			14: 14,
			15: 15,
			16: 16,
			17: 17,
		}

		solver := sat.NewKissatSolver()
		timetabler := NewTimetabler(solver)

		// Act
		timetable, err := timetabler.Build(curriculum, availability, rooms, professors)

		// Assert
		assert.Nil(t, err)
		assert.NotNil(t, timetable)
		assert.True(t, timetabler.Verify(timetable, curriculum, availability, rooms, professors))
	})

	t.Run("Test II", func(t *testing.T) {
		// Arrange
		curriculum := [][]uint64{
			{1, 1, 1, 0, 0, 0, 2, 0, 2, 0, 2, 0, 0, 0, 0, 0, 0, 0},
			{1, 1, 1, 0, 0, 0, 0, 2, 0, 2, 0, 2, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 1, 1, 1, 0, 0, 0, 0, 0, 0, 2, 0, 2, 0, 2, 0},
			{0, 0, 0, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 2, 0, 2, 0, 2},
		}

		professors := map[uint64]uint64{
			0:  0,
			1:  1,
			2:  2,
			3:  3,
			4:  4,
			5:  5,
			6:  0,
			7:  6,
			8:  7,
			9:  8,
			10: 9,
			11: 2,
			12: 3,
			13: 6,
			14: 10,
			15: 11,
			16: 12,
			17: 13,
		}

		availability := map[uint64][][]bool{
			0: {
				{true, false, true, false, true},
				{true, false, true, false, true},
				{true, false, true, false, true},
			},
			1: {
				{true, true, true, true, false},
				{true, true, true, true, false},
				{true, true, true, true, false},
			},
			2: {
				{true, true, false, false, true},
				{true, true, false, false, true},
				{true, true, false, false, true},
			},
			3: {
				{true, false, true, true, true},
				{true, false, true, true, true},
				{true, false, true, true, true},
			},
			4: {
				{true, true, true, true, false},
				{true, true, true, true, false},
				{true, true, true, true, false},
			},
			5: {
				{true, true, true, false, false},
				{true, true, true, false, false},
				{true, true, true, false, false},
			},
		}

		for i := range 14 {
			availability[uint64(i+5)] = [][]bool{
				{true, true, true, true, true},
				{true, true, true, true, true},
				{true, true, true, true, true},
			}
		}

		rooms := map[uint64]uint64{
			0:  0,
			1:  0,
			2:  0,
			3:  3,
			4:  3,
			5:  3,
			6:  0,
			7:  1,
			8:  0,
			9:  1,
			10: 0,
			11: 1,
			12: 3,
			13: 4,
			14: 3,
			15: 4,
			16: 3,
			17: 4,
		}

		solver := sat.NewKissatSolver()
		timetabler := NewTimetabler(solver)

		// Act
		timetable, err := timetabler.Build(curriculum, availability, rooms, professors)

		// Assert
		assert.Nil(t, err)
		assert.NotNil(t, timetable)
		assert.True(t, timetabler.Verify(timetable, curriculum, availability, rooms, professors))
	})

	t.Run("Test III", func(t *testing.T) {
		// Arrange
		curriculum := [][]uint64{
			{1, 1, 1, 1, 1, 0, 1, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{1, 1, 1, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 0, 1, 0, 1, 0, 1, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 1, 0, 1, 0, 1, 0, 1},
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
		timetabler := NewTimetabler(solver)

		// Act
		timetable, err := timetabler.Build(curriculum, availability, rooms, professors)

		// Assert
		assert.Nil(t, err)
		assert.NotNil(t, timetable)
		assert.True(t, timetabler.Verify(timetable, curriculum, availability, rooms, professors))
	})
}
