package model

// indexer interface is design to give a unique index to a combination of scheduling variable's attributes and vice versa
type indexer interface {
	// Returns a unique index to a combination of scheduling variable's attributes
	Index(period, day, lesson, subjectProfessor, group, room uint64) uint64
	// Returns a combination of scheduling variable's attributes from a unique index
	Attributes(index uint64) (period uint64, day uint64, lesson uint64, subjectProfessor uint64, group, room uint64)
}

func newIndexer(periods, days, lessons, subjectProfessors, groups, rooms uint64) indexer {
	return &indexerImplementation{
		periods:           periods,
		days:              days,
		subjectProfessors: subjectProfessors,
		lessons:           lessons,
		groups:            groups,
		rooms:             rooms,
	}
}
