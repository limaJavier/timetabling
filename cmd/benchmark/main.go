package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/limaJavier/timetabling/pkg/model"

	"github.com/samber/lo"
)

const (
	executablePath                     = "../../bin/timetabler"
	satisfiableTestDirectory           = "../../test/out/satisfiable/"
	unsatisfiableTestDirectory         = "../../test/out/unsatisfiable/"
	KB                                 = 1024
	MB                         float32 = 1024 * 1024
)

type TimetablerType int

const (
	pure TimetablerType = iota
	postponed
	hybrid
)

type SolverType int

const (
	kissat SolverType = iota
	cadical
	minisat
	cryptominisat
	slime
	ortoolsat
	glucosesimp
	glucosesyrup
)

type ResultType int

const (
	satisfiable ResultType = iota
	unsatisfiable
	verificationFailed
)

var (
	timetablerTypes = map[TimetablerType]string{
		pure:      "pure",
		postponed: "postponed",
		hybrid:    "hybrid",
	}
	solverTypes = map[SolverType]string{
		kissat:        "kissat",
		cadical:       "cadical",
		minisat:       "minisat",
		cryptominisat: "cryptominisat",
		slime:         "slime",
		ortoolsat:     "ortoolsat",
		glucosesimp:   "glucosesimp",
		glucosesyrup:  "glucosesyrup",
	}
	resultTypes = map[ResultType]string{
		satisfiable:        "satisfiable",
		unsatisfiable:      "unsatisfiable",
		verificationFailed: "verification-failed",
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
	Solver        SolverType
	Timetabler    TimetablerMetadata
	Test          TestMetadata
	Variables     int64
	Clauses       int64
	Duration      int64
	Memory        float32
	CpuPercentage int64
	Result        ResultType
}

func main() {
	tests := getTests()
	timetablers := getTimetablers()
	solvers := getSolvers()
	results := make([]BenchmarkResult, 0, len(tests)*len(timetablers)*len(solvers))

	for _, test := range tests {
		for _, timetabler := range timetablers {
			for _, solver := range solvers {
				log.Printf("Benchmarking test \"%v\" with strategy \"%v\", solver \"%v\" and similarity \"%v\"\n", test.Name, timetablerTypes[timetabler.Type], solverTypes[solver], timetabler.RoomSimilarityThreshold)

				variables, clauses, duration, maxMemory, cpuPercentage, result := measure(timetabler.Type, solver, timetabler.RoomSimilarityThreshold, test.Name)

				results = append(results, BenchmarkResult{
					Solver:        solver,
					Timetabler:    timetabler,
					Test:          test,
					Variables:     variables,
					Clauses:       clauses,
					Duration:      duration,
					Memory:        maxMemory,
					CpuPercentage: cpuPercentage,
					Result:        result,
				})
			}
		}
	}

	toCsv(results)
}

func getTests() []TestMetadata {
	tests := make([]TestMetadata, 0)
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
				log.Fatalf("cannot parse input file: %v", err)
			}

			tests = append(tests, TestMetadata{
				Name:              filename,
				Satisfiable:       satisfiable,
				Subjects:          len(input.Subjects),
				Professors:        len(input.Professors),
				SubjectProfessors: len(input.SubjectProfessors),
				Rooms:             len(input.Rooms),
				Classes:           len(input.Classes),
			})
		}
	}

	return tests
}

func getSolvers() []SolverType {
	return []SolverType{kissat, cadical, minisat, cryptominisat, slime, ortoolsat, glucosesimp, glucosesyrup}
}

func getTimetablers() []TimetablerMetadata {
	return []TimetablerMetadata{
		{
			Type: pure,
		},

		{
			Type: postponed,
		},

		{
			Type:                    hybrid,
			RoomSimilarityThreshold: 0.35,
		},

		{
			Type:                    hybrid,
			RoomSimilarityThreshold: 0.5,
		},

		{
			Type:                    hybrid,
			RoomSimilarityThreshold: 0.75,
		},
	}
}

