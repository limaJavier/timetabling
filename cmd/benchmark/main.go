package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
	"timetabling/internal/model"
	"timetabling/internal/sat"

	"github.com/samber/lo"
)

const satisfiableTestDirectory = "../../test/out/satisfiable/"
const unsatisfiableTestDirectory = "../../test/out/unsatisfiable/"
const KB = 1024

type TimetablerType int

const (
	Embedded TimetablerType = iota
	Isolated
	HybridIsolated
)

type SolverType int

const (
	Kissat SolverType = iota
	Cadical
	Minisat
	Cryptominisat
	Slime
)

type ResultType int

const (
	Solved ResultType = iota
	Unsatisfiable
	Timeout
)

var (
	timetablerTypes = map[TimetablerType]string{
		Embedded:       "Embedded",
		Isolated:       "Isolated",
		HybridIsolated: "Hybrid Isolated",
	}
	solverTypes = map[SolverType]string{
		Kissat:        "Kissat",
		Cadical:       "Cadical",
		Minisat:       "Minisat",
		Cryptominisat: "Cryptominisat",
		Slime:         "Slime",
	}
	resultTypes = map[ResultType]string{
		Solved:        "Solved",
		Unsatisfiable: "Unsatisfiable",
		Timeout:       "Timeout",
	}
)

type TestMetadata struct {
	Name              string
	Satisfiable       bool
	Subjects          int
	Professors        int
	SubjectProfessors int
	Rooms             int
	Classes           int
}

type TimetablerMetadata struct {
	Type                    TimetablerType
	RoomSimilarityThreshold float32
}

type BenchmarkResult struct {
	Solver     SolverType
	Timetabler TimetablerMetadata
	Test       TestMetadata
	Duration   int64
	Memory     uint64
	Result     ResultType
}

type timetablerMetadata struct {
	TimetablerMetadata
	Constructor func(solver sat.SATSolver) model.Timetabler
}

type solverMetadata struct {
	Type        SolverType
	Constructor func() sat.SATSolver
}

type testMetadata struct {
	Name        string
	Satisfiable bool
	Input       model.ModelInput
}

func main() {
	preprocessor := model.NewPreprocessor()

	tests := getTests()
	timetablers := getTimetablers()
	solvers := getSolvers()
	results := make([]BenchmarkResult, 0, len(tests)*len(timetablers)*len(solvers))

	for _, test := range tests {
		input := test.Input
		curriculum, groups, groupsGraph := preprocess(preprocessor, input)

		for _, solverMeta := range solvers {
			solver := solverMeta.Constructor()
			for _, timetablerMeta := range timetablers {
				timetabler := timetablerMeta.Constructor(solver)

				duration, memory, result := measure(timetabler, input, curriculum, groups, groupsGraph)

				results = append(results, BenchmarkResult{
					Solver:     solverMeta.Type,
					Timetabler: timetablerMeta.TimetablerMetadata,
					Test: TestMetadata{
						Name:              test.Name,
						Satisfiable:       test.Satisfiable,
						Subjects:          len(input.Subjects),
						Professors:        len(input.Professors),
						SubjectProfessors: len(input.SubjectProfessors),
						Rooms:             len(input.Rooms),
						Classes:           len(input.Classes),
					},
					Duration: duration,
					Memory:   memory,
					Result:   result,
				})
			}
		}
	}

	toCsv(results)
}

func getTests() []testMetadata {
	tests := make([]testMetadata, 0)
	for _, tuple := range lo.Zip2([]string{satisfiableTestDirectory, unsatisfiableTestDirectory}, []bool{true, false}) {
		directory, satisfiable := tuple.A, tuple.B
		testFiles, err := os.ReadDir(directory)
		if err != nil {
			log.Fatalf("cannot read directory: %v", err)
		}

		for _, file := range testFiles {
			filename := directory + file.Name()
			input, err := model.InputFromJson(filename)
			if err != nil {
				panic("cannot parse input file: " + err.Error())
			}

			name := file.Name()
			if !satisfiable {
				name = "unsatisfiable_" + name
			}
			tests = append(tests, testMetadata{
				Name:        name,
				Satisfiable: satisfiable,
				Input:       input,
			})
		}
	}

	return tests
}

