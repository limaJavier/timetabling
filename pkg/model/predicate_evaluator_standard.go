package model

import (
	"fmt"
	"slices"
	"strings"

	"github.com/samber/lo"
)

type predicateEvaluatorStandard struct {
	modelInput              ModelInput
	allocations             map[uint64][][]bool // Allocation matrix per group
	roomSimilarityThreshold float32             // Threshold for room similarity
}

func newPredicateEvaluator(modelInput ModelInput, roomSimilarityThreshold float32) predicateEvaluator {
	subjectProfessors := uint64(len(modelInput.SubjectProfessors))
	maxLessons := uint64(0)
	for _, value := range modelInput.Entries {
		if value.Lessons > maxLessons {
			maxLessons = value.Lessons
		}
	}

	evaluator := predicateEvaluatorStandard{
		modelInput:              modelInput,
		roomSimilarityThreshold: roomSimilarityThreshold,
	}

	evaluator.allocations = make(map[uint64][][]bool) // Initialize dictionary
	for group := range modelInput.Curriculum {        // For each group
		evaluator.allocations[uint64(group)] = make([][]bool, subjectProfessors) // Initialize allocation per group

		for subjectProfessor := range modelInput.Curriculum[group] { // For each subjectProfessor
			evaluator.allocations[uint64(group)][subjectProfessor] = make([]bool, maxLessons) // Initialize subjectProfessor row
			entryKey := [2]uint64{uint64(subjectProfessor), uint64(group)}

			for i := range modelInput.Entries[entryKey].Lessons {
				if modelInput.Curriculum[group][subjectProfessor] {
					evaluator.allocations[uint64(group)][subjectProfessor][i] = true // Set to true the first j lessons where j is the number of lessons assigned for "subjectProfessor" to teach to "group" (i.e. curriculum[group][subjectProfessor])
				}
			}
		}
	}

	return &evaluator
}

func (evaluator *predicateEvaluatorStandard) SameProfessor(subjectProfessor1, subjectProfessor2 uint64) bool {
	professor1 := evaluator.modelInput.SubjectProfessors[subjectProfessor1].Professor
	professor2 := evaluator.modelInput.SubjectProfessors[subjectProfessor2].Professor
	return professor1 == professor2
}

func (evaluator *predicateEvaluatorStandard) ProfessorAvailable(subjectProfessor, day, period uint64) bool {
	professorId := evaluator.modelInput.SubjectProfessors[subjectProfessor].Professor
	distribution := evaluator.modelInput.Professors[professorId].Availability
	return distribution[period][day]
}

func (evaluator *predicateEvaluatorStandard) Teaches(group, subjectProfessor, lesson uint64) bool {
	allocation, ok := evaluator.allocations[group]
	if !ok {
		panic("group not found")
	}
	return allocation[subjectProfessor][lesson]
}

func (evaluator *predicateEvaluatorStandard) Disjoint(group1, group2 uint64) bool {
	return !evaluator.modelInput.GroupsGraph[group1][group2]
}

func (evaluator *predicateEvaluatorStandard) Allowed(subjectProfessor, group, day, period uint64) bool {
	entryKey := [2]uint64{subjectProfessor, group}

	// Check if the pair subject-professor and group has an entry, if not then it's not allowed
	if _, ok := evaluator.modelInput.Entries[entryKey]; !ok {
		return false
	}

	distribution := evaluator.modelInput.Entries[entryKey].Permissibility
	return distribution[period][day]
}

func (evaluator *predicateEvaluatorStandard) Assigned(room, subjectProfessor, group uint64) bool {
	entryKey := [2]uint64{subjectProfessor, group}
	return slices.Contains(evaluator.modelInput.Entries[entryKey].Rooms, room)
}

func (evaluator *predicateEvaluatorStandard) Fits(group, room uint64) bool {
	groupSize := lo.Sum(
		lo.Map(evaluator.modelInput.Groups[group].Classes, func(class uint64, _ int) uint64 {
			return evaluator.modelInput.Classes[class].Size
		}),
	)
	return evaluator.modelInput.Rooms[room].Capacity >= groupSize
}

func (evaluator *predicateEvaluatorStandard) RoomSimilar(subjectProfessor1, subjectProfessor2, group1, group2 uint64) bool {
	entryKey1, entryKey2 := [2]uint64{subjectProfessor1, group1}, [2]uint64{subjectProfessor2, group2}
	rooms1 := evaluator.modelInput.Entries[entryKey1].Rooms
	rooms2 := evaluator.modelInput.Entries[entryKey2].Rooms

	// Filter rooms based on group-size and room-capacity
	rooms1 = lo.Filter(rooms1, func(room uint64, _ int) bool {
		return evaluator.Fits(group1, room)
	})
	rooms2 = lo.Filter(rooms2, func(room uint64, _ int) bool {
		return evaluator.Fits(group2, room)
	})

	// If there are no rooms for one of the subject-professors and its current group, panic with a descriptive error message
	if len(rooms1) == 0 {
		panic(evaluator.noRoomsErrorMessage(subjectProfessor1, group1))
	} else if len(rooms2) == 0 {
		panic(evaluator.noRoomsErrorMessage(subjectProfessor2, group2))
	}

	// Union set
	union := make(map[uint64]bool)
	lo.ForEach(slices.Concat(rooms1, rooms2), func(room uint64, _ int) {
		union[room] = true
	})

	// Intersection set
	intersection := make(map[uint64]bool)
	lo.ForEach(rooms1, func(room uint64, _ int) {
		if slices.Contains(rooms2, room) {
			intersection[room] = true
		}
	})
	lo.ForEach(rooms2, func(room uint64, _ int) {
		if slices.Contains(rooms1, room) {
			intersection[room] = true
		}
	})

	// Calculate Jaccard similarity
	jaccardSimilarity := float32(len(intersection)) / float32(len(union))

	return jaccardSimilarity >= evaluator.roomSimilarityThreshold
}

func (evaluator *predicateEvaluatorStandard) noRoomsErrorMessage(subjectProfessor, group uint64) string {
	var builder strings.Builder
	subjectName := evaluator.modelInput.Subjects[evaluator.modelInput.SubjectProfessors[subjectProfessor].Subject].Name
	professorName := evaluator.modelInput.Professors[evaluator.modelInput.SubjectProfessors[subjectProfessor].Professor].Name

	fmt.Fprintf(&builder, "There are not fitting rooms for: %v~%v to { ", subjectName, professorName)
	for _, class := range evaluator.modelInput.Groups[group].Classes {
		fmt.Fprintf(&builder, "%s, ", evaluator.modelInput.Classes[class].Name)
	}
	builder.WriteString("}")
	return builder.String()
}
