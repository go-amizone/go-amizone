package parse

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/ditsuke/go-amizone/amizone/models"
	"k8s.io/klog/v2"
)

const (
	scheduleJsonTimeFormat = "2006/01/02 03:04:05 PM"
)

// ClassSchedule attempts to parse the response of the Amizone diary events API endpoint into
// a models.ClassSchedule instance.
func ClassSchedule(body io.Reader) (models.ClassSchedule, error) {
	var diaryEvents models.AmizoneDiaryEvents
	if err := json.NewDecoder(body).Decode(&diaryEvents); err != nil {
		return nil, fmt.Errorf("JSON decode: %w", err)
	}

	var classSchedule models.ClassSchedule
	for _, entry := range diaryEvents {
		// Only add entries that are of type "C" (class)
		if entry.Type != "C" {
			continue
		}

		parseTime := func(timeStr string) time.Time {
			t, err := time.Parse(scheduleJsonTimeFormat, timeStr)
			if err != nil {
				klog.Warning("Failed to parse time for course %s: %s", entry.CourseCode, err.Error())
				return time.Unix(0, 0)
			}
			return t
		}

		class := models.ScheduledClass{
			Course: models.CourseRef{
				Code: cleanString(entry.CourseCode),
				Name: cleanString(entry.CourseName),
			},
			StartTime: parseTime(entry.Start),
			EndTime:   parseTime(entry.End),
			Faculty:   cleanString(entry.Faculty),
			Room:      cleanString(entry.Room),
			Attended:  entry.AttendanceState(),
		}

		classSchedule = append(classSchedule, class)
	}

	// We sort the parsed schedule by start time -- because the Amizone events endpoint does not guarantee order.
	classSchedule.Sort()

	return classSchedule, nil
}
