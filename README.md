# Timetabling

A SAT-based timetabling library written in Go, developed for the Faculty of Mathematics and Computer Science at the University of Havana. It allows automatic generation of class timetables under various constraints using modern SAT solvers.

## 📦 Installation

### Library

To install the library, run:

```console
$ go get github.com/limaJavier/timetabling
```

Or import it in your Go code as:

```go
import "github.com/limaJavier/timetabling/pkg/model"
```

Then run:

```console
$ go mod tidy
```

### CLI

To install the CLI, clone the repository and run the build script. The compiled binary and configuration files will be placed in the `bin` directory at the project root.

```console
$ git clone git@github.com:limaJavier/timetabling.git
$ cd timetabling
$ ./build.sh 
```

## ⚙️ How to Use

### Library

The library is organized into two main packages: `model` and `sat`.

- `model`: handles input processing and timetabler instantiation.
- `sat`: manages the SAT representation and solver interaction.

#### Example Usage

```go
package main

import (
	"log"
	"github.com/limaJavier/timetabling/pkg/model"
	"github.com/limaJavier/timetabling/pkg/sat"
)

func main() {
	sat.ConfigPath = "path/to/config.json" // Path to solver config

	input := model.ModelInput{
		// Populate input programmatically...
	}

	// Or load from JSON:
	// input, err := model.InputFromJson("input.json")

	solver := sat.NewCadicalSolver() // Initialize SAT solver
	timetabler := model.NewEmbeddedRoomTimetabler(solver) // Initialize timetabler

	timetable, variables, clauses, err := timetabler.Build(input)
	if err != nil {
		log.Fatal(err)
	}
	if timetable == nil {
		log.Fatal("Cannot generate timetable: instance is not satisfiable")
	}

	log.Printf("Timetable generated with %v variables and %v clauses\n", variables, clauses)

	// Process timetable...
}
```

> **Note**: Ensure that you provide a `config.json` file specifying the paths to the SAT solvers. Set the path to this file using the `sat.ConfigPath` global variable.

**Example `config.json`**:

```json
{
    "kissatPath": "kissat",
    "cadicalPath": "cadical",
    "minisatPath": "minisat",
    "cryptominisatPath": "cryptominisat",
    "glucoseSimpPath": "glucose-simp",
    "glucoseSyrupPath": "glucose-syrup",
    "slimePath": "slime",
    "ortoolsatPath": "ortoolsat"
}
```

---

### CLI

Use the `-help` flag to view available options.

Available flags:

- `-strategy`: Strategy to build the timetable.
- `-solver`: SAT solver to use.
- `-similarity`: Similarity threshold (0–1) used by the hybrid strategy.
- `-file`: Path to the input JSON file.
- `-out`: Output file path. If empty, the result is written to *stdout*.

#### Example Input File

The input is a JSON file structured as follows:

```json
{
  "subjects": [{"id": 0, "name": "Logica"}],
  "professors": [{"id": 0, "name": "Luciano", "availability": [[true, true, false], [true, true, false]]}],
  "rooms": [{"id": 0, "name": "Aula 6", "capacity": 50}],
  "classes": [{"id": 0, "name": "CC-111", "size": 30}],
  "entries": [{
    "subject": 0,
    "professor": 0,
    "classes": [0],
    "lessons": 2,
    "permissibility": [[true, true, true], [true, true, false]],
    "rooms": [0]
  }]
}
```

---

### 🐍 Python Wrapper

A minimal Python wrapper is available in `cmd/cli` on the **python_wrapper** branch.

---

## 🧪 Testing & Benchmarking

After running the `build.sh` script, test instances are created automatically.

To run all tests (from the project root):

```console
$ go test ./...
```

To benchmark solver performance, execute the benchmarking script:

```console
$ ./benchmarking.sh
```

Analysis notebooks for the resulting CSV data can be found in the **data_analysis** branch at `cmd/benchmarking`.
