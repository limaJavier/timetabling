package model

type Preprocessor interface {
	ExtractCurriculumAndGroups(modelInput ModelInput) ([][]bool, map[uint64][]uint64)
	BuildGroupsGraph(groups map[uint64][]uint64) [][]bool
}

func NewPreprocessor() Preprocessor {
	return &preprocessorImplementation{}
}
