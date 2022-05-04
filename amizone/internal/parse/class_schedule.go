package parse

import (
	"amizone/amizone/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"k8s.io/klog/v2"
	"time"
)

const (
	scheduleJsonTimeFormat = "2006/01/02 03:04:05 PM"
)

// ClassSchedule attempts to parse the response of the Amizone diary events API endpoint into
// a models.ClassSchedule instance.
func ClassSchedule(body io.Reader) (models.ClassSchedule, error) {
	var diaryEvents models.AmizoneDiaryEvents
	if err := json.NewDecoder(body).Decode(&diaryEvents); err != nil {
		return nil, errors.New(fmt.Sprintf("%s: %s", ErrFailedToParse, err.Error()))
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

		class := &models.ScheduledClass{
			Course: &models.Course{
				Code: entry.CourseCode,
				Name: entry.CourseName,
			},
			StartTime: parseTime(entry.Start),
			EndTime:   parseTime(entry.End),
			Faculty:   entry.Faculty,
			Room:      entry.Room,
		}

		classSchedule = append(classSchedule, class)
	}

	// We sort the parsed schedule by start time -- because the Amizone events endpoint does not guarantee order.
	classSchedule.Sort()

	return classSchedule, nil
}
