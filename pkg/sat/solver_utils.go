package sat

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/samber/lo"
)

const configPath = "../../config.json"

var Config = getConfig()

func parseSolution(solverOutput string) SATSolution {
	values := lo.Map(
		lo.Reduce(
			lo.Filter(strings.Split(solverOutput, "\n"), func(line string, _ int) bool {
				return len(line) > 0 && line[0] == 'v'
			}),
			func(values []string, line string, _ int) []string {
				return append(values, strings.Split(line[2:], " ")...)
			},
			[]string{},
		),
		func(valueStr string, _ int) int64 {
			value, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil && valueStr != "" {
				log.Panicf("invalid literal in solver output: %v", err)
			}
			return value
		},
	)
	return values[:len(values)-1]
}

func getConfig() map[string]string {
	bytes, _ := os.ReadFile(configPath)
	var inputJson map[string]any
	err := json.Unmarshal(bytes, &inputJson)
	if err != nil {
		log.Fatalf("cannot read config.json file: %v", err)
	}
	var config map[string]string
	mapstructure.Decode(inputJson, &config)
	return config
}
