package models_test

import (
	"amizone/amizone/internal/models"
	. "github.com/onsi/gomega"
	"testing"
	"time"
)

func TestClassSchedule_Sort(t *testing.T) {
	testCases := []struct {
		name     string
		schedule models.ClassSchedule
	}{
		{
			name: "2 classes - latter class in slice is earlier",
			schedule: models.ClassSchedule{
				{StartTime: time.Now()},
				{StartTime: time.Now().Add(-1 * time.Hour * 24)},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			tc.schedule.Sort()
			for i := 0; i < len(tc.schedule)-1; i++ {
				g.Expect(tc.schedule[i].StartTime.Before(tc.schedule[i+1].StartTime)).To(BeTrue())
			}
		})
	}
}
