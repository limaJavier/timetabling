package model

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"slices"

	"github.com/mitchellh/mapstructure"
	"github.com/samber/lo"
)

type RawEntry struct {
	Subject        uint64
	Professor      uint64
	Classes        []uint64
	Lessons        uint64
	Permissibility [][]bool
	Rooms          []uint64
}

type RawModelInput struct {
	Subjects   []Subject
	Professors []Professor
	Classes    []Class
	Rooms      []Room
	Entries    []RawEntry
}

type Subject struct {
	Id   uint64
	Name string
}

type Class struct {
	Id   uint64
	Name string
	Size uint64
}

type Group struct {
	Id      uint64
	Classes []uint64
}

type Room struct {
	Id       uint64
	Name     string
	Capacity uint64
}

type Professor struct {
	Id           uint64
	Name         string
	Availability [][]bool
}

type SubjectProfessor struct {
	Id        uint64
	Subject   uint64
	Professor uint64
}

type Entry struct {
	SubjectProfessor uint64
	Group            uint64
	Lessons          uint64
	Permissibility   [][]bool
	Rooms            []uint64
}

type ModelInput struct {
	Subjects          []Subject
	Professors        []Professor
	SubjectProfessors []SubjectProfessor
	Groups            []Group
	Entries           map[[2]uint64]Entry
	Classes           []Class
	Rooms             []Room
	Curriculum        [][]bool
	GroupsGraph       [][]bool // Groups matrix' coordinate (i, j) = true if and only if group_i and group_j have at least one class in common (i.e. it represents an undirected graph where an edge indicate that two groups share a common class). For completeness we assume that groups[i][i] = true for all i
}

func InputFromJson(file string) (ModelInput, error) {
	bytes, _ := os.ReadFile(file)
	var inputJson map[string]any
	err := json.Unmarshal(bytes, &inputJson)
	if err != nil {
		return ModelInput{}, err
	}

	var rawInput RawModelInput
	mapstructure.Decode(inputJson, &rawInput)
	return ProcessRawInput(rawInput)
}

func ProcessRawInput(rawInput RawModelInput) (ModelInput, error) {
	input := ModelInput{
		Subjects:   rawInput.Subjects,
		Professors: rawInput.Professors,
		Classes:    rawInput.Classes,
		Rooms:      rawInput.Rooms,
	}

	subjectProfessors := make([]SubjectProfessor, 0)
	associatedClasses := make(map[[2]uint64]map[uint64]bool)
	groups := make([]Group, 0)
	entries := make(map[[2]uint64]Entry)
	for _, rawEntry := range rawInput.Entries {
		//** Manage subject-professor
		// Find subject-professor
		subjectProfessor, ok := lo.Find(subjectProfessors, func(subjectProfessor SubjectProfessor) bool {
			return subjectProfessor.Subject == rawEntry.Subject && subjectProfessor.Professor == rawEntry.Professor
		})
		// Initialize subject-professor if it does not exist
		if !ok {
			subjectProfessor = SubjectProfessor{
				Id:        uint64(len(subjectProfessors)),
				Subject:   rawEntry.Subject,
				Professor: rawEntry.Professor,
			}
			subjectProfessors = append(subjectProfessors, subjectProfessor)
		}
		subjectProfessorName := fmt.Sprintf("%v~%v", rawInput.Subjects[subjectProfessor.Subject].Name, rawInput.Professors[subjectProfessor.Professor].Name)
		subjectProfessorKey := [2]uint64{subjectProfessor.Subject, subjectProfessor.Professor}

		// Initialize associated-classes for subject-professor if it does not exist
		if _, ok := associatedClasses[subjectProfessorKey]; !ok {
			associatedClasses[subjectProfessorKey] = make(map[uint64]bool)
		}

		// Make sure that groups associated to the same subject-professor are disjoint sets
		var conflictingClass uint64 = math.MaxUint64
		if lo.SomeBy(rawEntry.Classes, func(class uint64) bool {
			if _, ok := associatedClasses[subjectProfessorKey][class]; ok {
				conflictingClass = class
				return true
			}
			associatedClasses[subjectProfessorKey][class] = true
			return false
		}) {
			return ModelInput{}, fmt.Errorf("groups associated to the same subject-professor \"%v\" must be disjoint sets: class \"%v\" is present in more than one group or group \"%v\" is not a set", subjectProfessorName, rawInput.Classes[conflictingClass].Name, lo.Map(rawEntry.Classes, func(class uint64, _ int) string { return rawInput.Classes[class].Name }))
		}

		//** Manage group
		// Find group
		slices.Sort(rawEntry.Classes) // Sort classes to ensure uniqueness
		group, ok := lo.Find(groups, func(group Group) bool {
			return slices.Equal(group.Classes, rawEntry.Classes)
		})
		// Initialize group if it does not exist
		if !ok {
			group = Group{
				Id:      uint64(len(groups)),
				Classes: rawEntry.Classes,
			}
			groups = append(groups, group)
		}

		//** Manage entry
		entryKey := [2]uint64{subjectProfessor.Id, group.Id}
		// Make sure that can only be one entry for each subject-professor and group
		if _, ok := entries[entryKey]; ok {
			return ModelInput{}, fmt.Errorf("duplicate entry for subject %d and group %d", subjectProfessor.Id, group.Id)
		} else {
			entry := Entry{
				SubjectProfessor: subjectProfessor.Id,
				Group:            group.Id,
				Lessons:          rawEntry.Lessons,
				Permissibility:   rawEntry.Permissibility,
				Rooms:            rawEntry.Rooms,
			}
			entries[entryKey] = entry
		}
	}

	//** Manage curriculum
	// Initialize curriculum
	curriculum := make([][]bool, len(groups))
	for i, _ := range groups {
		curriculum[i] = make([]bool, len(subjectProfessors))
	}
	// Fill curriculum
	for _, entry := range entries {
		curriculum[entry.Group][entry.SubjectProfessor] = true
	}

	input.SubjectProfessors = subjectProfessors
	input.Groups = groups
	input.Entries = entries
	input.Curriculum = curriculum
	input.GroupsGraph = buildGroupsGraph(groups)
	return input, nil
}

func buildGroupsGraph(groups []Group) [][]bool {
	groupsGraph := make([][]bool, len(groups))

	for i := range len(groups) {
		groupsGraph[i] = make([]bool, len(groups))
	}

	for i := range len(groups) - 1 {
		groupsGraph[i][i] = true // For completeness we assume that groups[i][i] = true for all i
		for j := i + 1; j < len(groups); j++ {
			id1, id2 := groups[i].Id, groups[j].Id
			group1, group2 := groups[id1], groups[id2]

			// Verify group1 and group2 have a class in common
			if lo.SomeBy(group1.Classes, func(class uint64) bool {
				return slices.Contains(group2.Classes, class)
			}) {
				groupsGraph[id1][id2] = true
				groupsGraph[id2][id1] = true
			}
		}
	}
	groupsGraph[len(groups)-1][len(groups)-1] = true // Set last index from diagonal to true since the previous iteration does not account for it

	return groupsGraph
}
