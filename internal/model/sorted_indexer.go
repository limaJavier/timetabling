package model

type sortedIndexer struct {
	Periods           uint64
	Days              uint64
	SubjectProfessors uint64
	Lessons           uint64
	Classes           uint64
}

func (i *sortedIndexer) Index(period, day, lesson, subjectProfessor, class uint64) uint64 {
	return period + i.Periods*(day-1) + i.Periods*i.Days*(lesson-1) + i.Periods*i.Days*i.Lessons*(subjectProfessor-1) + i.Periods*i.Days*i.Lessons*i.SubjectProfessors*(class-1)
}

func (i *sortedIndexer) Attributes(index uint64) (period uint64, day uint64, lesson uint64, subjectProfessor uint64, class uint64) {
	if index == i.Periods*i.Days*i.Lessons*i.SubjectProfessors*i.Classes {
		return i.Periods, i.Days, i.Lessons, i.SubjectProfessors, i.Classes
	}

	period = index % i.Periods
	index = index / i.Periods

	day = index%i.Days + 1
	index = index / i.Days

	lesson = index%i.Lessons + 1
	index = index / i.Lessons

	subjectProfessor = index%i.SubjectProfessors + 1
	index = index / i.SubjectProfessors

	class = index%i.Classes + 1

	return period, day, lesson, subjectProfessor, class
}
