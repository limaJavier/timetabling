package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/limaJavier/timetabling/pkg/model"
	"github.com/limaJavier/timetabling/pkg/sat"
	"github.com/samber/lo"
)

var Days = map[uint64]string{
	0: "Monday",
	1: "Tuesday",
	2: "Wednesday",
	3: "Thursday",
	4: "Friday",
	5: "Saturday",
	6: "Sunday",
}

var (
	roomSimilarity  float32
	validStrategies = []string{"pure", "postponed", "hybrid"}
	validSolvers    = []string{"kissat", "cadical", "minisat", "cryptominisat", "glucosesimp", "glucosesyrup", "slime", "ortoolsat"}
	timetablers     = map[string]func(sat.SATSolver) model.Timetabler{
		"pure": model.NewEmbeddedRoomTimetabler,
		"postponed": func(solver sat.SATSolver) model.Timetabler {
			return model.NewIsolatedRoomTimetabler(solver, false, 0)
		},
		"hybrid": func(solver sat.SATSolver) model.Timetabler {
			return model.NewIsolatedRoomTimetabler(solver, true, roomSimilarity)
		},
	}
	solvers = map[string]func() sat.SATSolver{
		"kissat":        sat.NewKissatSolver,
		"cadical":       sat.NewCadicalSolver,
		"minisat":       sat.NewMinisatSolver,
		"cryptominisat": sat.NewCryptominisatSolver,
		"glucosesimp":   sat.NewGlucoseSimpSolver,
		"glucosesyrup":  sat.NewGlucoseSyrupSolver,
		"slime":         sat.NewSlimeSolver,
		"ortoolsat":     sat.NewOrtoolsatSolver,
	}
)

func main() {
	setConfigPath()
	// Define arguments
	strategyPtr := flag.String("strategy", "pure", `Strategy to build the timetable. Allowed values are: 
- "pure" (All restrictions and assigments are guarenteed by the SAT, therefore a solution will be found if it exists), 
- "postponed"(Room assigment will be postponed. Correctness is not guaranteed) and 
- "hybrid"(Room assignment is postponed, but similarity restriction are imposed in the SAT. Correctness is not guaranteed), where \"pure\" is the default`)
	solverPtr := flag.String("solver", "kissat", "SAT-Solver to use. Allowed values are: \"kissat\", \"cadical\", \"minisat\", \"cryptominisat\", \"glucosesimp\", \"glucosesyrup\", \"slime\", \"ortoolsat\", where \"kissat\" is the default")
	roomSimilarityPtr := flag.Float64("similarity", 0.5, "Similarity threshold (between 0 and 1) used by the hybrid strategy, where 0.5 is the default")
	filePathPtr := flag.String("file", "", "Path to the input file")
	outFilePathPtr := flag.String("out", "", "Path to the file where the output will be written; if empty, it'll be written into the Standard Output")
	flag.Parse()
	strategy := strings.ToLower(*strategyPtr)
	solverStr := strings.ToLower(*solverPtr)
	roomSimilarity = float32(*roomSimilarityPtr)
	filePath := *filePathPtr
	outFile := *outFilePathPtr

	// Validate arguments
	if !slices.Contains(validStrategies, strategy) {
		log.Fatalf("%v is not a valid strategy", strategy)
	} else if !slices.Contains(validSolvers, solverStr) {
		log.Fatalf("%v is not a valid solver", solverStr)
	} else if filePath == "" {
		log.Fatal("an input file must be specified")
	} else if strategy == "hybrid" && (roomSimilarity <= 0 || roomSimilarity >= 1) {
		log.Fatalf("room-similarity must be greater than 0 and smaller than 1: %v", roomSimilarity)
	}

	// Extract input
	input, err := model.InputFromJson(filePath)
	if err != nil {
		log.Fatalf("cannot parse input file: %v", err)
	}

	// Initialize engines
	solver := solvers[solverStr]()
	timetabler := timetablers[strategy](solver)

	// Build timetable
	timetable, variables, clauses, err := timetabler.Build(input)

	if err != nil {
		log.Fatalf("an error occurred during timetable construction: %v", err)
	} else if timetable == nil {
		fmt.Printf("Variables: %v\n", variables)
		fmt.Printf("Clauses: %v\n", clauses)
		os.Exit(20)
	}

	// Verify timetable correctness
	if !timetabler.Verify(timetable, input) {
		fmt.Printf("Variables: %v\n", variables)
		fmt.Printf("Clauses: %v\n", clauses)
		os.Exit(15)
	}

	compare := func(a, b [6]uint64, i int) int {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return 1
		}
		return 0
	}

	slices.SortFunc(timetable, func(a, b [6]uint64) int {
		dayComparison := compare(a, b, 1)
		if dayComparison != 0 {
			return dayComparison
		}
		return compare(a, b, 0)
	})

	// Build output from timetable
	perClassTimetable := make(map[uint64][]map[string]uint64)
	for _, positive := range timetable {
		period := positive[0]
		day := positive[1]
		subjectProfessor := positive[3]
		subject := input.Subjects[input.SubjectProfessors[subjectProfessor].Subject].Id
		professor := input.Professors[input.SubjectProfessors[subjectProfessor].Professor].Id
		group := positive[4]
		room := positive[5]

		// dayName := Days[day]
		// subjectProfessorName := fmt.Sprintf("%v~%v",
		// 	input.Subjects[input.SubjectProfessors[subjectProfessor].Subject].Name,
		// 	input.Professors[input.SubjectProfessors[subjectProfessor].Professor].Name,
		// )
		// roomName := input.Rooms[room].Name

		for _, class := range input.Groups[group].Classes {
			if _, ok := perClassTimetable[class]; !ok {
				perClassTimetable[class] = make([]map[string]uint64, 0)
			}
			perClassTimetable[class] = append(perClassTimetable[class], map[string]uint64{
				"period":    period,
				"day":       day,
				"subject":   subject,
				"professor": professor,
				"room":      room,
			})

			// className := input.Classes[class].Name
			// if !strings.Contains(className, "cc4") {
			// 	continue
			// }
			// fmt.Printf("Period: %v, Day: %v, Lesson: %v, SubjectProfessor: %v, Class: %v, Room: %v \n", period, dayName, positive[2], subjectProfessorName, className, roomName)
		}
	}

	// Marshal output into json
	perClassTimetableJson, err := json.Marshal(perClassTimetable)
	if err != nil {
		log.Fatalf("an error occurred while building output json: %v", err)
	}

	// Verify outfile is empty, if so then write the results to the Standard Output
	if outFile == "" {
		fmt.Println(string(perClassTimetableJson))
	} else {
		err := os.WriteFile(outFile, perClassTimetableJson, 0666)
		if err != nil {
			log.Fatalf("an error occurred while writing to the output file: %v", err)
		}
	}

	fmt.Printf("Variables: %v\n", variables)
	fmt.Printf("Clauses: %v\n", clauses)
	os.Exit(10)
}

func setConfigPath() {
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("cannot determine executable path: %v", err)
	}
	execPath = path.Dir(execPath)

	// Verify config.json exists
	files, err := os.ReadDir(execPath)
	if err != nil {
		log.Fatalf("cannot read executable's directory: %v", err)
	}
	fileNames := lo.Map(files, func(file os.DirEntry, _ int) string { return file.Name() })

	if !slices.Contains(fileNames, "config.json") {
		log.Fatalf("config.json file was not found: %v", fileNames)
	}

	sat.ConfigPath = execPath + "/config.json"
}
