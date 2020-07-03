package gts

import "time"

// Date represents a date stamp for record entries.
type Date struct {
	Year  int
	Month time.Month
	Day   int
}

// FromTime creates a Date object from a time.Time object.
func FromTime(t time.Time) Date {
	return Date{t.Year(), t.Month(), t.Day()}
}

// ToTime converts the Date object into a time.Time object.
func (d Date) ToTime() time.Time {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC)
}
