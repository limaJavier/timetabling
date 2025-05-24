package model

type Timetabler interface {
	Build(
		modelInput ModelInput,
	) (timetable [][6]uint64, variables uint64, clauses uint64, err error)

	Verify(
		timetable [][6]uint64,
		modelInput ModelInput,
	) bool
}
