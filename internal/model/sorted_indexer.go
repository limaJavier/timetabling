package model

type sortedIndexer struct {
	periods           uint64
	days              uint64
	subjectProfessors uint64
	lessons           uint64
	classes           uint64
}

func (i *sortedIndexer) Index(period, day, lesson, subjectProfessor, class uint64) uint64 {
	return period + i.periods*(day) + i.periods*i.days*(lesson) + i.periods*i.days*i.lessons*(subjectProfessor) + i.periods*i.days*i.lessons*i.subjectProfessors*(class)
}

func (i *sortedIndexer) Attributes(index uint64) (period uint64, day uint64, lesson uint64, subjectProfessor uint64, class uint64) {
	period = index % i.periods
	index = index / i.periods

	day = index % i.days
	index = index / i.days

	lesson = index % i.lessons
	index = index / i.lessons

	subjectProfessor = index % i.subjectProfessors
	index = index / i.subjectProfessors

	class = index % i.classes

	return period, day, lesson, subjectProfessor, class
}
