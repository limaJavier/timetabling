package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/limaJavier/timetabling/pkg/model"
	"github.com/limaJavier/timetabling/pkg/sat"

	"github.com/samber/lo"
)

const satisfiableTestDirectory = "../../test/out/satisfiable/"
const unsatisfiableTestDirectory = "../../test/out/unsatisfiable/"
const KB = 1024
const MB float32 = 1024 * 1024

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
	Ortoolsat
	GlucoseSimp
	GlucoseSyrup
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
		Ortoolsat:     "Ortoolsat",
		GlucoseSimp:   "GlucoseSimp",
		GlucoseSyrup:  "GlucoseSyrup",
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
	Memory     float32
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
	tests := getTests()
	timetablers := getTimetablers()
	solvers := getSolvers()
	results := make([]BenchmarkResult, 0, len(tests)*len(timetablers)*len(solvers))

	var testName, solverName, timetablerName string

	defer func() {
		if panic := recover(); panic != nil {
			fmt.Printf("a panic occurred with test \"%v\", solver \"%v\" and timetabler \"%v\": %v\n", testName, solverName, timetablerName, panic)
		}
	}()
	for _, test := range tests {
		testName = test.Name
		input := test.Input

		for _, solverMeta := range solvers {
			solverName = solverTypes[solverMeta.Type]
			solver := solverMeta.Constructor()
			for _, timetablerMeta := range timetablers {
				timetablerName = timetablerTypes[timetablerMeta.Type]
				timetabler := timetablerMeta.Constructor(solver)

				duration, memory, result := measure(timetabler, input)

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
		{
			Type:        Ortoolsat,
			Constructor: sat.NewOrtoolsatSolver,
		},
		{
			Type:        GlucoseSimp,
			Constructor: sat.NewGlucoseSimpSolver,
		},
		{
			Type:        GlucoseSyrup,
			Constructor: sat.NewGlucoseSyrupSolver,
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
				RoomSimilarityThreshold: 0.35,
			},
			Constructor: func(solver sat.SATSolver) model.Timetabler {
				return model.NewIsolatedRoomTimetabler(solver, true, 0.35)
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

		{
			TimetablerMetadata: TimetablerMetadata{
				Type:                    HybridIsolated,
				RoomSimilarityThreshold: 0.75,
			},
			Constructor: func(solver sat.SATSolver) model.Timetabler {
				return model.NewIsolatedRoomTimetabler(solver, true, 0.75)
			},
		},
	}
}

func measure(timetabler model.Timetabler, input model.ModelInput) (duration int64, memory float32, result ResultType) {

	var memoryStart, memoryEnd runtime.MemStats
	runtime.GC() // Clean up before measuring
	runtime.ReadMemStats(&memoryStart)

	start := time.Now()
	timetable, err := timetabler.Build(input)
	duration = time.Since(start).Milliseconds()
	runtime.ReadMemStats(&memoryEnd)
	memory = float32(memoryEnd.Alloc-memoryStart.Alloc) / MB
	if err != nil {
		panic("an error occurred during timetable-building: " + err.Error())
	} else if timetable == nil {
		return duration, memory, Unsatisfiable
	}

	if !timetabler.Verify(timetable, input) {
		panic("timetable verification failed")
	}
	return duration, memory, Solved
}

func toCsv(results []BenchmarkResult) {
	file, err := os.Create("benchmark_results.csv")
	if err != nil {
		log.Panicf("cannot create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"Solver", "Timetabler", "Room-Similarity Threshold", "Test", "Satisfiable", "Subjects", "Professors", "SubjectProfessors", "Rooms", "Classes", "Duration(ms)", "Memory(MB)", "Result"}
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
			fmt.Sprintf("%.1f", result.Memory),
			resultTypes[result.Result],
		}
		if err := writer.Write(record); err != nil {
			log.Panicf("cannot write CSV record: %v", err)
		}
	}
}
