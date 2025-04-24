package model

type indexerImplementation struct {
	periods           uint64
	days              uint64
	subjectProfessors uint64
	lessons           uint64
	groups            uint64
	rooms             uint64
}

func (indexer *indexerImplementation) Index(period, day, lesson, subjectProfessor, group, room uint64) uint64 {
	return period + indexer.periods*day + indexer.periods*indexer.days*lesson + indexer.periods*indexer.days*indexer.lessons*subjectProfessor + indexer.periods*indexer.days*indexer.lessons*indexer.subjectProfessors*group + indexer.periods*indexer.days*indexer.lessons*indexer.subjectProfessors*indexer.groups*room + 1
}

func (indexer *indexerImplementation) Attributes(index uint64) (period, day, lesson, subjectProfessor, group, room uint64) {
	index = index - 1
	period = index % indexer.periods
	index = index / indexer.periods

	day = index % indexer.days
	index = index / indexer.days

	lesson = index % indexer.lessons
	index = index / indexer.lessons

	subjectProfessor = index % indexer.subjectProfessors
	index = index / indexer.subjectProfessors

	group = index % indexer.groups
	index = index / indexer.groups

	room = index % indexer.rooms

	return period, day, lesson, subjectProfessor, group, room
}