func getSolvers() []solverMetadata {
	return []solverMetadata{
		{
			Type:        Kissat,
			Constructor: sat.NewKissatSolver,
		},
		{
			Type:        Cadical,
			Constructor: sat.NewCadicalSolver,
		},
		{
			Type:        Minisat,
			Constructor: sat.NewMinisatSolver,
		},
		{
			Type:        Cryptominisat,
			Constructor: sat.NewCryptominisatSolver,
		},
		{
			Type:        Slime,
			Constructor: sat.NewSlimeSolver,
		},
	}
}

func getTimetablers() []timetablerMetadata {
	return []timetablerMetadata{
		{
			TimetablerMetadata: TimetablerMetadata{
				Type: Embedded,
			},
			Constructor: model.NewEmbeddedRoomTimetabler,
		},

		{
			TimetablerMetadata: TimetablerMetadata{
				Type: Isolated,
			},
			Constructor: func(solver sat.SATSolver) model.Timetabler {
				return model.NewIsolatedRoomTimetabler(solver, false, 0)
			},
		},

		{
			TimetablerMetadata: TimetablerMetadata{
				Type:                    HybridIsolated,
				RoomSimilarityThreshold: 0.5,
			},
			Constructor: func(solver sat.SATSolver) model.Timetabler {
				return model.NewIsolatedRoomTimetabler(solver, true, 0.5)
			},
		},
	}
}

func preprocess(preprocessor model.Preprocessor, input model.ModelInput) (curriculum [][]bool, groups map[uint64][]uint64, groupsGraph [][]bool) {
	curriculum, groups = preprocessor.ExtractCurriculumAndGroups(input)
	groupsGraph = preprocessor.BuildGroupsGraph(groups)
	return curriculum, groups, groupsGraph
}

func measure(
	timetabler model.Timetabler,
	input model.ModelInput,
	curriculum [][]bool,
	groups map[uint64][]uint64,
	groupsGraph [][]bool,
) (duration int64, memory uint64, result ResultType) {

	var memoryStart, memoryEnd runtime.MemStats
	runtime.GC() // Clean up before measuring
	runtime.ReadMemStats(&memoryStart)

	start := time.Now()
	timetable, err := timetabler.Build(input, curriculum, groups, groupsGraph)
	duration = time.Since(start).Milliseconds()
	runtime.ReadMemStats(&memoryEnd)
	memory = memoryEnd.Alloc - memoryStart.Alloc
	if err != nil {
		panic("an error occurred during timetable-building: " + err.Error())
	} else if timetable == nil {
		return duration, memory, Unsatisfiable
	}

	if !timetabler.Verify(timetable, input, curriculum, groups, groupsGraph) {
		panic("timetable verification failed")
	}
	return duration, memory / KB, Solved
}

func toCsv(results []BenchmarkResult) {
	file, err := os.Create("benchmark_results.csv")
	if err != nil {
		log.Panicf("cannot create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"Solver", "Timetabler", "Room-Similarity Threshold", "Test", "Satisfiable", "Subjects", "Professors", "SubjectProfessors", "Rooms", "Classes", "Duration(ms)", "Memory(KB)", "Result"}
	if err := writer.Write(header); err != nil {
		log.Panicf("cannot write CSV header: %v", err)
	}

	for _, result := range results {
		record := []string{
			solverTypes[result.Solver],
			timetablerTypes[result.Timetabler.Type],
			fmt.Sprintf("%f", result.Timetabler.RoomSimilarityThreshold),
			result.Test.Name,
			fmt.Sprintf("%v", result.Test.Satisfiable),
			fmt.Sprintf("%d", result.Test.Subjects),
			fmt.Sprintf("%d", result.Test.Professors),
			fmt.Sprintf("%d", result.Test.SubjectProfessors),
			fmt.Sprintf("%d", result.Test.Rooms),
			fmt.Sprintf("%d", result.Test.Classes),
			fmt.Sprintf("%d", result.Duration),
			fmt.Sprintf("%d", result.Memory),
			resultTypes[result.Result],
		}
		if err := writer.Write(record); err != nil {
			log.Panicf("cannot write CSV record: %v", err)
		}
	}
}