func measure(timetable TimetablerType, solver SolverType, roomSimilarity float32, testFile string) (variables, clauses, duration int64, maxMemory float32, cpuPercentage int64, result ResultType) {
	cmd := exec.Command("/usr/bin/time", "-v", executablePath, "-strategy", timetablerTypes[timetable], "-solver", solverTypes[solver], "-similarity", fmt.Sprint(roomSimilarity), "-file", testFile)

	var stdOut bytes.Buffer
	cmd.Stdout = &stdOut
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr

	cmd.Run()
	// Exit code of 10 stands for satisfiable, 20 for unsatisfiable and 15 for verification-failed (that is a timetable was generated but it was not correct)
	if cmd.ProcessState.ExitCode() != 10 && cmd.ProcessState.ExitCode() != 20 && cmd.ProcessState.ExitCode() != 15 {
		log.Fatalf("an error occurred during the execution \"timetable\" at test \"%v\" using strategy \"%v\", solver \"%v\", room-similarity \"%v\": %v\n", testFile, timetablerTypes[timetable], solverTypes[solver], roomSimilarity, stdErr.String())
	} else if cmd.ProcessState.ExitCode() == 20 {
		result = unsatisfiable

	} else if cmd.ProcessState.ExitCode() == 15 {
		result = verificationFailed
	} else {
		result = satisfiable
	}

	splits := strings.Split(stdErr.String(), "\n")
	splits = append(splits, strings.Split(stdOut.String(), "\n")...)

	getLine := func(substr string) string {
		line, ok := lo.Find(splits, func(line string) bool {
			return strings.Contains(strings.ToLower(line), substr)
		})
		if !ok {
			log.Fatalf("Substring \"%v\" could not be found", substr)
		}
		return line
	}

	duration = parseDurationLine(getLine("wall clock"))
	maxMemory = parseMemoryLine(getLine("maximum resident set size"))
	cpuPercentage = parseCpuPercentageLine(getLine("percent of cpu"))
	variables = parseVariablesLine(getLine("variables:"))
	clauses = parseClausesLine(getLine("clauses:"))

	return variables, clauses, duration, maxMemory, cpuPercentage, result
}

func toCsv(results []BenchmarkResult) {
	file, err := os.Create("benchmark_results.csv")
	if err != nil {
		log.Panicf("cannot create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"Solver", "Strategy", "Room-Similarity Threshold", "Test", "Satisfiable", "Subjects", "Professors", "Subject-Professors", "Rooms", "Classes", "Variables", "Clauses", "Duration(ms)", "Memory(MB)", "CPU(%)", "Result"}
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
			fmt.Sprintf("%d", result.Variables),
			fmt.Sprintf("%d", result.Clauses),
			fmt.Sprintf("%d", result.Duration),
			fmt.Sprintf("%.1f", result.Memory),
			fmt.Sprintf("%d", result.CpuPercentage),
			resultTypes[result.Result],
		}
		if err := writer.Write(record); err != nil {
			log.Panicf("cannot write CSV record: %v", err)
		}
	}
}

func parseDurationLine(line string) int64 {
	durationStr := strings.Split(line, "(h:mm:ss or m:ss):")[1][1:]
	return parseDuration(durationStr)
}

func parseDuration(durationStr string) int64 {
	parts := strings.Split(durationStr, ":")
	secondsStr := parts[len(parts)-1]
	secondsParts := strings.Split(secondsStr, ".")

	var duration int64
	if len(parts) == 3 { // h:mm:ss
		hours := lo.Must(strconv.Atoi(parts[0]))
		minutes := lo.Must(strconv.Atoi(parts[1]))
		seconds := lo.Must(strconv.Atoi(secondsParts[0]))
		hundredthOfSeconds := lo.Must(strconv.Atoi(secondsParts[1]))
		duration = int64(hours*3600+minutes*60+seconds)*1000 + int64(hundredthOfSeconds*10)
	} else if len(parts) == 2 { // m:ss
		minutes := lo.Must(strconv.Atoi(parts[0]))
		seconds := lo.Must(strconv.Atoi(secondsParts[0]))
		hundredthOfSeconds := lo.Must(strconv.Atoi(secondsParts[1]))
		duration = int64(minutes*60+seconds)*1000 + int64(hundredthOfSeconds*10)
	} else {
		log.Fatalf("unexpected duration format: %v", durationStr)
	}
	return duration
}

func parseMemoryLine(line string) float32 {
	memoryStr := strings.Split(line, ":")[1][1:]
	return float32(lo.Must(strconv.ParseFloat(memoryStr, 32))) / 1024
}

func parseCpuPercentageLine(line string) int64 {
	percentageStr := strings.Split(line, ":")[1][1:]
	percentageStr = percentageStr[:len(percentageStr)-1]
	return int64(lo.Must(strconv.Atoi(percentageStr)))
}

func parseVariablesLine(line string) int64 {
	variablesStr := strings.Split(line, ":")[1][1:]
	return int64(lo.Must(strconv.Atoi(variablesStr)))
}

func parseClausesLine(line string) int64 {
	clausesStr := strings.Split(line, ":")[1][1:]
	return int64(lo.Must(strconv.Atoi(clausesStr)))
}
