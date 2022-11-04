package toproto

import (
	"github.com/ditsuke/go-amizone/amizone"
	v1 "github.com/ditsuke/go-amizone/server/gen/go/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

func Time(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

func CourseRef(a amizone.CourseRef) *v1.CourseRef {
	return &v1.CourseRef{
		Code: a.Code,
		Name: a.Name,
	}
}

func AttendanceRecord(a amizone.AttendanceRecord) *v1.AttendanceRecord {
	return &v1.AttendanceRecord{
		Course: CourseRef(amizone.CourseRef(a.Course)),
		Attendance: &v1.Attendance{
			Attended: a.ClassesAttended,
			Held:     a.ClassesHeld,
		},
	}
}

func AttendanceRecords(a amizone.AttendanceRecords) *v1.AttendanceRecords {
	t := make([]*v1.AttendanceRecord, len(a))
	for i, c := range a {
		t[i] = AttendanceRecord(amizone.AttendanceRecord(c))
	}
	return &v1.AttendanceRecords{
		Records: t,
	}
}

func ScheduledClasses(a amizone.ClassSchedule) *v1.ScheduledClasses {
	arr := make([]*v1.ScheduledClass, len(a))
	for i, c := range a {
		arr[i] = &v1.ScheduledClass{
			Course: &v1.CourseRef{
				Code: c.Course.Code,
				Name: c.Course.Name,
			},
			StartTime: Time(c.StartTime),
			EndTime:   Time(c.EndTime),
			Faculty:   c.Faculty,
			Room:      c.Room,
		}
	}
	return &v1.ScheduledClasses{
		Classes: arr,
	}
}

func ExamSchedule(a amizone.ExamSchedule) *v1.ExaminationSchedule {
	arr := make([]*v1.ScheduledExam, len(a.Exams))

	for i, c := range a.Exams {
		arr[i] = &v1.ScheduledExam{
			Course: CourseRef(amizone.CourseRef(c.Course)),
			Time:   Time(c.Time),
			Mode:   c.Mode,
		}
	}

	return &v1.ExaminationSchedule{
		Title: a.Title,
		Exams: arr,
	}
}

func SemesterList(a amizone.SemesterList) *v1.SemesterList {
	arr := make([]*v1.Semester, len(a))
	for i, c := range a {
		arr[i] = &v1.Semester{
			Name: c.Name,
			Ref:  c.Ref,
		}
	}
	return &v1.SemesterList{
		Semesters: arr,
	}
}

func Marks(marks amizone.Marks) *v1.Marks {
	return &v1.Marks{
		Have: marks.Have,
		Max:  marks.Max,
	}
}

func Courses(a amizone.Courses) *v1.Courses {
	arr := make([]*v1.Course, len(a))
	for i, c := range a {
		arr[i] = &v1.Course{
			Ref:  CourseRef(amizone.CourseRef(c.CourseRef)),
			Type: c.Type,
			Attendance: &v1.Attendance{
				Attended: c.Attendance.ClassesAttended,
				Held:     c.Attendance.ClassesHeld,
			},
			InternalMarks: Marks(amizone.Marks(c.InternalMarks)),
			SyllabusDoc:   c.SyllabusDoc,
		}
	}
	return &v1.Courses{
		Courses: arr,
	}
}
