package model

import (
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/limaJavier/timetabling/pkg/sat"

	"github.com/onsi/gomega/matchers/support/goraph/bipartitegraph"
	"github.com/samber/lo"
)

type unassignableError struct {
}

func (err unassignableError) Error() string {
	return "not all variables can be assigned a room"
}

func verify(timetable [][6]uint64, modelInput ModelInput) bool {
	//** Initialize dependencies
	evaluator := newPredicateEvaluator(
		modelInput,
		0,
	)

	//** Extract attributes's domains
	totalPeriods, totalDays, _, _, totalGroups, totalRooms := getAttributes(modelInput)

	//** Initialize derived-lessons
	derivedLessons := make(map[[2]uint64]uint64)

	//** Initialize professor-assistance
	professorAssistance := make(map[uint64][][]bool)
	for professor := range len(modelInput.Professors) {
		professorAssistance[uint64(professor)] = make([][]bool, totalPeriods)
		for i := range professorAssistance[uint64(professor)] {
			professorAssistance[uint64(professor)][i] = make([]bool, totalDays)
		}
	}

	//** Initialize group-assistance
	groupAssistance := make(map[uint64][][]bool)
	for group := range totalGroups {
		groupAssistance[group] = make([][]bool, totalPeriods)
		for i := range groupAssistance[group] {
			groupAssistance[group][i] = make([]bool, totalDays)
		}
	}

	//** Initialize room-assistance
	roomAssistance := make(map[uint64][][]bool)
	for room := range totalRooms {
		roomAssistance[room] = make([][]bool, totalPeriods)
		for i := range roomAssistance[room] {
			roomAssistance[room][i] = make([]bool, totalDays)
		}
	}

	lessonTaught := make(map[[3]uint64]bool)

	for _, positive := range timetable {
		period, day, subjectProfessor, group, room := positive[0], positive[1], positive[3], positive[4], positive[5]
		professor := modelInput.SubjectProfessors[subjectProfessor].Professor
		entryKey := [2]uint64{subjectProfessor, group}

		_, alreadyTaught := lessonTaught[[3]uint64{group, subjectProfessor, day}]
		// Check that:
		// - SubjectProfessor is allowed to teach (or to be taught) in the period and day
		// - Professor is available in the period and day
		// - Professor is not already assisting in the period and day
		// - A group with a common class is not already scheduled in the period and day (no collision)
		// - A subjectProfessor can only teach (or be taught to) a group once a day
		// - Room is not assigned to subjectProfessor
		// - Group does not fit in room
		// - Room must not be already assigned in the period and day
		if !modelInput.Entries[entryKey].Permissibility[period][day] ||
			!evaluator.ProfessorAvailable(subjectProfessor, day, period) ||
			professorAssistance[professor][period][day] ||
			collide(modelInput.GroupsGraph, groupAssistance, group, period, day) ||
			alreadyTaught ||
			!evaluator.Assigned(room, subjectProfessor, group) ||
			!evaluator.Fits(group, room) ||
			roomAssistance[room][period][day] {
			return false
		}

		professorAssistance[professor][period][day] = true           // Store professor assistance
		groupAssistance[group][period][day] = true                   // Store group assistance
		roomAssistance[room][period][day] = true                     // Store room assistance
		derivedLessons[entryKey]++                                   // Store lesson taught
		lessonTaught[[3]uint64{group, subjectProfessor, day}] = true // Store lesson taught
	}

	// Check whether the number of lessons taught for each subjectProfessor is equal to the number of lessons assigned in the curriculum
	for key, value := range modelInput.Entries {
		if derivedLessons[key] != value.Lessons {
			return false
		}
	}
	return true
}

func buildSat(variables uint64, constraints []func(state constraintState) [][]int64, state constraintState) (satInstance sat.SAT, explicitVariables map[int64]bool) {
	satInstance = sat.SAT{
		Variables: variables,
		Clauses:   [][]int64{},
	}

	explicitVariables = make(map[int64]bool)   // Variables that are explicitly stated in the clauses
	constraintsChannel := make(chan [][]int64) // Channel to collect constraints

	// Execute constraints functions on different goroutines to improve performance
	for _, constraint := range constraints {
		go func(constraint func(state constraintState) [][]int64) {
			constraintsChannel <- constraint(state)
		}(constraint)
	}

	// Collect generated constraints
	collectedConstraints := 0
	for clauses := range constraintsChannel {
		for _, clause := range clauses {
			for _, variable := range clause {
				// Check whether the variable is positive, since required explicit variables ought to be positive
				if variable > 0 {
					explicitVariables[variable] = true
				}
			}
		}
		// Append clauses to the SAT instance
		satInstance.Clauses = append(satInstance.Clauses, clauses...)

		// Check whether all constraints have been collected to properly close the channel
		if collectedConstraints++; collectedConstraints == len(constraints) {
			close(constraintsChannel)
		}
	}

	return satInstance, explicitVariables
}

