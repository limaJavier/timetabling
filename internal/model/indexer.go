package model

// Indexer interface is design to give a unique index to a combination of scheduling variable's attributes and vice versa
type Indexer interface {
	// Returns a unique index to a combination of scheduling variable's attributes
	Index(period, day, lesson, subjectProfessor, group uint64) uint64
	// Returns a combination of scheduling variable's attributes from a unique index
	Attributes(index uint64) (period uint64, day uint64, lesson uint64, subjectProfessor uint64, group uint64)
}

func NewIndexer(periods, days, lessons, subjectProfessors, groups uint64) Indexer {
	return &sortedIndexer{
		periods:           periods,
		days:              days,
		subjectProfessors: subjectProfessors,
		lessons:           lessons,
		groups:            groups,
	}
}
