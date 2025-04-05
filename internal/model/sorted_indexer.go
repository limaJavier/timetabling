package model

type sortedIndexer struct {
	periods           uint64
	days              uint64
	subjectProfessors uint64
	lessons           uint64
	groups            uint64
}

func (i *sortedIndexer) Index(period, day, lesson, subjectProfessor, group uint64) uint64 {
	return period + i.periods*(day) + i.periods*i.days*(lesson) + i.periods*i.days*i.lessons*(subjectProfessor) + i.periods*i.days*i.lessons*i.subjectProfessors*(group) + 1
}

func (i *sortedIndexer) Attributes(index uint64) (period uint64, day uint64, lesson uint64, subjectProfessor uint64, group uint64) {
	index = index - 1
	period = index % i.periods
	index = index / i.periods

	day = index % i.days
	index = index / i.days

	lesson = index % i.lessons
	index = index / i.lessons

	subjectProfessor = index % i.subjectProfessors
	index = index / i.subjectProfessors

	group = index % i.groups

	return period, day, lesson, subjectProfessor, group
}
