package amizone

import "time"

type Date struct {
	Year  int
	Month int
	Day   int
}

func DateFromTime(t time.Time) Date {
	return Date{Year: t.Year(), Month: int(t.Month()), Day: t.Day()}
}
