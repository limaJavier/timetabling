package model

// Curriculum matrix
// Professor available matrix
// Allocation matrix can be built from the curriculum matrix

type matrixPredicateEvaluator struct {
	availability    map[uint64][][]bool // Professor's availability for each period of each day
	rooms           map[uint64]uint64   // Room assigned to subjectProfessor
	professors      map[uint64]uint64   // Professor belonging to subjectProfessor
	allocations     map[uint64][][]bool // Allocation matrix per class
	togetherClasses map[[3]uint64]bool  // <subjectProfessor, class1, class2> format. It holds whether class1 and class2 must take subjectProfessor together
}

func newMatrixPredicateEvaluator(
	curriculum [][]uint64,
	groups map[uint64][][]uint64,
	availability map[uint64][][]bool,
	rooms map[uint64]uint64,
	professors map[uint64]uint64,
	lessons uint64,
	subjectProfessors uint64,
) *matrixPredicateEvaluator {

	evaluator := matrixPredicateEvaluator{
		availability: availability,
		rooms:        rooms,
		professors:   professors,
	}

	evaluator.buildTogetherClasses(groups)

	evaluator.buildAllocations(curriculum, lessons, subjectProfessors)

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

func (evaluator *matrixPredicateEvaluator) Teaches(class, subjectProfessor, lesson uint64) bool {
	allocation, ok := evaluator.allocations[class]
	if !ok {
		panic("class not found")
	}
	return allocation[subjectProfessor][lesson]
}

func (evaluator *matrixPredicateEvaluator) Together(subjectProfessor uint64, classes ...uint64) bool {
	// Take baseClass and check with all other classes, since "Together" relationship is transitive
	baseClass := classes[0]
	for _, class := range classes {
		if class != baseClass {
			_, ok1 := evaluator.togetherClasses[[3]uint64{subjectProfessor, baseClass, class}]
			_, ok2 := evaluator.togetherClasses[[3]uint64{subjectProfessor, class, baseClass}]

			if !ok1 && !ok2 {
				return false
			}
		}
	}
	return true
}

// TODO: (Optional) Test this method
func (evaluator *matrixPredicateEvaluator) buildAllocations(curriculum [][]uint64, lessons uint64, subjectProfessors uint64) {
	evaluator.allocations = make(map[uint64][][]bool) // Initialize dictionary
	for class := range curriculum {                   // For each class
		evaluator.allocations[uint64(class)] = make([][]bool, subjectProfessors) // Initialize allocation per class

		for subjectProfessor := range curriculum[class] { // For each subjectProfessor
			evaluator.allocations[uint64(class)][subjectProfessor] = make([]bool, lessons) // Initialize subjectProfessor row
			for i := range curriculum[class][subjectProfessor] {
				evaluator.allocations[uint64(class)][subjectProfessor][i] = true // Set to true the first j lessons where j is the number of lessons assigned for "subjectProfessor" to teach to "class" (i.e. curriculum[class][subjectProfessor])
			}
		}
	}
}

// TODO: Test this method
func (evaluator *matrixPredicateEvaluator) buildTogetherClasses(groups map[uint64][][]uint64) {
	evaluator.togetherClasses = make(map[[3]uint64]bool)
	for subjectProfessor, associatedGroups := range groups {
		for _, group := range associatedGroups {
			for i := 0; i < len(group)-1; i++ {
				for j := i + 1; j < len(group); j++ {
					evaluator.togetherClasses[[3]uint64{subjectProfessor, group[i], group[j]}] = true
				}
			}
		}
	}
}
