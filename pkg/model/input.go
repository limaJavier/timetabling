package model

import (
	"encoding/json"
	"os"

	"github.com/mitchellh/mapstructure"
)

type RawSubjectProfessor struct {
	Id             uint64
	Subject        uint64
	Professor      uint64
	Lessons        uint64
	Groups         [][]uint64
	Rooms          []uint64
	Permissibility [][]bool
}

type Subject struct {
	Id   uint64
	Name string
}

type Class struct {
	Id   uint64
	Name string
	Size uint64
}

type Group struct {
	Id      uint64
	Classes []uint64
}

type Room struct {
	Id       uint64
	Name     string
	Capacity uint64
}

type Professor struct {
	Id           uint64
	Name         string
	Availability [][]bool
}

type SubjectProfessor struct {
	Id        uint64
	Subject   uint64
	Professor uint64
}

type Entry struct {
	SubjectProfessor uint64
	Group            uint64
	Lessons          uint64
	Permissibility   [][]bool
	Rooms            []uint64
}

type ModelInput struct {
	Subjects          []Subject
	Professors        []Professor
	SubjectProfessors []SubjectProfessor
	Entries           map[[2]uint64]Entry
	Groups            []Group
	Classes           []Class
	Rooms             []Room
}

func InputFromJson(file string) (ModelInput, error) {
	bytes, _ := os.ReadFile(file)
	var inputJson map[string]any
	err := json.Unmarshal(bytes, &inputJson)
	if err != nil {
		return ModelInput{}, err
	}

	var input ModelInput
	mapstructure.Decode(inputJson, &input)

	return input, nil
}
