package model

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/mitchellh/mapstructure"
)

type Metadata struct {
	Professors        []string
	Rooms             []int
	Classes           []string
	SubjectProfessors []string `mapstructure:"subjectProfessors"`
}

type ModelInput struct {
	GroupsPerSubjectProfessors map[string][][]uint64 `mapstructure:"groupsPerSubjectProfessors"`
	Permissibility             map[string][][]bool
	Availability               map[string][][]bool
	Lessons                    map[string]uint64
	Rooms                      map[string]uint64
	Professors                 map[string]uint64
	Metadata                   Metadata
}

func (input ModelInput) GetGroupsPerSubjectProfessors() map[uint64][][]uint64 {
	result := make(map[uint64][][]uint64)
	for k, v := range input.GroupsPerSubjectProfessors {
		key, err := strconv.ParseUint(k, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("cannot transform dictionary: %v", err))
		}
		result[key] = v
	}
	return result
}

func (input ModelInput) GetLessons() map[uint64]uint64 {
	result := make(map[uint64]uint64)
	for k, v := range input.Lessons {
		key, err := strconv.ParseUint(k, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("cannot transform dictionary: %v", err))
		}
		result[key] = v
	}
	return result
}

func (input ModelInput) GetRooms() map[uint64]uint64 {
	result := make(map[uint64]uint64)
	for k, v := range input.Rooms {
		key, err := strconv.ParseUint(k, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("cannot transform dictionary: %v", err))
		}
		result[key] = v
	}
	return result
}

func (input ModelInput) GetProfessors() map[uint64]uint64 {
	result := make(map[uint64]uint64)
	for k, v := range input.Professors {
		key, err := strconv.ParseUint(k, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("cannot transform dictionary: %v", err))
		}
		result[key] = v
	}
	return result
}

func (input ModelInput) GetPermissibility() map[uint64][][]bool {
	result := make(map[uint64][][]bool)
	for k, v := range input.Permissibility {
		key, err := strconv.ParseUint(k, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("cannot transform dictionary: %v", err))
		}
		result[key] = v
	}
	return result
}

func (input ModelInput) GetAvailability() map[uint64][][]bool {
	result := make(map[uint64][][]bool)
	for k, v := range input.Availability {
		key, err := strconv.ParseUint(k, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("cannot transform dictionary: %v", err))
		}
		result[key] = v
	}
	return result
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
