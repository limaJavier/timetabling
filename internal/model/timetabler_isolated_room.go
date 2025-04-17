package model

import (
	"errors"
	"fmt"
	"log"
	"slices"
	"timetabling/internal/sat"

	"github.com/onsi/gomega/matchers/support/goraph/bipartitegraph"
	"github.com/samber/lo"
)

type isolatedRoomTimetabler struct {
	solver sat.SATSolver
}

func NewIsolatedRoomTimetabler(solver sat.SATSolver) Timetabler {
	return &isolatedRoomTimetabler{
		solver: solver,
	}
}

func (timetabler *isolatedRoomTimetabler) Build(
	modelInput ModelInput,
	curriculum [][]bool,
	groups map[uint64][]uint64,
	groupsGraph [][]bool,
) ([][6]uint64, error) {

	//** Extract attributes's domains
	totalRooms := uint64(1)
	totalPeriods, totalDays, totalLessons, totalSubjectProfessors, totalGroups, _ := getAttributes(modelInput, curriculum)

	//** Initialize dependencies
	evaluator := NewPredicateEvaluatorIsolatedRoom(
		modelInput,
		curriculum,
		groups,
		groupsGraph,
	)
	indexer := NewIndexer(totalPeriods, totalDays, totalLessons, totalSubjectProfessors, totalGroups, totalRooms)
	generator := NewPermutationGenerator(totalPeriods, totalDays, totalLessons, totalSubjectProfessors, totalGroups, totalRooms)

	//** Build SAT instance
	variables := totalPeriods * totalDays * totalLessons * totalSubjectProfessors * totalGroups * totalRooms

	// Constraints functions
	constraints := []func(state ConstraintState) [][]int64{
		professorConstraints,
		studentConstraints,
		subjectPermissibilityConstraints,
		professorAvailabilityConstraints,
		lessonConstraints,
		completenessConstraints,
		negationConstraints,
		uniquenessConstraints,
	}

	state := ConstraintState{
		evaluator:         evaluator,
		indexer:           indexer,
		generator:         generator,
		periods:           totalPeriods,
		days:              totalDays,
		lessons:           totalLessons,
		subjectProfessors: totalSubjectProfessors,
		groups:            totalGroups,
		rooms:             totalRooms,
	}

	satInstance, explicitVariables := buildSat(variables, constraints, state)

	log.Println("Start solver") // TODO: Remove when done debugging
	//** Solve SAT instance
	solution, err := timetabler.solver.Solve(satInstance)
	log.Println("Solver done") // TODO: Remove when done debugging
	if err != nil {
		return nil, err
	} else if solution == nil { // Return nil if the SAT instance is not satisfiable
		return nil, nil
	}

	// Filter solution by taking only positive and explicit variables
	solution = lo.Filter(solution, func(variable int64, _ int) bool {
		return variable > 0 && explicitVariables[variable]
	})

	return timetabler.assignRooms(solution, indexer, modelInput)
}

func (timetabler *isolatedRoomTimetabler) Verify(
	timetable [][6]uint64,
	modelInput ModelInput,
	curriculum [][]bool,
	groups map[uint64][]uint64,
	groupsGraph [][]bool,
) bool {
	return verify(timetable, modelInput, curriculum, groups, groupsGraph)
}

func (timetabler *isolatedRoomTimetabler) assignRooms(solution sat.SATSolution, indexer Indexer, modelInput ModelInput) ([][6]uint64, error) {
	simultaneousVariables, simultaneousRooms, simultaneousRelationships := make(map[[2]uint64][]int64), make(map[[2]uint64][]uint64), make(map[[2]uint64]map[[2]uint64]bool)

	for _, variable := range solution {
		period, day, _, subjectProfessor, _, _ := indexer.Attributes(uint64(variable))
		key := [2]uint64{period, day}

		// Initialize for each new key
		if _, ok := simultaneousVariables[key]; !ok {
			simultaneousVariables[key] = make([]int64, 0)
			simultaneousRooms[key] = make([]uint64, 0)
			simultaneousRelationships[key] = make(map[[2]uint64]bool)
		}

		// Add simultaneous variable
		simultaneousVariables[key] = append(simultaneousVariables[key], variable)

		for _, room := range modelInput.SubjectProfessors[subjectProfessor].Rooms {
			// Add simultaneous room
			if !slices.Contains(simultaneousRooms[key], room) {
				simultaneousRooms[key] = append(simultaneousRooms[key], room)
			}

			// Add simultaneous relationship
			pair := [2]uint64{uint64(variable), room}
			if _, ok := simultaneousRelationships[key][pair]; ok {
				panic(fmt.Sprintf("variable-room pair %v~%v must be added only once", variable, room))
			}
			simultaneousRelationships[key][pair] = true
		}
	}

	timetable := make([][6]uint64, 0, len(solution))
	for key, variables := range simultaneousVariables {
		rooms := simultaneousRooms[key]
		relationships := simultaneousRelationships[key]

		assignments, err := assignRooms(variables, rooms, relationships)
		if err != nil {
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
		return nil, errors.New("not all variables can be assigned a room")
	}

	for _, edge := range matching {
		variableIndex, roomIndex := edge.Node1, edge.Node2-len(variables)
		variable, room := variables[variableIndex], rooms[roomIndex]

		assignments = append(assignments, [2]uint64{uint64(variable), room})
	}

	return assignments, nil
}