func roomAssignment(solution sat.SATSolution, indexer indexer, evaluator predicateEvaluator, modelInput ModelInput) ([][6]uint64, error) {
	simultaneousVariables, simultaneousRooms, simultaneousRelationships := make(map[[2]uint64][]int64), make(map[[2]uint64][]uint64), make(map[[2]uint64]map[[2]uint64]bool)

	for _, variable := range solution {
		period, day, _, subjectProfessor, group, _ := indexer.Attributes(uint64(variable))
		key := [2]uint64{period, day}
		entryKey := [2]uint64{subjectProfessor, group}

		// Initialize for each new key
		if _, ok := simultaneousVariables[key]; !ok {
			simultaneousVariables[key] = make([]int64, 0)
			simultaneousRooms[key] = make([]uint64, 0)
			simultaneousRelationships[key] = make(map[[2]uint64]bool)
		}

		// Add simultaneous variable
		simultaneousVariables[key] = append(simultaneousVariables[key], variable)

		for _, room := range modelInput.Entries[entryKey].Rooms {
			// Add simultaneous room after verifying it fits the group
			if !slices.Contains(simultaneousRooms[key], room) && evaluator.Fits(group, room) {
				simultaneousRooms[key] = append(simultaneousRooms[key], room)
			}

			pair := [2]uint64{uint64(variable), room}
			if _, ok := simultaneousRelationships[key][pair]; ok {
				log.Panicf("variable-room pair %v~%v must be added only once", variable, room)
			}
			// Add simultaneous relationship after verifying room fits group
			if evaluator.Fits(group, room) {
				simultaneousRelationships[key][pair] = true
			}
		}
	}

	timetable := make([][6]uint64, 0, len(solution))
	for key, variables := range simultaneousVariables {
		rooms := simultaneousRooms[key]
		relationships := simultaneousRelationships[key]

		assignments, err := assignRooms(variables, rooms, relationships)
		if _, ok := err.(unassignableError); ok {
			var builder strings.Builder
			for _, variable := range variables {
				_, _, _, subjectProfessor, group, _ := indexer.Attributes(uint64(variable))
				entryKey := [2]uint64{subjectProfessor, group}

				subject := modelInput.Subjects[modelInput.SubjectProfessors[subjectProfessor].Subject].Name
				fmt.Fprintf(&builder, "\tsubject: %v -> { ", subject)

				for _, room := range modelInput.Entries[entryKey].Rooms {
					if evaluator.Fits(group, room) {
						roomName := modelInput.Rooms[room].Name
						fmt.Fprintf(&builder, "%v, ", roomName)
					}
				}
				builder.WriteString("}\n")
			}
			log.Printf("cannot assign rooms: \n%v\t%v", builder.String(), err)
			return nil, nil
		} else if err != nil {
			return nil, err
		}

		for _, assignment := range assignments {
			variable, room := assignment[0], assignment[1]

			positive := [6]uint64{}
			positive[5] = room
			positive[0], positive[1], positive[2], positive[3], positive[4], _ = indexer.Attributes(variable)

			timetable = append(timetable, positive)
		}
	}

	return timetable, nil
}

func assignRooms(variables []int64, rooms []uint64, relationships map[[2]uint64]bool) ([][2]uint64, error) {
	assignments := make([][2]uint64, 0, len(variables))

	// Build neighbors predicate based on relationships
	neighbors := func(variableAny any, roomAny any) (bool, error) {
		variable := variableAny.(int64)
		room := roomAny.(uint64)

		return relationships[[2]uint64{uint64(variable), room}], nil
	}

	// Transform variables and rooms to slices of any
	variablesAny, roomsAny := lo.Map(variables, func(variable int64, _ int) any { return variable }), lo.Map(rooms, func(room uint64, _ int) any { return room })

	graph, err := bipartitegraph.NewBipartiteGraph(variablesAny, roomsAny, neighbors)
	if err != nil {
		return nil, err
	}

	matching := graph.LargestMatching()

	// Check the matching is a maximum one
	if len(matching) < len(variables) {
		return nil, unassignableError{}
	}

	for _, edge := range matching {
		variableIndex, roomIndex := edge.Node1, edge.Node2-len(variables)
		variable, room := variables[variableIndex], rooms[roomIndex]

		assignments = append(assignments, [2]uint64{uint64(variable), room})
	}

	return assignments, nil
}

func getAttributes(modelInput ModelInput) (periods, days, lessons, subjectProfessors, groups, rooms uint64) {
	periods = uint64(len(modelInput.Professors[0].Availability))
	days = uint64(len(modelInput.Professors[0].Availability[0]))
	subjectProfessors = uint64(len(modelInput.SubjectProfessors))
	groups = uint64(len(modelInput.Groups))
	rooms = uint64(len(modelInput.Rooms))

	lessons = 0
	for _, value := range modelInput.Entries {
		if value.Lessons > lessons {
			lessons = value.Lessons
		}
	}

	return periods, days, lessons, subjectProfessors, groups, rooms
}

// Checks whether two groups sharing a common class (or classes) are scheduled in the same period and day
func collide(groupsGraph [][]bool, groupAssistance map[uint64][][]bool, group, period, day uint64) bool {
	for neighborGroup, notDisjoint := range groupsGraph[group] {
		if notDisjoint && groupAssistance[uint64(neighborGroup)][period][day] {
			return true
		}
	}
	return false
}
