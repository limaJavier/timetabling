package model

type Timetabler interface {
	Build(
		modelInput ModelInput,
	) ([][6]uint64, error)

	Verify(
		timetable [][6]uint64,
		modelInput ModelInput,
	) bool
}
