package model

type sortedIndexer struct {
	Periods         uint64
	Days            uint64
	SubjectTeachers uint64
	Lessons         uint64
	Classes         uint64
}

func (i *sortedIndexer) Index(period, day, lesson, subjectTeacher, class uint64) uint64 {
	return period + i.Periods*(day-1) + i.Periods*i.Days*(lesson-1) + i.Periods*i.Days*i.Lessons*(subjectTeacher-1) + i.Periods*i.Days*i.Lessons*i.SubjectTeachers*(class-1)
}

func (i *sortedIndexer) Attributes(index uint64) (period uint64, day uint64, lesson uint64, subjectTeacher uint64, class uint64) {
	if index == i.Periods*i.Days*i.Lessons*i.SubjectTeachers*i.Classes {
		return i.Periods, i.Days, i.Lessons, i.SubjectTeachers, i.Classes
	}

	period = index % i.Periods
	index = index / i.Periods

	day = index%i.Days + 1
	index = index / i.Days

	lesson = index%i.Lessons + 1
	index = index / i.Lessons

	subjectTeacher = index%i.SubjectTeachers + 1
	index = index / i.SubjectTeachers

	class = index%i.Classes + 1

	return period, day, lesson, subjectTeacher, class
}
