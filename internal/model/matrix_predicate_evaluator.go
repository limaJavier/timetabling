package model

// Curriculum matrix
// Professor available matrix
// Allocation matrix can be built from the curriculum matrix

type matrixPredicateEvaluator struct {
	availability map[uint64][][]bool // Professor's availability for each period of each day
	rooms        map[uint64]uint64   // Room assigned to subjectProfessor
	professors   map[uint64]uint64   // Professor belonging to subjectProfessor
	allocations  map[uint64][][]bool // Allocation matrix per group
}

func newMatrixPredicateEvaluator(
	availability map[uint64][][]bool,
	rooms map[uint64]uint64,
	professors map[uint64]uint64,
	curriculum [][]uint64,
	lessons uint64,
	subjectProfessors uint64,
) *matrixPredicateEvaluator {

	evaluator := matrixPredicateEvaluator{
		availability: availability,
		rooms:        rooms,
		professors:   professors,
	}

	evaluator.allocations = make(map[uint64][][]bool) // Initialize dictionary
	for group := range curriculum {                   // For each group
		evaluator.allocations[uint64(group)] = make([][]bool, subjectProfessors) // Initialize allocation per group

		for subjectProfessor := range curriculum[group] { // For each subjectProfessor
			evaluator.allocations[uint64(group)][subjectProfessor] = make([]bool, lessons) // Initialize subjectProfessor row
			for i := range curriculum[group][subjectProfessor] {
				evaluator.allocations[uint64(group)][subjectProfessor][i] = true // Set to true the first j lessons where j is the number of lessons assigned for "subjectProfessor" to teach to "group" (i.e. curriculum[group][subjectProfessor])
			}
		}
	}

	return &evaluator
}

func (evaluator *matrixPredicateEvaluator) SameProfessor(subjectProfessor1, subjectProfessor2 uint64) bool {
	professor1, ok1 := evaluator.professors[subjectProfessor1]
	professor2, ok2 := evaluator.professors[subjectProfessor2]
	if !ok1 || !ok2 {
		panic("subject-professor not found")
	}
	return professor1 == professor2
}

func (evaluator *matrixPredicateEvaluator) ProfessorAvailable(subjectProfessor, day, period uint64) bool {
	professor, ok := evaluator.professors[subjectProfessor]
	if !ok {
		panic("subject-professor not found")
	}

	distribution, ok := evaluator.availability[professor]
	if !ok {
		panic("professor not found")
	}

	return distribution[period][day]
}

func (evaluator *matrixPredicateEvaluator) SameRoom(subjectProfessor1, subjectProfessor2 uint64) bool {
	room1, ok1 := evaluator.rooms[subjectProfessor1]
	room2, ok2 := evaluator.rooms[subjectProfessor2]
	if !ok1 || !ok2 {
		panic("subject-professor not found")
	}
	return room1 == room2
}

func (evaluator *matrixPredicateEvaluator) Teaches(group, subjectProfessor, lesson uint64) bool {
	allocation, ok := evaluator.allocations[group]
	if !ok {
		panic("group not found")
	}
	return allocation[subjectProfessor][lesson]
}
