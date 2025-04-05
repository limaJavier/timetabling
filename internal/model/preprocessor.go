package model

type Preprocessor interface {
	AddSingletonGroups(classesCurriculum [][]bool, groupsPerSubjectProfessor map[uint64][][]uint64)
	ExtractCurriculumAndGroups(classesCurriculum [][]bool, groupsPerSubjectProfessor map[uint64][][]uint64) ([][]bool, map[uint64][]uint64)
	BuildGroupsGraph(groups map[uint64][]uint64) [][]bool
}

func NewPreprocessor() Preprocessor {
	return &preprocessorImplementation{}
}
