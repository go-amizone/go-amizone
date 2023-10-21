package toproto

import (
	"time"

	"github.com/ditsuke/go-amizone/amizone/models"
	v1 "github.com/ditsuke/go-amizone/server/gen/go/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TimeToProtoTS(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

func CourseRef(a models.CourseRef) *v1.CourseRef {
	return &v1.CourseRef{
		Code: a.Code,
		Name: a.Name,
	}
}

func AttendanceRecord(a models.AttendanceRecord) *v1.AttendanceRecord {
	return &v1.AttendanceRecord{
		Course: CourseRef(models.CourseRef(a.Course)),
		Attendance: &v1.Attendance{
			Attended: a.ClassesAttended,
			Held:     a.ClassesHeld,
		},
	}
}

func AttendanceRecords(a models.AttendanceRecords) *v1.AttendanceRecords {
	t := make([]*v1.AttendanceRecord, len(a))
	for i, c := range a {
		t[i] = AttendanceRecord(models.AttendanceRecord(c))
	}
	return &v1.AttendanceRecords{
		Records: t,
	}
}

func ScheduledClasses(a models.ClassSchedule) *v1.ScheduledClasses {
	arr := make([]*v1.ScheduledClass, len(a))
	for i, c := range a {
		arr[i] = &v1.ScheduledClass{
			Course: &v1.CourseRef{
				Code: c.Course.Code,
				Name: c.Course.Name,
			},
			StartTime: TimeToProtoTS(c.StartTime),
			EndTime:   TimeToProtoTS(c.EndTime),
			Faculty:   c.Faculty,
			Room:      c.Room,
			Attendance: func() v1.AttendanceState {
				switch c.Attended {
				case models.AttendanceStatePresent:
					return v1.AttendanceState_PRESENT
				case models.AttendanceStateAbsent:
					return v1.AttendanceState_ABSENT
				case models.AttendanceStatePending:
					return v1.AttendanceState_PENDING
				case models.AttendanceStateNA:
					return v1.AttendanceState_NA
				default:
					return v1.AttendanceState_INVALID
				}
			}(),
		}
	}
	return &v1.ScheduledClasses{
		Classes: arr,
	}
}

func ExamSchedule(a models.ExaminationSchedule) *v1.ExaminationSchedule {
	arr := make([]*v1.ScheduledExam, len(a.Exams))

	for i, c := range a.Exams {
		arr[i] = &v1.ScheduledExam{
			Course: CourseRef(models.CourseRef(c.Course)),
			Time:   TimeToProtoTS(c.Time),
			Mode:   c.Mode,
		}
	}

	return &v1.ExaminationSchedule{
		Title: a.Title,
		Exams: arr,
	}
}

func SemesterList(a models.SemesterList) *v1.SemesterList {
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

func Marks(marks models.Marks) *v1.Marks {
	return &v1.Marks{
		Have: marks.Have,
		Max:  marks.Max,
	}
}

func Courses(a models.Courses) *v1.Courses {
	arr := make([]*v1.Course, len(a))
	for i, c := range a {
		arr[i] = &v1.Course{
			Ref:  CourseRef(models.CourseRef(c.CourseRef)),
			Type: c.Type,
			Attendance: &v1.Attendance{
				Attended: c.Attendance.ClassesAttended,
				Held:     c.Attendance.ClassesHeld,
			},
			InternalMarks: Marks(models.Marks(c.InternalMarks)),
			SyllabusDoc:   c.SyllabusDoc,
		}
	}
	return &v1.Courses{
		Courses: arr,
	}
}

func Profile(a models.Profile) *v1.Profile {
	return &v1.Profile{
		Name:               a.Name,
		EnrollmentNumber:   a.EnrollmentNumber,
		EnrollmentValidity: TimeToProtoTS(a.EnrollmentValidity),
		Batch:              a.Batch,
		Program:            a.Program,
		DateOfBirth:        TimeToProtoTS(a.DateOfBirth),
		BloodGroup:         a.BloodGroup,
		IdCardNumber:       a.IDCardNumber,
		Uuid:               a.UUID,
	}
}

func WifiInfo(i models.WifiMacInfo) *v1.WifiMacInfo {
	return &v1.WifiMacInfo{
		Addresses: func() []string {
			addresses := make([]string, 0)
			for _, a := range i.RegisteredAddresses {
				addresses = append(addresses, a.String())
			}
			return addresses
		}(),
		Slots:     int32(i.Slots),
		FreeSlots: int32(i.FreeSlots),
	}
}

func AtpcPlacementDetails(a models.AtpcPlacementDetails) *v1.AtpcPlacementDetails {
	arr := make([]*v1.AtpcPlacementEntry, len(a))
	for i, c := range a {
		arr[i] = &v1.AtpcPlacementEntry{
			Company: c.Company,
			RegStartDate: TimeToProtoTS(c.RegStartDate),
			RegEndDate: TimeToProtoTS(c.RegEndDate),
		}
	}
	return &v1.AtpcPlacementDetails{
		Details: arr,
	}
}