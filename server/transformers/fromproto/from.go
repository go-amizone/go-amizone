package fromproto

import (
	"google.golang.org/genproto/googleapis/type/date"
	"time"
)

func Date(d *date.Date) time.Time {
	return time.Date(int((*d).Year), time.Month((*d).Month), int((*d).Day), 0, 0, 0, 0, time.UTC)
}
